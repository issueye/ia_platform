package builtin

import (
	"fmt"
	"path/filepath"
)

func newPathModule() Object {
	joinFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("path.join expects at least 1 arg")
		}
		parts := make([]string, 0, len(args))
		for i := range args {
			part, err := asStringArg("path.join", args, i)
			if err != nil {
				return nil, err
			}
			parts = append(parts, part)
		}
		return filepath.Join(parts...), nil
	})
	baseFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("path.base expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("path.base", args, 0)
		if err != nil {
			return nil, err
		}
		return filepath.Base(path), nil
	})
	dirFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("path.dir expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("path.dir", args, 0)
		if err != nil {
			return nil, err
		}
		return filepath.Dir(path), nil
	})
	extFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("path.ext expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("path.ext", args, 0)
		if err != nil {
			return nil, err
		}
		return filepath.Ext(path), nil
	})
	cleanFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("path.clean expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("path.clean", args, 0)
		if err != nil {
			return nil, err
		}
		return filepath.Clean(path), nil
	})
	absFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("path.abs expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("path.abs", args, 0)
		if err != nil {
			return nil, err
		}
		return filepath.Abs(path)
	})

	namespace := Object{
		"join":       joinFn,
		"base":       baseFn,
		"dir":        dirFn,
		"ext":        extFn,
		"clean":      cleanFn,
		"abs":        absFn,
		"resolve":    joinFn,
		"normalize":  cleanFn,
		"isAbsolute": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("path.isAbsolute expects 1 arg, got %d", len(args))
			}
			p, err := asStringArg("path.isAbsolute", args, 0)
			if err != nil {
				return nil, err
			}
			return filepath.IsAbs(p), nil
		}),
		"relative": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("path.relative expects 2 args, got %d", len(args))
			}
			basePath, err := asStringArg("path.relative", args, 0)
			if err != nil {
				return nil, err
			}
			targetPath, err := asStringArg("path.relative", args, 1)
			if err != nil {
				return nil, err
			}
			return filepath.Rel(basePath, targetPath)
		}),
		"sep":       string(filepath.Separator),
		"listSep":   string(filepath.ListSeparator),
	}
	module := cloneObject(namespace)
	module["path"] = namespace
	return module
}
