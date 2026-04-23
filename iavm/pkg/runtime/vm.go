package runtime

import (
	"context"
	"errors"
	"fmt"
	"iavm/pkg/core"
	"iavm/pkg/module"
	"strconv"
	"time"

	"iacommon/pkg/host/api"
)

type CompiledFunction struct {
	Name  string
	Index uint32
}

type BuiltinFunc func(args []core.Value) core.Value

type VM struct {
	mod                *module.Module
	options            Options
	runCtx             context.Context
	runCancel          context.CancelFunc
	stack              *Stack
	globals            []core.Value
	functions          []CompiledFunction
	handles            *HandleTable
	frames             []*Frame
	capabilityIDs      map[uint32]string
	lastCapabilityID   string
	lastCapabilityKind module.CapabilityKind
	exception          core.Value // current uncaught exception value
	suspension         *Suspension
	startedAt          int64
	stepCount          int64
	builtins           map[string]BuiltinFunc
}

type Suspension struct {
	Reason     string
	AwaitValue core.Value
	FrameDepth int
}

type capabilityTimeoutProfile struct {
	HostTimeout  time.Duration
	WaitTimeout  time.Duration
	RetryCount   int
	RetryBackoff time.Duration
}

func New(mod *module.Module, opts Options) (*VM, error) {
	vm := &VM{
		mod:       mod,
		options:   opts,
		stack:     NewStack(256),
		globals:   make([]core.Value, 0, 64),
		functions: make([]CompiledFunction, 0),
		handles:   NewHandleTable(),
		builtins:  make(map[string]BuiltinFunc),
	}

	// Index functions
	for i, fn := range mod.Functions {
		vm.functions = append(vm.functions, CompiledFunction{
			Name:  fn.Name,
			Index: uint32(i),
		})
	}

	// Register default builtins
	vm.registerBuiltin("print", builtinPrint)
	vm.registerBuiltin("len", builtinLen)
	vm.registerBuiltin("typeof", builtinTypeof)
	vm.registerBuiltin("str", builtinStr)
	vm.registerBuiltin("int", builtinInt)
	vm.registerBuiltin("float", builtinFloat)

	return vm, nil
}

func (vm *VM) registerBuiltin(name string, fn BuiltinFunc) {
	vm.builtins[name] = fn
}

func (vm *VM) GetBuiltin(name string) (BuiltinFunc, bool) {
	fn, ok := vm.builtins[name]
	return fn, ok
}

func (vm *VM) Run() error {
	ownsContext := vm.runCtx == nil
	if ownsContext {
		vm.beginRunContext(context.Background())
		defer vm.endRunContext()
	}
	vm.suspension = nil
	// Find entry function
	var entryIdx *uint32
	for i, fn := range vm.mod.Functions {
		if fn.IsEntryPoint || fn.Name == "main" || fn.Name == "entry" {
			idx := uint32(i)
			entryIdx = &idx
			break
		}
	}
	if entryIdx == nil && len(vm.mod.Functions) > 0 {
		idx := uint32(0)
		entryIdx = &idx
	}
	if entryIdx == nil {
		return core.ErrInvalidModule
	}

	return Interpret(vm, *entryIdx)
}

func (vm *VM) InvokeExport(name string, args ...any) (any, error) {
	for _, exp := range vm.mod.Exports {
		if exp.Name == name && exp.Kind == module.ExportFunction {
			frame := NewFrame(exp.Index, &vm.mod.Functions[exp.Index], uint32(vm.stack.Size()))
			vm.frames = append(vm.frames, frame)
			err := Interpret(vm, exp.Index)
			if err != nil {
				return nil, err
			}
			if vm.stack.Size() > 0 {
				return vm.stack.Pop(), nil
			}
			return nil, nil
		}
	}
	return nil, fmt.Errorf("export not found: %s", name)
}

func (vm *VM) PopResult() (core.Value, bool) {
	if vm.stack.Size() == 0 {
		return core.Value{}, false
	}
	return vm.stack.Pop(), true
}

func (vm *VM) StackSize() int {
	return vm.stack.Size()
}

func (vm *VM) SuspensionState() *Suspension {
	return vm.suspension
}

func (vm *VM) ResumeSuspension() error {
	if vm.suspension == nil {
		return fmt.Errorf("vm is not suspended")
	}

	resolved, err := vm.resolveSuspendedValue(vm.suspension.AwaitValue)
	if err != nil {
		if errors.Is(err, ErrPromisePending) {
			return err
		}
		vm.suspension = nil
		return err
	}

	vm.stack.Push(resolved)
	vm.suspension = nil
	if len(vm.frames) == 0 {
		return nil
	}
	return Interpret(vm, vm.frames[len(vm.frames)-1].FunctionIndex)
}

