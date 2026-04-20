package builtin

import (
	"fmt"
	"os"
)

func newFSModule(asyncRuntime AsyncRuntime) Object {
	readFileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("fs.readFile expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("fs.readFile", args, 0)
		if err != nil {
			return nil, err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		return string(raw), nil
	})

	writeFileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("fs.writeFile expects 2 args, got %d", len(args))
		}
		path, err := asStringArg("fs.writeFile", args, 0)
		if err != nil {
			return nil, err
		}
		content, err := asStringArg("fs.writeFile", args, 1)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return nil, err
		}
		return true, nil
	})

	appendFileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("fs.appendFile expects 2 args, got %d", len(args))
		}
		path, err := asStringArg("fs.appendFile", args, 0)
		if err != nil {
			return nil, err
		}
		content, err := asStringArg("fs.appendFile", args, 1)
		if err != nil {
			return nil, err
		}
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		if _, err := f.WriteString(content); err != nil {
			return nil, err
		}
		return true, nil
	})

	existsFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("fs.exists expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("fs.exists", args, 0)
		if err != nil {
			return nil, err
		}
		_, err = os.Stat(path)
		if err == nil {
			return true, nil
		}
		if os.IsNotExist(err) {
			return false, nil
		}
		return nil, err
	})

	mkdirFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("fs.mkdir expects 1-2 args: path, [recursive]")
		}
		path, err := asStringArg("fs.mkdir", args, 0)
		if err != nil {
			return nil, err
		}
		recursive := false
		if len(args) == 2 {
			recursive, err = asBoolArg("fs.mkdir", args, 1)
			if err != nil {
				return nil, err
			}
		}
		if recursive {
			if err := os.MkdirAll(path, 0o755); err != nil {
				return nil, err
			}
		} else {
			if err := os.Mkdir(path, 0o755); err != nil {
				return nil, err
			}
		}
		return true, nil
	})

	readDirFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("fs.readDir expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("fs.readDir", args, 0)
		if err != nil {
			return nil, err
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		out := make(Array, 0, len(entries))
		for _, e := range entries {
			out = append(out, e.Name())
		}
		return out, nil
	})

	statFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("fs.stat expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("fs.stat", args, 0)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		return Object{
			"name":        info.Name(),
			"isDir":       info.IsDir(),
			"size":        float64(info.Size()),
			"mode":        info.Mode().String(),
			"modTimeUnix": float64(info.ModTime().Unix()),
		}, nil
	})

	readFileAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return readFileFn(args)
		}), nil
	})
	writeFileAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return writeFileFn(args)
		}), nil
	})
	appendFileAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return appendFileFn(args)
		}), nil
	})
	existsAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return existsFn(args)
		}), nil
	})
	mkdirAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return mkdirFn(args)
		}), nil
	})
	readDirAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return readDirFn(args)
		}), nil
	})
	statAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return statFn(args)
		}), nil
	})

	renameFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("fs.rename expects 2 args, got %d", len(args))
		}
		oldPath, err := asStringArg("fs.rename", args, 0)
		if err != nil {
			return nil, err
		}
		newPath, err := asStringArg("fs.rename", args, 1)
		if err != nil {
			return nil, err
		}
		if err := os.Rename(oldPath, newPath); err != nil {
			return nil, err
		}
		return true, nil
	})

	removeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("fs.remove expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("fs.remove", args, 0)
		if err != nil {
			return nil, err
		}
		if err := os.Remove(path); err != nil {
			return nil, err
		}
		return true, nil
	})

	removeAllFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("fs.removeAll expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("fs.removeAll", args, 0)
		if err != nil {
			return nil, err
		}
		if err := os.RemoveAll(path); err != nil {
			return nil, err
		}
		return true, nil
	})

	copyFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("fs.copy expects 2 args, got %d", len(args))
		}
		src, err := asStringArg("fs.copy", args, 0)
		if err != nil {
			return nil, err
		}
		dst, err := asStringArg("fs.copy", args, 1)
		if err != nil {
			return nil, err
		}
		raw, err := os.ReadFile(src)
		if err != nil {
			return nil, err
		}
		if err := os.WriteFile(dst, raw, 0o644); err != nil {
			return nil, err
		}
		return true, nil
	})

	namespace := Object{
		"readFile":        readFileFn,
		"writeFile":       writeFileFn,
		"appendFile":      appendFileFn,
		"exists":          existsFn,
		"mkdir":           mkdirFn,
		"readDir":         readDirFn,
		"stat":            statFn,
		"rename":          renameFn,
		"remove":          removeFn,
		"removeAll":       removeAllFn,
		"copy":            copyFn,
		"readFileAsync":   readFileAsyncFn,
		"writeFileAsync":  writeFileAsyncFn,
		"appendFileAsync": appendFileAsyncFn,
		"existsAsync":     existsAsyncFn,
		"mkdirAsync":      mkdirAsyncFn,
		"readDirAsync":    readDirAsyncFn,
		"statAsync":       statAsyncFn,
	}
	module := cloneObject(namespace)
	module["fs"] = namespace
	return module
}
