package vm

import (
	bc "iacommon/pkg/ialang/bytecode"
	rt "ialang/pkg/lang/runtime"
)

type Value = rt.Value
type Object = rt.Object
type Array = rt.Array
type NativeFunction = rt.NativeFunction

type UserFunction = rt.UserFunction
type ClassValue = rt.ClassValue
type InstanceValue = rt.InstanceValue
type BoundMethod = rt.BoundMethod
type StringMethod = rt.StringMethod
type ArrayMethod = rt.ArrayMethod

type Awaitable = rt.Awaitable
type AsyncRuntime = rt.AsyncRuntime

type Environment = rt.Environment
type SandboxPolicy = rt.SandboxPolicy
type StepCounter = rt.StepCounter
type SandboxError = rt.SandboxError
type Promise = rt.Promise

var NewEnvironment = rt.NewEnvironment
var NewEnvironmentSized = rt.NewEnvironmentSized
var NewStepCounter = rt.NewStepCounter
var NewGoroutineRuntime = rt.NewGoroutineRuntime
var NewAsyncRuntimeErrorValue = rt.NewAsyncRuntimeErrorValue
var GetStringPrototype = rt.GetStringPrototype
var GetArrayPrototype = rt.GetArrayPrototype

var ErrAsyncTaskTimeout = rt.ErrAsyncTaskTimeout
var ErrAsyncAwaitTimeout = rt.ErrAsyncAwaitTimeout

const (
	AsyncErrorCodeTaskTimeout  = rt.AsyncErrorCodeTaskTimeout
	AsyncErrorCodeAwaitTimeout = rt.AsyncErrorCodeAwaitTimeout
	AsyncErrorKindTimeout      = rt.AsyncErrorKindTimeout
	RuntimeErrorCodeGeneric    = rt.RuntimeErrorCodeGeneric
	RuntimeErrorKindGeneric    = rt.RuntimeErrorKindGeneric
)

type Chunk = bc.Chunk
type FunctionTemplate = bc.FunctionTemplate
type Instruction = bc.Instruction
type OpCode = bc.OpCode

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
	OpGetGlobal        = bc.OpGetGlobal
	OpDefineGlobal     = bc.OpDefineGlobal
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
