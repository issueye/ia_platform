package lang_test

import (
	"fmt"
	"testing"
	"time"

	comp "ialang/pkg/lang/compiler"
	"ialang/pkg/lang/frontend"
	rt "ialang/pkg/lang/runtime"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	rvm "ialang/pkg/lang/runtime/vm"
)

// ============================================================
// Comprehensive Performance Benchmark Suite
// ============================================================

// --- Lexer Performance ---

func BenchmarkLexerEmptyString(b *testing.B) {
	input := ""
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		l.NextToken()
	}
}

func BenchmarkLexerSingleToken(b *testing.B) {
	input := "x"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		for l.NextToken().Type != frontend.EOF {
		}
	}
}

func BenchmarkLexerWhitespaceHandling(b *testing.B) {
	input := "   \n\n\t\t   \r\n   "
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		l.NextToken()
	}
}

func BenchmarkLexerManyIdentifiers(b *testing.B) {
	input := generateIdentifiers(5000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		for l.NextToken().Type != frontend.EOF {
		}
	}
}

func BenchmarkLexerManyNumbers(b *testing.B) {
	input := generateNumbers(5000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		for l.NextToken().Type != frontend.EOF {
		}
	}
}

func BenchmarkLexerManyStrings(b *testing.B) {
	input := generateStrings(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		for l.NextToken().Type != frontend.EOF {
		}
	}
}

// --- Parser Performance ---

func BenchmarkParserEmptyInput(b *testing.B) {
	input := ""
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		p := frontend.NewParser(l)
		p.ParseProgram()
	}
}

func BenchmarkParserManyLetStatements(b *testing.B) {
	input := generateLetStatements(2000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		p := frontend.NewParser(l)
		p.ParseProgram()
	}
}

func BenchmarkParserDeepNesting(b *testing.B) {
	input := generateDeepNesting(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		p := frontend.NewParser(l)
		p.ParseProgram()
	}
}

func BenchmarkParserWideExpressions(b *testing.B) {
	input := generateWideExpression(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		p := frontend.NewParser(l)
		p.ParseProgram()
	}
}

func BenchmarkParserManyFunctions(b *testing.B) {
	input := generateManyFunctions(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		p := frontend.NewParser(l)
		p.ParseProgram()
	}
}

func BenchmarkParserManyClasses(b *testing.B) {
	input := generateManyClasses(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		p := frontend.NewParser(l)
		p.ParseProgram()
	}
}

func BenchmarkParserComplexClassWithMethods(b *testing.B) {
	input := generateComplexClass(50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		p := frontend.NewParser(l)
		p.ParseProgram()
	}
}

// --- Compiler Performance ---

func BenchmarkCompilerEmptyProgram(b *testing.B) {
	input := ""
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileTestSource(b, input)
	}
}

func BenchmarkCompilerManyVariables(b *testing.B) {
	input := generateLetStatements(2000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileTestSource(b, input)
	}
}

func BenchmarkCompilerDeeplyNestedFunctions(b *testing.B) {
	input := generateNestedFunctions(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileTestSource(b, input)
	}
}

func BenchmarkCompilerManyClosures(b *testing.B) {
	input := generateManyClosures(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileTestSource(b, input)
	}
}

func BenchmarkCompilerLargeClassHierarchy(b *testing.B) {
	input := generateClassHierarchy(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileTestSource(b, input)
	}
}

func BenchmarkCompilerAsyncFunctions(b *testing.B) {
	input := generateAsyncFunctions(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileTestSource(b, input)
	}
}

// --- VM Execution Performance ---

