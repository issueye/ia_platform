package runtime

import "fmt"

var userFunctionCaller func(fn *UserFunction, args []Value) (Value, error)

// RegisterUserFunctionCaller allows runtime helpers to invoke script functions
// without importing the VM package and creating a package cycle.
func RegisterUserFunctionCaller(caller func(fn *UserFunction, args []Value) (Value, error)) {
	userFunctionCaller = caller
}

func callCallable(callback Value, args []Value, label string) (Value, error) {
	switch cb := callback.(type) {
	case NativeFunction:
		return cb(args)
	case *UserFunction:
		if userFunctionCaller == nil {
			return nil, fmt.Errorf("%s callback requires VM function caller", label)
		}
		return userFunctionCaller(cb, args)
	default:
		return nil, fmt.Errorf("%s callback expects function, got %T", label, callback)
	}
}