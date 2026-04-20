package bytecode

type OpCode uint8

const (
	OpConstant OpCode = iota
	OpAdd
	OpSub
	OpMul
	OpDiv
	OpMod
	OpNeg
	OpNot
	OpAnd
	OpOr
	OpBitAnd
	OpBitOr
	OpBitXor
	OpShl
	OpShr
	OpTruthy
	OpDup
	OpEqual
	OpNotEqual
	OpGreater
	OpLess
	OpGreaterEqual
	OpLessEqual
	OpPop
	OpGetName
	OpDefineName
	OpSetName
	OpClosure
	OpClass
	OpSetProperty
	OpNew
	OpGetGlobal
	OpDefineGlobal
	OpArray
	OpObject
	OpGetProperty
	OpIndex
	OpCall
	OpSpreadArray
	OpSpreadObject
	OpSpreadCall
	OpAwait
	OpPushTry
	OpPopTry
	OpThrow
	OpJumpIfFalse
	OpJumpIfTrue
	OpJumpIfNullish
	OpJumpIfNotNullish
	OpJump
	OpImportName
	OpImportNamespace
	OpImportDynamic
	OpExportName
	OpExportAs
	OpExportDefault
	OpExportAll
	OpSuper
	OpSuperCall
	OpTypeof
	OpObjectKeys
	OpReturn
)

type Instruction struct {
	Op OpCode
	A  int
	B  int
}

type Chunk struct {
	Code      []Instruction
	Constants []any
}

// FunctionTemplate is the compile-time function constant stored in bytecode.
// Runtime binds lexical env when OpClosure is executed.
type FunctionTemplate struct {
	Name          string
	Params        []string
	RestParam     string
	ParamDefaults []any // compiled default value constants (nil for params without defaults)
	Async         bool
	Chunk         *Chunk
}

func NewChunk() *Chunk {
	return &Chunk{}
}

func NewChunkSized(codeCap, constCap int) *Chunk {
	if codeCap < 0 {
		codeCap = 0
	}
	if constCap < 0 {
		constCap = 0
	}
	return &Chunk{
		Code:      make([]Instruction, 0, codeCap),
		Constants: make([]any, 0, constCap),
	}
}

func (c *Chunk) AddConstant(v any) int {
	c.Constants = append(c.Constants, v)
	return len(c.Constants) - 1
}

func (c *Chunk) Emit(op OpCode, a, b int) {
	c.Code = append(c.Code, Instruction{Op: op, A: a, B: b})
}
