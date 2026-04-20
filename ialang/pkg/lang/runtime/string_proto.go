package runtime

import (
	"fmt"
	"strconv"
	"strings"

	common "iacommon/pkg/ialang/value"
)

var stringPrototype common.Object

func GetStringPrototype() common.Object {
	if stringPrototype == nil {
		stringPrototype = buildStringPrototype()
	}
	return stringPrototype
}

func buildStringPrototype() common.Object {
	proto := common.Object{}

	proto["split"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("split expects 1 arg, got %d", len(args))
		}
		sep, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("split expects string separator")
		}
		s := args[len(args)-1].(string)
		parts := strings.Split(s, sep)
		arr := make(common.Array, 0, len(parts))
		for _, p := range parts {
			arr = append(arr, p)
		}
		return arr, nil
	})

	proto["trim"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		s := args[len(args)-1].(string)
		return strings.TrimSpace(s), nil
	})

	proto["trimLeft"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		s := args[len(args)-1].(string)
		if len(args) >= 2 {
			cutset, ok := args[0].(string)
			if ok {
				return strings.TrimLeft(s, cutset), nil
			}
		}
		return strings.TrimSpace(s), nil
	})

	proto["trimRight"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		s := args[len(args)-1].(string)
		if len(args) >= 2 {
			cutset, ok := args[0].(string)
			if ok {
				return strings.TrimRight(s, cutset), nil
			}
		}
		return strings.TrimSpace(s), nil
	})

	proto["replace"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("replace expects 2 args, got %d", len(args))
		}
		old, ok1 := args[0].(string)
		nw, ok2 := args[1].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("replace expects strings")
		}
		s := args[len(args)-1].(string)
		return strings.ReplaceAll(s, old, nw), nil
	})

	proto["replaceAll"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("replaceAll expects 2 args, got %d", len(args))
		}
		old, ok1 := args[0].(string)
		nw, ok2 := args[1].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("replaceAll expects strings")
		}
		s := args[len(args)-1].(string)
		return strings.ReplaceAll(s, old, nw), nil
	})

	proto["toLowerCase"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		s := args[len(args)-1].(string)
		return strings.ToLower(s), nil
	})

	proto["toUpperCase"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		s := args[len(args)-1].(string)
		return strings.ToUpper(s), nil
	})

	proto["startsWith"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("startsWith expects 1 arg, got %d", len(args))
		}
		prefix, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("startsWith expects string")
		}
		s := args[len(args)-1].(string)
		return strings.HasPrefix(s, prefix), nil
	})

	proto["endsWith"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("endsWith expects 1 arg, got %d", len(args))
		}
		suffix, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("endsWith expects string")
		}
		s := args[len(args)-1].(string)
		return strings.HasSuffix(s, suffix), nil
	})

	proto["contains"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("contains expects 1 arg, got %d", len(args))
		}
		substr, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("contains expects string")
		}
		s := args[len(args)-1].(string)
		return strings.Contains(s, substr), nil
	})

	proto["indexOf"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("indexOf expects 1 arg, got %d", len(args))
		}
		substr, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("indexOf expects string")
		}
		s := args[len(args)-1].(string)
		return float64(strings.Index(s, substr)), nil
	})

	proto["repeat"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("repeat expects 1 arg, got %d", len(args))
		}
		n, ok := args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("repeat expects number")
		}
		s := args[len(args)-1].(string)
		return strings.Repeat(s, int(n)), nil
	})

	proto["padStart"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("padStart expects 2 args, got %d", len(args))
		}
		length, ok1 := args[0].(float64)
		padStr, ok2 := args[1].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("padStart expects number and string")
		}
		s := args[len(args)-1].(string)
		if len(s) >= int(length) {
			return s, nil
		}
		padding := strings.Repeat(padStr, int(length)-len(s))
		return padding[:int(length)-len(s)] + s, nil
	})

	proto["padEnd"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("padEnd expects 2 args, got %d", len(args))
		}
		length, ok1 := args[0].(float64)
		padStr, ok2 := args[1].(string)
		if !ok1 || !ok2 {
			return nil, fmt.Errorf("padEnd expects number and string")
		}
		s := args[len(args)-1].(string)
		if len(s) >= int(length) {
			return s, nil
		}
		padding := strings.Repeat(padStr, int(length)-len(s))
		return s + padding[:int(length)-len(s)], nil
	})

	proto["parseInt"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		s := args[len(args)-1].(string)
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return float64(0), nil
		}
		return float64(int(n)), nil
	})

	proto["parseFloat"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		s := args[len(args)-1].(string)
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return float64(0), nil
		}
		return n, nil
	})

	proto["substring"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("substring expects 1 arg, got %d", len(args))
		}
		s := args[len(args)-1].(string)
		start, ok := args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("substring expects number for start")
		}
		st := int(start)
		if st < 0 {
			st = 0
		}
		if st >= len(s) {
			return "", nil
		}
		if len(args) >= 2 {
			end, ok := args[1].(float64)
			if ok {
				en := int(end)
				if en > len(s) {
					en = len(s)
				}
				if en < st {
					en = st
				}
				return s[st:en], nil
			}
		}
		return s[st:], nil
	})

	proto["charAt"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("charAt expects 1 arg, got %d", len(args))
		}
		s := args[len(args)-1].(string)
		idx, ok := args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("charAt expects number")
		}
		i := int(idx)
		if i < 0 || i >= len(s) {
			return "", nil
		}
		return string(s[i]), nil
	})

	proto["slice"] = common.NativeFunction(func(args []common.Value) (common.Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("slice expects 1 arg, got %d", len(args))
		}
		s := args[len(args)-1].(string)
		start, ok := args[0].(float64)
		if !ok {
			return nil, fmt.Errorf("slice expects number for start")
		}
		st := int(start)
		if st < 0 {
			st = len(s) + st
		}
		if st < 0 {
			st = 0
		}
		if st >= len(s) {
			return "", nil
		}
		if len(args) >= 2 {
			end, ok := args[1].(float64)
			if ok {
				en := int(end)
				if en < 0 {
					en = len(s) + en
				}
				if en > len(s) {
					en = len(s)
				}
				if en <= st {
					return "", nil
				}
				return s[st:en], nil
			}
		}
		return s[st:], nil
	})

	return proto
}
