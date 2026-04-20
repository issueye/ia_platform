package builtin

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func newRandModule() Object {
	src := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(src)
	var mu sync.Mutex

	intFn := NativeFunction(func(args []Value) (Value, error) {
		mu.Lock()
		defer mu.Unlock()

		switch len(args) {
		case 0:
			return float64(rng.Intn(1 << 30)), nil
		case 1:
			max, err := asIntValue("rand.int arg[0]", args[0])
			if err != nil {
				return nil, err
			}
			if max <= 0 {
				return nil, fmt.Errorf("rand.int arg[0] expects positive integer, got %d", max)
			}
			return float64(rng.Intn(max)), nil
		case 2:
			min, err := asIntValue("rand.int arg[0]", args[0])
			if err != nil {
				return nil, err
			}
			max, err := asIntValue("rand.int arg[1]", args[1])
			if err != nil {
				return nil, err
			}
			if max <= min {
				return nil, fmt.Errorf("rand.int expects max > min, got min=%d max=%d", min, max)
			}
			return float64(min + rng.Intn(max-min)), nil
		default:
			return nil, fmt.Errorf("rand.int expects 0-2 args, got %d", len(args))
		}
	})

	floatFn := NativeFunction(func(args []Value) (Value, error) {
		mu.Lock()
		defer mu.Unlock()

		switch len(args) {
		case 0:
			return rng.Float64(), nil
		case 2:
			min, ok1 := args[0].(float64)
			max, ok2 := args[1].(float64)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("rand.float expects number args")
			}
			if max <= min {
				return nil, fmt.Errorf("rand.float expects max > min, got min=%v max=%v", min, max)
			}
			return min + rng.Float64()*(max-min), nil
		default:
			return nil, fmt.Errorf("rand.float expects 0 or 2 args, got %d", len(args))
		}
	})

	pickFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("rand.pick expects 1 arg, got %d", len(args))
		}
		arr, ok := args[0].(Array)
		if !ok {
			return nil, fmt.Errorf("rand.pick expects array, got %T", args[0])
		}
		if len(arr) == 0 {
			return nil, fmt.Errorf("rand.pick expects non-empty array")
		}
		mu.Lock()
		defer mu.Unlock()
		return arr[rng.Intn(len(arr))], nil
	})

	stringFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("rand.string expects 1-2 args: length, [charset]")
		}
		length, err := asIntValue("rand.string arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		if length < 0 {
			return nil, fmt.Errorf("rand.string length must be >= 0, got %d", length)
		}
		charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		if len(args) == 2 && args[1] != nil {
			s, err := asStringValue("rand.string arg[1]", args[1])
			if err != nil {
				return nil, err
			}
			if s == "" {
				return nil, fmt.Errorf("rand.string charset must not be empty")
			}
			charset = s
		}

		mu.Lock()
		defer mu.Unlock()
		b := make([]byte, length)
		for i := 0; i < length; i++ {
			b[i] = charset[rng.Intn(len(charset))]
		}
		return string(b), nil
	})

	namespace := Object{
		"int":    intFn,
		"float":  floatFn,
		"pick":   pickFn,
		"string": stringFn,
		"shuffle": NativeFunction(func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("rand.shuffle expects 1 arg, got %d", len(args))
			}
			arr, ok := args[0].(Array)
			if !ok {
				return nil, fmt.Errorf("rand.shuffle expects array, got %T", args[0])
			}
			if len(arr) <= 1 {
				return arr, nil
			}
			result := make(Array, len(arr))
			copy(result, arr)
			mu.Lock()
			defer mu.Unlock()
			rng.Shuffle(len(result), func(i, j int) {
				result[i], result[j] = result[j], result[i]
			})
			return result, nil
		}),
		"seed": NativeFunction(func(args []Value) (Value, error) {
			mu.Lock()
			defer mu.Unlock()
			if len(args) == 0 {
				rng.Seed(time.Now().UnixNano())
			} else {
				n, ok := args[0].(float64)
				if !ok {
					return nil, fmt.Errorf("rand.seed expects number, got %T", args[0])
				}
				rng.Seed(int64(n))
			}
			return true, nil
		}),
	}
	module := cloneObject(namespace)
	module["rand"] = namespace
	return module
}
