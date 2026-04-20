package vm

import (
	"errors"
	"fmt"
	"time"
)

func (v *VM) callEntryCallable(callee Value) (Value, error) {
	switch fn := callee.(type) {
	case NativeFunction:
		return fn(nil)
	case *UserFunction:
		if fn.Async {
			return v.callUserFunctionAsync(fn, nil, nil), nil
		}
		return v.callUserFunctionSync(fn, nil, nil)
	case *BoundMethod:
		if fn.Method.Async {
			return v.callUserFunctionAsync(fn.Method, nil, fn.Receiver), nil
		}
		return v.callUserFunctionSync(fn.Method, nil, fn.Receiver)
	case *StringMethod:
		return fn.Method([]Value{fn.Value})
	case *ArrayMethod:
		return fn.Method([]Value{fn.Value})
	default:
		return nil, fmt.Errorf("value is not callable: %T", callee)
	}
}

func (v *VM) execSuperProperty(propName string) error {
	receiver, err := v.getReceiver()
	if err != nil {
		return err
	}

	instance, ok := receiver.(*InstanceValue)
	if !ok {
		return fmt.Errorf("super can only be used on instance methods")
	}

	method := v.lookupMethodInParent(instance.Class, propName)
	if method == nil {
		return fmt.Errorf("super property %s not found", propName)
	}

	v.push(&BoundMethod{
		Method:   method,
		Receiver: instance,
	})
	return nil
}

func (v *VM) execSuperCall(argc int) error {
	receiver, err := v.getReceiver()
	if err != nil {
		return err
	}

	instance, ok := receiver.(*InstanceValue)
	if !ok {
		return fmt.Errorf("super() can only be called on instance methods")
	}

	if len(v.stack) < argc {
		return fmt.Errorf("stack underflow on super call")
	}
	argStart := len(v.stack) - argc
	args := append([]Value(nil), v.stack[argStart:]...)
	v.stack = v.stack[:argStart]

	constructor := v.lookupMethodInParent(instance.Class, "constructor")
	if constructor == nil {
		return fmt.Errorf("parent class has no constructor")
	}

	if constructor.Async {
		v.push(v.callUserFunctionAsync(constructor, args, instance))
		return nil
	}

	_, callErr := v.callUserFunctionSync(constructor, args, instance)
	if callErr != nil {
		return callErr
	}
	v.push(nil)
	return nil
}

func (v *VM) getReceiver() (Value, error) {
	if v.env == nil {
		return nil, fmt.Errorf("no environment for super access")
	}
	val, ok := v.env.Get("this")
	if !ok {
		return nil, fmt.Errorf("no 'this' receiver for super access")
	}
	return val, nil
}

func (v *VM) lookupMethodInParent(class *ClassValue, name string) *UserFunction {
	current := class.Parent
	for current != nil {
		if method, ok := current.Methods[name]; ok {
			return method
		}
		current = current.Parent
	}
	return nil
}

func (v *VM) lookupMethodInClassHierarchy(class *ClassValue, name string) *UserFunction {
	current := class
	for current != nil {
		if method, ok := current.Methods[name]; ok {
			return method
		}
		current = current.Parent
	}
	return nil
}

func (v *VM) lookupGetterInClassHierarchy(class *ClassValue, name string) *UserFunction {
	current := class
	for current != nil {
		if getter, ok := current.Getters[name]; ok {
			return getter
		}
		current = current.Parent
	}
	return nil
}

func (v *VM) lookupSetterInClassHierarchy(class *ClassValue, name string) *UserFunction {
	current := class
	for current != nil {
		if setter, ok := current.Setters[name]; ok {
			return setter
		}
		current = current.Parent
	}
	return nil
}

func (v *VM) execNew(argc int) error {
	if len(v.stack) < argc+1 {
		return errors.New("stack underflow on new")
	}
	argStart := len(v.stack) - argc
	args := v.stack[argStart:]
	v.stack = v.stack[:argStart]

	callee, err := v.pop()
	if err != nil {
		return err
	}
	classVal, ok := callee.(*ClassValue)
	if !ok {
		return fmt.Errorf("new target is not a class: %T", callee)
	}

	instance := &InstanceValue{
		Class:  classVal,
		Fields: Object{},
	}
	for _, privateName := range classVal.PrivateFields {
		instance.Fields[privateName] = nil
	}
	if ctor, ok := classVal.Methods["constructor"]; ok {
		if ctor.Async {
			return fmt.Errorf("async constructor is not supported: %s", classVal.Name)
		}
		if _, err := v.callUserFunctionSync(ctor, args, instance); err != nil {
			return err
		}
	}
	v.push(instance)
	return nil
}

