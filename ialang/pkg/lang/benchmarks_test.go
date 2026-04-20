package lang_test

import (
	"testing"

	"ialang/pkg/lang/frontend"
	rt "ialang/pkg/lang/runtime"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	rvm "ialang/pkg/lang/runtime/vm"
)

// ============================================================
// Lexer Benchmarks - Different input sizes
// ============================================================

func BenchmarkLexerSmallInput(b *testing.B) {
	input := "let x = 1 + 2 * 3;"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		for l.NextToken().Type != frontend.EOF {
		}
	}
}

func BenchmarkLexerMediumInput(b *testing.B) {
	input := generateBenchmarkProgram(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		for l.NextToken().Type != frontend.EOF {
		}
	}
}

func BenchmarkLexerLargeInput(b *testing.B) {
	input := generateBenchmarkProgram(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		for l.NextToken().Type != frontend.EOF {
		}
	}
}

func BenchmarkLexerHugeInput(b *testing.B) {
	input := generateBenchmarkProgram(10000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		for l.NextToken().Type != frontend.EOF {
		}
	}
}

func BenchmarkLexerStringLiterals(b *testing.B) {
	input := generateStringBenchmark(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		for l.NextToken().Type != frontend.EOF {
		}
	}
}

// ============================================================
// Parser Benchmarks - Different program complexities
// ============================================================

func BenchmarkParserSimpleProgram(b *testing.B) {
	input := "let x = 1 + 2 * 3;"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		p := frontend.NewParser(l)
		p.ParseProgram()
	}
}

func BenchmarkParserMediumProgram(b *testing.B) {
	input := generateBenchmarkProgram(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		p := frontend.NewParser(l)
		p.ParseProgram()
	}
}

func BenchmarkParserLargeProgram(b *testing.B) {
	input := generateBenchmarkProgram(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		p := frontend.NewParser(l)
		p.ParseProgram()
	}
}

func BenchmarkParserComplexClassProgram(b *testing.B) {
	input := generateClassBenchmark(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		p := frontend.NewParser(l)
		p.ParseProgram()
	}
}

func BenchmarkParserDeeplyNestedProgram(b *testing.B) {
	input := generateNestedBenchmark(50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := frontend.NewLexer(input)
		p := frontend.NewParser(l)
		p.ParseProgram()
	}
}

// ============================================================
// Compiler Benchmarks - Different AST sizes
// ============================================================

func BenchmarkCompilerSmallAST(b *testing.B) {
	input := "let x = 1 + 2 * 3;"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileTestSource(b, input)
	}
}

func BenchmarkCompilerMediumAST(b *testing.B) {
	input := generateBenchmarkProgram(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileTestSource(b, input)
	}
}

func BenchmarkCompilerLargeAST(b *testing.B) {
	input := generateBenchmarkProgram(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileTestSource(b, input)
	}
}

func BenchmarkCompilerClassCompilation(b *testing.B) {
	input := generateClassBenchmark(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileTestSource(b, input)
	}
}

func BenchmarkCompilerClosureCompilation(b *testing.B) {
	input := generateClosureBenchmark(50)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compileTestSource(b, input)
	}
}

// ============================================================
// VM Benchmarks - Different execution patterns
// ============================================================

func BenchmarkVMSimpleLoop(b *testing.B) {
	source := `
let sum = 0;
for (let i = 0; i < 1000; i = i + 1) {
  sum = sum + i;
}
`
	benchmarkVMExec(b, source)
}

func BenchmarkVMNestedLoop(b *testing.B) {
	source := `
let sum = 0;
for (let i = 0; i < 100; i = i + 1) {
  for (let j = 0; j < 10; j = j + 1) {
    sum = sum + 1;
  }
}
`
	benchmarkVMExec(b, source)
}

func BenchmarkVMFunctionCall(b *testing.B) {
	source := `
function add(a, b) {
  return a + b;
}
let result = 0;
for (let i = 0; i < 1000; i = i + 1) {
  result = add(result, i);
}
`
	benchmarkVMExec(b, source)
}

func BenchmarkVMRecursiveFunction(b *testing.B) {
	source := `
function fib(n) {
  if (n <= 1) { return n; }
  return fib(n - 1) + fib(n - 2);
}
fib(15);
`
	benchmarkVMExec(b, source)
}

func BenchmarkVMRecursiveTail(b *testing.B) {
	source := `
function sumTo(n, acc) {
  if (n <= 0) { return acc; }
  return sumTo(n - 1, acc + n);
}
sumTo(500, 0);
`
	benchmarkVMExec(b, source)
}

func BenchmarkVMObjectCreation(b *testing.B) {
	source := `
let total = 0;
for (let i = 0; i < 100; i = i + 1) {
  let obj = {key: i, doubled: i * 2};
  total = total + obj.doubled;
}
`
	benchmarkVMExec(b, source)
}

func BenchmarkVMClassInstantiation(b *testing.B) {
	source := `
class Point {
  constructor(x, y) {
    this.x = x;
    this.y = y;
  }
  distance() {
    return this.x + this.y;
  }
}
let sum = 0;
for (let i = 0; i < 100; i = i + 1) {
  let p = new Point(i, i * 2);
  sum = sum + p.distance();
}
`
	benchmarkVMExec(b, source)
}

