package types

import (
	"context"
	common "iacommon/pkg/ialang/value"
)

type Value = common.Value
type Object = common.Object
type Array = common.Array

type NativeFunction = common.NativeFunction

type Awaitable interface {
	Await() (Value, error)
	IsDone() bool
}

type ContextAwaitable interface {
	AwaitContext(ctx context.Context) (Value, error)
}

type AsyncTask func() (Value, error)

type AsyncRuntime interface {
	Spawn(task AsyncTask) Awaitable
	AwaitValue(v Value) (Value, error)
	Name() string
}
