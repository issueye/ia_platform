package types

type UserFunction struct {
	Name      string
	Params    []string
	RestParam string
	Async     bool
	Chunk     any
	Env       any
}

type ClassValue struct {
	Name          string
	Parent        *ClassValue
	Methods       map[string]*UserFunction // instance methods
	StaticMethods map[string]*UserFunction // static methods (called on class itself)
	Getters       map[string]*UserFunction // getter functions
	Setters       map[string]*UserFunction // setter functions
	PrivateFields []string
}

type InstanceValue struct {
	Class  *ClassValue
	Fields Object
}

type BoundMethod struct {
	Method   *UserFunction
	Receiver *InstanceValue
}

type StringMethod struct {
	Method NativeFunction
	Value  string
}

type ArrayMethod struct {
	Method NativeFunction
	Value  Array
}
