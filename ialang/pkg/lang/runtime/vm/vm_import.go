package vm

import (
	"errors"
	"fmt"
)

func (v *VM) execImportName(moduleIdx, symbolIdx int) error {
	moduleName, ok := v.chunk.Constants[moduleIdx].(string)
	if !ok {
		return errors.New("module name is not a string constant")
	}
	symbol, ok := v.chunk.Constants[symbolIdx].(string)
	if !ok {
		return errors.New("symbol name is not a string constant")
	}
	moduleObj, err := v.loadModuleObject(moduleName)
	if err != nil {
		return err
	}
	val, exists := moduleObj[symbol]
	if !exists {
		return fmt.Errorf("symbol %s not found in module %s", symbol, moduleName)
	}
	v.defineName(symbol, val)
	return nil
}

func (v *VM) execImportNamespace(moduleIdx, targetIdx int) error {
	moduleName, ok := v.chunk.Constants[moduleIdx].(string)
	if !ok {
		return errors.New("module name is not a string constant")
	}
	target, ok := v.chunk.Constants[targetIdx].(string)
	if !ok {
		return errors.New("namespace target is not a string constant")
	}
	moduleObj, err := v.loadModuleObject(moduleName)
	if err != nil {
		return err
	}
	v.defineName(target, moduleObj)
	return nil
}

func (v *VM) execImportDynamic() error {
	moduleNameVal, err := v.pop()
	if err != nil {
		return err
	}
	moduleName, ok := moduleNameVal.(string)
	if !ok {
		return fmt.Errorf("dynamic import expects string module name, got %T", moduleNameVal)
	}
	promise := v.asyncRuntime.Spawn(func() (Value, error) {
		return v.loadModuleObject(moduleName)
	})
	v.push(promise)
	return nil
}

func (v *VM) execExportAll(moduleIdx int) error {
	moduleName, ok := v.chunk.Constants[moduleIdx].(string)
	if !ok {
		return errors.New("module name is not a string constant")
	}
	moduleObj, err := v.loadModuleObject(moduleName)
	if err != nil {
		return err
	}
	for name, val := range moduleObj {
		if name == "default" {
			continue
		}
		v.exports[name] = val
	}
	return nil
}

func (v *VM) loadModuleObject(moduleName string) (Object, error) {
	// Check sandbox policy
	if v.sandbox != nil && !v.sandbox.IsModuleAllowed(moduleName) {
		return nil, &SandboxError{
			Violation: "module not allowed",
			Limit:     "sandbox policy",
			Current:   moduleName,
		}
	}

	moduleVal, exists := v.modules[moduleName]
	if !exists {
		if v.resolveImport == nil {
			return nil, fmt.Errorf("module not found: %s", moduleName)
		}
		var err error
		moduleVal, err = v.resolveImport(v.modulePath, moduleName)
		if err != nil {
			return nil, err
		}
	}
	moduleObj, ok := moduleVal.(Object)
	if !ok {
		return nil, fmt.Errorf("module is not an object: %s", moduleName)
	}
	return moduleObj, nil
}
