package core

type OpCode uint16

const (
	OpNop OpCode = iota
	OpConst
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpNeg
	OpNot
	OpEq
	OpNe
	OpLt
	OpGt
	OpLe
	OpGe
	OpJump
	OpJumpIfFalse
	OpCall
	OpReturn
	OpLoadLocal
	OpStoreLocal
	OpLoadGlobal
	OpStoreGlobal
	OpMakeArray
	OpMakeObject
	OpGetProp
	OpSetProp
	OpImportFunc
	OpImportCap
	OpHostCall
	OpHostPoll
)
