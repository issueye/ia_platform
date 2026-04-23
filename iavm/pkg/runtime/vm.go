package runtime

import (
	"fmt"
	"iavm/pkg/core"
	"iavm/pkg/module"
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
