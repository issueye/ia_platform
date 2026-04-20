package compiler

import bc "ialang/pkg/lang/bytecode"

type Chunk = bc.Chunk
type FunctionTemplate = bc.FunctionTemplate
type OpCode = bc.OpCode
type Instruction = bc.Instruction

const (
	OpConstant         = bc.OpConstant
	OpAdd              = bc.OpAdd
	OpSub              = bc.OpSub
	OpMul              = bc.OpMul
	OpDiv              = bc.OpDiv
	OpMod              = bc.OpMod
	OpNeg              = bc.OpNeg
	OpNot              = bc.OpNot
	OpAnd              = bc.OpAnd
	OpOr               = bc.OpOr
	OpJumpIfNullish    = bc.OpJumpIfNullish
	OpJumpIfNotNullish = bc.OpJumpIfNotNullish
	OpBitAnd           = bc.OpBitAnd
	OpBitOr            = bc.OpBitOr
	OpBitXor           = bc.OpBitXor
	OpShl              = bc.OpShl
	OpShr              = bc.OpShr
	OpTruthy           = bc.OpTruthy
	OpDup              = bc.OpDup
	OpEqual            = bc.OpEqual
	OpNotEqual         = bc.OpNotEqual
	OpGreater          = bc.OpGreater
	OpLess             = bc.OpLess
	OpGreaterEqual     = bc.OpGreaterEqual
	OpLessEqual        = bc.OpLessEqual
	OpPop              = bc.OpPop
	OpGetName          = bc.OpGetName
	OpDefineName       = bc.OpDefineName
	OpSetName          = bc.OpSetName
	OpClosure          = bc.OpClosure
	OpClass            = bc.OpClass
	OpSetProperty      = bc.OpSetProperty
	OpNew              = bc.OpNew
	OpArray            = bc.OpArray
	OpObject           = bc.OpObject
	OpGetProperty      = bc.OpGetProperty
	OpIndex            = bc.OpIndex
	OpCall             = bc.OpCall
	OpSpreadArray      = bc.OpSpreadArray
	OpSpreadObject     = bc.OpSpreadObject
	OpSpreadCall       = bc.OpSpreadCall
	OpAwait            = bc.OpAwait
	OpPushTry          = bc.OpPushTry
	OpPopTry           = bc.OpPopTry
	OpThrow            = bc.OpThrow
	OpJumpIfFalse      = bc.OpJumpIfFalse
	OpJumpIfTrue       = bc.OpJumpIfTrue
	OpJump             = bc.OpJump
	OpImportName       = bc.OpImportName
	OpImportNamespace  = bc.OpImportNamespace
	OpImportDynamic    = bc.OpImportDynamic
	OpExportName       = bc.OpExportName
	OpExportAs         = bc.OpExportAs
	OpExportDefault    = bc.OpExportDefault
	OpExportAll        = bc.OpExportAll
	OpSuper            = bc.OpSuper
	OpSuperCall        = bc.OpSuperCall
	OpTypeof           = bc.OpTypeof
	OpObjectKeys       = bc.OpObjectKeys
	OpReturn           = bc.OpReturn
)

func NewChunk() *Chunk {
	return bc.NewChunk()
}

func NewChunkSized(codeCap, constCap int) *Chunk {
	return bc.NewChunkSized(codeCap, constCap)
}
