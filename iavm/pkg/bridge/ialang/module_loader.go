package ialang

import moduleapi "iacommon/pkg/ialang/module"

var (
	ErrUnknownModule = moduleapi.ErrUnknownModule
	builtinRegistry  = newBuiltinRegistry()
)

func ResolveModule(name string) (any, error) {
	if moduleValue, ok := builtinRegistry.Resolve(name); ok {
		return moduleValue, nil
	}
	return nil, moduleapi.UnknownModuleError(name)
}

func newBuiltinRegistry() *moduleapi.BuiltinRegistry[any] {
	registry := moduleapi.NewBuiltinRegistry[any]()
	registry.RegisterProvider(PlatformFSModuleName, func() any {
		return BuildPlatformFSModule()
	})
	registry.RegisterProvider(PlatformHTTPModuleName, func() any {
		return BuildPlatformHTTPModule()
	})
	return registry
}
