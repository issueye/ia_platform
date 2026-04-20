package builtin

import (
	"encoding/hex"
	"fmt"
)

func hexDecodeString(s string) ([]byte, error) {
	return hex.DecodeString(s)
}

func newHexModule() Object {
	encodeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("hex.encode expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("hex.encode", args, 0)
		if err != nil {
			return nil, err
		}
		return hex.EncodeToString([]byte(text)), nil
	})

	decodeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("hex.decode expects 1 arg, got %d", len(args))
		}
		rawHex, err := asStringArg("hex.decode", args, 0)
		if err != nil {
			return nil, err
		}
		raw, err := hex.DecodeString(rawHex)
		if err != nil {
			return nil, err
		}
		return string(raw), nil
	})

	encodeBytesFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("hex.encodeBytes expects 1 arg, got %d", len(args))
		}
		arr, ok := args[0].(Array)
		if !ok {
			return nil, fmt.Errorf("hex.encodeBytes expects array<number>, got %T", args[0])
		}
		buf := make([]byte, 0, len(arr))
		for i, v := range arr {
			n, err := asIntValue(fmt.Sprintf("hex.encodeBytes arg[0][%d]", i), v)
			if err != nil {
				return nil, err
			}
			if n < 0 || n > 255 {
				return nil, fmt.Errorf("hex.encodeBytes byte out of range at index %d: %d", i, n)
			}
			buf = append(buf, byte(n))
		}
		return hex.EncodeToString(buf), nil
	})

	decodeBytesFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("hex.decodeBytes expects 1 arg, got %d", len(args))
		}
		rawHex, err := asStringArg("hex.decodeBytes", args, 0)
		if err != nil {
			return nil, err
		}
		raw, err := hex.DecodeString(rawHex)
		if err != nil {
			return nil, err
		}
		out := make(Array, 0, len(raw))
		for _, b := range raw {
			out = append(out, float64(b))
		}
		return out, nil
	})

	namespace := Object{
		"encode":      encodeFn,
		"decode":      decodeFn,
		"encodeBytes": encodeBytesFn,
		"decodeBytes": decodeBytesFn,
	}
	module := cloneObject(namespace)
	module["hex"] = namespace
	return module
}
