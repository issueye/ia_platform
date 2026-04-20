package vm

import (
	"errors"
	"fmt"
	"sort"
)

func (v *VM) resolveName(name string) (Value, error) {
	if v.env != nil {
		if val, ok := v.env.Get(name); ok {
			return val, nil
		}
	}
	if val, ok := v.globals[name]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("undefined variable: %s", name)
}

func (v *VM) defineName(name string, val Value) {
	if v.env != nil {
		v.env.Define(name, val)
		return
	}
	v.globals[name] = val
}

func (v *VM) setName(name string, val Value) error {
	if v.env != nil && v.env.Set(name, val) {
		return nil
	}
	if _, ok := v.globals[name]; ok {
		v.globals[name] = val
		return nil
	}
	return fmt.Errorf("assignment to undefined variable: %s", name)
}

func (v *VM) push(val Value) {
	v.stack = append(v.stack, val)
}

func (v *VM) pop() (Value, error) {
	if len(v.stack) == 0 {
		return nil, errors.New("stack underflow")
	}
	last := len(v.stack) - 1
	val := v.stack[last]
	v.stack = v.stack[:last]
	return val, nil
}

func isTruthy(v Value) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case float64:
		return x != 0
	case string:
		return x != ""
	default:
		return true
	}
}

func valueEqual(a, b Value) bool {
	switch av := a.(type) {
	case nil:
		return b == nil
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case float64:
		bv, ok := b.(float64)
		return ok && av == bv
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case *Promise:
		bv, ok := b.(*Promise)
		return ok && av == bv
	case *UserFunction:
		bv, ok := b.(*UserFunction)
		return ok && av == bv
	case *ClassValue:
		bv, ok := b.(*ClassValue)
		return ok && av == bv
	case *InstanceValue:
		bv, ok := b.(*InstanceValue)
		return ok && av == bv
	case *BoundMethod:
		bv, ok := b.(*BoundMethod)
		return ok && av == bv
	default:
		return false
	}
}

func toString(v Value) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case string:
		return x
	case float64:
		return fmt.Sprintf("%g", x)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case *Promise:
		if x.IsDone() {
			return "[Promise resolved]"
		}
		return "[Promise pending]"
	case *UserFunction:
		if x.Name != "" {
			if x.Async {
				return fmt.Sprintf("[AsyncFunction %s]", x.Name)
			}
			return fmt.Sprintf("[Function %s]", x.Name)
		}
		return "[Function]"
	case *ClassValue:
		return fmt.Sprintf("[Class %s]", x.Name)
	case *InstanceValue:
		if x.Class != nil {
			return fmt.Sprintf("[Instance %s]", x.Class.Name)
		}
		return "[Instance]"
	case *BoundMethod:
		if x.Method != nil && x.Method.Name != "" {
			return fmt.Sprintf("[BoundMethod %s]", x.Method.Name)
		}
		return "[BoundMethod]"
	case Array:
		return fmt.Sprintf("%v", []Value(x))
	case Object:
		return fmt.Sprintf("%v", map[string]Value(x))
	default:
		return fmt.Sprintf("%v", x)
	}
}

func (v *VM) execTypeof() error {
	val, err := v.pop()
	if err != nil {
		return err
	}

	typeStr := typeofString(val)
	v.push(typeStr)
	return nil
}

func (v *VM) execObjectKeys() error {
	val, err := v.pop()
	if err != nil {
		return err
	}

	obj, ok := val.(Object)
	if !ok {
		// If it's already an array, return as-is (for iterating arrays)
		if arr, ok := val.(Array); ok {
			v.push(arr)
			return nil
		}
		// For non-objects, push empty array
		v.push(Array{})
		return nil
	}

	// Extract keys from object
	keys := make(Array, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		left, _ := keys[i].(string)
		right, _ := keys[j].(string)
		return left < right
	})
	v.push(keys)
	return nil
}

func typeofString(val Value) string {
	switch val.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case string:
		return "string"
	case *UserFunction, *BoundMethod:
		return "function"
	case Array:
		return "array"
	case *ClassValue, *InstanceValue, Object:
		return "object"
	default:
		return "object"
	}
}
