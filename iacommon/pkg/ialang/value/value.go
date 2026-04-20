package value

type Value any

type Object map[string]Value

type Array []Value

type NativeFunction func(args []Value) (Value, error)
