package vm

import (
	"errors"
	"fmt"
	"strings"
)

func (v *VM) execClass(methodCount int) error {
	privateFieldCount := (methodCount >> 16) & 0xF
	hasParent := (methodCount >> 20) & 1
	instanceMethodCount := methodCount & 0xF
	staticMethodCount := (methodCount >> 4) & 0xF
	getterCount := (methodCount >> 8) & 0xF
	setterCount := (methodCount >> 12) & 0xF

	instanceMethods := map[string]*UserFunction{}
	staticMethods := map[string]*UserFunction{}
	getters := map[string]*UserFunction{}
	setters := map[string]*UserFunction{}
	privateFields := make([]string, 0, privateFieldCount)

	// The compiler emits in order: instance, static, getters, setters
	// then private fields.
	// But the stack is LIFO, so we pop in reverse order: private, setters, getters, static, instance

	// Read private field names (emitted last, popped first)
	for i := 0; i < privateFieldCount; i++ {
		fieldNameVal, err := v.pop()
		if err != nil {
			return err
		}
		fieldName, ok := fieldNameVal.(string)
		if !ok {
			return fmt.Errorf("private field name must be string, got %T", fieldNameVal)
		}
		privateFields = append(privateFields, fieldName)
	}

	// Read setters (emitted last, popped first)
	for i := 0; i < setterCount; i++ {
		name, fn, err := v.popMethodPair()
		if err != nil {
			return err
		}
		setters[name] = fn
	}

	// Read getters
	for i := 0; i < getterCount; i++ {
		name, fn, err := v.popMethodPair()
		if err != nil {
			return err
		}
		getters[name] = fn
	}

	// Read static methods
	for i := 0; i < staticMethodCount; i++ {
		name, fn, err := v.popMethodPair()
		if err != nil {
			return err
		}
		staticMethods[name] = fn
	}

	// Read instance methods (emitted first, popped last)
	for i := 0; i < instanceMethodCount; i++ {
		name, fn, err := v.popMethodPair()
		if err != nil {
			return err
		}
		instanceMethods[name] = fn
	}

	classNameVal, err := v.pop()
	if err != nil {
		return err
	}
	className, ok := classNameVal.(string)
	if !ok {
		return fmt.Errorf("class name must be string, got %T", classNameVal)
	}

	var parentClass *ClassValue
	if hasParent == 1 {
		parentNameVal, err := v.pop()
		if err != nil {
			return err
		}
		parentName, ok := parentNameVal.(string)
		if !ok {
			return fmt.Errorf("parent class name must be string, got %T", parentNameVal)
		}
		parentVal, exists := v.resolveClass(parentName)
		if !exists {
			return fmt.Errorf("parent class not found: %s", parentName)
		}
		var ok2 bool
		parentClass, ok2 = parentVal.(*ClassValue)
		if !ok2 {
			return fmt.Errorf("parent class must be ClassValue, got %T", parentVal)
		}
	}

	v.push(&ClassValue{
		Name:          className,
		Parent:        parentClass,
		Methods:       instanceMethods,
		StaticMethods: staticMethods,
		Getters:       getters,
		Setters:       setters,
		PrivateFields: privateFields,
	})
	return nil
}

// popMethodPair pops a method name and function from the stack.
func (v *VM) popMethodPair() (string, *UserFunction, error) {
	methodVal, err := v.pop()
	if err != nil {
		return "", nil, err
	}
	methodNameVal, err := v.pop()
	if err != nil {
		return "", nil, err
	}

	methodName, ok := methodNameVal.(string)
	if !ok {
		return "", nil, fmt.Errorf("class method name must be string, got %T", methodNameVal)
	}
	method, ok := methodVal.(*UserFunction)
	if !ok {
		return "", nil, fmt.Errorf("class method must be function, got %T", methodVal)
	}
	return methodName, method, nil
}

func (v *VM) resolveClass(name string) (Value, bool) {
	if v.env != nil {
		if val, ok := v.env.Get(name); ok {
			return val, true
		}
	}
	val, ok := v.globals[name]
	return val, ok
}

