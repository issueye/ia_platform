package builtin

import "fmt"

func newSetModule() Object {
	unionFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("set.union expects 2 args, got %d", len(args))
		}
		a, err := asArrayArg("set.union", args, 0)
		if err != nil {
			return nil, err
		}
		b, err := asArrayArg("set.union", args, 1)
		if err != nil {
			return nil, err
		}
		seen := map[string]bool{}
		out := make(Array, 0, len(a)+len(b))
		for _, v := range a {
			k := setKey(v)
			if seen[k] {
				continue
			}
			seen[k] = true
			out = append(out, v)
		}
		for _, v := range b {
			k := setKey(v)
			if seen[k] {
				continue
			}
			seen[k] = true
			out = append(out, v)
		}
		return out, nil
	})

	intersectFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("set.intersect expects 2 args, got %d", len(args))
		}
		a, err := asArrayArg("set.intersect", args, 0)
		if err != nil {
			return nil, err
		}
		b, err := asArrayArg("set.intersect", args, 1)
		if err != nil {
			return nil, err
		}
		inB := map[string]bool{}
		for _, v := range b {
			inB[setKey(v)] = true
		}
		seen := map[string]bool{}
		out := make(Array, 0)
		for _, v := range a {
			k := setKey(v)
			if !inB[k] || seen[k] {
				continue
			}
			seen[k] = true
			out = append(out, v)
		}
		return out, nil
	})

	diffFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("set.diff expects 2 args, got %d", len(args))
		}
		a, err := asArrayArg("set.diff", args, 0)
		if err != nil {
			return nil, err
		}
		b, err := asArrayArg("set.diff", args, 1)
		if err != nil {
			return nil, err
		}
		inB := map[string]bool{}
		for _, v := range b {
			inB[setKey(v)] = true
		}
		out := make(Array, 0, len(a))
		for _, v := range a {
			if !inB[setKey(v)] {
				out = append(out, v)
			}
		}
		return out, nil
	})

	hasFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("set.has expects 2 args, got %d", len(args))
		}
		arr, err := asArrayArg("set.has", args, 0)
		if err != nil {
			return nil, err
		}
		target := setKey(args[1])
		for _, v := range arr {
			if setKey(v) == target {
				return true, nil
			}
		}
		return false, nil
	})

	symmetricDiffFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("set.symmetricDiff expects 2 args, got %d", len(args))
		}
		a, err := asArrayArg("set.symmetricDiff", args, 0)
		if err != nil {
			return nil, err
		}
		b, err := asArrayArg("set.symmetricDiff", args, 1)
		if err != nil {
			return nil, err
		}
		inA := map[string]Value{}
		for _, v := range a {
			inA[setKey(v)] = v
		}
		inB := map[string]Value{}
		for _, v := range b {
			inB[setKey(v)] = v
		}
		out := make(Array, 0)
		for k, v := range inA {
			if _, ok := inB[k]; !ok {
				out = append(out, v)
			}
		}
		for k, v := range inB {
			if _, ok := inA[k]; !ok {
				out = append(out, v)
			}
		}
		return out, nil
	})

	fromArrayFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("set.fromArray expects 1 arg, got %d", len(args))
		}
		arr, err := asArrayArg("set.fromArray", args, 0)
		if err != nil {
			return nil, err
		}
		seen := map[string]bool{}
		out := make(Array, 0, len(arr))
		for _, v := range arr {
			k := setKey(v)
			if !seen[k] {
				seen[k] = true
				out = append(out, v)
			}
		}
		return out, nil
	})

	toArrayFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("set.toArray expects 1 arg, got %d", len(args))
		}
		return fromArrayFn(args)
	})

	namespace := Object{
		"union":         unionFn,
		"intersect":     intersectFn,
		"diff":          diffFn,
		"has":           hasFn,
		"symmetricDiff": symmetricDiffFn,
		"fromArray":     fromArrayFn,
		"toArray":       toArrayFn,
	}
	module := cloneObject(namespace)
	module["set"] = namespace
	return module
}

func asArrayArg(fn string, args []Value, idx int) (Array, error) {
	if idx < 0 || idx >= len(args) {
		return nil, fmt.Errorf("%s arg[%d] missing", fn, idx)
	}
	arr, ok := args[idx].(Array)
	if !ok {
		return nil, fmt.Errorf("%s arg[%d] expects array, got %T", fn, idx, args[idx])
	}
	return arr, nil
}

func setKey(v Value) string {
	return fmt.Sprintf("%T:%v", v, v)
}
