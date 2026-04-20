package builtin

import (
	"encoding/base64"
	"fmt"
	netURL "net/url"
)

func newEncodingModule() Object {
	base64EncodeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("encoding.base64Encode expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("encoding.base64Encode", args, 0)
		if err != nil {
			return nil, err
		}
		return base64.StdEncoding.EncodeToString([]byte(text)), nil
	})
	base64DecodeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("encoding.base64Decode expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("encoding.base64Decode", args, 0)
		if err != nil {
			return nil, err
		}
		raw, err := base64.StdEncoding.DecodeString(text)
		if err != nil {
			return nil, err
		}
		return string(raw), nil
	})
	urlEncodeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("encoding.urlEncode expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("encoding.urlEncode", args, 0)
		if err != nil {
			return nil, err
		}
		return netURL.QueryEscape(text), nil
	})
	urlDecodeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("encoding.urlDecode expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("encoding.urlDecode", args, 0)
		if err != nil {
			return nil, err
		}
		v, err := netURL.QueryUnescape(text)
		if err != nil {
			return nil, err
		}
		return v, nil
	})
	hexEncodeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("encoding.hexEncode expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("encoding.hexEncode", args, 0)
		if err != nil {
			return nil, err
		}
		return fmt.Sprintf("%x", text), nil
	})
	hexDecodeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("encoding.hexDecode expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("encoding.hexDecode", args, 0)
		if err != nil {
			return nil, err
		}
		import_hex, err := hexDecodeString(text)
		if err != nil {
			return nil, err
		}
		return string(import_hex), nil
	})

	namespace := Object{
		"base64Encode": base64EncodeFn,
		"base64Decode": base64DecodeFn,
		"urlEncode":    urlEncodeFn,
		"urlDecode":    urlDecodeFn,
		"hexEncode":    hexEncodeFn,
		"hexDecode":    hexDecodeFn,
	}
	module := cloneObject(namespace)
	module["encoding"] = namespace
	return module
}
