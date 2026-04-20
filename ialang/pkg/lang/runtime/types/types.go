package types

import "context"

type Value any
type Object map[string]Value
type Array []Value

type NativeFunction func(args []Value) (Value, error)

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
