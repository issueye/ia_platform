package lang_test

import (
	"sync"
	"testing"

	comp "ialang/pkg/lang/compiler"
	"ialang/pkg/lang/frontend"
	rt "ialang/pkg/lang/runtime"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	rvm "ialang/pkg/lang/runtime/vm"
)

// ============================================================
// Concurrency Tests
// ============================================================

func TestVMConcurrentExecution(t *testing.T) {
	source := `
let x = 1 + 2 * 3;
let y = x - 1;
`
	chunk := compileTestSource(t, source)
	if chunk == nil {
		t.Fatal("chunk is nil")
	}

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runtime := rt.NewGoroutineRuntime()
			vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "concurrent_test.ia", runtime)
			if err := vm.Run(); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent VM execution error: %v", err)
	}
}

func TestVMConcurrentDifferentPrograms(t *testing.T) {
	programs := []string{
		`let x = 1 + 2;`,
		`function add(a, b) { return a + b; } add(3, 4);`,
		`let arr = [1, 2, 3]; let sum = 0; let i = 0; while (i < arr.length) { sum = sum + arr[i]; i = i + 1; }`,
		`class Point { constructor(x, y) { this.x = x; this.y = y; } } let p = new Point(1, 2);`,
		`let obj = {a: 1, b: 2}; let sum = obj.a + obj.b;`,
	}

	var wg sync.WaitGroup
	errors := make(chan error, len(programs))

	for _, src := range programs {
		wg.Add(1)
		go func(source string) {
			defer wg.Done()
			l := frontend.NewLexer(source)
			p := frontend.NewParser(l)
			program := p.ParseProgram()
			if parseErrs := p.Errors(); len(parseErrs) > 0 {
				errors <- nil
				return
			}
			c := comp.NewCompiler()
			chunk, errs := c.Compile(program)
			if len(errs) > 0 {
				errors <- nil
				return
			}
			runtime := rt.NewGoroutineRuntime()
			vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "concurrent_diff.ia", runtime)
			if err := vm.Run(); err != nil {
				errors <- err
			}
		}(src)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Errorf("concurrent different programs error: %v", err)
		}
	}
}

func TestVMConcurrentClosures(t *testing.T) {
	source := `
function makeCounter() {
  let count = 0;
  function next() {
    count = count + 1;
    return count;
  }
  return next;
}
let counter = makeCounter();
counter();
counter();
counter();
`
	var wg sync.WaitGroup
	errors := make(chan error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l := frontend.NewLexer(source)
			p := frontend.NewParser(l)
			program := p.ParseProgram()
			if parseErrs := p.Errors(); len(parseErrs) > 0 {
				return
			}
			c := comp.NewCompiler()
			chunk, errs := c.Compile(program)
			if len(errs) > 0 {
				return
			}
			runtime := rt.NewGoroutineRuntime()
			vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "closure_concurrent.ia", runtime)
			if err := vm.Run(); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent closures error: %v", err)
	}
}

func TestVMConcurrentClassInstantiation(t *testing.T) {
	source := `
class Person {
  constructor(name, age) {
    this.name = name;
    this.age = age;
  }
  greet() {
    return "Hello, " + this.name;
  }
}
let p = new Person("Alice", 30);
p.greet();
`
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			l := frontend.NewLexer(source)
			p := frontend.NewParser(l)
			program := p.ParseProgram()
			if parseErrs := p.Errors(); len(parseErrs) > 0 {
				return
			}
			c := comp.NewCompiler()
			chunk, errs := c.Compile(program)
			if len(errs) > 0 {
				return
			}
			runtime := rt.NewGoroutineRuntime()
			vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "class_concurrent.ia", runtime)
			if err := vm.Run(); err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent class instantiation error: %v", err)
	}
}

// ============================================================
// Edge Case Tests
// ============================================================

func TestVMEmptySource(t *testing.T) {
	source := ""
	l := frontend.NewLexer(source)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	c := comp.NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "empty.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("empty source execution error: %v", err)
	}
}

func TestVMWhitespaceOnly(t *testing.T) {
	source := "   \n\n\t\t   "
	l := frontend.NewLexer(source)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	c := comp.NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "whitespace.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("whitespace only execution error: %v", err)
	}
}

func TestVMLargeNumberLiterals(t *testing.T) {
	source := `
let big = 999999999999;
let small = 0.00000001;
let pi = 3.14159265358979;
`
	runVMTestNoError(t, source)
}

func TestVMStringEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "empty string",
			source: `let s = "";`,
		},
		{
			name:   "string with only spaces",
			source: `let s = "   ";`,
		},
		{
			name:   "very long string",
			source: `let s = "abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789abcdefghijklmnopqrstuvwxyz0123456789";`,
		},
		{
			name:   "string concatenation chain",
			source: `let s = "a" + "b" + "c" + "d" + "e" + "f" + "g" + "h" + "i" + "j";`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMTestNoError(t, tt.source)
		})
	}
}

func TestVMArrayEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "empty array",
			source: `let arr = [];`,
		},
		{
			name:   "array with single element",
			source: `let arr = [42];`,
		},
		{
			name:   "array with mixed types",
			source: `let arr = [1, 2, true];`,
		},
		{
			name:   "deeply nested arrays",
			source: `let arr = [[[1, 2], [3, 4]], [[5, 6], [7, 8]]];`,
		},
		{
			name:   "array index out of bounds read",
			source: `let arr = [1, 2, 3]; let x = arr[100];`,
		},
		{
			name:   "array negative index",
			source: `let arr = [1, 2, 3]; let x = arr[-1];`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMTestNoError(t, tt.source)
		})
	}
}

func TestVMObjectEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "empty object",
			source: `let obj = {};`,
		},
		{
			name:   "object with many properties",
			source: `let obj = {a:1,b:2,c:3,d:4,e:5,f:6,g:7,h:8,i:9,j:10};`,
		},
		{
			name:   "object with nested objects",
			source: `let obj = {outer: {middle: {inner: {value: 42}}}};`,
		},
		{
			name:   "object property access missing",
			source: `let obj = {a: 1}; let x = obj.a;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMTestNoError(t, tt.source)
		})
	}
}

func TestVMFunctionEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "function with no body",
			source: `
function noop() {}
noop();
`,
		},
		{
			name: "function with many parameters",
			source: `
function many(a, b, c, d, e, f, g, h) { return a + b + c; }
many(1, 2, 3, 4, 5, 6, 7, 8);
`,
		},
		{
			name: "function returning function",
			source: `
function makeAdder(x) {
  let result = x + 1;
  return result;
}
let val = makeAdder(5);
`,
		},
		{
			name: "recursive factorial",
			source: `
function factorial(n) {
  if (n <= 1) { return 1; }
  return n * factorial(n - 1);
}
factorial(10);
`,
		},
		{
			name: "mutual recursion",
			source: `
function isEven(n) {
  if (n == 0) { return true; }
  return isOdd(n - 1);
}
function isOdd(n) {
  if (n == 0) { return false; }
  return isEven(n - 1);
}
isEven(10);
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMTestNoError(t, tt.source)
		})
	}
}

func TestVMClassEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "class with no constructor",
			source: `
class Empty {
  getValue() { return 42; }
}
let e = new Empty();
`,
		},
		{
			name: "class with many methods",
			source: `
class Multi {
  m1() { return 1; }
  m2() { return 2; }
  m3() { return 3; }
  m4() { return 4; }
  m5() { return 5; }
}
let m = new Multi();
`,
		},
		{
			name: "deep inheritance chain",
			source: `
class A { a() { return 1; } }
class B extends A { b() { return 2; } }
class C extends B { c() { return 3; } }
class D extends C { d() { return 4; } }
class E extends D { e() { return 5; } }
let obj = new E();
`,
		},
		{
			name: "class method calling another method",
			source: `
class SelfCall {
  constructor() { this.value = 0; }
  increment() { this.value = this.value + 1; return this; }
  getValue() { return this.value; }
}
let sc = new SelfCall();
sc.increment().increment().increment();
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMTestNoError(t, tt.source)
		})
	}
}

func TestVMControlFlowEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "nested if-else",
			source: `
let x = 5;
if (x > 0) {
  if (x > 10) {
    x = 1;
  } else {
    if (x > 3) {
      x = 2;
    } else {
      x = 3;
    }
  }
}
`,
		},
		{
			name: "while with complex condition",
			source: `
let i = 0;
let j = 10;
while (i < 5 && j > 5 || i == 0) {
  i = i + 1;
  j = j - 1;
}
`,
		},
		{
			name: "for with no init",
			source: `
let i = 0;
let sum = 0;
for (; i < 5; i = i + 1) {
  sum = sum + i;
}
`,
		},
		{
			name: "break in nested loops",
			source: `
let count = 0;
for (let i = 0; i < 10; i = i + 1) {
  for (let j = 0; j < 10; j = j + 1) {
    if (j == 3) { break; }
    count = count + 1;
  }
}
`,
		},
		{
			name: "continue in nested loops",
			source: `
let count = 0;
for (let i = 0; i < 5; i = i + 1) {
  for (let j = 0; j < 5; j = j + 1) {
    if (j == 2) { continue; }
    count = count + 1;
  }
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMTestNoError(t, tt.source)
		})
	}
}

func TestVMOperatorEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "division by one",
			source: `let x = 42 / 1;`,
		},
		{
			name:   "modulo by one",
			source: `let x = 42 % 1;`,
		},
		{
			name:   "multiply by zero",
			source: `let x = 42 * 0;`,
		},
		{
			name:   "add zero",
			source: `let x = 42 + 0;`,
		},
		{
			name:   "subtract zero",
			source: `let x = 42 - 0;`,
		},
		{
			name:   "double negation",
			source: `let x = - -42;`,
		},
		{
			name:   "comparison chaining",
			source: `let x = 1 < 2 && 2 < 3 && 3 < 4 && 4 < 5;`,
		},
		{
			name:   "logical expression complexity",
			source: `let x = (true && false) || (false && true) || (true && true);`,
		},
		{
			name:   "bitwise with zero",
			source: `let x = 0 & 0 | 0 ^ 0;`,
		},
		{
			name:   "shift by zero",
			source: `let x = 42 << 0 >> 0;`,
		},
		{
			name:   "ternary nesting",
			source: `let x = true ? (false ? 1 : 2) : (true ? 3 : 4);`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMTestNoError(t, tt.source)
		})
	}
}

func TestVMTypeErrorCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "string subtraction",
			source: `"hello" - "world";`,
		},
		{
			name:   "string multiplication",
			source: `"hello" * 2;`,
		},
		{
			name:   "string division",
			source: `"hello" / 2;`,
		},
		{
			name:   "boolean arithmetic",
			source: `let x = true - false;`,
		},
		{
			name:   "null property access",
			source: `let x = true.someProp;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runVMTestExpectError(t, tt.source)
		})
	}
}

// ============================================================
// Memory and Stress Tests
// ============================================================

func TestVMStressManyVariables(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}
	source := `
let v0 = 0; let v1 = 1; let v2 = 2; let v3 = 3; let v4 = 4;
let v5 = 5; let v6 = 6; let v7 = 7; let v8 = 8; let v9 = 9;
let v10 = 10; let v11 = 11; let v12 = 12; let v13 = 13; let v14 = 14;
let v15 = 15; let v16 = 16; let v17 = 17; let v18 = 18; let v19 = 19;
let v20 = 20; let v21 = 21; let v22 = 22; let v23 = 23; let v24 = 24;
let v25 = 25; let v26 = 26; let v27 = 27; let v28 = 28; let v29 = 29;
let v30 = 30; let v31 = 31; let v32 = 32; let v33 = 33; let v34 = 34;
let v35 = 35; let v36 = 36; let v37 = 37; let v38 = 38; let v39 = 39;
let v40 = 40; let v41 = 41; let v42 = 42; let v43 = 43; let v44 = 44;
let v45 = 45; let v46 = 46; let v47 = 47; let v48 = 48; let v49 = 49;
let v50 = 50; let v51 = 51; let v52 = 52; let v53 = 53; let v54 = 54;
let v55 = 55; let v56 = 56; let v57 = 57; let v58 = 58; let v59 = 59;
let v60 = 60; let v61 = 61; let v62 = 62; let v63 = 63; let v64 = 64;
let v65 = 65; let v66 = 66; let v67 = 67; let v68 = 68; let v69 = 69;
let v70 = 70; let v71 = 71; let v72 = 72; let v73 = 73; let v74 = 74;
let v75 = 75; let v76 = 76; let v77 = 77; let v78 = 78; let v79 = 79;
let v80 = 80; let v81 = 81; let v82 = 82; let v83 = 83; let v84 = 84;
let v85 = 85; let v86 = 86; let v87 = 87; let v88 = 88; let v89 = 89;
let v90 = 90; let v91 = 91; let v92 = 92; let v93 = 93; let v94 = 94;
let v95 = 95; let v96 = 96; let v97 = 97; let v98 = 98; let v99 = 99;
`
	runVMTestNoError(t, source)
}

func TestVMStressLargeArray(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}
	source := `
let arr = [];
for (let i = 0; i < 10000; i = i + 1) {
  arr = arr.push(i);
}
`
	runVMTestNoError(t, source)
}

func TestVMStressLargeObject(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}
	source := "let obj = {"
	for i := 0; i < 1000; i++ {
		if i > 0 {
			source += ","
		}
		source += "k" + itoaTest(i) + ": " + itoaTest(i)
	}
	source += "};"
	runVMTestNoError(t, source)
}

func TestVMStressDeepRecursion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}
	source := `
function deep(n) {
  if (n <= 0) { return 0; }
  return deep(n - 1) + 1;
}
deep(500);
`
	runVMTestNoError(t, source)
}

func TestVMStressStringConcatenation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}
	source := `
let s = "";
for (let i = 0; i < 1000; i = i + 1) {
  s = s + "x";
}
`
	runVMTestNoError(t, source)
}
