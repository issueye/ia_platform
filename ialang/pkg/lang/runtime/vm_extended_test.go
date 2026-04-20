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
// Arithmetic operation tests
// ============================================================

func TestVMArithmeticOperations(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"addition", "let x = 1 + 2;"},
		{"subtraction", "let x = 10 - 3;"},
		{"multiplication", "let x = 4 * 5;"},
		{"division", "let x = 20 / 4;"},
		{"modulo", "let x = 17 % 5;"},
		{"negative", "let x = -5;"},
		{"compound +=", "let x = 10; x += 5;"},
		{"compound -=", "let x = 10; x -= 3;"},
		{"compound *=", "let x = 2; x *= 6;"},
		{"compound /=", "let x = 20; x /= 4;"},
		{"compound %=", "let x = 17; x %= 5;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

func TestVMDivisionByZero(t *testing.T) {
	source := "let x = 1 / 0;"
	chunk := compileChunk(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "div_zero_test.ia", runtime)
	err := vm.Run()
	if err == nil {
		t.Error("expected error for division by zero, got nil")
	}
}

func TestVMModuloByZero(t *testing.T) {
	source := "let x = 1 % 0;"
	chunk := compileChunk(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "mod_zero_test.ia", runtime)
	err := vm.Run()
	if err == nil {
		t.Error("expected error for modulo by zero, got nil")
	}
}

// ============================================================
// Comparison and logical tests
// ============================================================

func TestVMComparisonOperations(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"equal true", "let x = 5 == 5;"},
		{"equal false", "let x = 5 == 3;"},
		{"not equal true", "let x = 5 != 3;"},
		{"not equal false", "let x = 5 != 5;"},
		{"less than true", "let x = 3 < 5;"},
		{"less than false", "let x = 5 < 3;"},
		{"greater than true", "let x = 5 > 3;"},
		{"greater than false", "let x = 3 > 5;"},
		{"less or equal true", "let x = 5 <= 5;"},
		{"less or equal false", "let x = 6 <= 5;"},
		{"greater or equal true", "let x = 5 >= 5;"},
		{"greater or equal false", "let x = 4 >= 5;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

func TestVMLogicalOperations(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"and true", "let x = true && true;"},
		{"and false", "let x = true && false;"},
		{"or true", "let x = true || false;"},
		{"or false", "let x = false || false;"},
		{"not true", "let x = !true;"},
		{"not false", "let x = !false;"},
		{"short-circuit and", "let x = false && true;"},
		{"short-circuit or", "let x = true || false;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

// ============================================================
// Bitwise operation tests
// ============================================================

func TestVMBitwiseOperations(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"bitwise and", "let x = 5 & 3;"},
		{"bitwise or", "let x = 5 | 3;"},
		{"bitwise xor", "let x = 5 ^ 3;"},
		{"left shift", "let x = 5 << 1;"},
		{"right shift", "let x = 10 >> 1;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

// ============================================================
// Control flow tests
// ============================================================

func TestVMIfStatement(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "if then",
			source: `
let x = 0;
if (true) { x = 42; }
`,
		},
		{
			name: "if else",
			source: `
let x = 0;
if (false) { x = 1; } else { x = 2; }
`,
		},
		{
			name: "nested if",
			source: `
let x = 0;
if (true) {
  if (true) { x = 99; }
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

func TestVMWhileLoop(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "simple while",
			source: `
let i = 0;
let sum = 0;
while (i < 5) {
  sum = sum + i;
  i = i + 1;
}
`,
		},
		{
			name: "while with break",
			source: `
let i = 0;
while (true) {
  if (i >= 3) { break; }
  i = i + 1;
}
`,
		},
		{
			name: "while with continue",
			source: `
let i = 0;
let sum = 0;
while (i < 5) {
  i = i + 1;
  if (i == 3) { continue; }
  sum = sum + i;
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

func TestVMForLoop(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "classic for loop",
			source: `
let sum = 0;
for (let i = 0; i < 5; i = i + 1) {
  sum = sum + i;
}
`,
		},
		{
			name: "for with break",
			source: `
let sum = 0;
for (let i = 0; i < 10; i = i + 1) {
  if (i >= 3) { break; }
  sum = sum + i;
}
`,
		},
		{
			name: "for with continue",
			source: `
let sum = 0;
for (let i = 0; i < 5; i = i + 1) {
  if (i == 2) { continue; }
  sum = sum + i;
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

// ============================================================
// Function call tests
// ============================================================

func TestVMFunctionCalls(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "simple function",
			source: `
function add(a, b) { return a + b; }
add(3, 4);
`,
		},
		{
			name: "nested function calls",
			source: `
function double(x) { return x * 2; }
function square(x) { return x * x; }
double(square(3));
`,
		},
		{
			name: "recursive function",
			source: `
function factorial(n) {
  if (n <= 1) { return 1; }
  return n * factorial(n - 1);
}
factorial(5);
`,
		},
		{
			name: "function with no return",
			source: `
function noop() { let x = 1; }
noop();
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

// ============================================================
// Class and inheritance tests
// ============================================================

func TestVMClassAndInheritance(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "simple class",
			source: `
class Point {
  constructor(x, y) {
    this.x = x;
    this.y = y;
  }
  sum() {
    let s = this.x + this.y;
  }
}
let p = new Point(3, 4);
`,
		},
		{
			name: "class inheritance",
			source: `
class Animal {
  constructor(name) {
    this.name = name;
  }
  speak() {
    let s = 1;
  }
}

class Dog extends Animal {
  bark() {
    let b = 2;
  }
}

let dog = new Dog("Buddy");
`,
		},
		{
			name: "class method chaining",
			source: `
class Counter {
  constructor() {
    this.value = 0;
  }
  increment() {
    this.value = this.value + 1;
    return this;
  }
  getValue() {
    let v = this.value;
  }
}

let c = new Counter();
c.increment().increment().increment();
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

// ============================================================
// Array and Object tests
// ============================================================

func TestVMArrayOperations(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "array creation and indexing",
			source: `
let arr = [10, 20, 30];
let x = arr[1];
`,
		},
		{
			name: "array length",
			source: `
let arr = [1, 2, 3, 4, 5];
let len = arr.length;
`,
		},
		{
			name: "array sum",
			source: `
let arr = [1, 2, 3, 4, 5];
let sum = 0;
let i = 0;
while (i < arr.length) {
  sum = sum + arr[i];
  i = i + 1;
}
`,
		},
		{
			name: "nested arrays",
			source: `
let matrix = [[1, 2], [3, 4]];
let x = matrix[0][0];
let y = matrix[1][1];
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

func TestVMObjectOperations(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "object property access",
			source: `
let obj = {x: 10, y: 20};
let z = obj.x + obj.y;
`,
		},
		{
			name: "object property assignment",
			source: `
let obj = {};
obj.x = 42;
`,
		},
		{
			name: "nested objects",
			source: `
let obj = {outer: {inner: 99}};
let x = obj.outer.inner;
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

// ============================================================
// Async/Promise tests
// ============================================================

func TestVMAsyncFunctions(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "simple async function",
			source: `
async function getValue() {
  return 42;
}
let p = getValue();
`,
		},
		{
			name: "async function with await",
			source: `
async function double(x) {
  return x * 2;
}

async function main() {
  let result = await double(21);
  return result;
}

let p = main();
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

// ============================================================
// Exception handling tests
// ============================================================

func TestVMTryCatch(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "catch exception",
			source: `
let caught = false;
try {
  throw "error";
} catch (e) {
  caught = true;
}
`,
		},
		{
			name: "try without exception",
			source: `
let x = 0;
try {
  x = 42;
} catch (e) {
  x = -1;
}
`,
		},
		{
			name: "finally block",
			source: `
let x = 0;
try {
  x = 1;
} finally {
  x = x + 10;
}
`,
		},
		{
			name: "catch and finally",
			source: `
let x = 0;
try {
  throw "error";
} catch (e) {
  x = 5;
} finally {
  x = x + 10;
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

// ============================================================
// String operation tests
// ============================================================

func TestVMStringOperations(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "string concatenation",
			source: `
let s = "hello" + " " + "world";
`,
		},
		{
			name: "string indexing",
			source: `
let s = "hello";
let c = s[0];
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

// ============================================================
// Ternary expression tests
// ============================================================

func TestVMTernaryExpression(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "ternary true",
			source: `
let x = true ? 10 : 20;
`,
		},
		{
			name: "ternary false",
			source: `
let x = false ? 10 : 20;
`,
		},
		{
			name: "nested ternary",
			source: `
let x = 5;
let y = x > 10 ? 1 : x > 0 ? 2 : 3;
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectNoError(t, tt.source)
		})
	}
}

// ============================================================
// Boundary condition tests
// ============================================================

func TestVMUndefinedVariable(t *testing.T) {
	source := "let x = undefinedVar;"
	runVMExpectError(t, source)
}

func TestVMArrayTypeIndex(t *testing.T) {
	source := `
let arr = [1, 2, 3];
let x = arr[10];
`
	runVMExpectNoError(t, source) // Out of bounds returns nil, not error
}

func TestVMStringIndexBounds(t *testing.T) {
	source := `
let s = "hi";
let c = s[10];
`
	runVMExpectNoError(t, source) // Out of bounds returns nil, not error
}

func TestVMTypeErrors(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "subtract strings",
			source: `
let x = "hello" - "world";
`,
		},
		{
			name: "multiply strings",
			source: `
let x = "hello" * 2;
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMExpectError(t, tt.source)
		})
	}
}

// ============================================================
// Module import/export tests
// ============================================================

func TestVMModuleImportExport(t *testing.T) {
	source := `
import { add } from "math_module";
let result = add(3, 4);
`

	// Create a mock math module
	runtime := rt.NewGoroutineRuntime()
	modules := map[string]rt.Value{
		"math_module": rt.Object{
			"add": rt.NativeFunction(func(args []rt.Value) (rt.Value, error) {
				if len(args) != 2 {
					return nil, nil
				}
				a, _ := args[0].(float64)
				b, _ := args[1].(float64)
				return a + b, nil
			}),
		},
	}

	chunk := compileChunk(t, source)
	vm := rvm.NewVM(chunk, modules, nil, "module_test.ia", runtime)
	err := vm.Run()
	if err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
}

func TestVMTemplateStringInterpolation(t *testing.T) {
	source := `
let name = "ialang";
let version = 2;
let msg = ` + "`" + `hello ${name} v${version}` + "`" + `;
if (msg != "hello ialang v2") {
  throw "bad-template-result";
}
`
	runVMExpectNoError(t, source)
}

func TestVMFunctionRestParams(t *testing.T) {
	source := `
function sum(...nums) {
  let i = 0;
  let total = 0;
  while (i < nums.length) {
    total = total + nums[i];
    i = i + 1;
  }
  return total;
}

function headPlusCount(head, ...rest) {
  return head + rest.length;
}

let a = sum(1, 2, 3, 4);
let b = headPlusCount(10, "x", "y", "z");
if (a != 10 || b != 13) {
  throw "bad-rest-result";
}
`
	runVMExpectNoError(t, source)
}

func TestVMExportAlias(t *testing.T) {
	source := `
let value = 42;
export { value as answer };
`
	chunk := compileChunk(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "export_alias_test.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
	exports := vm.Exports()
	got, ok := exports["answer"]
	if !ok {
		t.Fatal("expected export alias answer to exist")
	}
	if got != float64(42) {
		t.Fatalf("export answer = %#v, want 42", got)
	}
}

func TestVMImportNamespace(t *testing.T) {
	source := `
import * as mathmod from "math_module";
let result = mathmod.add(5, 7);
if (result != 12) {
  throw "bad-namespace-import";
}
`
	runtime := rt.NewGoroutineRuntime()
	modules := map[string]rt.Value{
		"math_module": rt.Object{
			"add": rt.NativeFunction(func(args []rt.Value) (rt.Value, error) {
				if len(args) != 2 {
					return nil, nil
				}
				a, _ := args[0].(float64)
				b, _ := args[1].(float64)
				return a + b, nil
			}),
		},
	}

	chunk := compileChunk(t, source)
	vm := rvm.NewVM(chunk, modules, nil, "import_namespace_test.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
}

func TestVMExportDefault(t *testing.T) {
	source := `
let value = 99;
export default value + 1;
`
	chunk := compileChunk(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "export_default_test.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
	exports := vm.Exports()
	got, ok := exports["default"]
	if !ok {
		t.Fatal("expected default export to exist")
	}
	if got != float64(100) {
		t.Fatalf("default export = %#v, want 100", got)
	}
}

func TestVMDestructuringLet(t *testing.T) {
	source := `
let arr = [2, 3];
let [a, b] = arr;
let obj = {x: 4, y: 5};
let {x, y: z} = obj;
if (a != 2 || b != 3 || x != 4 || z != 5) {
  throw "bad-destructure";
}
`
	runVMExpectNoError(t, source)
}

func TestVMDestructuringAssign(t *testing.T) {
	source := `
let arr = [2, 3];
let a = 0;
let b = 0;
[a, b] = arr;

let obj = {x: 4, y: 5};
let x = 0;
let z = 0;
({x, y: z} = obj);

if (a != 2 || b != 3 || x != 4 || z != 5) {
  throw "bad-destructure-assign";
}
`
	runVMExpectNoError(t, source)
}

func TestVMDynamicImportExpression(t *testing.T) {
	source := `
let mod = await import("math_module");
let result = mod.add(8, 9);
if (result != 17) {
  throw "bad-dynamic-import";
}
`
	runtime := rt.NewGoroutineRuntime()
	modules := map[string]rt.Value{
		"math_module": rt.Object{
			"add": rt.NativeFunction(func(args []rt.Value) (rt.Value, error) {
				if len(args) != 2 {
					return nil, nil
				}
				a, _ := args[0].(float64)
				b, _ := args[1].(float64)
				return a + b, nil
			}),
		},
	}

	chunk := compileChunk(t, source)
	vm := rvm.NewVM(chunk, modules, nil, "dynamic_import_test.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
}

func TestVMExportAll(t *testing.T) {
	source := `
export * from "dep_mod";
`
	runtime := rt.NewGoroutineRuntime()
	modules := map[string]rt.Value{
		"dep_mod": rt.Object{
			"a":       float64(1),
			"b":       float64(2),
			"default": float64(99),
		},
	}

	chunk := compileChunk(t, source)
	vm := rvm.NewVM(chunk, modules, nil, "export_all_test.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
	exports := vm.Exports()
	if exports["a"] != float64(1) || exports["b"] != float64(2) {
		t.Fatalf("unexpected export-all result: %#v", exports)
	}
	if _, exists := exports["default"]; exists {
		t.Fatalf("export * should skip default, got %#v", exports["default"])
	}
}

func TestVMClassPrivateField(t *testing.T) {
	source := `
class Counter {
  #value;
  constructor() { this.#value = 7; }
  inc() { this.#value += 1; return this.#value; }
}
let c = new Counter();
let n = c.inc();
if (n != 8) {
  throw "bad-private-field";
}
`
	runVMExpectNoError(t, source)
}

func TestVMClassPrivateFieldStrictAccess(t *testing.T) {
	source := `
class Counter {
  #value;
  constructor() { this.#value = 1; }
}
let c = new Counter();
let leaked = c.__private_value;
`
	runVMExpectError(t, source)
}

func TestVMExportDefaultClass(t *testing.T) {
	source := `
export default class Counter {
  constructor() { this.v = 2; }
}
`
	chunk := compileChunk(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "export_default_class_test.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
	exports := vm.Exports()
	def, ok := exports["default"]
	if !ok {
		t.Fatal("expected default export to exist")
	}
	if _, ok := def.(*rt.ClassValue); !ok {
		t.Fatalf("default export type = %T, want *ClassValue", def)
	}
}

func TestVMExportNamedClass(t *testing.T) {
	source := `
export class Counter {
  constructor() { this.v = 2; }
}
`
	chunk := compileChunk(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "export_named_class_test.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
	exports := vm.Exports()
	value, ok := exports["Counter"]
	if !ok {
		t.Fatal("expected named class export to exist")
	}
	classValue, ok := value.(*rt.ClassValue)
	if !ok {
		t.Fatalf("named class export type = %T, want *ClassValue", value)
	}
	if classValue.Name != "Counter" {
		t.Fatalf("named class export name = %q, want Counter", classValue.Name)
	}
}

// ============================================================
// Helper functions
// ============================================================

func runVMExpectNoError(t *testing.T, source string) {
	t.Helper()
	chunk := compileChunk(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "test.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
}

func runVMExpectError(t *testing.T, source string) {
	t.Helper()
	chunk := compileChunk(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "test.ia", runtime)
	if err := vm.Run(); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ============================================================
// Benchmark Tests
// ============================================================

func BenchmarkVMSimpleArithmetic(b *testing.B) {
	source := "let x = 1 + 2 * 3 - 4 / 2;"
	benchmarkVM(b, source)
}

func BenchmarkVMFunctionCall(b *testing.B) {
	source := `
function add(a, b) { return a + b; }
add(1, 2);
`
	benchmarkVM(b, source)
}

func BenchmarkVMRecursiveFunction(b *testing.B) {
	source := `
function fib(n) {
  if (n <= 1) { return n; }
  return fib(n - 1) + fib(n - 2);
}
fib(10);
`
	benchmarkVM(b, source)
}

func BenchmarkVMClassInstantiation(b *testing.B) {
	source := `
class Point {
  constructor(x, y) {
    this.x = x;
    this.y = y;
  }
}
let p = new Point(1, 2);
`
	benchmarkVM(b, source)
}

func BenchmarkVMLoop(b *testing.B) {
	source := `
let sum = 0;
for (let i = 0; i < 1000; i = i + 1) {
  sum = sum + i;
}
`
	benchmarkVM(b, source)
}

func BenchmarkVMArrayOperations(b *testing.B) {
	source := `
let arr = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10];
let sum = 0;
let i = 0;
while (i < arr.length) {
  sum = sum + arr[i];
  i = i + 1;
}
`
	benchmarkVM(b, source)
}

func benchmarkVM(b *testing.B, source string) {
	b.Helper()
	chunk := compileChunkBenchmark(source)
	runtime := rt.NewGoroutineRuntime()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "bench.ia", runtime)
		_ = vm.Run()
	}
}

func compileChunkBenchmark(source string) *rt.Chunk {
	l := frontend.NewLexer(source)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	c := comp.NewCompiler()
	chunk, _ := c.Compile(program)
	return chunk
}
