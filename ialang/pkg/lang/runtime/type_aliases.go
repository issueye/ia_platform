package runtime

import (
	common "iacommon/pkg/ialang/value"
	rttypes "ialang/pkg/lang/runtime/types"
)

type Value = common.Value
type Object = common.Object
type Array = common.Array
type NativeFunction = common.NativeFunction

type UserFunction = rttypes.UserFunction
type ClassValue = rttypes.ClassValue
type InstanceValue = rttypes.InstanceValue
type BoundMethod = rttypes.BoundMethod
type StringMethod = rttypes.StringMethod
type ArrayMethod = rttypes.ArrayMethod