func (vm *VM) WaitSuspension(ctx context.Context) error {
	if vm.suspension == nil {
		return fmt.Errorf("vm is not suspended")
	}
	if vm.options.Host == nil {
		return fmt.Errorf("no host configured for suspended wait")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	promise := vm.suspension.AwaitValue
	if promise.Kind != core.ValuePromise {
		return vm.ResumeSuspension()
	}
	state, ok := promise.Raw.(*promiseState)
	if !ok || state == nil {
		return fmt.Errorf("invalid promise value")
	}
	if state.Status != promiseStatusPending || state.PollHandleID == 0 {
		return vm.ResumeSuspension()
	}

	waiter, ok := vm.options.Host.(api.Waiter)
	if ok {
		waitTimeout := vm.options.WaitTimeout
		if state.WaitTimeout > 0 {
			waitTimeout = state.WaitTimeout
		}
		result, err := vm.retryPollLike(ctx, state.RetryCount, state.RetryBackoff, func() (api.PollResult, error) {
			waitCtx, cancel := vm.hostOperationContext(ctx, waitTimeout)
			defer cancel()
			return waiter.Wait(waitCtx, state.PollHandleID)
		})
		if err != nil {
			return err
		}
		switch {
		case !result.Done:
			return ErrPromisePending
		case result.Error != "":
			state.Status = promiseStatusRejected
			state.Error = result.Error
		default:
			state.Status = promiseStatusResolved
			state.Result = coreValueFromHostPoll(result)
		}
		return vm.ResumeSuspension()
	}

	return vm.waitSuspensionByPolling(ctx, state)
}

func (vm *VM) RunUntilSettled(ctx context.Context) error {
	vm.beginRunContext(ctx)
	defer vm.endRunContext()
	if err := vm.Run(); err != nil {
		if !errors.Is(err, ErrPromisePending) {
			return err
		}
	} else {
		return nil
	}

	for {
		if err := vm.WaitSuspension(vm.hostContext()); err != nil {
			if errors.Is(err, ErrPromisePending) {
				continue
			}
			return err
		}
		return nil
	}
}

func (vm *VM) waitSuspensionByPolling(ctx context.Context, state *promiseState) error {
	interval := vm.options.WaitInterval
	if interval <= 0 {
		interval = 10 * time.Millisecond
	}
	for {
		pollTimeout := vm.options.HostTimeout
		if state.HostTimeout > 0 {
			pollTimeout = state.HostTimeout
		}
		result, err := vm.retryPollLike(ctx, state.RetryCount, state.RetryBackoff, func() (api.PollResult, error) {
			pollCtx, cancel := vm.hostOperationContext(ctx, pollTimeout)
			defer cancel()
			return vm.options.Host.Poll(pollCtx, state.PollHandleID)
		})
		if err != nil {
			return fmt.Errorf("host.poll failed during wait: %w", err)
		}
		switch {
		case result.Done && result.Error != "":
			state.Status = promiseStatusRejected
			state.Error = result.Error
			return vm.ResumeSuspension()
		case result.Done:
			state.Status = promiseStatusResolved
			state.Result = coreValueFromHostPoll(result)
			return vm.ResumeSuspension()
		}

		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (vm *VM) hostContext() context.Context {
	if vm.runCtx != nil {
		return vm.runCtx
	}
	return context.Background()
}

func (vm *VM) hostOperationContext(base context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if base == nil {
		base = vm.hostContext()
	}
	if timeout <= 0 {
		return base, func() {}
	}
	return context.WithTimeout(base, timeout)
}

func (vm *VM) retryPollLike(ctx context.Context, retryCount int, retryBackoff time.Duration, op func() (api.PollResult, error)) (api.PollResult, error) {
	attempts := retryCount + 1
	if attempts <= 0 {
		attempts = 1
	}
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		result, err := op()
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !vm.shouldRetryPollLike(ctx, err) || attempt == attempts-1 {
			return api.PollResult{}, err
		}
		if err := vm.sleepBackoff(ctx, retryBackoff); err != nil {
			return api.PollResult{}, err
		}
	}
	return api.PollResult{}, lastErr
}

func (vm *VM) shouldRetryPollLike(ctx context.Context, err error) bool {
	if err == nil {
		return false
	}
	if ctx != nil && ctx.Err() != nil {
		return false
	}
	return errors.Is(err, context.DeadlineExceeded)
}

func (vm *VM) sleepBackoff(ctx context.Context, backoff time.Duration) error {
	if backoff <= 0 {
		return nil
	}
	timer := time.NewTimer(backoff)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (vm *VM) beginRunContext(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	if vm.runCancel != nil {
		vm.runCancel()
		vm.runCancel = nil
	}
	if vm.options.MaxDuration > 0 {
		vm.runCtx, vm.runCancel = context.WithTimeout(ctx, vm.options.MaxDuration)
		return
	}
	vm.runCtx = ctx
}

func (vm *VM) endRunContext() {
	if vm.runCancel != nil {
		vm.runCancel()
		vm.runCancel = nil
	}
	vm.runCtx = nil
}

func (vm *VM) resolveStringConstant(fn *module.Function, index uint32) (string, bool) {
	var value any
	if len(vm.mod.Constants) > 0 {
		if int(index) >= len(vm.mod.Constants) {
			return "", false
		}
		value = vm.mod.Constants[index]
	} else {
		if int(index) >= len(fn.Constants) {
			return "", false
		}
		value = fn.Constants[index]
	}
	text, ok := value.(string)
	return text, ok
}

func (vm *VM) capabilityConfig(kind module.CapabilityKind) map[string]any {
	for _, capability := range vm.mod.Capabilities {
		if capability.Kind != kind {
			continue
		}
		if len(capability.Config) == 0 {
			return nil
		}
		result := make(map[string]any, len(capability.Config))
		for key, value := range capability.Config {
			result[key] = value
		}
		return result
	}
	return nil
}

func (vm *VM) capabilityTimeoutProfile(kind module.CapabilityKind) capabilityTimeoutProfile {
	profile := capabilityTimeoutProfile{
		HostTimeout:  vm.options.HostTimeout,
		WaitTimeout:  vm.options.WaitTimeout,
		RetryCount:   vm.options.RetryCount,
		RetryBackoff: vm.options.RetryBackoff,
	}
	config := vm.capabilityConfig(kind)
	if len(config) == 0 {
		return profile
	}
	if timeout, ok := readDurationMS(config, "host_timeout_ms", "hostTimeoutMS"); ok {
		profile.HostTimeout = timeout
	}
	if timeout, ok := readDurationMS(config, "wait_timeout_ms", "waitTimeoutMS"); ok {
		profile.WaitTimeout = timeout
	}
	if retryCount, ok := readInt(config, "retry_count", "retryCount"); ok {
		profile.RetryCount = retryCount
	}
	if backoff, ok := readDurationMS(config, "retry_backoff_ms", "retryBackoffMS"); ok {
		profile.RetryBackoff = backoff
	}
	return profile
}

func readInt(values map[string]any, keys ...string) (int, bool) {
	for _, key := range keys {
		value, ok := values[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case int:
			return typed, true
		case int64:
			return int(typed), true
		case uint64:
			return int(typed), true
		case float64:
			return int(typed), true
		case string:
			parsed, err := strconv.ParseInt(typed, 10, 64)
			if err == nil {
				return int(parsed), true
			}
		}
	}
	return 0, false
}

func readDurationMS(values map[string]any, keys ...string) (time.Duration, bool) {
	for _, key := range keys {
		value, ok := values[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case int:
			return time.Duration(typed) * time.Millisecond, true
		case int64:
			return time.Duration(typed) * time.Millisecond, true
		case uint64:
			return time.Duration(typed) * time.Millisecond, true
		case float64:
			return time.Duration(typed * float64(time.Millisecond)), true
		case string:
			parsed, err := strconv.ParseInt(typed, 10, 64)
			if err == nil {
				return time.Duration(parsed) * time.Millisecond, true
			}
		}
	}
	return 0, false
}

func (vm *VM) runFunctionSync(fnIdx uint32, args []core.Value, fnRef core.Value) (core.Value, error) {
	child := &VM{
		mod:                vm.mod,
		options:            vm.options,
		stack:              NewStack(256),
		globals:            vm.globals,
		functions:          vm.functions,
		handles:            vm.handles,
		capabilityIDs:      vm.capabilityIDs,
		lastCapabilityID:   vm.lastCapabilityID,
		lastCapabilityKind: vm.lastCapabilityKind,
		builtins:           vm.builtins,
	}
	if err := child.pushCallFrame(fnIdx, args, fnRef); err != nil {
		return core.Value{}, err
	}
	if err := Interpret(child, fnIdx); err != nil {
		return core.Value{}, err
	}
	vm.capabilityIDs = child.capabilityIDs
	vm.lastCapabilityID = child.lastCapabilityID
	vm.lastCapabilityKind = child.lastCapabilityKind
	if child.stack.Size() == 0 {
		return core.Value{Kind: core.ValueNull}, nil
	}
	return child.stack.Pop(), nil
}

func (vm *VM) resolveSuspendedValue(v core.Value) (core.Value, error) {
	if v.Kind != core.ValuePromise {
		return v, nil
	}

	state, ok := v.Raw.(*promiseState)
	if !ok || state == nil {
		return core.Value{}, fmt.Errorf("invalid promise value")
	}
	if state.Status == promiseStatusPending && state.PollHandleID != 0 {
		if vm.options.Host == nil {
			return core.Value{}, fmt.Errorf("no host configured for suspended poll")
		}
		pollTimeout := vm.options.HostTimeout
		if state.HostTimeout > 0 {
			pollTimeout = state.HostTimeout
		}
		result, err := vm.retryPollLike(vm.hostContext(), state.RetryCount, state.RetryBackoff, func() (api.PollResult, error) {
			pollCtx, cancel := vm.hostOperationContext(vm.hostContext(), pollTimeout)
			defer cancel()
			return vm.options.Host.Poll(pollCtx, state.PollHandleID)
		})
		if err != nil {
			return core.Value{}, fmt.Errorf("host.poll failed during resume: %w", err)
		}
		switch {
		case !result.Done:
			return core.Value{}, ErrPromisePending
		case result.Error != "":
			state.Status = promiseStatusRejected
			state.Error = result.Error
		default:
			state.Status = promiseStatusResolved
			state.Result = coreValueFromHostPoll(result)
		}
	}

	return awaitValue(v)
}
