package builtin

import (
	common "iacommon/pkg/ialang/value"
	rttypes "ialang/pkg/lang/runtime/types"
)

type Value = common.Value
type Object = common.Object
type Array = common.Array
type NativeFunction = common.NativeFunction
type UserFunction = rttypes.UserFunction
type BoundMethod = rttypes.BoundMethod
type AsyncRuntime = rttypes.AsyncRuntime
