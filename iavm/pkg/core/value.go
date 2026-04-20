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
	Kind ValueKind
	Raw  any
}
