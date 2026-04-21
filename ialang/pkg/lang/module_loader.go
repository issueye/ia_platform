package lang

import (
	moduleapi "iacommon/pkg/ialang/module"
	comp "ialang/pkg/lang/compiler"
	"ialang/pkg/lang/frontend"
	"os"
	"strings"
	"sync"
)

type ModuleLoader struct {
	mu              sync.RWMutex
	builtins        map[string]Value
	builtinRegistry *moduleapi.BuiltinRegistry[Value]
	asyncRuntime    AsyncRuntime
	vmOptions       VMOptions
	resolverOptions ModuleResolverOptions
	cache           map[string]Object
	inflight        map[string]*moduleLoad
}

type moduleLoad struct {
	done    chan struct{}
	exports Object
	err     error
}

func NewModuleLoader(builtins map[string]Value, asyncRuntime AsyncRuntime) *ModuleLoader {
	return NewModuleLoaderWithOptions(builtins, asyncRuntime, VMOptions{})
}

func NewModuleLoaderWithOptions(builtins map[string]Value, asyncRuntime AsyncRuntime, vmOptions VMOptions) *ModuleLoader {
	return NewModuleLoaderWithResolverOptions(builtins, asyncRuntime, vmOptions, ModuleResolverOptions{})
}

func NewModuleLoaderWithResolverOptions(builtins map[string]Value, asyncRuntime AsyncRuntime, vmOptions VMOptions, resolverOptions ModuleResolverOptions) *ModuleLoader {
	if asyncRuntime == nil {
		asyncRuntime = NewGoroutineRuntime()
	}
	return &ModuleLoader{
		builtins:        builtins,
		builtinRegistry: moduleapi.BuiltinRegistryFromValues(builtins),
		asyncRuntime:    asyncRuntime,
		vmOptions:       vmOptions,
		resolverOptions: resolverOptions,
		cache:           map[string]Object{},
		inflight:        map[string]*moduleLoad{},
	}
}

func (m *ModuleLoader) Resolve(fromPath, moduleName string) (Value, error) {
	return m.resolveWithStack(fromPath, moduleName, map[string]bool{})
}

func (m *ModuleLoader) resolveWithStack(fromPath, moduleName string, stack map[string]bool) (Value, error) {
	if val, ok := m.builtinRegistry.Resolve(moduleName); ok {
		return val, nil
	}

	resolvedPath, err := ResolveModulePathWithOptions(fromPath, moduleName, m.resolverOptions)
	if err != nil {
		return nil, err
	}

	if stack[resolvedPath] {
		return nil, moduleapi.CyclicImportError(resolvedPath)
	}

	m.mu.RLock()
	if exports, ok := m.cache[resolvedPath]; ok {
		m.mu.RUnlock()
		return exports, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	if exports, ok := m.cache[resolvedPath]; ok {
		m.mu.Unlock()
		return exports, nil
	}
	if loading, ok := m.inflight[resolvedPath]; ok {
		m.mu.Unlock()
		<-loading.done
		if loading.err != nil {
			return nil, loading.err
		}
		return cloneObject(loading.exports), nil
	}
	loading := &moduleLoad{done: make(chan struct{})}
	m.inflight[resolvedPath] = loading
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.inflight, resolvedPath)
		m.mu.Unlock()
		close(loading.done)
	}()

	stack[resolvedPath] = true
	defer delete(stack, resolvedPath)

	src, err := os.ReadFile(resolvedPath)
	if err != nil {
		loading.err = moduleapi.ReadModuleError(resolvedPath, err)
		return nil, loading.err
	}

	l := frontend.NewLexer(string(src))
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if parseErrs := p.Errors(); len(parseErrs) > 0 {
		loading.err = moduleapi.ParseModuleError(resolvedPath, strings.Join(parseErrs, "; "))
		return nil, loading.err
	}

	c := comp.NewCompiler()
	chunk, compileErrs := c.Compile(program)
	if len(compileErrs) > 0 {
		loading.err = moduleapi.CompileModuleError(resolvedPath, strings.Join(compileErrs, "; "))
		return nil, loading.err
	}

	vm := NewVMWithOptions(
		chunk,
		m.builtins,
		func(fp, mn string) (Value, error) { return m.resolveWithStack(fp, mn, stack) },
		resolvedPath,
		m.asyncRuntime,
		m.vmOptions,
	)
	if err := vm.Run(); err != nil {
		loading.err = moduleapi.RuntimeModuleError(resolvedPath, err)
		return nil, loading.err
	}

	exports := cloneObject(vm.Exports())
	loading.exports = exports
	m.mu.Lock()
	m.cache[resolvedPath] = exports
	m.mu.Unlock()
	return exports, nil
}

func cloneObject(obj Object) Object {
	out := Object{}
	for k, v := range obj {
		out[k] = v
	}
	return out
}