func BenchmarkVMEmptyProgram(b *testing.B) {
	source := ""
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMSimpleVariableAccess(b *testing.B) {
	source := `
let x = 42;
let y = x;
let z = y;
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMLargeLoop(b *testing.B) {
	source := `
let sum = 0;
for (let i = 0; i < 10000; i = i + 1) {
  sum = sum + 1;
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMNestedLoops(b *testing.B) {
	source := `
let count = 0;
for (let i = 0; i < 100; i = i + 1) {
  for (let j = 0; j < 100; j = j + 1) {
    count = count + 1;
  }
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMManyFunctionCalls(b *testing.B) {
	source := `
function identity(x) { return x; }
let result = 0;
for (let i = 0; i < 1000; i = i + 1) {
  result = identity(i);
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMDeepRecursion(b *testing.B) {
	source := `
function countdown(n) {
  if (n <= 0) { return 0; }
  return countdown(n - 1) + 1;
}
countdown(200);
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMTailRecursion(b *testing.B) {
	source := `
function factorial(n, acc) {
  if (n <= 1) { return acc; }
  return factorial(n - 1, acc * n);
}
factorial(50, 1);
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMArrayPush(b *testing.B) {
	source := `
let arr = [];
for (let i = 0; i < 1000; i = i + 1) {
  arr = arr.push(i);
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMArrayIteration(b *testing.B) {
	source := `
let arr = [0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99];
let sum = 0;
let i = 0;
while (i < arr.length) {
  sum = sum + arr[i];
  i = i + 1;
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMObjectPropertyAccess(b *testing.B) {
	source := `
let obj = {a:1,b:2,c:3,d:4,e:5,f:6,g:7,h:8,i:9,j:10};
let sum = 0;
for (let i = 0; i < 1000; i = i + 1) {
  sum = sum + obj.a + obj.b + obj.c + obj.d + obj.e;
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkPerfVMObjectCreation(b *testing.B) {
	source := `
for (let i = 0; i < 100; i = i + 1) {
  let obj = {x: i, y: i * 2, z: i * 3};
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMStringConcatenation(b *testing.B) {
	source := `
let s = "";
for (let i = 0; i < 100; i = i + 1) {
  s = s + "a";
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMStringOperations(b *testing.B) {
	source := `
let s = "hello world foo bar baz";
let upper = s.toUpperCase();
let lower = s.toLowerCase();
let len = s.length;
let idx = s.indexOf("world");
let sub = s.substring(0, 5);
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMClassMethodCall(b *testing.B) {
	source := `
class Calculator {
  constructor(v) { this.value = v; }
  add(x) { this.value = this.value + x; return this; }
  getValue() { return this.value; }
}
let calc = new Calculator(0);
for (let i = 0; i < 1000; i = i + 1) {
  calc.add(1);
}
calc.getValue();
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMClassInheritanceChain(b *testing.B) {
	source := `
class A { methodA() { return 1; } }
class B extends A { methodB() { return 2; } }
class C extends B { methodC() { return 3; } }
let obj = new C();
let sum = 0;
for (let i = 0; i < 500; i = i + 1) {
  sum = sum + obj.methodA() + obj.methodB() + obj.methodC();
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMClosureCapture(b *testing.B) {
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
let result = 0;
for (let i = 0; i < 500; i = i + 1) {
  result = counter();
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkPerfVMLogicalShortCircuit(b *testing.B) {
	source := `
function sideEffect() {
  return true;
}
let count = 0;
for (let i = 0; i < 5000; i = i + 1) {
  if (false && sideEffect()) {
    count = count + 1;
  }
  if (true || sideEffect()) {
    count = count + 1;
  }
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkPerfVMTernaryOperator(b *testing.B) {
	source := `
let result = 0;
for (let i = 0; i < 5000; i = i + 1) {
  result = i % 2 == 0 ? result + 1 : result - 1;
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkPerfVMIfElseChain(b *testing.B) {
	source := `
function classify(x) {
  if (x < 10) { return 1; }
  if (x < 20) { return 2; }
  if (x < 30) { return 3; }
  if (x < 40) { return 4; }
  if (x < 50) { return 5; }
  if (x < 60) { return 6; }
  if (x < 70) { return 7; }
  if (x < 80) { return 8; }
  if (x < 90) { return 9; }
  return 10;
}
let sum = 0;
for (let i = 0; i < 1000; i = i + 1) {
  sum = sum + classify(i);
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkPerfVMBitwiseOperations(b *testing.B) {
	source := `
let result = 0;
for (let i = 0; i < 5000; i = i + 1) {
  result = result ^ (i & 255) | (i << 1) >> 2;
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMCompoundAssignment(b *testing.B) {
	source := `
let x = 0;
for (let i = 0; i < 5000; i = i + 1) {
  x += i;
  x -= 1;
  x *= 2;
  x /= 2;
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMMathOperations(b *testing.B) {
	source := `
let result = 0;
for (let i = 0; i < 1000; i = i + 1) {
  result = result + (i * 3.14159 / 2.71828);
}
`
	benchmarkPerfVMExec(b, source)
}

// --- Memory / Allocation benchmarks ---

func BenchmarkVMManyObjectsAllocation(b *testing.B) {
	source := `
for (let i = 0; i < 500; i = i + 1) {
  let obj = {id: i, name: "item" + i, value: i * 10};
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMManyArraysAllocation(b *testing.B) {
	source := `
for (let i = 0; i < 500; i = i + 1) {
  let arr = [i, i+1, i+2, i+3, i+4];
}
`
	benchmarkPerfVMExec(b, source)
}

func BenchmarkVMManyStringsAllocation(b *testing.B) {
	source := `
for (let i = 0; i < 500; i = i + 1) {
  let s = "string_number_" + i;
}
`
	benchmarkPerfVMExec(b, source)
}

// --- End-to-End Pipeline benchmarks ---

func BenchmarkPipelineSmall(b *testing.B) {
	source := "let x = 1 + 2 * 3;"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(source)
		p := frontend.NewParser(l)
		program := p.ParseProgram()
		c := comp.NewCompiler()
		chunk, _ := c.Compile(program)
		runtime := rt.NewGoroutineRuntime()
		vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "bench.ia", runtime)
		_ = vm.Run()
	}
}

func BenchmarkPipelineMedium(b *testing.B) {
	source := generatePerfBenchmarkProgram(200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(source)
		p := frontend.NewParser(l)
		program := p.ParseProgram()
		c := comp.NewCompiler()
		chunk, _ := c.Compile(program)
		runtime := rt.NewGoroutineRuntime()
		vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "bench.ia", runtime)
		_ = vm.Run()
	}
}

func BenchmarkPipelineLarge(b *testing.B) {
	source := generatePerfBenchmarkProgram(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(source)
		p := frontend.NewParser(l)
		program := p.ParseProgram()
		c := comp.NewCompiler()
		chunk, _ := c.Compile(program)
		runtime := rt.NewGoroutineRuntime()
		vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "bench.ia", runtime)
		_ = vm.Run()
	}
}

// --- Scalability benchmarks ---

func BenchmarkScalabilityLexerLinear(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000}
	for _, size := range sizes {
		input := generateIdentifiers(size)
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				l := frontend.NewLexer(input)
				for l.NextToken().Type != frontend.EOF {
				}
			}
		})
	}
}

func BenchmarkScalabilityParserLinear(b *testing.B) {
	sizes := []int{50, 100, 200, 500}
	for _, size := range sizes {
		input := generateLetStatements(size)
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				l := frontend.NewLexer(input)
				p := frontend.NewParser(l)
				p.ParseProgram()
			}
		})
	}
}

func BenchmarkScalabilityCompilerLinear(b *testing.B) {
	sizes := []int{50, 100, 200, 500}
	for _, size := range sizes {
		input := generateLetStatements(size)
		b.Run(fmt.Sprintf("size=%d", size), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				compileTestSource(b, input)
			}
		})
	}
}

func BenchmarkScalabilityVMLoopIterations(b *testing.B) {
	sizes := []int{100, 500, 1000, 5000}
	for _, size := range sizes {
		source := fmt.Sprintf(`
let sum = 0;
for (let i = 0; i < %d; i = i + 1) {
  sum = sum + i;
}
`, size)
		b.Run(fmt.Sprintf("iterations=%d", size), func(b *testing.B) {
			benchmarkPerfVMExec(b, source)
		})
	}
}

// ============================================================
// Helper functions
// ============================================================

func benchmarkPerfVMExec(b *testing.B, source string) {
	b.Helper()
	chunk := compileTestSource(b, source)
	runtime := rt.NewGoroutineRuntime()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "bench.ia", runtime)
		_ = vm.Run()
	}
}

func generateIdentifiers(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += fmt.Sprintf("var%d ", i)
	}
	return result
}

func generateNumbers(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += fmt.Sprintf("%d ", i)
	}
	return result
}

func generateStrings(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += fmt.Sprintf("\"string%d\" ", i)
	}
	return result
}

func generateLetStatements(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += fmt.Sprintf("let x%d = %d + %d;\n", i, i, i*2)
	}
	return result
}

func generateDeepNesting(depth int) string {
	result := ""
	for i := 0; i < depth; i++ {
		result += "if (true) {\n"
	}
	result += "let x = 42;\n"
	for i := 0; i < depth; i++ {
		result += "}\n"
	}
	return result
}

func generateWideExpression(n int) string {
	result := "let x = "
	for i := 0; i < n; i++ {
		if i > 0 {
			result += " + "
		}
		result += fmt.Sprintf("%d", i)
	}
	result += ";"
	return result
}

func generateManyFunctions(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += fmt.Sprintf("function fn%d(x) { return x + %d; }\n", i, i)
	}
	return result
}

func generateManyClasses(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += fmt.Sprintf("class Class%d { constructor() { this.value = %d; } getValue() { return this.value; } }\n", i, i)
	}
	return result
}

