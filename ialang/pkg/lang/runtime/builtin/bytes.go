package builtin

import (
	"encoding/base64"
	"fmt"
)

func newBytesModule() Object {
	fromStringFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("bytes.fromString expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("bytes.fromString", args, 0)
		if err != nil {
			return nil, err
		}
		return byteSliceToArray([]byte(text)), nil
	})

	toStringFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("bytes.toString expects 1 arg, got %d", len(args))
		}
		buf, err := asByteArray("bytes.toString arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		return string(buf), nil
	})

	fromBase64Fn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("bytes.fromBase64 expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("bytes.fromBase64", args, 0)
		if err != nil {
			return nil, err
		}
		raw, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			return nil, err
		}
		return byteSliceToArray(raw), nil
	})

	toBase64Fn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("bytes.toBase64 expects 1 arg, got %d", len(args))
		}
		buf, err := asByteArray("bytes.toBase64 arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		return base64.StdEncoding.EncodeToString(buf), nil
	})

	concatFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("bytes.concat expects >=1 args")
		}
		out := make([]byte, 0)
		for i, arg := range args {
			buf, err := asByteArray(fmt.Sprintf("bytes.concat arg[%d]", i), arg)
			if err != nil {
				return nil, err
			}
			out = append(out, buf...)
		}
		return byteSliceToArray(out), nil
	})

	sliceFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 || len(args) > 3 {
			return nil, fmt.Errorf("bytes.slice expects 2-3 args: bytes, start, [end]")
		}
		buf, err := asByteArray("bytes.slice arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		start, err := asIntValue("bytes.slice arg[1]", args[1])
		if err != nil {
			return nil, err
		}
		end := len(buf)
		if len(args) == 3 {
			end, err = asIntValue("bytes.slice arg[2]", args[2])
			if err != nil {
				return nil, err
			}
		}
		if start < 0 {
			start = 0
		}
		if end > len(buf) {
			end = len(buf)
		}
		if start > end {
			start = end
		}
		return byteSliceToArray(buf[start:end]), nil
	})

	lengthFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("bytes.length expects 1 arg, got %d", len(args))
		}
		buf, err := asByteArray("bytes.length arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		return float64(len(buf)), nil
	})

	namespace := Object{
		"fromString": fromStringFn,
		"toString":   toStringFn,
		"fromBase64": fromBase64Fn,
		"toBase64":   toBase64Fn,
		"concat":     concatFn,
		"slice":      sliceFn,
		"length":     lengthFn,
	}
	module := cloneObject(namespace)
	module["bytes"] = namespace
	return module
}

func asByteArray(label string, v Value) ([]byte, error) {
	arr, ok := v.(Array)
	if !ok {
		return nil, fmt.Errorf("%s expects array<number>, got %T", label, v)
	}
	out := make([]byte, 0, len(arr))
	for i, item := range arr {
		n, err := asIntValue(fmt.Sprintf("%s[%d]", label, i), item)
		if err != nil {
			return nil, err
		}
		if n < 0 || n > 255 {
			return nil, fmt.Errorf("%s[%d] out of byte range: %d", label, i, n)
		}
		out = append(out, byte(n))
	}
	return out, nil
}

func byteSliceToArray(buf []byte) Array {
	out := make(Array, 0, len(buf))
	for _, b := range buf {
		out = append(out, float64(b))
	}
	return out
}
