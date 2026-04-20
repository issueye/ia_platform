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
// Missing property access returns null (JS-style)
// ============================================================

func compileChunkForPropertyTest(t *testing.T, source string) *rt.Chunk {
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

func runVMExpectNoErrorForPropertyTest(t *testing.T, source string) {
	t.Helper()
	chunk := compileChunkForPropertyTest(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "property_test.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
}

func TestVMMissingPropertyReturnsNull(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		// Object missing property
		{"object missing property", `
			let obj = {"a": 1};
			let x = obj["b"];
			if (x == null) {
				let result = 1;
			}
		`},
		{"object missing property with || default", `
			let obj = {"a": 1};
			let x = obj["b"] || 42;
			if (x == 42) {
				let result = 1;
			}
		`},
		{"nested object missing property", `
			let obj = {"a": {"b": 2}};
			let x = obj["a"]["c"];
			if (x == null) {
				let result = 1;
			}
		`},
		
		// Instance missing property
		{"instance missing property", `
			class Foo {
				constructor() {
					this.existing = 1;
				}
			}
			let obj = new Foo();
			let x = obj.missing;
			if (x == null) {
				let result = 1;
			}
		`},
		{"instance missing property with || default", `
			class Foo {
				constructor() {
					this.existing = 1;
				}
			}
			let obj = new Foo();
			let x = obj.missing || 99;
			if (x == 99) {
				let result = 1;
			}
		`},
		
		// String missing property
		{"string missing property", `
			let s = "hello";
			let x = s.nonExistentMethod;
			if (x == null) {
				let result = 1;
			}
		`},
		
		// Array missing property
		{"array missing property", `
			let arr = [1, 2, 3];
			let x = arr.nonExistentMethod;
			if (x == null) {
				let result = 1;
			}
		`},
		{"array has length", `
			let arr = [1, 2, 3];
			let x = arr.length;
			if (x == 3) {
				let result = 1;
			}
		`},
		
		// Real-world patterns
		{"config default pattern", `
			let config = {};
			let timeout = config["timeout"] || 30;
			if (timeout == 30) {
				let result = 1;
			}
		`},
		{"optional chaining simulation", `
			let obj = null;
			let x = obj || {};
			let y = x["missing"] || "default";
			if (y == "default") {
				let result = 1;
			}
		`},
		{"deep optional chaining", `
			let data = {"user": {"name": "Alice"}};
			let name = data["user"]["name"] || "Unknown";
			if (name == "Alice") {
				let result = 1;
			}
		`},
		{"deep optional chaining with missing", `
			let data = {"user": {}};
			let name = data["user"]["name"] || "Unknown";
			if (name == "Unknown") {
				let result = 1;
			}
		`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoErrorForPropertyTest(t, tt.source)
		})
	}
}

func TestVMPropertyAccessStillWorks(t *testing.T) {
	// Ensure that existing properties still work correctly
	tests := []struct {
		name   string
		source string
	}{
		{"object existing property", `
			let obj = {"a": 1, "b": 2};
			let x = obj["a"];
			if (x == 1) {
				let result = 1;
			}
		`},
		{"instance existing property", `
			class Foo {
				constructor() {
					this.value = 42;
				}
			}
			let obj = new Foo();
			let x = obj.value;
			if (x == 42) {
				let result = 1;
			}
		`},
		{"array length", `
			let arr = [1, 2, 3, 4];
			let x = arr.length;
			if (x == 4) {
				let result = 1;
			}
		`},
		{"string methods", `
			let s = "hello";
			let x = s.length;
			// String prototype should have methods
		`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoErrorForPropertyTest(t, tt.source)
		})
	}
}
