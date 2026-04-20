package vm

import (
	"errors"
	"fmt"
)

func (v *VM) stringConstant(idx int, label string) (string, error) {
	if idx < 0 || idx >= len(v.chunk.Constants) {
		return "", fmt.Errorf("%s constant index out of range: %d", label, idx)
	}
	name, ok := v.chunk.Constants[idx].(string)
	if !ok {
		return "", fmt.Errorf("%s is not a string constant", label)
	}
	return name, nil
}

func (v *VM) functionConstant(idx int) (*FunctionTemplate, error) {
	if idx < 0 || idx >= len(v.chunk.Constants) {
		return nil, fmt.Errorf("function constant index out of range: %d", idx)
	}
	switch fn := v.chunk.Constants[idx].(type) {
	case *FunctionTemplate:
		return fn, nil
	case *UserFunction:
		// Backward compatibility for legacy chunks that encode UserFunction constants.
		typedChunk, ok := fn.Chunk.(*Chunk)
		if !ok || typedChunk == nil {
			return nil, fmt.Errorf("legacy function constant has invalid chunk: %T", fn.Chunk)
		}
		return &FunctionTemplate{
			Name:      fn.Name,
			Params:    append([]string(nil), fn.Params...),
			RestParam: fn.RestParam,
			Async:     fn.Async,
			Chunk:     typedChunk,
		}, nil
	default:
		return nil, errors.New("constant is not a function template")
	}
}

func (v *VM) bindFunctionTemplate(tmpl *FunctionTemplate) *UserFunction {
	return &UserFunction{
		Name:      tmpl.Name,
		Params:    append([]string(nil), tmpl.Params...),
		RestParam: tmpl.RestParam,
		Async:     tmpl.Async,
		Chunk:     tmpl.Chunk,
		Env:       v.env,
	}
}

func (v *VM) execDup() error {
	if len(v.stack) == 0 {
		return errors.New("stack underflow")
	}
	val := v.stack[len(v.stack)-1]
	v.stack = append(v.stack, val)
	return nil
}
