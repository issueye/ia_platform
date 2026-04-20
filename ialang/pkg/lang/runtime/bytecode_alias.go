package runtime

import bc "iacommon/pkg/ialang/bytecode"

type OpCode = bc.OpCode
type Instruction = bc.Instruction
type Chunk = bc.Chunk
type FunctionTemplate = bc.FunctionTemplate

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
	OpGetGlobal        = bc.OpGetGlobal
	OpDefineGlobal     = bc.OpDefineGlobal
	OpArray            = bc.OpArray
	OpObject           = bc.OpObject
	OpGetProperty      = bc.OpGetProperty
	OpIndex            = bc.OpIndex
	OpCall             = bc.OpCall
	OpAwait            = bc.OpAwait
	OpPushTry          = bc.OpPushTry
	OpPopTry           = bc.OpPopTry
	OpThrow            = bc.OpThrow
	OpJumpIfFalse      = bc.OpJumpIfFalse
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
	OpReturn           = bc.OpReturn
)
