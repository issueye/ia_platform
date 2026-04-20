package builtin

import (
	"fmt"
	"strconv"
	"strings"

	rttypes "ialang/pkg/lang/runtime/types"
)

func stringSingleArg(name string, fn func(string) (rttypes.Value, error)) rttypes.NativeFunction {
	return rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("string.%s expects 1 arg, got %d", name, len(args))
		}
		s, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("string.%s expects string", name)
		}
		return fn(s)
	})
}

func stringDualArg(name string, fn func(string, string) (rttypes.Value, error)) rttypes.NativeFunction {
	return rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("string.%s expects 2 args, got %d", name, len(args))
		}
		s, ok1 := args[0].(string)
		t, ok2 := args[1].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("string.%s expects strings", name)
		}
		return fn(s, t)
	})
}

func newStringModule() rttypes.Value {
	module := rttypes.Object{
		"split": stringDualArg("split", func(s, sep string) (rttypes.Value, error) {
			parts := strings.Split(s, sep)
			arr := make(rttypes.Array, 0, len(parts))
			for _, p := range parts {
				arr = append(arr, p)
			}
			return arr, nil
		}),
		"join": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) < 1 || len(args) > 2 {
				return nil, fmt.Errorf("string.join expects 1-2 args, got %d", len(args))
			}
			arr, ok := args[0].(rttypes.Array)
			if !ok {
				return nil, fmt.Errorf("string.join expects array")
			}
			sep := ""
			if len(args) == 2 {
				if s, ok := args[1].(string); ok {
					sep = s
				}
			}
			strs := make([]string, 0, len(arr))
			for _, v := range arr {
				strs = append(strs, toString(v))
			}
			return strings.Join(strs, sep), nil
		}),
		"parseInt": stringSingleArg("parseInt", func(s string) (rttypes.Value, error) {
			n, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return float64(0), nil
			}
			return float64(int(n)), nil
		}),
		"parseFloat": stringSingleArg("parseFloat", func(s string) (rttypes.Value, error) {
			n, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return float64(0), nil
			}
			return n, nil
		}),
		"fromCodePoint": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("string.fromCodePoint expects 1 arg, got %d", len(args))
			}
			code, ok := args[0].(float64)
			if !ok {
				return nil, fmt.Errorf("string.fromCodePoint expects number")
			}
			return string(rune(code)), nil
		}),
		"trim":        stringSingleArg("trim", func(s string) (rttypes.Value, error) { return strings.TrimSpace(s), nil }),
		"toLowerCase": stringSingleArg("toLowerCase", func(s string) (rttypes.Value, error) { return strings.ToLower(s), nil }),
		"toUpperCase": stringSingleArg("toUpperCase", func(s string) (rttypes.Value, error) { return strings.ToUpper(s), nil }),
		"contains":    stringDualArg("contains", func(s, substr string) (rttypes.Value, error) { return strings.Contains(s, substr), nil }),
		"indexOf":     stringDualArg("indexOf", func(s, substr string) (rttypes.Value, error) { return float64(strings.Index(s, substr)), nil }),
		"startsWith":  stringDualArg("startsWith", func(s, prefix string) (rttypes.Value, error) { return strings.HasPrefix(s, prefix), nil }),
		"endsWith":    stringDualArg("endsWith", func(s, suffix string) (rttypes.Value, error) { return strings.HasSuffix(s, suffix), nil }),
		"replace": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) != 3 {
				return nil, fmt.Errorf("string.replace expects 3 args, got %d", len(args))
			}
			s, ok1 := args[0].(string)
			old, ok2 := args[1].(string)
			newStr, ok3 := args[2].(string)
			if !ok1 || !ok2 || !ok3 {
				return nil, fmt.Errorf("string.replace expects 3 strings")
			}
			return strings.ReplaceAll(s, old, newStr), nil
		}),
		"length": stringSingleArg("length", func(s string) (rttypes.Value, error) { return float64(len([]rune(s))), nil }),
		"lastIndexOf": stringDualArg("lastIndexOf", func(s, substr string) (rttypes.Value, error) {
			return float64(strings.LastIndex(s, substr)), nil
		}),
		"substring": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) < 2 || len(args) > 3 {
				return nil, fmt.Errorf("string.substring expects 2-3 args, got %d", len(args))
			}
			s, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("string.substring expects string")
			}
			runes := []rune(s)
			start, ok1 := args[1].(float64)
			if !ok1 {
				return nil, fmt.Errorf("string.substring start expects number")
			}
			si := int(start)
			if si < 0 {
				si = 0
			}
			if si > len(runes) {
				si = len(runes)
			}
			if len(args) == 3 {
				end, ok2 := args[2].(float64)
				if !ok2 {
					return nil, fmt.Errorf("string.substring end expects number")
				}
				ei := int(end)
				if ei < 0 {
					ei = 0
				}
				if ei > len(runes) {
					ei = len(runes)
				}
				if si > ei {
					si, ei = ei, si
				}
				return string(runes[si:ei]), nil
			}
			return string(runes[si:]), nil
		}),
		"padStart": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) < 2 || len(args) > 3 {
				return nil, fmt.Errorf("string.padStart expects 2-3 args, got %d", len(args))
			}
			s, ok1 := args[0].(string)
			length, ok2 := args[1].(float64)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("string.padStart expects (string, number[, string])")
			}
			padStr := " "
			if len(args) == 3 {
				if p, ok := args[2].(string); ok && p != "" {
					padStr = p
				}
			}
			runes := []rune(s)
			targetLen := int(length)
			padLen := targetLen - len(runes)
			if padLen <= 0 {
				return s, nil
			}
			padRunes := []rune(padStr)
			var result []rune
			for len(result) < padLen {
				remaining := padLen - len(result)
				if remaining >= len(padRunes) {
					result = append(result, padRunes...)
				} else {
					result = append(result, padRunes[:remaining]...)
				}
			}
			return string(append(result, runes...)), nil
		}),
		"padEnd": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) < 2 || len(args) > 3 {
				return nil, fmt.Errorf("string.padEnd expects 2-3 args, got %d", len(args))
			}
			s, ok1 := args[0].(string)
			length, ok2 := args[1].(float64)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("string.padEnd expects (string, number[, string])")
			}
			padStr := " "
			if len(args) == 3 {
				if p, ok := args[2].(string); ok && p != "" {
					padStr = p
				}
			}
			runes := []rune(s)
			targetLen := int(length)
			padLen := targetLen - len(runes)
			if padLen <= 0 {
				return s, nil
			}
			padRunes := []rune(padStr)
			var result []rune
			result = append(result, runes...)
			for len(result) < targetLen {
				remaining := targetLen - len(result)
				if remaining >= len(padRunes) {
					result = append(result, padRunes...)
				} else {
					result = append(result, padRunes[:remaining]...)
				}
			}
			return string(result), nil
		}),
		"charCodeAt": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("string.charCodeAt expects 2 args, got %d", len(args))
			}
			s, ok1 := args[0].(string)
			idx, ok2 := args[1].(float64)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("string.charCodeAt expects (string, number)")
			}
			runes := []rune(s)
			i := int(idx)
			if i < 0 || i >= len(runes) {
				return float64(-1), nil
			}
			return float64(runes[i]), nil
		}),
		"repeat": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("string.repeat expects 2 args, got %d", len(args))
			}
			s, ok := args[0].(string)
			if !ok {
				return nil, fmt.Errorf("string.repeat expects string")
			}
			count, ok2 := args[1].(float64)
			if !ok2 {
				return nil, fmt.Errorf("string.repeat count expects number")
			}
			return strings.Repeat(s, int(count)), nil
		}),
	}

	return cloneObject(module)
}
