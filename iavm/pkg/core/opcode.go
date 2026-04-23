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
	OpJumpIfTrue
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
	OpDup
	OpPop
	OpBitAnd
	OpBitOr
	OpBitXor
	OpShl
	OpShr
	OpAnd
	OpOr
	OpTypeof
	OpPushTry
	OpPopTry
	OpIndex
	OpThrow
	OpTruthy
	OpObjectKeys
	OpJumpIfNullish
	OpJumpIfNotNullish
	OpClosure
	OpNewInstance
)