func (v *VM) execGetProperty(name string) error {
	targetVal, err := v.pop()
	if err != nil {
		return err
	}
	switch target := targetVal.(type) {
	case Object:
		if val, ok := target[name]; ok {
			v.push(val)
			return nil
		}
		// JS-style: return null for missing properties
		v.push(nil)
		return nil
	case *InstanceValue:
		if isPrivateFieldName(name) {
			if !v.canAccessPrivateField(target) {
				return fmt.Errorf("private field access denied: %s", demanglePrivateName(name))
			}
			if val, ok := target.Fields[name]; ok {
				v.push(val)
				return nil
			}
			return fmt.Errorf("private field not found: %s", demanglePrivateName(name))
		}
		// Check for getter first
		if getter := v.lookupGetterInClassHierarchy(target.Class, name); getter != nil {
			// Call the getter with the instance as receiver
			ret, callErr := v.callUserFunctionSync(getter, nil, target)
			if callErr != nil {
				return fmt.Errorf("getter call error: %w", callErr)
			}
			v.push(ret)
			return nil
		}
		if val, ok := target.Fields[name]; ok {
			v.push(val)
			return nil
		}
		if method := v.lookupMethodInClassHierarchy(target.Class, name); method != nil {
			v.push(&BoundMethod{
				Method:   method,
				Receiver: target,
			})
			return nil
		}
		// JS-style: return null for missing properties
		v.push(nil)
		return nil
	case *ClassValue:
		// Accessing property on a class (for static methods)
		if staticMethod, ok := target.StaticMethods[name]; ok {
			v.push(staticMethod)
			return nil
		}
		// Inherit static methods from parent
		current := target.Parent
		for current != nil {
			if staticMethod, ok := current.StaticMethods[name]; ok {
				v.push(staticMethod)
				return nil
			}
			current = current.Parent
		}
		v.push(nil)
		return nil
	default:
		if str, ok := targetVal.(string); ok {
			if name == "length" {
				v.push(float64(len(str)))
				return nil
			}
			proto := GetStringPrototype()
			if method, exists := proto[name]; exists {
				v.push(&StringMethod{
					Method: method.(NativeFunction),
					Value:  str,
				})
				return nil
			}
			// JS-style: return null for missing string properties
			v.push(nil)
			return nil
		}
		if arr, ok := targetVal.(Array); ok {
			proto := GetArrayPrototype()
			if method, exists := proto[name]; exists {
				v.push(&ArrayMethod{
					Method: method.(NativeFunction),
					Value:  arr,
				})
				return nil
			}
			if name == "length" {
				v.push(float64(len(arr)))
				return nil
			}
			// JS-style: return null for missing array properties
			v.push(nil)
			return nil
		}
		// Type error: this should still be an error
		return fmt.Errorf("not an object for property access: %T", targetVal)
	}
}

func (v *VM) execSetProperty(name string) error {
	value, err := v.pop()
	if err != nil {
		return err
	}
	targetVal, err := v.pop()
	if err != nil {
		return err
	}
	switch target := targetVal.(type) {
	case Object:
		target[name] = value
		return nil
	case *InstanceValue:
		if isPrivateFieldName(name) {
			if !v.canAccessPrivateField(target) {
				return fmt.Errorf("private field access denied: %s", demanglePrivateName(name))
			}
			target.Fields[name] = value
			return nil
		}
		// Check for setter first
		if setter := v.lookupSetterInClassHierarchy(target.Class, name); setter != nil {
			// Call the setter with the instance as receiver and the value as argument
			_, callErr := v.callUserFunctionSync(setter, []Value{value}, target)
			return callErr
		}
		target.Fields[name] = value
		return nil
	default:
		return fmt.Errorf("not an object for property set: %T", targetVal)
	}
}

func isPrivateFieldName(name string) bool {
	return strings.HasPrefix(name, "__private_")
}

func demanglePrivateName(name string) string {
	return strings.TrimPrefix(name, "__private_")
}

func (v *VM) canAccessPrivateField(target *InstanceValue) bool {
	if v.env == nil {
		return false
	}
	thisVal, ok := v.env.Get("this")
	if !ok {
		return false
	}
	receiver, ok := thisVal.(*InstanceValue)
	if !ok {
		return false
	}
	return receiver == target
}

func (v *VM) execArray(count int) error {
	if len(v.stack) < count {
		return errors.New("stack underflow on array construction")
	}
	start := len(v.stack) - count
	elems := append([]Value(nil), v.stack[start:]...)
	v.stack = v.stack[:start]
	v.push(Array(elems))
	return nil
}

