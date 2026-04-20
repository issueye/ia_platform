package builtin

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func asStringArg(fn string, args []Value, idx int) (string, error) {
	if idx < 0 || idx >= len(args) {
		return "", fmt.Errorf("%s arg[%d] is missing", fn, idx)
	}
	return asStringValue(fn+" arg["+strconv.Itoa(idx)+"]", args[idx])
}

func asStringValue(label string, v Value) (string, error) {
	switch vv := v.(type) {
	case string:
		return vv, nil
	case float64:
		return strconv.FormatFloat(vv, 'f', -1, 64), nil
	case bool:
		if vv {
			return "true", nil
		}
		return "false", nil
	case nil:
		return "", nil
	default:
		return "", fmt.Errorf("%s expects string-like value, got %T", label, v)
	}
}

func asBoolArg(fn string, args []Value, idx int) (bool, error) {
	if idx < 0 || idx >= len(args) {
		return false, fmt.Errorf("%s arg[%d] is missing", fn, idx)
	}
	v, ok := args[idx].(bool)
	if !ok {
		return false, fmt.Errorf("%s arg[%d] expects bool, got %T", fn, idx, args[idx])
	}
	return v, nil
}

func asIntArg(fn string, args []Value, idx int) (int, error) {
	if idx < 0 || idx >= len(args) {
		return 0, fmt.Errorf("%s arg[%d] is missing", fn, idx)
	}
	return asIntValue(fn+" arg["+strconv.Itoa(idx)+"]", args[idx])
}

func asIntValue(label string, v Value) (int, error) {
	switch vv := v.(type) {
	case float64:
		return int(vv), nil
	case string:
		i, err := strconv.Atoi(vv)
		if err != nil {
			return 0, fmt.Errorf("%s invalid integer string %q", label, vv)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("%s expects number or integer string, got %T", label, v)
	}
}

func headersToObject(headers http.Header) Object {
	out := Object{}
	for k, vals := range headers {
		out[k] = strings.Join(vals, ",")
	}
	return out
}

func envObject() Object {
	out := Object{}
	for _, pair := range os.Environ() {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			out[parts[0]] = parts[1]
		}
	}
	return out
}

func argsArray() Array {
	out := make(Array, 0, len(os.Args))
	for _, a := range os.Args {
		out = append(out, a)
	}
	return out
}

func toRuntimeJSONValue(v any) Value {
	return yamlToValue(v)
}

func cloneObject(src Object) Object {
	out := Object{}
	for k, v := range src {
		out[k] = v
	}
	return out
}

func toString(v Value) string {
	if v == nil {
		return "null"
	}
	switch vv := v.(type) {
	case string:
		return vv
	case float64:
		return fmt.Sprintf("%g", vv)
	case bool:
		if vv {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("%v", vv)
	}
}

func makeCwdFn(prefix string) NativeFunction {
	return NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("%s.cwd expects 0 args, got %d", prefix, len(args))
		}
		return os.Getwd()
	})
}

func makeGetEnvFn(prefix string) NativeFunction {
	return NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("%s.getEnv expects 1 arg, got %d", prefix, len(args))
		}
		key, err := asStringArg(prefix+".getEnv", args, 0)
		if err != nil {
			return nil, err
		}
		v, ok := os.LookupEnv(key)
		if !ok {
			return nil, nil
		}
		return v, nil
	})
}

func makeSetEnvFn(prefix string) NativeFunction {
	return NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("%s.setEnv expects 2 args, got %d", prefix, len(args))
		}
		key, err := asStringArg(prefix+".setEnv", args, 0)
		if err != nil {
			return nil, err
		}
		value, err := asStringArg(prefix+".setEnv", args, 1)
		if err != nil {
			return nil, err
		}
		if err := os.Setenv(key, value); err != nil {
			return nil, err
		}
		return true, nil
	})
}

func makeEnvFn(prefix string) NativeFunction {
	return NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("%s.env expects 0 args, got %d", prefix, len(args))
		}
		return envObject(), nil
	})
}
