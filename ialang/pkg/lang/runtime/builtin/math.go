package builtin

import (
	"fmt"
	"math"
	"math/rand"

	rttypes "ialang/pkg/lang/runtime/types"
)

func mathSingleArg(name string, fn func(float64) float64) rttypes.NativeFunction {
	return rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("math.%s expects 1 arg, got %d", name, len(args))
		}
		n, ok := args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("math.%s expects number, got %T", name, args[0])
		}
		return fn(n), nil
	})
}

func mathDualArg(name string, fn func(float64, float64) float64) rttypes.NativeFunction {
	return rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("math.%s expects 2 args, got %d", name, len(args))
		}
		a, ok1 := args[0].(float64)
		b, ok2 := args[1].(float64)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("math.%s expects numbers", name)
		}
		return fn(a, b), nil
	})
}

func mathVariadic(name string, fn func(...float64) float64) rttypes.NativeFunction {
	return rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("math.%s expects at least 1 arg", name)
		}
		nums := make([]float64, len(args))
		for i, a := range args {
			n, ok := a.(float64)
			if !ok {
				return nil, fmt.Errorf("math.%s expects numbers", name)
			}
			nums[i] = n
		}
		return fn(nums...), nil
	})
}

func newMathModule() rttypes.Value {
	module := rttypes.Object{
		"abs":   mathSingleArg("abs", math.Abs),
		"ceil":  mathSingleArg("ceil", math.Ceil),
		"floor": mathSingleArg("floor", math.Floor),
		"round": mathSingleArg("round", math.Round),
		"sqrt":  mathSingleArg("sqrt", math.Sqrt),
		"log":   mathSingleArg("log", math.Log),
		"log10": mathSingleArg("log10", math.Log10),
		"log2":  mathSingleArg("log2", math.Log2),
		"sin":   mathSingleArg("sin", math.Sin),
		"cos":   mathSingleArg("cos", math.Cos),
		"tan":   mathSingleArg("tan", math.Tan),
		"asin":  mathSingleArg("asin", math.Asin),
		"acos":  mathSingleArg("acos", math.Acos),
		"atan":  mathSingleArg("atan", math.Atan),
		"atan2": mathDualArg("atan2", math.Atan2),
		"exp":   mathSingleArg("exp", math.Exp),
		"trunc": mathSingleArg("trunc", math.Trunc),
		"sign":  mathSingleArg("sign", func(x float64) float64 {
			if x > 0 {
				return 1
			} else if x < 0 {
				return -1
			}
			return 0
		}),
		"pow":   mathDualArg("pow", math.Pow),
		"max": mathVariadic("max", func(nums ...float64) float64 {
			m := nums[0]
			for _, n := range nums[1:] {
				if n > m {
					m = n
				}
			}
			return m
		}),
		"min": mathVariadic("min", func(nums ...float64) float64 {
			m := nums[0]
			for _, n := range nums[1:] {
				if n < m {
					m = n
				}
			}
			return m
		}),
		"mod": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("math.mod expects 2 args, got %d", len(args))
			}
			a, ok1 := args[0].(float64)
			b, ok2 := args[1].(float64)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("math.mod expects numbers")
			}
			if b == 0 {
				return nil, fmt.Errorf("math.mod: division by zero")
			}
			return math.Mod(a, b), nil
		}),
		"random": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) > 2 {
				return nil, fmt.Errorf("math.random expects 0-2 args, got %d", len(args))
			}
			if len(args) == 0 {
				return rand.Float64(), nil
			}
			minVal, ok1 := args[0].(float64)
			maxVal, ok2 := args[1].(float64)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("math.random expects numbers")
			}
			r := rand.Float64()
			return minVal + r*(maxVal-minVal), nil
		}),
		"PI":        math.Pi,
		"E":         math.E,
		"sqrt2":     math.Sqrt2,
		"NaN":       math.NaN(),
		"Infinity":  math.Inf(1),
		"NEG_INFINITY": math.Inf(-1),
	}

	return cloneObject(module)
}