func (v *VM) execCall(argc int) error {
	if len(v.stack) < argc+1 {
		return errors.New("stack underflow on call")
	}

	argStart := len(v.stack) - argc
	args := v.stack[argStart:]
	v.stack = v.stack[:argStart]

	callee, err := v.pop()
	if err != nil {
		return err
	}

	switch fn := callee.(type) {
	case NativeFunction:
		ret, callErr := fn(args)
		if callErr != nil {
			return callErr
		}
		v.push(ret)
		return nil
	case *UserFunction:
		if fn.Async {
			v.push(v.callUserFunctionAsync(fn, cloneValues(args), nil))
			return nil
		}
		ret, callErr := v.callUserFunctionSync(fn, args, nil)
		if callErr != nil {
			return callErr
		}
		v.push(ret)
		return nil
	case *BoundMethod:
		if fn.Method.Async {
			v.push(v.callUserFunctionAsync(fn.Method, cloneValues(args), fn.Receiver))
			return nil
		}
		ret, callErr := v.callUserFunctionSync(fn.Method, args, fn.Receiver)
		if callErr != nil {
			return callErr
		}
		v.push(ret)
		return nil
	case *StringMethod:
		allArgs := make([]Value, len(args)+1)
		copy(allArgs, args)
		allArgs[len(args)] = fn.Value
		ret, callErr := fn.Method(allArgs)
		if callErr != nil {
			return callErr
		}
		v.push(ret)
		return nil
	case *ArrayMethod:
		allArgs := make([]Value, len(args)+1)
		copy(allArgs, args)
		allArgs[len(args)] = fn.Value
		ret, callErr := fn.Method(allArgs)
		if callErr != nil {
			return callErr
		}
		v.push(ret)
		return nil
	default:
		return fmt.Errorf("value is not callable: %T", callee)
	}
}

// execSpreadCall handles function calls with spread arguments.
// A operand: total argument count on stack
// B operand: number of spread arguments
func (v *VM) execSpreadCall(totalArgc, spreadArgc int) error {
	if spreadArgc < 1 {
		return v.execCall(totalArgc)
	}

	if len(v.stack) < totalArgc+1 {
		return errors.New("stack underflow on spread call")
	}

	// Collect arguments from stack (in reverse order)
	rawArgs := make([]Value, totalArgc)
	for i := 0; i < totalArgc; i++ {
		val, err := v.pop()
		if err != nil {
			return err
		}
		rawArgs[totalArgc-1-i] = val
	}

	// Pop the callee
	callee, err := v.pop()
	if err != nil {
		return err
	}

	// Flatten spread arguments
	flatArgs := make([]Value, 0, totalArgc)
	for _, arg := range rawArgs {
		if arr, isArray := arg.(Array); isArray {
			// This is a spread argument - flatten it
			for _, elem := range arr {
				flatArgs = append(flatArgs, elem)
			}
		} else {
			// Regular argument
			flatArgs = append(flatArgs, arg)
		}
	}

	// Call the function with flattened arguments
	switch fn := callee.(type) {
	case NativeFunction:
		ret, callErr := fn(flatArgs)
		if callErr != nil {
			return callErr
		}
		v.push(ret)
		return nil
	case *UserFunction:
		if fn.Async {
			v.push(v.callUserFunctionAsync(fn, cloneValues(flatArgs), nil))
			return nil
		}
		ret, callErr := v.callUserFunctionSync(fn, flatArgs, nil)
		if callErr != nil {
			return callErr
		}
		v.push(ret)
		return nil
	case *BoundMethod:
		if fn.Method.Async {
			v.push(v.callUserFunctionAsync(fn.Method, cloneValues(flatArgs), fn.Receiver))
			return nil
		}
		ret, callErr := v.callUserFunctionSync(fn.Method, flatArgs, fn.Receiver)
		if callErr != nil {
			return callErr
		}
		v.push(ret)
		return nil
	case *StringMethod:
		allArgs := make([]Value, len(flatArgs)+1)
		copy(allArgs, flatArgs)
		allArgs[len(flatArgs)] = fn.Value
		ret, callErr := fn.Method(allArgs)
		if callErr != nil {
			return callErr
		}
		v.push(ret)
		return nil
	case *ArrayMethod:
		allArgs := make([]Value, len(flatArgs)+1)
		copy(allArgs, flatArgs)
		allArgs[len(flatArgs)] = fn.Value
		ret, callErr := fn.Method(allArgs)
		if callErr != nil {
			return callErr
		}
		v.push(ret)
		return nil
	default:
		return fmt.Errorf("value is not callable: %T", callee)
	}
}

