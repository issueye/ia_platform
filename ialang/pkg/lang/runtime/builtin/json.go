package builtin

import (
	goJSON "encoding/json"
	"fmt"
	"os"
)

func newJSONModule(asyncRuntime AsyncRuntime) Object {
	parseFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("json.parse expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("json.parse", args, 0)
		if err != nil {
			return nil, err
		}
		var out any
		if err := goJSON.Unmarshal([]byte(text), &out); err != nil {
			return nil, err
		}
		return toRuntimeJSONValue(out), nil
	})

	fromFileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("json.fromFile expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("json.fromFile", args, 0)
		if err != nil {
			return nil, err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("json.fromFile failed: %w", err)
		}
		var out any
		if err := goJSON.Unmarshal(data, &out); err != nil {
			return nil, fmt.Errorf("json.fromFile parse error: %w", err)
		}
		return toRuntimeJSONValue(out), nil
	})

	stringifyFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("json.stringify expects 1-2 args: value, [pretty]")
		}
		pretty := false
		if len(args) == 2 {
			v, ok := args[1].(bool)
			if !ok {
				return nil, fmt.Errorf("json.stringify arg[1] expects bool, got %T", args[1])
			}
			pretty = v
		}
		var (
			raw []byte
			err error
		)
		if pretty {
			raw, err = goJSON.MarshalIndent(args[0], "", "  ")
		} else {
			raw, err = goJSON.Marshal(args[0])
		}
		if err != nil {
			return nil, err
		}
		return string(raw), nil
	})
	validFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("json.valid expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("json.valid", args, 0)
		if err != nil {
			return nil, err
		}
		return goJSON.Valid([]byte(text)), nil
	})

	saveToFileFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 || len(args) > 3 {
			return nil, fmt.Errorf("json.saveToFile expects 2-3 args: value, path, [pretty]")
		}
		path, err := asStringArg("json.saveToFile", args, 1)
		if err != nil {
			return nil, err
		}
		pretty := false
		if len(args) == 3 {
			v, ok := args[2].(bool)
			if !ok {
				return nil, fmt.Errorf("json.saveToFile arg[2] expects bool, got %T", args[2])
			}
			pretty = v
		}
		var raw []byte
		if pretty {
			raw, err = goJSON.MarshalIndent(args[0], "", "  ")
		} else {
			raw, err = goJSON.Marshal(args[0])
		}
		if err != nil {
			return nil, fmt.Errorf("json.saveToFile marshal error: %w", err)
		}
		if err := os.WriteFile(path, raw, 0644); err != nil {
			return nil, fmt.Errorf("json.saveToFile write error: %w", err)
		}
		return true, nil
	})

	fromFileAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("json.fromFileAsync expects 1 arg, got %d", len(args))
		}
		path, err := asStringArg("json.fromFileAsync", args, 0)
		if err != nil {
			return nil, err
		}
		return asyncRuntime.Spawn(func() (Value, error) {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("json.fromFileAsync failed: %w", err)
			}
			var out any
			if err := goJSON.Unmarshal(data, &out); err != nil {
				return nil, fmt.Errorf("json.fromFileAsync parse error: %w", err)
			}
			return toRuntimeJSONValue(out), nil
		}), nil
	})

	saveToFileAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 || len(args) > 3 {
			return nil, fmt.Errorf("json.saveToFileAsync expects 2-3 args: value, path, [pretty]")
		}
		path, err := asStringArg("json.saveToFileAsync", args, 1)
		if err != nil {
			return nil, err
		}
		pretty := false
		if len(args) == 3 {
			v, ok := args[2].(bool)
			if !ok {
				return nil, fmt.Errorf("json.saveToFileAsync arg[2] expects bool, got %T", args[2])
			}
			pretty = v
		}
		val := args[0]
		return asyncRuntime.Spawn(func() (Value, error) {
			var raw []byte
			if pretty {
				raw, err = goJSON.MarshalIndent(val, "", "  ")
			} else {
				raw, err = goJSON.Marshal(val)
			}
			if err != nil {
				return nil, fmt.Errorf("json.saveToFileAsync marshal error: %w", err)
			}
			if err := os.WriteFile(path, raw, 0644); err != nil {
				return nil, fmt.Errorf("json.saveToFileAsync write error: %w", err)
			}
			return true, nil
		}), nil
	})

	namespace := Object{
		"parse":           parseFn,
		"fromFile":        fromFileFn,
		"fromFileAsync":   fromFileAsyncFn,
		"stringify":       stringifyFn,
		"saveToFile":      saveToFileFn,
		"saveToFileAsync": saveToFileAsyncFn,
		"valid":           validFn,
	}
	module := cloneObject(namespace)
	module["json"] = namespace
	return module
}
