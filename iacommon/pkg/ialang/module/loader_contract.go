package module

import (
	"errors"
	"fmt"
)

const (
	PlatformFSModuleName   = "@platform/fs"
	PlatformHTTPModuleName = "@platform/http"
)

var (
	ErrUnknownModule  = errors.New("unknown module")
	ErrModuleNotFound = errors.New("module not found")
	ErrCyclicImport   = errors.New("cyclic import detected")
	ErrReadModule     = errors.New("read module error")
	ErrParseModule    = errors.New("parse module error")
	ErrCompileModule  = errors.New("compile module error")
	ErrRuntimeModule  = errors.New("runtime module error")
)

func UnknownModuleError(name string) error {
	return fmt.Errorf("%w: %s", ErrUnknownModule, name)
}

func ModuleNotFoundError(name string) error {
	return fmt.Errorf("%w: %s", ErrModuleNotFound, name)
}

func CyclicImportError(modulePath string) error {
	return fmt.Errorf("%w: %s", ErrCyclicImport, modulePath)
}

func ReadModuleError(modulePath string, err error) error {
	return fmt.Errorf("%w (%s): %w", ErrReadModule, modulePath, err)
}

func ParseModuleError(modulePath, details string) error {
	return fmt.Errorf("%w (%s): %s", ErrParseModule, modulePath, details)
}

func CompileModuleError(modulePath, details string) error {
	return fmt.Errorf("%w (%s): %s", ErrCompileModule, modulePath, details)
}

func RuntimeModuleError(modulePath string, err error) error {
	return fmt.Errorf("%w (%s): %w", ErrRuntimeModule, modulePath, err)
}
