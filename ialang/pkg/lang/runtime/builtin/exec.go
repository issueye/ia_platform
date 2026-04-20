package builtin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	osexec "os/exec"
	goRuntime "runtime"
	"strings"
	"sync"
	"time"
)

type execRunConfig struct {
	Command       string
	Args          []string
	Cwd           string
	Env           Object
	Stdin         string
	Timeout       time.Duration
	Shell         bool
	InheritOutput bool
}

type execChild struct {
	cmd           *osexec.Cmd
	ctx           context.Context
	cancel        context.CancelFunc
	inheritOutput bool
	stdout        bytes.Buffer
	stderr        bytes.Buffer

	waitOnce   sync.Once
	waitDone   chan struct{}
	waitResult Object
	waitErr    error
}

func newExecModule(asyncRuntime AsyncRuntime) Object {
	runFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseExecRunArgs("exec.run", args)
		if err != nil {
			return nil, err
		}
		return runExecCommand(cfg)
	})
	runAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return runFn(args)
		}), nil
	})
	startFn := NativeFunction(func(args []Value) (Value, error) {
		cfg, err := parseExecRunArgs("exec.start", args)
		if err != nil {
			return nil, err
		}
		return startExecCommand(cfg, asyncRuntime)
	})
	lookPathFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("exec.lookPath expects 1 arg, got %d", len(args))
		}
		name, err := asStringArg("exec.lookPath", args, 0)
		if err != nil {
			return nil, err
		}
		path, err := osexec.LookPath(name)
		if err != nil {
			if errors.Is(err, osexec.ErrNotFound) {
				return nil, nil
			}
			return nil, err
		}
		return path, nil
	})

	namespace := Object{
		"run":      runFn,
		"runAsync": runAsyncFn,
		"start":    startFn,
		"lookPath": lookPathFn,
		"which":    lookPathFn,
	}
	module := cloneObject(namespace)
	module["exec"] = namespace
	return module
}

func parseExecRunArgs(fn string, args []Value) (execRunConfig, error) {
	if len(args) < 1 || len(args) > 2 {
		return execRunConfig{}, fmt.Errorf("%s expects 1-2 args: command, [options]", fn)
	}
	command, err := asStringArg(fn, args, 0)
	if err != nil {
		return execRunConfig{}, err
	}
	cfg := execRunConfig{
		Command:       command,
		Args:          nil,
		Cwd:           "",
		Env:           nil,
		Stdin:         "",
		Timeout:       0,
		Shell:         false,
		InheritOutput: false,
	}

	if len(args) == 1 || args[1] == nil {
		return cfg, nil
	}
	options, ok := args[1].(Object)
	if !ok {
		return execRunConfig{}, fmt.Errorf("%s arg[1] expects object options, got %T", fn, args[1])
	}
	if v, ok := options["args"]; ok && v != nil {
		rawArgs, ok := v.(Array)
		if !ok {
			return execRunConfig{}, fmt.Errorf("exec options.args expects array, got %T", v)
		}
		parsed := make([]string, 0, len(rawArgs))
		for i, item := range rawArgs {
			s, err := asStringValue(fmt.Sprintf("exec options.args[%d]", i), item)
			if err != nil {
				return execRunConfig{}, err
			}
			parsed = append(parsed, s)
		}
		cfg.Args = parsed
	}
	if v, ok := options["cwd"]; ok && v != nil {
		cwd, err := asStringValue("exec options.cwd", v)
		if err != nil {
			return execRunConfig{}, err
		}
		cfg.Cwd = cwd
	}
	if v, ok := options["env"]; ok && v != nil {
		env, ok := v.(Object)
		if !ok {
			return execRunConfig{}, fmt.Errorf("exec options.env expects object, got %T", v)
		}
		cfg.Env = cloneObject(env)
	}
	if v, ok := options["stdin"]; ok && v != nil {
		stdin, err := asStringValue("exec options.stdin", v)
		if err != nil {
			return execRunConfig{}, err
		}
		cfg.Stdin = stdin
	}
	if v, ok := options["timeoutMs"]; ok && v != nil {
		timeoutMs, err := asIntValue("exec options.timeoutMs", v)
		if err != nil {
			return execRunConfig{}, err
		}
		if timeoutMs <= 0 {
			return execRunConfig{}, fmt.Errorf("exec options.timeoutMs expects positive integer, got %d", timeoutMs)
		}
		cfg.Timeout = time.Duration(timeoutMs) * time.Millisecond
	}
	if v, ok := options["shell"]; ok && v != nil {
		shell, ok := v.(bool)
		if !ok {
			return execRunConfig{}, fmt.Errorf("exec options.shell expects bool, got %T", v)
		}
		cfg.Shell = shell
	}
	if v, ok := options["inheritOutput"]; ok && v != nil {
		inheritOutput, ok := v.(bool)
		if !ok {
			return execRunConfig{}, fmt.Errorf("exec options.inheritOutput expects bool, got %T", v)
		}
		cfg.InheritOutput = inheritOutput
	}
	return cfg, nil
}

