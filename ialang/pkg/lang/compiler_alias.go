package lang

import comp "ialang/pkg/lang/compiler"

type Compiler = comp.Compiler

func NewCompiler() *Compiler {
	return comp.NewCompiler()
}
