package lang

import (
	bc "iacommon/pkg/ialang/bytecode"
	commonrt "iacommon/pkg/ialang/runtime"
	rt "ialang/pkg/lang/runtime"
	rvm "ialang/pkg/lang/runtime/vm"
)

type Value = rt.Value
type Object = rt.Object
type AsyncRuntime = commonrt.AsyncRuntime
type GoroutineRuntimeOptions = rt.GoroutineRuntimeOptions

type VM = rvm.VM
type ImportResolver = rvm.ImportResolver
type VMOptions = rvm.VMOptions

func NewGoroutineRuntime() AsyncRuntime {
	return rt.NewGoroutineRuntime()
}

func NewGoroutineRuntimeWithOptions(options GoroutineRuntimeOptions) AsyncRuntime {
	return rt.NewGoroutineRuntimeWithOptions(options)
}

func NewVM(chunk *bc.Chunk, modules map[string]Value, resolve ImportResolver, modulePath string, asyncRuntime AsyncRuntime) *VM {
	return rvm.NewVM(chunk, modules, resolve, modulePath, asyncRuntime)
}

func NewVMWithOptions(chunk *bc.Chunk, modules map[string]Value, resolve ImportResolver, modulePath string, asyncRuntime AsyncRuntime, options VMOptions) *VM {
	return rvm.NewVMWithOptions(chunk, modules, resolve, modulePath, asyncRuntime, options)
}
