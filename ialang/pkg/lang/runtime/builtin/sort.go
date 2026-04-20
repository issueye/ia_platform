package builtin

import (
	"fmt"
	"sort"

	rtvm "ialang/pkg/lang/runtime/vm"
)

func newSortModule() Object {
	ascFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("sort.asc expects 1 arg, got %d", len(args))
		}
		arr, ok := args[0].(Array)
		if !ok {
			return nil, fmt.Errorf("sort.asc expects array, got %T", args[0])
		}
		out := cloneArray(arr)
		if allNumbers(out) {
			sort.Slice(out, func(i, j int) bool {
				return out[i].(float64) < out[j].(float64)
			})
			return out, nil
		}
		if allStrings(out) {
			sort.Slice(out, func(i, j int) bool {
				return out[i].(string) < out[j].(string)
			})
			return out, nil
		}
		return nil, fmt.Errorf("sort.asc only supports number[] or string[]")
	})

	descFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("sort.desc expects 1 arg, got %d", len(args))
		}
		sorted, err := ascFn(args)
		if err != nil {
			return nil, err
		}
		arr := sorted.(Array)
		reverseInPlace(arr)
		return arr, nil
	})

	reverseFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("sort.reverse expects 1 arg, got %d", len(args))
		}
		arr, ok := args[0].(Array)
		if !ok {
			return nil, fmt.Errorf("sort.reverse expects array, got %T", args[0])
		}
		out := cloneArray(arr)
		reverseInPlace(out)
		return out, nil
	})

	uniqueFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("sort.unique expects 1 arg, got %d", len(args))
		}
		arr, ok := args[0].(Array)
		if !ok {
			return nil, fmt.Errorf("sort.unique expects array, got %T", args[0])
		}
		seen := map[string]bool{}
		out := make(Array, 0, len(arr))
		for _, v := range arr {
			key := fmt.Sprintf("%T:%v", v, v)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, v)
		}
		return out, nil
	})

	sortByFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("sort.sortBy expects 2 args: array, keyFn")
		}
		arr, ok := args[0].(Array)
		if !ok {
			return nil, fmt.Errorf("sort.sortBy expects array, got %T", args[0])
		}
		out := cloneArray(arr)
		callKey := func(v Value) (Value, error) {
			switch cb := args[1].(type) {
			case NativeFunction:
				return cb([]Value{v})
			case *UserFunction:
				return rtvm.CallUserFunctionSync(cb, []Value{v})
			default:
				return nil, fmt.Errorf("sort.sortBy keyFn expects function, got %T", args[1])
			}
		}
		sort.SliceStable(out, func(i, j int) bool {
			ki, err1 := callKey(out[i])
			kj, err2 := callKey(out[j])
			if err1 != nil || err2 != nil {
				return false
			}
			return compareValues(ki, kj) < 0
		})
		return out, nil
	})

	namespace := Object{
		"asc":     ascFn,
		"desc":    descFn,
		"reverse": reverseFn,
		"unique":  uniqueFn,
		"sortBy":  sortByFn,
	}
	module := cloneObject(namespace)
	module["sort"] = namespace
	return module
}

func cloneArray(arr Array) Array {
	out := make(Array, len(arr))
	copy(out, arr)
	return out
}

func reverseInPlace(arr Array) {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
}

func allNumbers(arr Array) bool {
	for _, v := range arr {
		if _, ok := v.(float64); !ok {
			return false
		}
	}
	return true
}

func allStrings(arr Array) bool {
	for _, v := range arr {
		if _, ok := v.(string); !ok {
			return false
		}
	}
	return true
}