func generateComplexClass(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += fmt.Sprintf("class Service%d {\n", i)
		result += fmt.Sprintf("  constructor() { this.id = %d; this.data = {}; }\n", i)
		result += fmt.Sprintf("  getId() { return this.id; }\n")
		result += fmt.Sprintf("  setData(val) { this.data = {value: val}; }\n")
		result += fmt.Sprintf("  getData() { return this.data.value; }\n")
		result += fmt.Sprintf("}\n")
	}
	return result
}

func generateNestedFunctions(depth int) string {
	result := ""
	for i := 0; i < depth; i++ {
		result += fmt.Sprintf("function level%d(x) {\n", i)
	}
	result += "return x;\n"
	for i := depth - 1; i >= 0; i-- {
		result += fmt.Sprintf("}\n")
	}
	return result
}

func generateManyClosures(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += fmt.Sprintf("function makeAdder%d(x) { function addY(y) { return x + y + %d; } return addY; }\n", i, i)
	}
	return result
}

func generateClassHierarchy(depth int) string {
	result := "class Base { baseMethod() { return 0; } }\n"
	for i := 0; i < depth; i++ {
		parent := "Base"
		if i > 0 {
			parent = fmt.Sprintf("Level%d", i-1)
		}
		result += fmt.Sprintf("class Level%d extends %s { method%d() { return %d; } }\n", i, parent, i, i)
	}
	return result
}

