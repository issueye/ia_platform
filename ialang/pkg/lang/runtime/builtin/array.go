package builtin

import (
	"fmt"
	"math/rand"
	"sort"

	rttypes "ialang/pkg/lang/runtime/types"
	rtvm "ialang/pkg/lang/runtime/vm"
)

func newArrayModule() rttypes.Value {
	callCallback := func(fn rttypes.Value, args []rttypes.Value) (rttypes.Value, error) {
		switch cb := fn.(type) {
		case rttypes.NativeFunction:
			return cb(args)
		case *rttypes.UserFunction:
			return rtvm.CallUserFunctionSync(cb, args)
		default:
			return nil, fmt.Errorf("expects function, got %T", fn)
		}
	}

	expectCallback := func(fn rttypes.Value) error {
		switch fn.(type) {
		case rttypes.NativeFunction, *rttypes.UserFunction:
			return nil
		default:
			return fmt.Errorf("expects function, got %T", fn)
		}
	}

	// map: array.map(callback)
	mapFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("array.map expects 2 args: array, callback")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.map arg[0] expects array, got %T", args[0])
		}
		if err := expectCallback(args[1]); err != nil {
			return nil, fmt.Errorf("array.map arg[1] %w", err)
		}
		result := make(rttypes.Array, len(arr))
		for i, item := range arr {
			res, err := callCallback(args[1], []rttypes.Value{item, float64(i), arr})
			if err != nil {
				return nil, err
			}
			result[i] = res
		}
		return result, nil
	})

	// filter: array.filter(callback)
	filterFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("array.filter expects 2 args: array, callback")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.filter arg[0] expects array, got %T", args[0])
		}
		if err := expectCallback(args[1]); err != nil {
			return nil, fmt.Errorf("array.filter arg[1] %w", err)
		}
		result := make(rttypes.Array, 0)
		for i, item := range arr {
			res, err := callCallback(args[1], []rttypes.Value{item, float64(i), arr})
			if err != nil {
				return nil, err
			}
			if isTruthyValue(res) {
				result = append(result, item)
			}
		}
		return result, nil
	})

	// find: array.find(callback)
	findFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("array.find expects 2 args: array, callback")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.find arg[0] expects array, got %T", args[0])
		}
		if err := expectCallback(args[1]); err != nil {
			return nil, fmt.Errorf("array.find arg[1] %w", err)
		}
		for i, item := range arr {
			res, err := callCallback(args[1], []rttypes.Value{item, float64(i), arr})
			if err != nil {
				return nil, err
			}
			if isTruthyValue(res) {
				return item, nil
			}
		}
		return nil, nil
	})

	// findIndex: array.findIndex(callback)
	findIndexFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("array.findIndex expects 2 args: array, callback")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.findIndex arg[0] expects array, got %T", args[0])
		}
		if err := expectCallback(args[1]); err != nil {
			return nil, fmt.Errorf("array.findIndex arg[1] %w", err)
		}
		for i, item := range arr {
			res, err := callCallback(args[1], []rttypes.Value{item, float64(i), arr})
			if err != nil {
				return nil, err
			}
			if isTruthyValue(res) {
				return float64(i), nil
			}
		}
		return float64(-1), nil
	})

	// forEach: array.forEach(callback)
	forEachFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("array.forEach expects 2 args: array, callback")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.forEach arg[0] expects array, got %T", args[0])
		}
		if err := expectCallback(args[1]); err != nil {
			return nil, fmt.Errorf("array.forEach arg[1] %w", err)
		}
		for i, item := range arr {
			_, err := callCallback(args[1], []rttypes.Value{item, float64(i), arr})
			if err != nil {
				return nil, fmt.Errorf("array.forEach callback error: %w", err)
			}
		}
		return true, nil
	})

	// some: array.some(callback)
	someFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("array.some expects 2 args: array, callback")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.some arg[0] expects array, got %T", args[0])
		}
		if err := expectCallback(args[1]); err != nil {
			return nil, fmt.Errorf("array.some arg[1] %w", err)
		}
		for i, item := range arr {
			res, err := callCallback(args[1], []rttypes.Value{item, float64(i), arr})
			if err != nil {
				return nil, err
			}
			if isTruthyValue(res) {
				return true, nil
			}
		}
		return false, nil
	})

	// every: array.every(callback)
	everyFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("array.every expects 2 args: array, callback")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.every arg[0] expects array, got %T", args[0])
		}
		if err := expectCallback(args[1]); err != nil {
			return nil, fmt.Errorf("array.every arg[1] %w", err)
		}
		for i, item := range arr {
			res, err := callCallback(args[1], []rttypes.Value{item, float64(i), arr})
			if err != nil {
				return nil, err
			}
			if !isTruthyValue(res) {
				return false, nil
			}
		}
		return true, nil
	})

	// reduce: array.reduce(callback, [initialValue])
	reduceFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 2 || len(args) > 3 {
			return nil, fmt.Errorf("array.reduce expects 2-3 args: array, callback, [initialValue]")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.reduce arg[0] expects array, got %T", args[0])
		}
		if err := expectCallback(args[1]); err != nil {
			return nil, fmt.Errorf("array.reduce arg[1] %w", err)
		}

		var accumulator rttypes.Value
		startIndex := 0

		if len(args) == 3 {
			accumulator = args[2]
		} else {
			if len(arr) == 0 {
				return nil, fmt.Errorf("array.reduce of empty array with no initial value")
			}
			accumulator = arr[0]
			startIndex = 1
		}

		for i := startIndex; i < len(arr); i++ {
			res, err := callCallback(args[1], []rttypes.Value{accumulator, arr[i], float64(i), arr})
			if err != nil {
				return nil, err
			}
			accumulator = res
		}

		return accumulator, nil
	})

	// includes: array.includes(value)
	includesFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("array.includes expects 2 args: array, value")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.includes arg[0] expects array, got %T", args[0])
		}
		value := args[1]
		for _, item := range arr {
			if valuesEqual(item, value) {
				return true, nil
			}
		}
		return false, nil
	})

	// indexOf: array.indexOf(value)
	indexOfFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("array.indexOf expects 2 args: array, value")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.indexOf arg[0] expects array, got %T", args[0])
		}
		value := args[1]
		for i, item := range arr {
			if valuesEqual(item, value) {
				return float64(i), nil
			}
		}
		return float64(-1), nil
	})

	// lastIndexOf: array.lastIndexOf(value)
	lastIndexOfFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("array.lastIndexOf expects 2 args: array, value")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.lastIndexOf arg[0] expects array, got %T", args[0])
		}
		value := args[1]
		for i := len(arr) - 1; i >= 0; i-- {
			if valuesEqual(arr[i], value) {
				return float64(i), nil
			}
		}
		return float64(-1), nil
	})

	// flat: array.flat([depth])
	flatFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("array.flat expects 1-2 args: array, [depth]")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.flat arg[0] expects array, got %T", args[0])
		}
		depth := 1.0
		if len(args) == 2 {
			if d, ok := args[1].(float64); ok {
				depth = d
			}
		}
		return flattenArray(arr, int(depth)), nil
	})

	// flatMap: array.flatMap(callback)
	flatMapFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("array.flatMap expects 2 args: array, callback")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.flatMap arg[0] expects array, got %T", args[0])
		}
		if err := expectCallback(args[1]); err != nil {
			return nil, fmt.Errorf("array.flatMap arg[1] %w", err)
		}
		result := make(rttypes.Array, 0)
		for i, item := range arr {
			res, err := callCallback(args[1], []rttypes.Value{item, float64(i), arr})
			if err != nil {
				return nil, err
			}
			if resArr, ok := res.(rttypes.Array); ok {
				result = append(result, resArr...)
			} else {
				result = append(result, res)
			}
		}
		return result, nil
	})

	// slice: array.slice([start], [end])
	sliceFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 1 || len(args) > 3 {
			return nil, fmt.Errorf("array.slice expects 1-3 args: array, [start], [end]")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.slice arg[0] expects array, got %T", args[0])
		}
		start := 0
		end := len(arr)
		if len(args) >= 2 {
			if s, ok := args[1].(float64); ok {
				start = int(s)
				if start < 0 {
					start = len(arr) + start
				}
			}
		}
		if len(args) >= 3 {
			if e, ok := args[2].(float64); ok {
				end = int(e)
				if end < 0 {
					end = len(arr) + end
				}
			}
		}
		if start < 0 {
			start = 0
		}
		if end > len(arr) {
			end = len(arr)
		}
		if start > end {
			start = end
		}
		result := make(rttypes.Array, end-start)
		copy(result, arr[start:end])
		return result, nil
	})

	// splice: array.splice(start, [deleteCount], [items...])
	spliceFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("array.splice expects at least 2 args: array, start")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.splice arg[0] expects array, got %T", args[0])
		}
		startIdx := 0
		if s, ok := args[1].(float64); ok {
			startIdx = int(s)
			if startIdx < 0 {
				startIdx = len(arr) + startIdx
			}
		}
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx > len(arr) {
			startIdx = len(arr)
		}

		deleteCount := len(arr) - startIdx
		if len(args) >= 3 {
			if d, ok := args[2].(float64); ok {
				deleteCount = int(d)
			}
		}
		if deleteCount < 0 {
			deleteCount = 0
		}
		if startIdx+deleteCount > len(arr) {
			deleteCount = len(arr) - startIdx
		}

		deleted := make(rttypes.Array, deleteCount)
		copy(deleted, arr[startIdx:startIdx+deleteCount])

		newItems := make(rttypes.Array, 0)
		for i := 3; i < len(args); i++ {
			newItems = append(newItems, args[i])
		}

		result := make(rttypes.Array, 0, len(arr)-deleteCount+len(newItems))
		result = append(result, arr[:startIdx]...)
		result = append(result, newItems...)
		result = append(result, arr[startIdx+deleteCount:]...)

		return deleted, nil
	})

	// join: array.join([separator])
	joinFn := rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("array.join expects 1-2 args: array, [separator]")
		}
		arr, ok := args[0].(rttypes.Array)
		if !ok {
			return nil, fmt.Errorf("array.join arg[0] expects array, got %T", args[0])
		}
		sep := ","
		if len(args) == 2 {
			if s, ok := args[1].(string); ok {
				sep = s
			}
		}
		parts := make([]string, len(arr))
		for i, item := range arr {
			parts[i] = toString(item)
		}
		return joinStrings(parts, sep), nil
	})

	module := rttypes.Object{
		"concat": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("array.concat expects at least 1 arg, got %d", len(args))
			}
			result := make(rttypes.Array, 0)
			for _, arg := range args {
				if arr, ok := arg.(rttypes.Array); ok {
					result = append(result, arr...)
				} else {
					result = append(result, arg)
				}
			}
			return result, nil
		}),
		"range": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) < 1 || len(args) > 3 {
				return nil, fmt.Errorf("array.range expects 1-3 args, got %d", len(args))
			}
			start := 0.0
			end := 0.0
			step := 1.0
			if len(args) == 1 {
				if e, ok := args[0].(float64); ok {
					end = e
				}
			} else if len(args) == 2 {
				if s, ok := args[0].(float64); ok {
					start = s
				}
				if e, ok := args[1].(float64); ok {
					end = e
				}
			} else {
				if s, ok := args[0].(float64); ok {
					start = s
				}
				if e, ok := args[1].(float64); ok {
					end = e
				}
				if st, ok := args[2].(float64); ok {
					step = st
				}
			}
			result := make(rttypes.Array, 0)
			if step > 0 {
				for i := start; i < end; i += step {
					result = append(result, i)
				}
			} else if step < 0 {
				for i := start; i > end; i += step {
					result = append(result, i)
				}
			}
			return result, nil
		}),
		"from": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("array.from expects 1 arg, got %d", len(args))
			}
			if arr, ok := args[0].(rttypes.Array); ok {
				result := make(rttypes.Array, len(arr))
				copy(result, arr)
				return result, nil
			}
			if s, ok := args[0].(string); ok {
				result := make(rttypes.Array, len(s))
				for i, c := range s {
					result[i] = string(c)
				}
				return result, nil
			}
			return rttypes.Array{}, nil
		}),
		"isArray": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("array.isArray expects 1 arg, got %d", len(args))
			}
			_, ok := args[0].(rttypes.Array)
			return ok, nil
		}),
		"of": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			result := make(rttypes.Array, len(args))
			for i, arg := range args {
				result[i] = arg
			}
			return result, nil
		}),
		// 新增方法
		"map":         mapFn,
		"filter":      filterFn,
		"find":        findFn,
		"findIndex":   findIndexFn,
		"forEach":     forEachFn,
		"some":        someFn,
		"every":       everyFn,
		"reduce":      reduceFn,
		"includes":    includesFn,
		"indexOf":     indexOfFn,
		"lastIndexOf": lastIndexOfFn,
		"flat":        flatFn,
		"flatMap":     flatMapFn,
		"slice":       sliceFn,
		"splice":      spliceFn,
		"join":        joinFn,
		"sort": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("array.sort expects at least 1 arg: array")
			}
			arr, ok := args[0].(rttypes.Array)
			if !ok {
				return nil, fmt.Errorf("array.sort arg[0] expects array, got %T", args[0])
			}
			result := make(rttypes.Array, len(arr))
			copy(result, arr)
			sort.Slice(result, func(i, j int) bool {
				return compareValues(result[i], result[j]) < 0
			})
			return result, nil
		}),
		"reverse": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("array.reverse expects 1 arg: array")
			}
			arr, ok := args[0].(rttypes.Array)
			if !ok {
				return nil, fmt.Errorf("array.reverse arg[0] expects array, got %T", args[0])
			}
			// 创建副本避免修改原数组
			result := make(rttypes.Array, len(arr))
			copy(result, arr)

			// 反转数组
			for i := 0; i < len(result)/2; i++ {
				result[i], result[len(result)-1-i] = result[len(result)-1-i], result[i]
			}
			return result, nil
		}),
		"shuffle": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("array.shuffle expects 1 arg: array")
			}
			arr, ok := args[0].(rttypes.Array)
			if !ok {
				return nil, fmt.Errorf("array.shuffle arg[0] expects array, got %T", args[0])
			}
			result := make(rttypes.Array, len(arr))
			copy(result, arr)
			rand.Shuffle(len(result), func(i, j int) {
				result[i], result[j] = result[j], result[i]
			})
			return result, nil
		}),
		"fill": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) < 2 {
				return nil, fmt.Errorf("array.fill expects at least 2 args: array, value")
			}
			arr, ok := args[0].(rttypes.Array)
			if !ok {
				return nil, fmt.Errorf("array.fill arg[0] expects array, got %T", args[0])
			}
			value := args[1]
			result := make(rttypes.Array, len(arr))
			for i := range result {
				result[i] = value
			}
			return result, nil
		}),
	}

	return cloneObject(module)
}

