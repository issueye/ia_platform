package lang_test

import (
	"strconv"
	"testing"

	"ialang/pkg/lang/bytecode"
	comp "ialang/pkg/lang/compiler"
	"ialang/pkg/lang/frontend"
	rt "ialang/pkg/lang/runtime"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	rvm "ialang/pkg/lang/runtime/vm"
)

func compileTestSource(tb testing.TB, source string) *bytecode.Chunk {
	tb.Helper()

	l := frontend.NewLexer(source)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		tb.Fatalf("parse errors: %v", errs)
	}

	c := comp.NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		tb.Fatalf("compile errors: %v", errs)
	}
	return chunk
}

func itoaTest(v int) string {
	return strconv.Itoa(v)
}

func createTestVM(chunk *bytecode.Chunk) *rvm.VM {
	runtime := rt.NewGoroutineRuntime()
	return rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "test.ia", runtime)
}

func runVMTestNoError(tb testing.TB, source string) {
	tb.Helper()
	chunk := compileTestSource(tb, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		tb.Fatalf("vm.Run() unexpected error: %v", err)
	}
}

func runVMTestExpectError(tb testing.TB, source string) {
	tb.Helper()
	chunk := compileTestSource(tb, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err == nil {
		tb.Fatal("vm.Run() expected error, got nil")
	}
}