func generateAsyncFunctions(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += fmt.Sprintf("async function task%d() { return %d; }\n", i, i)
	}
	return result
}

func generatePerfBenchmarkProgram(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += fmt.Sprintf("let x%d = %d + %d;\n", i, i, i*2)
	}
	return result
}

// ============================================================
// Timing benchmarks (measure actual execution time)
// ============================================================

func TestPerfLexerTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timing test in short mode")
	}
	input := generateIdentifiers(10000)
	start := time.Now()
	l := frontend.NewLexer(input)
	for l.NextToken().Type != frontend.EOF {
	}
	elapsed := time.Since(start)
	t.Logf("Lexing 10000 identifiers took %v", elapsed)
}

func TestPerfParserTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timing test in short mode")
	}
	input := generateLetStatements(5000)
	start := time.Now()
	l := frontend.NewLexer(input)
	p := frontend.NewParser(l)
	p.ParseProgram()
	elapsed := time.Since(start)
	t.Logf("Parsing 5000 let statements took %v", elapsed)
}

func TestPerfCompilerTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timing test in short mode")
	}
	input := generateLetStatements(2000)
	start := time.Now()
	compileTestSource(t, input)
	elapsed := time.Since(start)
	t.Logf("Compiling 2000 let statements took %v", elapsed)
}

func TestPerfVMExecutionTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timing test in short mode")
	}
	source := `
let sum = 0;
for (let i = 0; i < 100000; i = i + 1) {
  sum = sum + i;
}
`
	chunk := compileTestSource(t, source)
	runtime := rt.NewGoroutineRuntime()
	start := time.Now()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "bench.ia", runtime)
	_ = vm.Run()
	elapsed := time.Since(start)
	t.Logf("VM execution of 100000 iterations took %v", elapsed)
}

func TestPerfFullPipelineTiming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timing test in short mode")
	}
	source := generatePerfBenchmarkProgram(1000)
	start := time.Now()
	l := frontend.NewLexer(source)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	c := comp.NewCompiler()
	chunk, _ := c.Compile(program)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "bench.ia", runtime)
	_ = vm.Run()
	elapsed := time.Since(start)
	t.Logf("Full pipeline (lex+parse+compile+exec) for 1000 statements took %v", elapsed)
}