func runExecCommand(cfg execRunConfig) (Value, error) {
	ctx := context.Background()
	cancel := func() {}
	if cfg.Timeout > 0 {
		ctxWithTimeout, cancelFn := context.WithTimeout(ctx, cfg.Timeout)
		ctx = ctxWithTimeout
		cancel = cancelFn
	}
	defer cancel()

	cmd := buildExecCommand(ctx, cfg)
	if cfg.Cwd != "" {
		cmd.Dir = cfg.Cwd
	}
	if len(cfg.Env) > 0 {
		cmd.Env = mergeExecEnv(cfg.Env)
	}
	if cfg.Stdin != "" {
		cmd.Stdin = strings.NewReader(cfg.Stdin)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if cfg.InheritOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}

	err := cmd.Run()
	stdoutText := ""
	stderrText := ""
	if !cfg.InheritOutput {
		stdoutText = stdout.String()
		stderrText = stderr.String()
	}
	return buildExecResult(ctx, stdoutText, stderrText, err), nil
}

func startExecCommand(cfg execRunConfig, asyncRuntime AsyncRuntime) (Value, error) {
	ctx := context.Background()
	cancel := func() {}
	if cfg.Timeout > 0 {
		ctxWithTimeout, cancelFn := context.WithTimeout(ctx, cfg.Timeout)
		ctx = ctxWithTimeout
		cancel = cancelFn
	}

	cmd := buildExecCommand(ctx, cfg)
	if cfg.Cwd != "" {
		cmd.Dir = cfg.Cwd
	}
	if len(cfg.Env) > 0 {
		cmd.Env = mergeExecEnv(cfg.Env)
	}
	if cfg.Stdin != "" {
		cmd.Stdin = strings.NewReader(cfg.Stdin)
	}

	child := &execChild{
		cmd:           cmd,
		ctx:           ctx,
		cancel:        cancel,
		inheritOutput: cfg.InheritOutput,
		waitDone:      make(chan struct{}),
	}
	if cfg.InheritOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &child.stdout
		cmd.Stderr = &child.stderr
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}
	child.beginWait()

	return newExecChildObject(child, asyncRuntime), nil
}

func newExecChildObject(child *execChild, asyncRuntime AsyncRuntime) Object {
	pidFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("exec.child.pid expects 0 args, got %d", len(args))
		}
		if child.cmd.Process == nil {
			return float64(0), nil
		}
		return float64(child.cmd.Process.Pid), nil
	})
	waitFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("exec.child.wait expects 0 args, got %d", len(args))
		}
		return child.wait()
	})
	waitAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return waitFn(args)
		}), nil
	})
	killFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("exec.child.kill expects 0 args, got %d", len(args))
		}
		return child.kill()
	})
	isRunningFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("exec.child.isRunning expects 0 args, got %d", len(args))
		}
		return child.isRunning(), nil
	})

	return Object{
		"pid":       pidFn,
		"wait":      waitFn,
		"waitAsync": waitAsyncFn,
		"kill":      killFn,
		"isRunning": isRunningFn,
	}
}

func (c *execChild) beginWait() {
	go func() {
		_, _ = c.wait()
	}()
}

func (c *execChild) wait() (Value, error) {
	c.waitOnce.Do(func() {
		err := c.cmd.Wait()
		stdoutText := ""
		stderrText := ""
		if !c.inheritOutput {
			stdoutText = c.stdout.String()
			stderrText = c.stderr.String()
		}
		c.waitResult = buildExecResult(c.ctx, stdoutText, stderrText, err)
		c.cancel()
		close(c.waitDone)
	})
	<-c.waitDone
	return cloneObject(c.waitResult), c.waitErr
}

func (c *execChild) kill() (Value, error) {
	if c.cmd.Process == nil {
		return false, fmt.Errorf("exec.child.kill process not started")
	}
	if !c.isRunning() {
		return false, nil
	}
	if err := c.cmd.Process.Kill(); err != nil {
		return nil, err
	}
	return true, nil
}

func (c *execChild) isRunning() bool {
	select {
	case <-c.waitDone:
		return false
	default:
		return true
	}
}

func buildExecResult(ctx context.Context, stdoutText string, stderrText string, err error) Object {
	timedOut := errors.Is(ctx.Err(), context.DeadlineExceeded)
	exitCode := 0
	if err != nil {
		exitCode = -1
		var exitErr *osexec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		}
	}
	var errText Value
	if err != nil {
		errText = err.Error()
	} else {
		errText = nil
	}
	return makeExecResult(err == nil, exitCode, stdoutText, stderrText, timedOut, errText)
}

func makeExecResult(ok bool, exitCode int, stdoutText string, stderrText string, timedOut bool, errText Value) Object {
	return Object{
		"ok":       ok,
		"code":     float64(exitCode),
		"stdout":   stdoutText,
		"stderr":   stderrText,
		"combined": stdoutText + stderrText,
		"timedOut": timedOut,
		"error":    errText,
	}
}

func buildExecCommand(ctx context.Context, cfg execRunConfig) *osexec.Cmd {
	if cfg.Shell {
		line := cfg.Command
		if len(cfg.Args) > 0 {
			line = line + " " + strings.Join(cfg.Args, " ")
		}
		if goRuntime.GOOS == "windows" {
			return osexec.CommandContext(ctx, "cmd", "/C", line)
		}
		return osexec.CommandContext(ctx, "sh", "-c", line)
	}
	return osexec.CommandContext(ctx, cfg.Command, cfg.Args...)
}

func mergeExecEnv(overrides Object) []string {
	base := map[string]string{}
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			base[parts[0]] = parts[1]
		}
	}
	for k, v := range overrides {
		s, err := asStringValue("exec options.env["+k+"]", v)
		if err != nil {
			continue
		}
		base[k] = s
	}
	out := make([]string, 0, len(base))
	for k, v := range base {
		out = append(out, k+"="+v)
	}
	return out
}
