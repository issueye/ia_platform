package builtin

import (
	"fmt"
)

func newCryptoModule() Object {
	namespace := Object{
		"sha256": newHashFn("crypto.sha256", "sha256"),
		"md5":    newHashFn("crypto.md5", "md5"),
	}
	module := cloneObject(namespace)
	module["crypto"] = namespace
	return module
}

func newHashFn(label, algo string) NativeFunction {
	return NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("%s expects 1 arg, got %d", label, len(args))
		}
		text, err := asStringArg(label, args, 0)
		if err != nil {
			return nil, err
		}
		return computeHash(algo, text)
	})
}
