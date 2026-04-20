package lang_test

import (
	"testing"

	rt "ialang/pkg/lang/runtime"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	rvm "ialang/pkg/lang/runtime/vm"
)

// ============================================================
// Optional Chaining Tests
// ============================================================

func TestOptionalChaining_NullBase(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "null?.prop returns null",
			source: "let x = null?.prop;",
		},
		{
			name:   "null?.[0] returns null",
			source: "let x = null?.[0];",
		},
		{
			name:   "null?.() returns null",
			source: "let x = null?.();",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMTestNoError(t, tt.source)
		})
	}
}

func TestOptionalChaining_UndefinedBase(t *testing.T) {
	// Test with undefined variable (will be null in the runtime)
	runVMTestNoError(t, "let obj = null; let x = obj?.prop;")
}

func TestOptionalChaining_ValidBase(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "obj?.prop returns obj.prop when obj is not null",
			source: `let obj = {"name": "test"}; let x = obj?.name;`,
		},
		{
			name:   "arr?.[index] returns arr[index] when arr is not null",
			source: "let arr = [1, 2, 3]; let x = arr?.[1];",
		},
		{
			name:   "fn?.() calls fn when fn is not null",
			source: "let fn = function() { return 42; }; let x = fn?.();",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMTestNoError(t, tt.source)
		})
	}
}

func TestOptionalChaining_Chained(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "chained optional access with null in middle",
			source: "let obj = null; let x = obj?.a?.b;",
		},
		{
			name:   "chained optional access all valid",
			source: "let obj = {a: {b: 123}}; let x = obj?.a?.b;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMTestNoError(t, tt.source)
		})
	}
}

func TestOptionalChaining_WithMethod(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "optional method call on null",
			source: "let obj = null; let x = obj?.method();",
		},
		{
			name:   "optional method call on valid object",
			source: `let obj = {"getValue": function() { return 99; }}; let x = obj?.getValue();`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Debug: print what's being compiled
			t.Logf("Compiling: %s", tt.source)
			runVMTestNoError(t, tt.source)
		})
	}
}

func TestOptionalChaining_MixedWithNullish(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "optional chaining with nullish coalescing",
			source: `let obj = null; let x = obj?.prop ?? "default";`,
		},
		{
			name:   "optional chaining returns value for nullish coalescing",
			source: `let obj = {"name": "test"}; let x = obj?.name ?? "default";`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMTestNoError(t, tt.source)
		})
	}
}

func TestOptionalChaining_Array(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "optional index on array",
			source: "let arr = [10, 20, 30]; let x = arr?.[2];",
		},
		{
			name:   "optional index on null array",
			source: "let arr = null; let x = arr?.[0];",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMTestNoError(t, tt.source)
		})
	}
}

func TestOptionalChaining_FunctionCall(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "optional function call with arguments",
			source: "let add = function(a, b) { return a + b; }; let x = add?.(3, 4);",
		},
		{
			name:   "optional function call on null with arguments",
			source: "let fn = null; let x = fn?.(1, 2, 3);",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMTestNoError(t, tt.source)
		})
	}
}

// Test to verify the actual runtime values
func TestOptionalChaining_RuntimeValues(t *testing.T) {
	// Test that null?.prop actually returns null
	source := `
let x = null?.prop;
if (x == null) {
  let y = 1;
} else {
  throw "expected null";
}
`
	runVMTestNoError(t, source)

	// Test that obj?.prop returns the actual property value
	source2 := `
let obj = {"value": 42};
let x = obj?.value;
if (x == 42) {
  let y = 1;
} else {
  throw "expected 42";
}
`
	runVMTestNoError(t, source2)
}

// Benchmark for optional chaining performance
func BenchmarkOptionalChaining(b *testing.B) {
	benchmarks := []struct {
		name   string
		source string
	}{
		{
			name:   "null_base_property",
			source: "for (let i = 0; i < 1000; i = i + 1) { let x = null?.prop; }",
		},
		{
			name:   "valid_base_property",
			source: "let obj = {value: 1}; for (let i = 0; i < 1000; i = i + 1) { let x = obj?.value; }",
		},
		{
			name:   "null_base_index",
			source: "for (let i = 0; i < 1000; i = i + 1) { let x = null?.[0]; }",
		},
		{
			name:   "valid_base_index",
			source: "let arr = [1, 2, 3]; for (let i = 0; i < 1000; i = i + 1) { let x = arr?.[0]; }",
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				chunk := compileTestSource(b, bm.source)
				runtime := rt.NewGoroutineRuntime()
				vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "bench.ia", runtime)
				if err := vm.Run(); err != nil {
					b.Fatalf("vm.Run() error: %v", err)
				}
			}
		})
	}
}