func (v *VM) execObject(count int) error {
	if len(v.stack) < count*2 {
		return errors.New("stack underflow on object construction")
	}
	obj := Object{}
	for i := 0; i < count; i++ {
		val, err := v.pop()
		if err != nil {
			return err
		}
		keyVal, err := v.pop()
		if err != nil {
			return err
		}
		key, ok := keyVal.(string)
		if !ok {
			return fmt.Errorf("object key must be string, got %T", keyVal)
		}
		obj[key] = val
	}
	v.push(obj)
	return nil
}

// execSpreadArray creates an array from stack values, flattening spread elements.
// A operand: total element count on stack
// B operand: number of spread elements
func (v *VM) execSpreadArray(totalCount, spreadCount int) error {
	if spreadCount < 1 {
		return v.execArray(totalCount)
	}

	// Collect all elements from stack (they're in reverse order)
	elements := make([]Value, 0, totalCount)
	for i := 0; i < totalCount; i++ {
		val, err := v.pop()
		if err != nil {
			return err
		}
		elements = append(elements, val)
	}

	// Reverse to get correct order
	for i, j := 0, len(elements)-1; i < j; i, j = i+1, j-1 {
		elements[i], elements[j] = elements[j], elements[i]
	}

	// Build result array with spread flattening
	result := Array{}
	for _, elem := range elements {
		if arr, isArray := elem.(Array); isArray {
			// Flatten spread array
			result = append(result, arr...)
		} else {
			result = append(result, elem)
		}
	}

	v.push(result)
	return nil
}

// execSpreadObject creates an object from stack values, merging spread objects.
// A operand: regular property count
// B operand: spread property count
func (v *VM) execSpreadObject(regCount, spreadCount int) error {
	obj := Object{}

	// First, process spread objects (they come after regular properties on stack)
	// Spread objects are pushed in order, later ones override earlier properties
	spreadElements := make([]Value, spreadCount)
	for i := spreadCount - 1; i >= 0; i-- {
		val, err := v.pop()
		if err != nil {
			return err
		}
		spreadElements[i] = val
	}

	// Merge spread objects into result
	for _, spreadVal := range spreadElements {
		if spreadObj, isObject := spreadVal.(Object); isObject {
			// Copy all properties from spread object
			for key, val := range spreadObj {
				obj[key] = val
			}
		} else {
			return fmt.Errorf("spread value must be object, got %T", spreadVal)
		}
	}

	// Then process regular properties (they override spread properties)
	for i := 0; i < regCount; i++ {
		val, err := v.pop()
		if err != nil {
			return err
		}
		keyVal, err := v.pop()
		if err != nil {
			return err
		}
		key, ok := keyVal.(string)
		if !ok {
			return fmt.Errorf("object key must be string, got %T", keyVal)
		}
		obj[key] = val
	}

	v.push(obj)
	return nil
}

func (v *VM) execIndex() error {
	indexVal, err := v.pop()
	if err != nil {
		return err
	}
	targetVal, err := v.pop()
	if err != nil {
		return err
	}

	switch target := targetVal.(type) {
	case Array:
		idx, ok := indexVal.(float64)
		if !ok {
			return fmt.Errorf("array index must be number, got %T", indexVal)
		}
		i := int(idx)
		if float64(i) != idx {
			return fmt.Errorf("array index must be integer, got %v", idx)
		}
		if i < 0 || i >= len(target) {
			v.push(nil)
			return nil
		}
		v.push(target[i])
		return nil
	case Object:
		key, ok := indexVal.(string)
		if !ok {
			return fmt.Errorf("object index must be string, got %T", indexVal)
		}
		if val, exists := target[key]; exists {
			v.push(val)
		} else {
			v.push(nil)
		}
		return nil
	case string:
		idx, ok := indexVal.(float64)
		if !ok {
			return fmt.Errorf("string index must be number, got %T", indexVal)
		}
		i := int(idx)
		if float64(i) != idx {
			return fmt.Errorf("string index must be integer, got %v", idx)
		}
		if i < 0 || i >= len(target) {
			v.push(nil)
			return nil
		}
		v.push(string(target[i]))
		return nil
	default:
		return fmt.Errorf("index operator not supported on %T", targetVal)
	}
}
