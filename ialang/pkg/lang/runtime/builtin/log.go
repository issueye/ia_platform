package builtin

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type logModuleState struct {
	mu         sync.RWMutex
	level      slog.Level
	asJSON     bool
	outputPath string // 日志输出路径，空字符串表示 stdout
	file       *os.File
	logger     *slog.Logger
}

func newLogModule() Object {
	state := newLogModuleState()
	namespace := newLogNamespace(state, nil)
	// Note: log namespace already has "log" function, so we don't add self-reference
	// to avoid overwriting the function
	return cloneObject(namespace)
}

func newLogModuleState() *logModuleState {
	s := &logModuleState{
		level:      slog.LevelInfo,
		asJSON:     false,
		outputPath: "",
	}
	s.rebuildLoggerLocked()
	return s
}

func (s *logModuleState) rebuildLoggerLocked() {
	// 关闭旧文件句柄
	if s.file != nil {
		s.file.Close()
		s.file = nil
	}

	options := &slog.HandlerOptions{
		Level: s.level,
	}

	var writer io.Writer
	if s.outputPath == "" {
		writer = os.Stdout
	} else {
		f, err := os.OpenFile(s.outputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			// 如果打开文件失败，回退到 stdout
			writer = os.Stdout
		} else {
			s.file = f
			writer = f
		}
	}

	var handler slog.Handler
	if s.asJSON {
		handler = slog.NewJSONHandler(writer, options)
	} else {
		handler = slog.NewTextHandler(writer, options)
	}
	s.logger = slog.New(handler)
}

func (s *logModuleState) setOutputPath(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 关闭旧文件（如果不是 stdout）
	if s.file != nil {
		s.file.Close()
		s.file = nil
	}

	if path == "" || path == "stdout" {
		s.outputPath = ""
		s.rebuildLoggerLocked()
		return nil
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	s.outputPath = path
	s.rebuildLoggerLocked()
	return nil
}

func (s *logModuleState) getOutputPath() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.outputPath == "" {
		return "stdout"
	}
	return s.outputPath
}

func (s *logModuleState) setLevel(level slog.Level) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.level = level
	s.rebuildLoggerLocked()
}

func (s *logModuleState) getLevel() slog.Level {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.level
}

func (s *logModuleState) setJSON(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.asJSON = enabled
	s.rebuildLoggerLocked()
}

func (s *logModuleState) isJSON() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.asJSON
}

func (s *logModuleState) getLogger() *slog.Logger {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.logger
}

