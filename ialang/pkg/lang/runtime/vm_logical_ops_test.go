package runtime_test

import (
	"testing"

	comp "ialang/pkg/lang/compiler"
	"ialang/pkg/lang/frontend"
	rt "ialang/pkg/lang/runtime"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	rvm "ialang/pkg/lang/runtime/vm"
)

// ============================================================
// JS-style logical operator return value tests
// ============================================================

func runVMAndGetResult(t *testing.T, source string) any {
	t.Helper()
	chunk := compileChunkForLogicalOps(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "logical_ops_test.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
	// Get the value of 'x' from VM globals
	if val, ok := vm.Globals()["x"]; ok {
		return val
	}
	t.Fatal("variable 'x' not found in globals")
	return nil
}

func runVMExpectNoErrorForLogicalOps(t *testing.T, source string) {
	t.Helper()
	chunk := compileChunkForLogicalOps(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "logical_ops_test.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
}

func compileChunkForLogicalOps(t *testing.T, source string) *rt.Chunk {
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

func TestVMLogicalAndReturnValue(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		// Both truthy: returns right operand
		{"true && true", "let x = true && true;"},
		{"true && 5", "let x = true && 5;"},
		{"5 && true", "let x = 5 && true;"},
		{"5 && 3", "let x = 5 && 3;"},
		{`"hello" && "world"`, `let x = "hello" && "world";`},
		
		// Left falsy: returns left operand
		{"false && true", "let x = false && true;"},
		{"false && 5", "let x = false && 5;"},
		{"0 && true", "let x = 0 && true;"},
		{"0 && 5", "let x = 0 && 5;"},
		{`"" && "world"`, `let x = "" && "world";`},
		{"null && true", "let x = null && true;"},
		
		// Complex expressions
		{"5 && 3 && 7", "let x = 5 && 3 && 7;"},
		{"5 && 0 && 7", "let x = 5 && 0 && 7;"},
		{"false && 0 && 7", "let x = false && 0 && 7;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoErrorForLogicalOps(t, tt.source)
		})
	}
}

func TestVMLogicalOrReturnValue(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		// Left truthy: returns left operand
		{"true || false", "let x = true || false;"},
		{"5 || 0", "let x = 5 || 0;"},
		{`"hello" || ""`, `let x = "hello" || "";`},
		{"true || 5", "let x = true || 5;"},
		
		// Left falsy: returns right operand
		{"false || true", "let x = false || true;"},
		{"false || 5", "let x = false || 5;"},
		{"0 || 5", "let x = 0 || 5;"},
		{`"" || "world"`, `let x = "" || "world";`},
		{"0 || false", "let x = 0 || false;"},
		{"null || 42", "let x = null || 42;"},
		
		// Complex expressions
		{"false || 0 || 7", "let x = false || 0 || 7;"},
		{`0 || "" || "hi"`, `let x = 0 || "" || "hi";`},
		{"true || 5 || 7", "let x = true || 5 || 7;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoErrorForLogicalOps(t, tt.source)
		})
	}
}

func TestVMLogicalOperatorWithVariables(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"and with variables", "let a = 5; let b = 10; let x = a && b;"},
		{"or with variables", "let a = 0; let b = 10; let x = a || b;"},
		{"mixed and/or", "let a = 5; let b = 0; let c = 10; let x = a && b || c;"},
		{"default value pattern", `let config = {}; let x = config["timeout"] || 30;`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoErrorForLogicalOps(t, tt.source)
		})
	}
}

func TestVMLogicalOperatorInConditions(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"if with &&", `
			let x = 5 && 10;
			if (x) {
				let result = 1;
			} else {
				let result = 0;
			}
		`},
		{"if with ||", `
			let x = 0 || "";
			if (!x) {
				let result = 1;
			} else {
				let result = 0;
			}
		`},
		{"while with &&", `
			let i = 0;
			let limit = 5;
			while (i < limit && limit > 0) {
				i = i + 1;
			}
		`},
		{"complex condition", `
			let a = 5;
			let b = 10;
			let c = 0;
			let result = (a && b) || c;
		`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoErrorForLogicalOps(t, tt.source)
		})
	}
}

func TestVMLogicalOperatorEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"nested &&", "let x = (1 && 2) && (3 && 4);"},
		{"nested ||", `let x = (0 || "") || (false || 5);`},
		{"mixed precedence", "let x = 1 || 2 && 3;"},
		{"with comparison", "let x = (5 > 3) && 10;"},
		{"with equality", "let x = (5 == 5) || 0;"},
		{"triple && chain", "let a = 1; let b = 2; let c = 3; let x = a && b && c;"},
		{"triple || chain", `let a = 0; let b = ""; let c = 42; let x = a || b || c;`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoErrorForLogicalOps(t, tt.source)
		})
	}
}
