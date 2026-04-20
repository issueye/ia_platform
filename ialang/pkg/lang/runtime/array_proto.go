package runtime

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	common "iacommon/pkg/ialang/value"
)

var arrayPrototype common.Object

func GetArrayPrototype() common.Object {
	if arrayPrototype == nil {
		arrayPrototype = buildArrayPrototype()
	}
	return arrayPrototype
}

func buildArrayPrototype() common.Object {
	proto := common.Object{}

	proto["sort"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		arr := args[len(args)-1].(common.Array)
		result := make(common.Array, len(arr))
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

	proto["reverse"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		arr := args[len(args)-1].(common.Array)
		result := make(common.Array, len(arr))
		for i := 0; i < len(arr); i++ {
			result[i] = arr[len(arr)-1-i]
		}
		return result, nil
	})

	proto["map"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("map expects 1 callback arg, got %d", len(args)-1)
		}
		arr := args[len(args)-1].(common.Array)
		callback := args[0]
		result := make(common.Array, len(arr))
		for i, item := range arr {
			mapped, err := callCallable(callback, []common.Value{item, float64(i), arr}, "array.map")
			if err != nil {
				return nil, err
			}
			result[i] = mapped
		}
		return result, nil
	})

	proto["filter"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("filter expects 1 callback arg, got %d", len(args)-1)
		}
		arr := args[len(args)-1].(common.Array)
		callback := args[0]
		result := make(common.Array, 0, len(arr))
		for i, item := range arr {
			keep, err := callCallable(callback, []common.Value{item, float64(i), arr}, "array.filter")
			if err != nil {
				return nil, err
			}
			if isTruthy(keep) {
				result = append(result, item)
			}
		}
		return result, nil
	})

	proto["find"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("find expects 1 callback arg, got %d", len(args)-1)
		}
		arr := args[len(args)-1].(common.Array)
		callback := args[0]
		for i, item := range arr {
			matched, err := callCallable(callback, []common.Value{item, float64(i), arr}, "array.find")
			if err != nil {
				return nil, err
			}
			if isTruthy(matched) {
				return item, nil
			}
		}
		return nil, nil
	})

	proto["findIndex"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("findIndex expects 1 callback arg, got %d", len(args)-1)
		}
		arr := args[len(args)-1].(common.Array)
		callback := args[0]
		for i, item := range arr {
			matched, err := callCallable(callback, []common.Value{item, float64(i), arr}, "array.findIndex")
			if err != nil {
				return nil, err
			}
			if isTruthy(matched) {
				return float64(i), nil
			}
		}
		return float64(-1), nil
	})

	proto["forEach"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("forEach expects 1 callback arg, got %d", len(args)-1)
		}
		arr := args[len(args)-1].(common.Array)
		callback := args[0]
		for i, item := range arr {
			if _, err := callCallable(callback, []common.Value{item, float64(i), arr}, "array.forEach"); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})

	proto["some"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("some expects 1 callback arg, got %d", len(args)-1)
		}
		arr := args[len(args)-1].(common.Array)
		callback := args[0]
		for i, item := range arr {
			matched, err := callCallable(callback, []common.Value{item, float64(i), arr}, "array.some")
			if err != nil {
				return nil, err
			}
			if isTruthy(matched) {
				return true, nil
			}
		}
		return false, nil
	})

	proto["every"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("every expects 1 callback arg, got %d", len(args)-1)
		}
		arr := args[len(args)-1].(common.Array)
		callback := args[0]
		for i, item := range arr {
			matched, err := callCallable(callback, []common.Value{item, float64(i), arr}, "array.every")
			if err != nil {
				return nil, err
			}
			if !isTruthy(matched) {
				return false, nil
			}
		}
		return true, nil
	})

	proto["reduce"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 2 || len(args) > 3 {
			return nil, fmt.Errorf("reduce expects callback and optional initial value, got %d args", len(args)-1)
		}
		arr := args[len(args)-1].(common.Array)
		callback := args[0]
		if len(arr) == 0 && len(args) < 3 {
			return nil, fmt.Errorf("reduce of empty array with no initial value")
		}

		var acc common.Value
		start := 0
		if len(args) == 3 {
			acc = args[1]
		} else {
			acc = arr[0]
			start = 1
		}
		for i := start; i < len(arr); i++ {
			next, err := callCallable(callback, []common.Value{acc, arr[i], float64(i), arr}, "array.reduce")
			if err != nil {
				return nil, err
			}
			acc = next
		}
		return acc, nil
	})

	proto["includes"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("includes expects 1 arg, got %d", len(args))
		}
		arr := args[len(args)-1].(common.Array)
		target := args[0]
		for _, v := range arr {
			if valueEqual(v, target) {
				return true, nil
			}
		}
		return false, nil
	})

	proto["join"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		arr := args[len(args)-1].(common.Array)
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

	proto["indexOf"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("indexOf expects 1 arg, got %d", len(args))
		}
		arr := args[len(args)-1].(common.Array)
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

	proto["lastIndexOf"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("lastIndexOf expects 1 arg, got %d", len(args))
		}
		arr := args[len(args)-1].(common.Array)
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

	proto["slice"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("slice expects 1 arg, got %d", len(args))
		}
		arr := args[len(args)-1].(common.Array)
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
			return common.Array{}, nil
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
			return common.Array{}, nil
		}
		result := make(common.Array, end-st)
		copy(result, arr[st:end])
		return result, nil
	})

	proto["flat"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		arr := args[len(args)-1].(common.Array)
		depth := 1
		if len(args) >= 2 {
			if d, ok := args[0].(float64); ok {
				depth = int(d)
			}
		}
		return flattenArray(arr, depth), nil
	})

	proto["fill"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("fill expects 1 arg, got %d", len(args))
		}
		arr := args[len(args)-1].(common.Array)
		val := args[0]
		result := make(common.Array, len(arr))
		for i := 0; i < len(arr); i++ {
			result[i] = val
		}
		return result, nil
	})

	proto["shuffle"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		arr := args[len(args)-1].(common.Array)
		result := make(common.Array, len(arr))
		copy(result, arr)
		rand.Shuffle(len(result), func(i, j int) {
			result[i], result[j] = result[j], result[i]
		})
		return result, nil
	})

	proto["concat"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		arr := args[len(args)-1].(common.Array)
		totalLen := len(arr)
		for _, arg := range args[:len(args)-1] {
			if a, ok := arg.(common.Array); ok {
				totalLen += len(a)
				continue
			}
			totalLen++
		}

		result := make(common.Array, len(arr), totalLen)
		copy(result, arr)
		for _, arg := range args[:len(args)-1] {
			if a, ok := arg.(common.Array); ok {
				result = append(result, a...)
				continue
			}
			result = append(result, arg)
		}
		return result, nil
	})

	proto["push"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		arr := args[len(args)-1].(common.Array)
		addCount := len(args) - 1
		result := append(arr, args[:addCount]...)
		return result, nil
	})

	proto["pop"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		arr := args[len(args)-1].(common.Array)
		if len(arr) == 0 {
			return common.Array{}, nil
		}
		result := make(common.Array, len(arr)-1)
		copy(result, arr[:len(arr)-1])
		return result, nil
	})

	proto["unshift"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		arr := args[len(args)-1].(common.Array)
		prefixCount := len(args) - 1
		result := make(common.Array, prefixCount+len(arr))
		copy(result[:prefixCount], args[:prefixCount])
		copy(result[prefixCount:], arr)
		return result, nil
	})

	proto["shift"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		arr := args[len(args)-1].(common.Array)
		if len(arr) == 0 {
			return common.Array{}, nil
		}
		result := make(common.Array, len(arr)-1)
		copy(result, arr[1:])
		return result, nil
	})

	proto["at"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("at expects 1 arg, got %d", len(args))
		}
		arr := args[len(args)-1].(common.Array)
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

func flattenArray(arr common.Array, depth int) common.Array {
	if depth <= 0 {
		result := make(common.Array, len(arr))
		copy(result, arr)
		return result
	}
	result := make(common.Array, 0)
	for _, v := range arr {
		if subArr, ok := v.(common.Array); ok {
			flat := flattenArray(subArr, depth-1)
			result = append(result, flat...)
		} else {
			result = append(result, v)
		}
	}
	return result
}

func arrayProtoLength() common.NativeFunction {
	return func(args []common.Value) (common.Value, error) {
		arr := args[len(args)-1].(common.Array)
		return float64(len(arr)), nil
	}
}
