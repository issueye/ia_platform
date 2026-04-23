package runtime

import (
	"context"
	"errors"
	"fmt"
	"iavm/pkg/core"
	"iavm/pkg/module"
	"time"

	"iacommon/pkg/host/api"
)

type CompiledFunction struct {
	Name  string
	Index uint32
}

type BuiltinFunc func(args []core.Value) core.Value

type VM struct {
	mod              *module.Module
	options          Options
	stack            *Stack
	globals          []core.Value
	functions        []CompiledFunction
	handles          *HandleTable
	frames           []*Frame
	capabilityIDs    map[uint32]string
	lastCapabilityID string
	exception        core.Value // current uncaught exception value
	suspension       *Suspension
	startedAt        int64
	stepCount        int64
	builtins         map[string]BuiltinFunc
}

type Suspension struct {
	Reason     string
	AwaitValue core.Value
	FrameDepth int
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
		result, err := waiter.Wait(ctx, state.PollHandleID)
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
	if err := vm.Run(); err != nil {
		if !errors.Is(err, ErrPromisePending) {
			return err
		}
	} else {
		return nil
	}

	for {
		if err := vm.WaitSuspension(ctx); err != nil {
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
		result, err := vm.options.Host.Poll(ctx, state.PollHandleID)
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

func (vm *VM) runFunctionSync(fnIdx uint32, args []core.Value, fnRef core.Value) (core.Value, error) {
	child := &VM{
		mod:              vm.mod,
		options:          vm.options,
		stack:            NewStack(256),
		globals:          vm.globals,
		functions:        vm.functions,
		handles:          vm.handles,
		capabilityIDs:    vm.capabilityIDs,
		lastCapabilityID: vm.lastCapabilityID,
		builtins:         vm.builtins,
	}
	if err := child.pushCallFrame(fnIdx, args, fnRef); err != nil {
		return core.Value{}, err
	}
	if err := Interpret(child, fnIdx); err != nil {
		return core.Value{}, err
	}
	vm.capabilityIDs = child.capabilityIDs
	vm.lastCapabilityID = child.lastCapabilityID
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
		result, err := vm.options.Host.Poll(context.Background(), state.PollHandleID)
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
