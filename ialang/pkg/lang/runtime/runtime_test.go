package runtime_test

import (
	"testing"

	comp "ialang/pkg/lang/compiler"
	"ialang/pkg/lang/frontend"
	rt "ialang/pkg/lang/runtime"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	rvm "ialang/pkg/lang/runtime/vm"
)

func TestVMRunBasicProgram(t *testing.T) {
	src := `
let x = 1 + 2;
let y = x - 1;
`
	chunk := compileChunk(t, src)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "runtime_basic_test.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
}

func TestVMStructuredRuntimeErrorOption(t *testing.T) {
	src := `
try {
  not_exists = 1;
} catch (e) {
  if (e.code != "RUNTIME_ERROR") {
    throw "unexpected-runtime-error-code";
  }
}
`
	chunk := compileChunk(t, src)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVMWithOptions(
		chunk,
		rtbuiltin.DefaultModules(runtime),
		nil,
		"runtime_structured_error_test.ia",
		runtime,
		rvm.VMOptions{StructuredRuntimeErrors: true},
	)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
}

func compileChunk(t *testing.T, source string) *rt.Chunk {
	t.Helper()
	l := frontend.NewLexer(source)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	c := comp.NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	return chunk
}
