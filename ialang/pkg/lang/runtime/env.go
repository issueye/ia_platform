package runtime

type Environment struct {
	values map[string]Value
	parent *Environment
}

func NewEnvironment(parent *Environment) *Environment {
	return NewEnvironmentSized(parent, 0)
}

func NewEnvironmentSized(parent *Environment, size int) *Environment {
	if size < 0 {
		size = 0
	}
	return &Environment{
		values: make(map[string]Value, size),
		parent: parent,
	}
}

func (e *Environment) Get(name string) (Value, bool) {
	for cur := e; cur != nil; cur = cur.parent {
		if val, ok := cur.values[name]; ok {
			return val, true
		}
	}
	return nil, false
}

func (e *Environment) Define(name string, val Value) {
	e.values[name] = val
}

func (e *Environment) Set(name string, val Value) bool {
	for cur := e; cur != nil; cur = cur.parent {
		if _, ok := cur.values[name]; ok {
			cur.values[name] = val
			return true
		}
	}
	return false
}