// valuesEqual 比较两个值是否相等
func valuesEqual(a, b rttypes.Value) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	// 类型相同，直接比较
	switch va := a.(type) {
	case bool:
		vb, ok := b.(bool)
		return ok && va == vb
	case float64:
		vb, ok := b.(float64)
		return ok && va == vb
	case string:
		vb, ok := b.(string)
		return ok && va == vb
	default:
		return false
	}
}

// compareValues 比较两个值，返回 -1, 0, 或 1
func compareValues(a, b rttypes.Value) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	// 数字比较
	if na, ok := a.(float64); ok {
		if nb, ok := b.(float64); ok {
			if na < nb {
				return -1
			}
			if na > nb {
				return 1
			}
			return 0
		}
	}

	// 字符串比较
	if sa, ok := a.(string); ok {
		if sb, ok := b.(string); ok {
			if sa < sb {
				return -1
			}
			if sa > sb {
				return 1
			}
			return 0
		}
	}

	// 其他类型：基于类型名称排序
	typeA := fmt.Sprintf("%T", a)
	typeB := fmt.Sprintf("%T", b)
	if typeA < typeB {
		return -1
	}
	if typeA > typeB {
		return 1
	}
	return 0
}

// flattenArray 扁平化数组
func flattenArray(arr rttypes.Array, depth int) rttypes.Array {
	if depth <= 0 {
		result := make(rttypes.Array, len(arr))
		copy(result, arr)
		return result
	}
	result := make(rttypes.Array, 0)
	for _, item := range arr {
		if subArr, ok := item.(rttypes.Array); ok {
			flattened := flattenArray(subArr, depth-1)
			result = append(result, flattened...)
		} else {
			result = append(result, item)
		}
	}
	return result
}

func isTruthyValue(v rttypes.Value) bool {
	switch val := v.(type) {
	case nil:
		return false
	case bool:
		return val
	case float64:
		return val != 0
	case string:
		return val != ""
	default:
		return true
	}
}

// joinStrings 连接字符串
func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += sep + parts[i]
	}
	return result
}