func (v *VM) callUserFunctionSync(fn *UserFunction, args []Value, receiver *InstanceValue) (Value, error) {
	if fn.RestParam == "" && len(args) > len(fn.Params) {
		return nil, fmt.Errorf("function %s expects %d args, got %d", fn.Name, len(fn.Params), len(args))
	}
	parentEnv := (*Environment)(nil)
	if fn.Env != nil {
		resolvedEnv, ok := fn.Env.(*Environment)
		if !ok {
			return nil, fmt.Errorf("function %s has invalid env type %T", fn.Name, fn.Env)
		}
		parentEnv = resolvedEnv
	}
	fnChunk, ok := fn.Chunk.(*Chunk)
	if !ok || fnChunk == nil {
		return nil, fmt.Errorf("function %s has invalid chunk type %T", fn.Name, fn.Chunk)
	}

	envSize := len(fn.Params) + 1
	if fn.RestParam != "" {
		envSize++
	}
	env := NewEnvironmentSized(parentEnv, envSize)
	if receiver != nil {
		env.Define("this", receiver)
	}
	for i, p := range fn.Params {
		if i < len(args) {
			env.Define(p, args[i])
		} else {
			env.Define(p, nil) // Will be replaced by default value in compiled code
		}
	}
	if fn.RestParam != "" {
		rest := Array{}
		if len(args) > len(fn.Params) {
			rest = append(rest, args[len(fn.Params):]...)
		}
		env.Define(fn.RestParam, rest)
	}

	child := vmPool.Get().(*VM)
	child.chunk = fnChunk
	child.ip = 0
	child.stack = child.stack[:0]
	child.globals = v.globals
	child.modules = v.modules
	child.modulePath = v.modulePath
	child.resolveImport = v.resolveImport
	child.asyncRuntime = v.asyncRuntime
	child.options = v.options
	child.env = env
	child.exports = nil
	child.tryStack = child.tryStack[:0]
	child.sandbox = nil
	child.stepCounter = nil
	child.startTime = time.Time{}

	ret, runErr := child.runChunk()

	for i := range child.stack {
		child.stack[i] = nil
	}
	child.stack = child.stack[:0]
	child.chunk = nil
	child.globals = nil
	child.modules = nil
	child.resolveImport = nil
	child.asyncRuntime = nil
	child.env = nil
	child.exports = nil
	child.tryStack = child.tryStack[:0]
	child.sandbox = nil
	child.stepCounter = nil
	child.options = VMOptions{}
	child.modulePath = ""
	child.startTime = time.Time{}
	vmPool.Put(child)

	return ret, runErr
}

func cloneValues(vals []Value) []Value {
	if len(vals) == 0 {
		return nil
	}
	out := make([]Value, len(vals))
	copy(out, vals)
	return out
}

func (v *VM) callUserFunctionAsync(fn *UserFunction, args []Value, receiver *InstanceValue) Awaitable {
	return v.asyncRuntime.Spawn(func() (Value, error) {
		return v.callUserFunctionSync(fn, args, receiver)
	})
}

func CallUserFunctionSync(fn *UserFunction, args []Value) (Value, error) {
	if fn == nil {
		return nil, fmt.Errorf("function is nil")
	}
	if fn.RestParam == "" && len(args) > len(fn.Params) {
		return nil, fmt.Errorf("function %s expects %d args, got %d", fn.Name, len(fn.Params), len(args))
	}

	parentEnv := (*Environment)(nil)
	if fn.Env != nil {
		resolvedEnv, ok := fn.Env.(*Environment)
		if !ok {
			return nil, fmt.Errorf("function %s has invalid env type %T", fn.Name, fn.Env)
		}
		parentEnv = resolvedEnv
	}
	fnChunk, ok := fn.Chunk.(*Chunk)
	if !ok || fnChunk == nil {
		return nil, fmt.Errorf("function %s has invalid chunk type %T", fn.Name, fn.Chunk)
	}

	envSize := len(fn.Params)
	if fn.RestParam != "" {
		envSize++
	}
	env := NewEnvironmentSized(parentEnv, envSize)
	for i, p := range fn.Params {
		if i < len(args) {
			env.Define(p, args[i])
		} else {
			env.Define(p, nil) // Will be replaced by default value in compiled code
		}
	}
	if fn.RestParam != "" {
		rest := Array{}
		if len(args) > len(fn.Params) {
			rest = append(rest, args[len(fn.Params):]...)
		}
		env.Define(fn.RestParam, rest)
	}

	runner := NewVM(fnChunk, nil, nil, "", NewGoroutineRuntime())
	runner.env = env
	return runner.runChunk()
}

func CallBoundMethodSync(fn *BoundMethod, args []Value) (Value, error) {
	if fn == nil {
		return nil, fmt.Errorf("bound method is nil")
	}
	if fn.Method == nil {
		return nil, fmt.Errorf("bound method function is nil")
	}

	runner := NewVM(nil, nil, nil, "", NewGoroutineRuntime())
	return runner.callUserFunctionSync(fn.Method, args, fn.Receiver)
}
