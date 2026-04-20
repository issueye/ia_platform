package bytecode

import common "iacommon/pkg/ialang/bytecode"

type OpCode = common.OpCode

type Instruction = common.Instruction

type Chunk = common.Chunk

// FunctionTemplate is the compile-time function constant stored in bytecode.
// Runtime binds lexical env when OpClosure is executed.
type FunctionTemplate = common.FunctionTemplate

const (
	OpConstant = common.OpConstant
	OpAdd = common.OpAdd
	OpSub = common.OpSub
	OpMul = common.OpMul
	OpDiv = common.OpDiv
	OpMod = common.OpMod
	OpNeg = common.OpNeg
	OpNot = common.OpNot
	OpAnd = common.OpAnd
	OpOr = common.OpOr
	OpBitAnd = common.OpBitAnd
	OpBitOr = common.OpBitOr
	OpBitXor = common.OpBitXor
	OpShl = common.OpShl
	OpShr = common.OpShr
	OpTruthy = common.OpTruthy
	OpDup = common.OpDup
	OpEqual = common.OpEqual
	OpNotEqual = common.OpNotEqual
	OpGreater = common.OpGreater
	OpLess = common.OpLess
	OpGreaterEqual = common.OpGreaterEqual
	OpLessEqual = common.OpLessEqual
	OpPop = common.OpPop
	OpGetName = common.OpGetName
	OpDefineName = common.OpDefineName
	OpSetName = common.OpSetName
	OpClosure = common.OpClosure
	OpClass = common.OpClass
	OpSetProperty = common.OpSetProperty
	OpNew = common.OpNew
	OpGetGlobal = common.OpGetGlobal
	OpDefineGlobal = common.OpDefineGlobal
	OpArray = common.OpArray
	OpObject = common.OpObject
	OpGetProperty = common.OpGetProperty
	OpIndex = common.OpIndex
	OpCall = common.OpCall
	OpSpreadArray = common.OpSpreadArray
	OpSpreadObject = common.OpSpreadObject
	OpSpreadCall = common.OpSpreadCall
	OpAwait = common.OpAwait
	OpPushTry = common.OpPushTry
	OpPopTry = common.OpPopTry
	OpThrow = common.OpThrow
	OpJumpIfFalse = common.OpJumpIfFalse
	OpJumpIfTrue = common.OpJumpIfTrue
	OpJumpIfNullish = common.OpJumpIfNullish
	OpJumpIfNotNullish = common.OpJumpIfNotNullish
	OpJump = common.OpJump
	OpImportName = common.OpImportName
	OpImportNamespace = common.OpImportNamespace
	OpImportDynamic = common.OpImportDynamic
	OpExportName = common.OpExportName
	OpExportAs = common.OpExportAs
	OpExportDefault = common.OpExportDefault
	OpExportAll = common.OpExportAll
	OpSuper = common.OpSuper
	OpSuperCall = common.OpSuperCall
	OpTypeof = common.OpTypeof
	OpObjectKeys = common.OpObjectKeys
	OpReturn = common.OpReturn
)

func NewChunk() *Chunk {
	return common.NewChunk()
}

func NewChunkSized(codeCap, constCap int) *Chunk {
	return common.NewChunkSized(codeCap, constCap)
}
