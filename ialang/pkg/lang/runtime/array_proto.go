package runtime

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	rttypes "ialang/pkg/lang/runtime/types"
)

var arrayPrototype rttypes.Object

func GetArrayPrototype() rttypes.Object {
	if arrayPrototype == nil {
		arrayPrototype = buildArrayPrototype()
	}
	return arrayPrototype
}

func buildArrayPrototype() rttypes.Object {
	proto := rttypes.Object{}

	proto["sort"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		arr := args[len(args)-1].(rttypes.Array)
		result := make(rttypes.Array, len(arr))
		copy(result, arr)
		sort.Slice(result, func(i, j int) bool {
			a, aok := result[i].(float64)
			b, bok := result[j].(float64)
			if aok && bok {
				return a < b
			}
			return toString(result[i]) < toString(result[j])
		})
		return result, nil
	})

	proto["reverse"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		arr := args[len(args)-1].(rttypes.Array)
		result := make(rttypes.Array, len(arr))
		for i := 0; i < len(arr); i++ {
			result[i] = arr[len(arr)-1-i]
		}
		return result, nil
	})

	proto["map"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("map expects 1 callback arg, got %d", len(args)-1)
		}
		arr := args[len(args)-1].(rttypes.Array)
		callback := args[0]
		result := make(rttypes.Array, len(arr))
		for i, item := range arr {
			mapped, err := callCallable(callback, []rttypes.Value{item, float64(i), arr}, "array.map")
			if err != nil {
				return nil, err
			}
			result[i] = mapped
		}
		return result, nil
	})

	proto["filter"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("filter expects 1 callback arg, got %d", len(args)-1)
		}
		arr := args[len(args)-1].(rttypes.Array)
		callback := args[0]
		result := make(rttypes.Array, 0, len(arr))
		for i, item := range arr {
			keep, err := callCallable(callback, []rttypes.Value{item, float64(i), arr}, "array.filter")
			if err != nil {
				return nil, err
			}
			if isTruthy(keep) {
				result = append(result, item)
			}
		}
		return result, nil
	})

	proto["find"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("find expects 1 callback arg, got %d", len(args)-1)
		}
		arr := args[len(args)-1].(rttypes.Array)
		callback := args[0]
		for i, item := range arr {
			matched, err := callCallable(callback, []rttypes.Value{item, float64(i), arr}, "array.find")
			if err != nil {
				return nil, err
			}
			if isTruthy(matched) {
				return item, nil
			}
		}
		return nil, nil
	})

	proto["findIndex"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("findIndex expects 1 callback arg, got %d", len(args)-1)
		}
		arr := args[len(args)-1].(rttypes.Array)
		callback := args[0]
		for i, item := range arr {
			matched, err := callCallable(callback, []rttypes.Value{item, float64(i), arr}, "array.findIndex")
			if err != nil {
				return nil, err
			}
			if isTruthy(matched) {
				return float64(i), nil
			}
		}
		return float64(-1), nil
	})

	proto["forEach"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("forEach expects 1 callback arg, got %d", len(args)-1)
		}
		arr := args[len(args)-1].(rttypes.Array)
		callback := args[0]
		for i, item := range arr {
			if _, err := callCallable(callback, []rttypes.Value{item, float64(i), arr}, "array.forEach"); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})

	proto["some"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("some expects 1 callback arg, got %d", len(args)-1)
		}
		arr := args[len(args)-1].(rttypes.Array)
		callback := args[0]
		for i, item := range arr {
			matched, err := callCallable(callback, []rttypes.Value{item, float64(i), arr}, "array.some")
			if err != nil {
				return nil, err
			}
			if isTruthy(matched) {
				return true, nil
			}
		}
		return false, nil
	})

	proto["every"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("every expects 1 callback arg, got %d", len(args)-1)
		}
		arr := args[len(args)-1].(rttypes.Array)
		callback := args[0]
		for i, item := range arr {
			matched, err := callCallable(callback, []rttypes.Value{item, float64(i), arr}, "array.every")
			if err != nil {
				return nil, err
			}
			if !isTruthy(matched) {
				return false, nil
			}
		}
		return true, nil
	})

	proto["reduce"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 2 || len(args) > 3 {
			return nil, fmt.Errorf("reduce expects callback and optional initial value, got %d args", len(args)-1)
		}
		arr := args[len(args)-1].(rttypes.Array)
		callback := args[0]
		if len(arr) == 0 && len(args) < 3 {
			return nil, fmt.Errorf("reduce of empty array with no initial value")
		}

		var acc rttypes.Value
		start := 0
		if len(args) == 3 {
			acc = args[1]
		} else {
			acc = arr[0]
			start = 1
		}
		for i := start; i < len(arr); i++ {
			next, err := callCallable(callback, []rttypes.Value{acc, arr[i], float64(i), arr}, "array.reduce")
			if err != nil {
				return nil, err
			}
			acc = next
		}
		return acc, nil
	})

	proto["includes"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("includes expects 1 arg, got %d", len(args))
		}
		arr := args[len(args)-1].(rttypes.Array)
		target := args[0]
		for _, v := range arr {
			if valueEqual(v, target) {
				return true, nil
			}
		}
		return false, nil
	})

	proto["join"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		arr := args[len(args)-1].(rttypes.Array)
		sep := ","
		if len(args) >= 2 {
			if s, ok := args[0].(string); ok {
				sep = s
			}
		}
		strs := make([]string, 0, len(arr))
		for _, v := range arr {
			strs = append(strs, toString(v))
		}
		return strings.Join(strs, sep), nil
	})

	proto["indexOf"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("indexOf expects 1 arg, got %d", len(args))
		}
		arr := args[len(args)-1].(rttypes.Array)
		target := args[0]
		fromIdx := 0
		if len(args) >= 2 {
			if f, ok := args[1].(float64); ok {
				fromIdx = int(f)
				if fromIdx < 0 {
					fromIdx = len(arr) + fromIdx
				}
				if fromIdx < 0 {
					fromIdx = 0
				}
			}
		}
		for i := fromIdx; i < len(arr); i++ {
			if valueEqual(arr[i], target) {
				return float64(i), nil
			}
		}
		return float64(-1), nil
	})

	proto["lastIndexOf"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("lastIndexOf expects 1 arg, got %d", len(args))
		}
		arr := args[len(args)-1].(rttypes.Array)
		target := args[0]
		fromIdx := len(arr) - 1
		if len(args) >= 2 {
			if f, ok := args[1].(float64); ok {
				fromIdx = int(f)
				if fromIdx < 0 {
					fromIdx = len(arr) + fromIdx
				}
				if fromIdx >= len(arr) {
					fromIdx = len(arr) - 1
				}
			}
		}
		for i := fromIdx; i >= 0; i-- {
			if valueEqual(arr[i], target) {
				return float64(i), nil
			}
		}
		return float64(-1), nil
	})

	proto["slice"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("slice expects 1 arg, got %d", len(args))
		}
		arr := args[len(args)-1].(rttypes.Array)
		start, ok := args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("slice expects number for start")
		}
		st := int(start)
		if st < 0 {
			st = len(arr) + st
		}
		if st < 0 {
			st = 0
		}
		if st >= len(arr) {
			return rttypes.Array{}, nil
		}
		end := len(arr)
		if len(args) >= 2 {
			if e, ok := args[1].(float64); ok {
				end = int(e)
				if end < 0 {
					end = len(arr) + end
				}
			}
		}
		if end > len(arr) {
			end = len(arr)
		}
		if st >= end {
			return rttypes.Array{}, nil
		}
		result := make(rttypes.Array, end-st)
		copy(result, arr[st:end])
		return result, nil
	})

	proto["flat"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		arr := args[len(args)-1].(rttypes.Array)
		depth := 1
		if len(args) >= 2 {
			if d, ok := args[0].(float64); ok {
				depth = int(d)
			}
		}
		return flattenArray(arr, depth), nil
	})

	proto["fill"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("fill expects 1 arg, got %d", len(args))
		}
		arr := args[len(args)-1].(rttypes.Array)
		val := args[0]
		result := make(rttypes.Array, len(arr))
		for i := 0; i < len(arr); i++ {
			result[i] = val
		}
		return result, nil
	})

	proto["shuffle"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		arr := args[len(args)-1].(rttypes.Array)
		result := make(rttypes.Array, len(arr))
		copy(result, arr)
		rand.Shuffle(len(result), func(i, j int) {
			result[i], result[j] = result[j], result[i]
		})
		return result, nil
	})

	proto["concat"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		arr := args[len(args)-1].(rttypes.Array)
		totalLen := len(arr)
		for _, arg := range args[:len(args)-1] {
			if a, ok := arg.(rttypes.Array); ok {
				totalLen += len(a)
				continue
			}
			totalLen++
		}

		result := make(rttypes.Array, len(arr), totalLen)
		copy(result, arr)
		for _, arg := range args[:len(args)-1] {
			if a, ok := arg.(rttypes.Array); ok {
				result = append(result, a...)
				continue
			}
			result = append(result, arg)
		}
		return result, nil
	})

	proto["push"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		arr := args[len(args)-1].(rttypes.Array)
		addCount := len(args) - 1
		result := append(arr, args[:addCount]...)
		return result, nil
	})

	proto["pop"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		arr := args[len(args)-1].(rttypes.Array)
		if len(arr) == 0 {
			return rttypes.Array{}, nil
		}
		result := make(rttypes.Array, len(arr)-1)
		copy(result, arr[:len(arr)-1])
		return result, nil
	})

	proto["unshift"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		arr := args[len(args)-1].(rttypes.Array)
		prefixCount := len(args) - 1
		result := make(rttypes.Array, prefixCount+len(arr))
		copy(result[:prefixCount], args[:prefixCount])
		copy(result[prefixCount:], arr)
		return result, nil
	})

	proto["shift"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		arr := args[len(args)-1].(rttypes.Array)
		if len(arr) == 0 {
			return rttypes.Array{}, nil
		}
		result := make(rttypes.Array, len(arr)-1)
		copy(result, arr[1:])
		return result, nil
	})

	proto["at"] = rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("at expects 1 arg, got %d", len(args))
		}
		arr := args[len(args)-1].(rttypes.Array)
		idx, ok := args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("at expects number")
		}
		i := int(idx)
		if i < 0 {
			i = len(arr) + i
		}
		if i < 0 || i >= len(arr) {
			return nil, nil
		}
		return arr[i], nil
	})

	return proto
}

func flattenArray(arr rttypes.Array, depth int) rttypes.Array {
	if depth <= 0 {
		result := make(rttypes.Array, len(arr))
		copy(result, arr)
		return result
	}
	result := make(rttypes.Array, 0)
	for _, v := range arr {
		if subArr, ok := v.(rttypes.Array); ok {
			flat := flattenArray(subArr, depth-1)
			result = append(result, flat...)
		} else {
			result = append(result, v)
		}
	}
	return result
}

func arrayProtoLength() rttypes.NativeFunction {
	return func(args []rttypes.Value) (rttypes.Value, error) {
		arr := args[len(args)-1].(rttypes.Array)
		return float64(len(arr)), nil
	}
}