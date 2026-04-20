package runtime

import "iavm/pkg/module"

type CompiledFunction struct {
	Name  string
	Index uint32
}

type VM struct {
	mod       *module.Module
	options   Options
	stack     []any
	globals   map[string]any
	functions []CompiledFunction
	handles   *HandleTable
	startedAt int64
	stepCount int64
}

func New(mod *module.Module, opts Options) (*VM, error) {
	vm := &VM{
		mod:       mod,
		options:   opts,
		stack:     make([]any, 0, 64),
		globals:   map[string]any{},
		functions: make([]CompiledFunction, 0),
		handles:   NewHandleTable(),
	}
	return vm, nil
}

func (vm *VM) Run() error {
	_ = vm
	return nil
}

func (vm *VM) InvokeExport(name string, args ...any) (any, error) {
	_ = name
	_ = args
	return nil, nil
}