func newLogNamespace(state *logModuleState, baseFields []any) Object {
	logWithLevel := func(level slog.Level, fn string, args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("%s expects 1-2 args: message, [fields]", fn)
		}
		message, err := asStringArg(fn, args, 0)
		if err != nil {
			return nil, err
		}
		fields := append([]any{}, baseFields...)
		if len(args) == 2 {
			extra, err := parseLogFieldsArg(fn+" fields", args[1])
			if err != nil {
				return nil, err
			}
			fields = append(fields, extra...)
		}
		state.getLogger().Log(context.Background(), level, message, fields...)
		return true, nil
	}

	customLogFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 || len(args) > 3 {
			return nil, fmt.Errorf("log.log expects 2-3 args: level, message, [fields]")
		}
		level, err := parseLogLevelValue(args[0])
		if err != nil {
			return nil, err
		}
		message, err := asStringValue("log.log arg[1]", args[1])
		if err != nil {
			return nil, err
		}
		fields := append([]any{}, baseFields...)
		if len(args) == 3 {
			extra, err := parseLogFieldsArg("log.log fields", args[2])
			if err != nil {
				return nil, err
			}
			fields = append(fields, extra...)
		}
		state.getLogger().Log(context.Background(), level, message, fields...)
		return true, nil
	})

	withFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("log.with expects 1 arg: fields")
		}
		fields, err := parseLogFieldsArg("log.with arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		nextBase := append(append([]any{}, baseFields...), fields...)
		return newLogNamespace(state, nextBase), nil
	})

	setLevelFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("log.setLevel expects 1 arg, got %d", len(args))
		}
		level, err := parseLogLevelValue(args[0])
		if err != nil {
			return nil, err
		}
		state.setLevel(level)
		return levelName(state.getLevel()), nil
	})

	getLevelFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("log.getLevel expects 0 args, got %d", len(args))
		}
		return levelName(state.getLevel()), nil
	})

	setJSONFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("log.setJSON expects 1 arg, got %d", len(args))
		}
		enabled, ok := args[0].(bool)
		if !ok {
			return nil, fmt.Errorf("log.setJSON arg[0] expects bool, got %T", args[0])
		}
		state.setJSON(enabled)
		return enabled, nil
	})

	isJSONFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("log.isJSON expects 0 args, got %d", len(args))
		}
		return state.isJSON(), nil
	})

	setOutputPathFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("log.setOutputPath expects 1 arg, got %d", len(args))
		}
		path, err := asStringValue("log.setOutputPath", args[0])
		if err != nil {
			return nil, err
		}
		if err := state.setOutputPath(path); err != nil {
			return nil, fmt.Errorf("log.setOutputPath failed: %w", err)
		}
		return path, nil
	})

	getOutputPathFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("log.getOutputPath expects 0 args, got %d", len(args))
		}
		return state.getOutputPath(), nil
	})

	return Object{
		"debug": NativeFunction(func(args []Value) (Value, error) {
			return logWithLevel(slog.LevelDebug, "log.debug", args)
		}),
		"info": NativeFunction(func(args []Value) (Value, error) {
			return logWithLevel(slog.LevelInfo, "log.info", args)
		}),
		"warn": NativeFunction(func(args []Value) (Value, error) {
			return logWithLevel(slog.LevelWarn, "log.warn", args)
		}),
		"error": NativeFunction(func(args []Value) (Value, error) {
			return logWithLevel(slog.LevelError, "log.error", args)
		}),
		"log":           customLogFn,
		"with":          withFn,
		"setLevel":      setLevelFn,
		"getLevel":      getLevelFn,
		"setJSON":       setJSONFn,
		"isJSON":        isJSONFn,
		"setOutputPath": setOutputPathFn,
		"getOutputPath": getOutputPathFn,
	}
}

func parseLogLevelValue(v Value) (slog.Level, error) {
	switch vv := v.(type) {
	case string:
		switch strings.ToLower(strings.TrimSpace(vv)) {
		case "debug":
			return slog.LevelDebug, nil
		case "info":
			return slog.LevelInfo, nil
		case "warn", "warning":
			return slog.LevelWarn, nil
		case "error":
			return slog.LevelError, nil
		default:
			return 0, fmt.Errorf("unsupported log level: %q", vv)
		}
	case float64:
		return slog.Level(int(vv)), nil
	default:
		return 0, fmt.Errorf("log level expects string or number, got %T", v)
	}
}

func levelName(level slog.Level) string {
	switch {
	case level <= slog.LevelDebug:
		return "debug"
	case level < slog.LevelWarn:
		return "info"
	case level < slog.LevelError:
		return "warn"
	default:
		return "error"
	}
}

func parseLogFieldsArg(label string, v Value) ([]any, error) {
	if v == nil {
		return nil, nil
	}
	obj, ok := v.(Object)
	if !ok {
		return nil, fmt.Errorf("%s expects object, got %T", label, v)
	}
	out := make([]any, 0, len(obj)*2)
	for k, val := range obj {
		out = append(out, k, toLogValue(val))
	}
	return out, nil
}

func toLogValue(v Value) any {
	switch vv := v.(type) {
	case nil, string, bool, float64:
		return vv
	case Object:
		out := map[string]any{}
		for k, x := range vv {
			out[k] = toLogValue(x)
		}
		return out
	case Array:
		out := make([]any, 0, len(vv))
		for _, x := range vv {
			out = append(out, toLogValue(x))
		}
		return out
	default:
		return fmt.Sprintf("%v", vv)
	}
}
