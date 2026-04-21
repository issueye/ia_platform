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

type VM struct {
	mod           *module.Module
	options       Options
	stack         *Stack
	globals       []core.Value
	functions     []CompiledFunction
	handles       *HandleTable
	frames        []*Frame
	capabilityIDs map[uint32]string
	tryStack      []uint32
	startedAt     int64
	stepCount     int64
}

func New(mod *module.Module, opts Options) (*VM, error) {
	vm := &VM{
		mod:       mod,
		options:   opts,
		stack:     NewStack(256),
		globals:   make([]core.Value, 0, 64),
		functions: make([]CompiledFunction, 0),
		handles:   NewHandleTable(),
	}

	// Index functions
	for i, fn := range mod.Functions {
		vm.functions = append(vm.functions, CompiledFunction{
			Name:  fn.Name,
			Index: uint32(i),
		})
	}

	return vm, nil
}

func (vm *VM) Run() error {
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
