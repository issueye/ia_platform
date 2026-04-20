package core

type Instruction struct {
	Op OpCode
	A  uint32
	B  uint32
	C  uint32
}
