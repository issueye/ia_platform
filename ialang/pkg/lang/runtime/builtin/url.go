package builtin

import (
	"fmt"
	neturl "net/url"
)

func newURLModule() Object {
	parseFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("url.parse expects 1 arg: raw")
		}
		raw, err := asStringArg("url.parse", args, 0)
		if err != nil {
			return nil, err
		}
		u, err := neturl.Parse(raw)
		if err != nil {
			return nil, err
		}
		out := Object{
			"scheme":   u.Scheme,
			"host":     u.Host,
			"path":     u.Path,
			"query":    u.RawQuery,
			"fragment": u.Fragment,
			"opaque":   u.Opaque,
		}
		if u.User != nil {
			out["userinfo"] = u.User.String()
		} else {
			out["userinfo"] = nil
		}
		return out, nil
	})

	escapeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("url.escape expects 1 arg: text")
		}
		text, err := asStringArg("url.escape", args, 0)
		if err != nil {
			return nil, err
		}
		return neturl.QueryEscape(text), nil
	})

	unescapeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("url.unescape expects 1 arg: text")
		}
		text, err := asStringArg("url.unescape", args, 0)
		if err != nil {
			return nil, err
		}
		decoded, err := neturl.QueryUnescape(text)
		if err != nil {
			return nil, err
		}
		return decoded, nil
	})

	queryEncodeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("url.queryEncode expects 1 arg: obj")
		}
		obj, ok := args[0].(Object)
		if !ok {
			return nil, fmt.Errorf("url.queryEncode arg[0] expects object, got %T", args[0])
		}
		values := neturl.Values{}
		for k, v := range obj {
			s, err := asStringValue("url.queryEncode value", v)
			if err != nil {
				return nil, err
			}
			values.Set(k, s)
		}
		return values.Encode(), nil
	})

	queryDecodeFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("url.queryDecode expects 1 arg: query")
		}
		query, err := asStringArg("url.queryDecode", args, 0)
		if err != nil {
			return nil, err
		}
		values, err := neturl.ParseQuery(query)
		if err != nil {
			return nil, err
		}
		out := Object{}
		for k, vals := range values {
			if len(vals) == 0 {
				out[k] = ""
				continue
			}
			if len(vals) == 1 {
				out[k] = vals[0]
				continue
			}
			arr := make(Array, 0, len(vals))
			for _, v := range vals {
				arr = append(arr, v)
			}
			out[k] = arr
		}
		return out, nil
	})

	namespace := Object{
		"parse":       parseFn,
		"escape":      escapeFn,
		"unescape":    unescapeFn,
		"queryEncode": queryEncodeFn,
		"queryDecode": queryDecodeFn,
	}
	module := cloneObject(namespace)
	module["url"] = namespace
	return module
}
