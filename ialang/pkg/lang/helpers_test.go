package lang_test

import (
	"strconv"
	"testing"

	comp "ialang/pkg/lang/compiler"
	"ialang/pkg/lang/frontend"
	rt "ialang/pkg/lang/runtime"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	rvm "ialang/pkg/lang/runtime/vm"
)

func compileTestSource(tb testing.TB, source string) *comp.Chunk {
	tb.Helper()
	l := frontend.NewLexer(source)
	p := frontend.NewParser(l)
	program := p.ParseProgram()

	c := comp.NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		tb.Fatalf("compilation errors: %v", errs)
	}
	return chunk
}

func itoaTest(i int) string {
	return strconv.Itoa(i)
}

func createTestVM(chunk *comp.Chunk) *rvm.VM {
	runtime := rt.NewGoroutineRuntime()
	return rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "test.ia", runtime)
}

func runVMTestNoError(t *testing.T, source string) {
	t.Helper()
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
}

func runVMTestExpectError(t *testing.T, source string) {
	t.Helper()
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err == nil {
		t.Fatal("vm.Run() expected error, got nil")
	}
}
