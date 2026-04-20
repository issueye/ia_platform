package builtin

import (
	"fmt"
	"os"
)

func newProcessModule() Object {
	pidFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("process.pid expects 0 args, got %d", len(args))
		}
		return float64(os.Getpid()), nil
	})
	ppidFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("process.ppid expects 0 args, got %d", len(args))
		}
		return float64(os.Getppid()), nil
	})
	argsFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("process.args expects 0 args, got %d", len(args))
		}
		return argsArray(), nil
	})
	chdirFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("process.chdir expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("process.chdir", args, 0)
		if err != nil {
			return nil, err
		}
		if err := os.Chdir(path); err != nil {
			return nil, err
		}
		return true, nil
	})
	exitFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) > 1 {
			return nil, fmt.Errorf("process.exit expects 0-1 args, got %d", len(args))
		}
		code := 0
		if len(args) == 1 {
			parsed, err := asIntArg("process.exit", args, 0)
			if err != nil {
				return nil, err
			}
			code = parsed
		}
		return nil, fmt.Errorf("process.exit(%d) is disabled in embedded runtime", code)
	})

	namespace := Object{
		"pid":    pidFn,
		"ppid":   ppidFn,
		"args":   argsFn,
		"cwd":    makeCwdFn("process"),
		"chdir":  chdirFn,
		"getEnv": makeGetEnvFn("process"),
		"setEnv": makeSetEnvFn("process"),
		"env":    makeEnvFn("process"),
		"exit":   exitFn,
	}
	module := cloneObject(namespace)
	module["process"] = namespace
	return module
}