func BenchmarkVMClassInheritance(b *testing.B) {
	source := `
class Animal {
  constructor(name) {
    this.name = name;
  }
  speak() {
    return 1;
  }
}

class Dog extends Animal {
  bark() {
    return 2;
  }
}

let sum = 0;
for (let i = 0; i < 100; i = i + 1) {
  let dog = new Dog("Buddy");
  sum = sum + dog.bark();
}
`
	benchmarkVMExec(b, source)
}

func BenchmarkVMClosureExecution(b *testing.B) {
	source := `
function makeAdder(x) {
  function addY(y) {
    return x + y;
  }
  return addY;
}
let add5 = makeAdder(5);
let sum = 0;
for (let i = 0; i < 1000; i = i + 1) {
  sum = add5(sum);
}
`
	benchmarkVMExec(b, source)
}

func BenchmarkVMArrayOperations(b *testing.B) {
	source := `
let arr = [];
for (let i = 0; i < 100; i = i + 1) {
  arr = arr.push(i * 2);
}
let sum = 0;
let i = 0;
while (i < arr.length) {
  sum = sum + arr[i];
  i = i + 1;
}
`
	benchmarkVMExec(b, source)
}

func BenchmarkVMIfElseChain(b *testing.B) {
	source := `
function classify(x) {
  if (x < 10) { return 1; }
  if (x < 20) { return 2; }
  if (x < 30) { return 3; }
  if (x < 40) { return 4; }
  if (x < 50) { return 5; }
  return 6;
}
let sum = 0;
for (let i = 0; i < 100; i = i + 1) {
  sum = sum + classify(i);
}
`
	benchmarkVMExec(b, source)
}

func BenchmarkVMLogicalShortCircuit(b *testing.B) {
	source := `
function sideEffect() {
  return true;
}
let count = 0;
for (let i = 0; i < 1000; i = i + 1) {
  if (false && sideEffect()) {
    count = count + 1;
  }
  if (true || sideEffect()) {
    count = count + 1;
  }
}
`
	benchmarkVMExec(b, source)
}

func BenchmarkVMBitwiseOperations(b *testing.B) {
	source := `
let result = 0;
for (let i = 0; i < 1000; i = i + 1) {
  result = result ^ (i & 255) | (i << 1) >> 2;
}
`
	benchmarkVMExec(b, source)
}

func BenchmarkVMTernaryExpression(b *testing.B) {
	source := `
let result = 0;
for (let i = 0; i < 1000; i = i + 1) {
  result = i % 2 == 0 ? result + i : result - i;
}
`
	benchmarkVMExec(b, source)
}

// ============================================================
// Standard library benchmarks
// ============================================================

func BenchmarkStdlibMathFunctions(b *testing.B) {
	source := `
import { sin, cos, sqrt, abs, pow, floor, ceil, round, max, min, PI } from "math";
let result = 0;
result = result + sin(PI / 2);
result = result + cos(0);
result = result + sqrt(16);
result = result + abs(-5);
result = result + pow(2, 10);
result = result + floor(3.7);
result = result + ceil(3.2);
result = result + round(3.5);
result = result + max(1, 2);
result = result + min(1, 2);
`
	benchmarkVMExec(b, source)
}

func BenchmarkStdlibStringOperations(b *testing.B) {
	source := `
let s = "hello world";
let upper = s.toUpperCase();
let lower = s.toLowerCase();
let trimmed = "  trim me  ".trim();
let substr = s.substring(0, 5);
let idx = s.indexOf("world");
let replaced = s.replace("world", "ialang");
let split_result = s.split(" ");
`
	benchmarkVMExec(b, source)
}

func BenchmarkStdlibArrayOperations(b *testing.B) {
	source := `
let arr = [3, 1, 4, 1, 5, 9, 2, 6];
let arr2 = arr.concat([7, 8]);
let arr3 = arr.slice(1, 4);
let idx = arr.indexOf(4);
let hasFive = arr.includes(5);
let joined = arr.join(",");
`
	benchmarkVMExec(b, source)
}

// ============================================================
// Helper functions
// ============================================================

func benchmarkVMExec(b *testing.B, source string) {
	b.Helper()
	chunk := compileTestSource(b, source)
	runtime := rt.NewGoroutineRuntime()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "bench.ia", runtime)
		_ = vm.Run()
	}
}

func generateBenchmarkProgram(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += "let x" + itoaTest(i) + " = " + itoaTest(i) + " + " + itoaTest(i*2) + ";\n"
	}
	return result
}

func generateClassBenchmark(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += "class Class" + itoaTest(i) + " {\n"
		result += "  constructor() { this.value = " + itoaTest(i) + "; }\n"
		result += "  getValue() { return this.value; }\n"
		result += "}\n"
	}
	return result
}

func generateNestedBenchmark(depth int) string {
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

func generateClosureBenchmark(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += "function makeAdder" + itoaTest(i) + "(x) {\n"
		result += "  function addY(y) { return x + y + " + itoaTest(i) + "; }\n"
		result += "  return addY;\n"
		result += "}\n"
	}
	return result
}

func generateStringBenchmark(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += "let s" + itoaTest(i) + " = \"hello world " + itoaTest(i) + "\";\n"
	}
	return result
}
