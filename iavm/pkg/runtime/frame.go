package runtime

type Frame struct {
	FunctionIndex uint32
	IP            uint32
	Locals        []any
	BasePointer   uint32
}
