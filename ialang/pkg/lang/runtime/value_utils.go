package runtime

import "fmt"

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
