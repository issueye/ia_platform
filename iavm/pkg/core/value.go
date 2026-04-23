package core

type ValueKind uint8

const (
	ValueNull ValueKind = iota
	ValueBool
	ValueI64
	ValueF64
	ValueString
	ValueBytes
	ValueArrayRef
	ValueObjectRef
	ValueFuncRef
	ValueHostHandle
)

type Value struct {
	Kind          ValueKind
	Raw           any
	Upvalues      []*Upvalue // captured upvalue cells (for ValueFuncRef closures)
	BoundReceiver *Value     // bound receiver for instance method calls
}

type Upvalue struct {
	Closed bool
	Value  Value
}
