package lang_test

import (
	goRuntime "runtime"
	"testing"
	"time"

	comp "ialang/pkg/lang/compiler"
	"ialang/pkg/lang/frontend"
	rt "ialang/pkg/lang/runtime"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	rvm "ialang/pkg/lang/runtime/vm"
)

// ============================================================
// Async/Await Integration Tests
// ============================================================

func TestAsyncAwaitSimpleValue(t *testing.T) {
	source := `
async function getValue() {
  return 42;
}
let p = getValue();
`
	chunk := compileTestSource(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "async_simple.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("async simple value error: %v", err)
	}
}

func TestAsyncAwaitChained(t *testing.T) {
	source := `
async function double(x) {
  return x * 2;
}
async function main() {
  let result = await double(21);
  return result;
}
let p = main();
`
	chunk := compileTestSource(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "async_chain.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("async chained error: %v", err)
	}
}

func TestAsyncAwaitMultipleFunctions(t *testing.T) {
	source := `
async function step1() { return 1; }
async function step2() { return 2; }
async function step3() { return 3; }
let p1 = step1();
let p2 = step2();
let p3 = step3();
`
	chunk := compileTestSource(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "async_multi.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("async multiple functions error: %v", err)
	}
}

// ============================================================
// Try-Catch Integration Tests
// ============================================================

func TestTryCatchThrowAndCatch(t *testing.T) {
	source := `
let caught = false;
try {
  throw "error message";
} catch (e) {
  caught = true;
}
`
	runVMTestNoError(t, source)
}

func TestTryCatchFinally(t *testing.T) {
	source := `
let x = 0;
try {
  x = 1;
} finally {
  x = x + 10;
}
`
	runVMTestNoError(t, source)
}

func TestTryCatchNested(t *testing.T) {
	source := `
try {
  try {
    let x = 1;
  } catch (e) {
    let y = 2;
  }
} catch (e) {
  let z = 3;
}
`
	runVMTestNoError(t, source)
}

// ============================================================
// Math Module Integration Tests
// ============================================================

func TestMathModuleFunctions(t *testing.T) {
	source := `
import { sin, cos, sqrt, abs, pow, floor, ceil, round, max, min, PI } from "math";
let result = 0;
result = sin(PI / 2);
result = cos(0);
result = sqrt(16);
result = abs(-5);
result = pow(2, 10);
result = floor(3.7);
result = ceil(3.2);
result = round(3.5);
result = max(1, 2);
result = min(1, 2);
`
	runVMTestNoError(t, source)
}

func TestMathModuleConstants(t *testing.T) {
	source := `
import { PI, E } from "math";
let pi = PI;
let e = E;
`
	runVMTestNoError(t, source)
}

// ============================================================
// String Module Integration Tests
// ============================================================

func TestStringModuleOperations(t *testing.T) {
	source := `
let s = "hello world";
let upper = s.toUpperCase();
let lower = s.toLowerCase();
let trimmed = "  trim  ".trim();
let substr = s.substring(0, 5);
let idx = s.indexOf("world");
let replaced = s.replace("world", "ialang");
let split_result = s.split(" ");
`
	runVMTestNoError(t, source)
}

func TestStringModuleLength(t *testing.T) {
	source := `
let s = "hello";
let first = s[0];
`
	runVMTestNoError(t, source)
}

// ============================================================
// Array Module Integration Tests
// ============================================================

func TestArrayModuleOperations(t *testing.T) {
	source := `
let arr = [3, 1, 4, 1, 5, 9, 2, 6];
let arr2 = arr.concat([7, 8]);
let arr3 = arr.slice(1, 4);
let idx = arr.indexOf(4);
let hasFive = arr.includes(5);
let joined = arr.join(",");
`
	runVMTestNoError(t, source)
}

func TestArrayModuleLength(t *testing.T) {
	source := `
let arr = [1, 2, 3, 4, 5];
let len = arr.length;
`
	runVMTestNoError(t, source)
}

func TestArrayModulePush(t *testing.T) {
	source := `
let arr = [];
arr = arr.push(1);
arr = arr.push(2);
arr = arr.push(3);
`
	runVMTestNoError(t, source)
}

func TestArrayPrototypeMethodsReturnExpectedValues(t *testing.T) {
	source := `
function assert(cond, msg) {
  if (!cond) {
    throw msg;
  }
}

let base = [3, 1, 2];
let sorted = base.sort();
assert(sorted.length == 3, "sort length");
assert(sorted[0] == 1 && sorted[1] == 2 && sorted[2] == 3, "sort values");
assert(base[0] == 3 && base[1] == 1 && base[2] == 2, "sort should not mutate source");

let reversed = base.reverse();
assert(reversed[0] == 2 && reversed[1] == 1 && reversed[2] == 3, "reverse values");
assert(base[0] == 3 && base[2] == 2, "reverse should not mutate source");

let search = [1, 2, 1, 2];
assert(search.includes(2), "includes existing value");
assert(!search.includes(9), "includes missing value");
assert(search.indexOf(1) == 0, "indexOf first match");
assert(search.indexOf(1, 1) == 2, "indexOf from index");
assert(search.indexOf(9) == -1, "indexOf missing value");
assert(search.lastIndexOf(1) == 2, "lastIndexOf last match");
assert(search.lastIndexOf(1, 1) == 0, "lastIndexOf from index");

assert((["a", "b", "c"]).join("-") == "a-b-c", "join custom separator");
assert(([1, 2, 3]).join() == "1,2,3", "join default separator");

let sliced = ([1, 2, 3, 4, 5]).slice(1, 4);
assert(sliced.length == 3 && sliced[0] == 2 && sliced[2] == 4, "slice range");
let tail = ([1, 2, 3, 4, 5]).slice(-2);
assert(tail.length == 2 && tail[0] == 4 && tail[1] == 5, "slice negative start");
let emptySlice = ([1, 2, 3]).slice(3);
assert(emptySlice.length == 0, "slice past end");

let flatOnce = ([1, [2, [3]], 4]).flat();
assert(flatOnce.length == 4 && flatOnce[1] == 2 && flatOnce[2][0] == 3, "flat default depth");
let flatDeep = ([1, [2, [3]], 4]).flat(2);
assert(flatDeep.length == 4 && flatDeep[0] == 1 && flatDeep[2] == 3, "flat depth 2");

let concatBase = [1, 2];
let concatenated = concatBase.concat([3, 4], 5);
assert(concatenated.length == 5 && concatenated[0] == 1 && concatenated[4] == 5, "concat arrays and values");
assert(concatBase.length == 2, "concat should not mutate source");

let pushed = ([1, 2]).push(3, 4);
assert(pushed.length == 4 && pushed[2] == 3 && pushed[3] == 4, "push multiple values");
let popped = ([1, 2, 3]).pop();
assert(popped.length == 2 && popped[0] == 1 && popped[1] == 2, "pop removes last");
let emptyPop = ([]).pop();
assert(emptyPop.length == 0, "pop empty array");

let unshifted = ([2, 3]).unshift(0, 1);
assert(unshifted.length == 4 && unshifted[0] == 0 && unshifted[3] == 3, "unshift multiple values");
let shifted = ([1, 2, 3]).shift();
assert(shifted.length == 2 && shifted[0] == 2 && shifted[1] == 3, "shift removes first");
let emptyShift = ([]).shift();
assert(emptyShift.length == 0, "shift empty array");

let atArr = [10, 20, 30];
assert(atArr.at(0) == 10, "at first");
assert(atArr.at(-1) == 30, "at negative index");
assert(atArr.at(3) == null, "at out of range");

let fillBase = [1, 2, 3];
let filled = fillBase.fill(0);
assert(filled.length == 3 && filled[0] == 0 && filled[2] == 0, "fill values");
assert(fillBase[0] == 1 && fillBase[2] == 3, "fill should not mutate source");

let shuffled = ([1, 2, 3, 4]).shuffle();
assert(shuffled.length == 4, "shuffle length");
assert(shuffled.includes(1) && shuffled.includes(2) && shuffled.includes(3) && shuffled.includes(4), "shuffle preserves elements");
`
	runVMTestNoError(t, source)
}

func TestArrayNativeModuleMethods(t *testing.T) {
	source := `
import * as array from "array";

function assert(cond, msg) {
  if (!cond) {
    throw msg;
  }
}

let ranged = array.range(1, 5);
assert(ranged.length == 4 && ranged[0] == 1 && ranged[3] == 4, "range");

let fromString = array.from("abc");
assert(fromString.length == 3 && fromString[0] == "a" && fromString[2] == "c", "from string");

let copied = array.from([1, 2, 3]);
assert(copied.length == 3 && copied[1] == 2, "from array");

let built = array.of(4, 5, 6);
assert(built.length == 3 && built[0] == 4 && built[2] == 6, "of");
assert(array.isArray(built), "isArray true");
assert(!array.isArray("no"), "isArray false");

let merged = array.concat([1, 2], [3], 4);
assert(merged.length == 4 && merged[2] == 3 && merged[3] == 4, "concat");

assert(array.includes([1, 2, 3], 2), "includes true");
assert(!array.includes([1, 2, 3], 9), "includes false");
assert(array.indexOf([1, 2, 1], 1) == 0, "indexOf");
assert(array.lastIndexOf([1, 2, 1], 1) == 2, "lastIndexOf");

let flat = array.flat([1, [2, [3]], 4], 2);
assert(flat.length == 4 && flat[1] == 2 && flat[2] == 3, "flat");

let sliced = array.slice([1, 2, 3, 4], 1, 3);
assert(sliced.length == 2 && sliced[0] == 2 && sliced[1] == 3, "slice");

let joined = array.join(["a", "b", "c"], "-");
assert(joined == "a-b-c", "join");

let sorted = array.sort([3, 1, 2]);
assert(sorted.length == 3 && sorted[0] == 1 && sorted[2] == 3, "sort");

let reversed = array.reverse([1, 2, 3]);
assert(reversed.length == 3 && reversed[0] == 3 && reversed[2] == 1, "reverse");

let filled = array.fill([1, 2, 3], 9);
assert(filled.length == 3 && filled[0] == 9 && filled[2] == 9, "fill");

let shuffled = array.shuffle([1, 2, 3, 4]);
assert(shuffled.length == 4, "shuffle length");
assert(array.includes(shuffled, 1) && array.includes(shuffled, 2) && array.includes(shuffled, 3) && array.includes(shuffled, 4), "shuffle values");
`
	runVMTestNoError(t, source)
}

func TestArrayNativeModuleHigherOrderMethods(t *testing.T) {
	source := `
import * as array from "array";

function assert(cond, msg) {
  if (!cond) {
    throw msg;
  }
}

let nums = [1, 2, 3, 4];
let doubled = array.map(nums, function(x, i, arr) { return x * 2; });
assert(doubled.length == 4 && doubled[0] == 2 && doubled[3] == 8, "map values");

let evens = array.filter(nums, function(x, i, arr) { return x % 2 == 0; });
assert(evens.length == 2 && evens[0] == 2 && evens[1] == 4, "filter values");

assert(array.find(nums, function(x, i, arr) { return x > 2; }) == 3, "find match");
assert(array.find(nums, function(x, i, arr) { return x > 9; }) == null, "find missing");
assert(array.findIndex(nums, function(x, i, arr) { return x == 3; }) == 2, "findIndex match");
assert(array.findIndex(nums, function(x, i, arr) { return x == 9; }) == -1, "findIndex missing");

let total = 0;
assert(array.forEach(nums, function(x, i, arr) {
  total = total + x + i;
}) == true, "forEach return");
assert(total == 16, "forEach side effects");

assert(array.some(nums, function(x, i, arr) { return x == 4; }), "some true");
assert(!array.some(nums, function(x, i, arr) { return x == 9; }), "some false");
assert(array.every(nums, function(x, i, arr) { return x > 0; }), "every true");
assert(!array.every(nums, function(x, i, arr) { return x < 4; }), "every false");

assert(array.reduce(nums, function(acc, x, i, arr) { return acc + x; }, 0) == 10, "reduce with initial value");
assert(array.reduce(nums, function(acc, x, i, arr) { return acc + x; }) == 10, "reduce without initial value");

let flatMapped = array.flatMap(nums, function(x, i, arr) { return [x, x * 10]; });
assert(flatMapped.length == 8 && flatMapped[0] == 1 && flatMapped[1] == 10 && flatMapped[6] == 4 && flatMapped[7] == 40, "flatMap values");
`
	runVMTestNoError(t, source)
}

func TestArrayNativeModuleHigherOrderMethodsErrors(t *testing.T) {
	t.Run("reduce empty array without initial value", func(t *testing.T) {
		source := `
import * as array from "array";
array.reduce([], function(acc, x, i, arr) { return acc + x; });
`
		chunk := compileTestSource(t, source)
		vm := createTestVM(chunk)
		if err := vm.Run(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("map rejects non function callback", func(t *testing.T) {
		source := `
import * as array from "array";
array.map([1, 2, 3], 123);
`
		chunk := compileTestSource(t, source)
		vm := createTestVM(chunk)
		if err := vm.Run(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("flatMap rejects non function callback", func(t *testing.T) {
		source := `
import * as array from "array";
array.flatMap([1, 2, 3], "noop");
`
		chunk := compileTestSource(t, source)
		vm := createTestVM(chunk)
		if err := vm.Run(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestArrayPrototypeHigherOrderMethods(t *testing.T) {
	source := `
function assert(cond, msg) {
  if (!cond) {
    throw msg;
  }
}

let nums = [1, 2, 3, 4];
let doubled = nums.map(function(x, i, arr) { return x * 2; });
assert(doubled.length == 4 && doubled[0] == 2 && doubled[3] == 8, "map values");

let evens = nums.filter(function(x, i, arr) { return x % 2 == 0; });
assert(evens.length == 2 && evens[0] == 2 && evens[1] == 4, "filter values");

assert(nums.find(function(x, i, arr) { return x > 2; }) == 3, "find match");
assert(nums.find(function(x, i, arr) { return x > 9; }) == null, "find missing");
assert(nums.findIndex(function(x, i, arr) { return x == 3; }) == 2, "findIndex match");
assert(nums.findIndex(function(x, i, arr) { return x == 9; }) == -1, "findIndex missing");

let total = 0;
nums.forEach(function(x, i, arr) {
  total = total + x + i;
});
assert(total == 16, "forEach side effects");

assert(nums.some(function(x, i, arr) { return x == 4; }), "some true");
assert(!nums.some(function(x, i, arr) { return x == 9; }), "some false");
assert(nums.every(function(x, i, arr) { return x > 0; }), "every true");
assert(!nums.every(function(x, i, arr) { return x < 4; }), "every false");

assert(nums.reduce(function(acc, x, i, arr) { return acc + x; }, 0) == 10, "reduce with initial value");
assert(nums.reduce(function(acc, x, i, arr) { return acc + x; }) == 10, "reduce without initial value");
`
	runVMTestNoError(t, source)
}

// ============================================================
// Module Import/Export Integration Tests
// ============================================================

func TestModuleImportFromMath(t *testing.T) {
	source := `
import { PI } from "math";
let x = PI;
`
	runVMTestNoError(t, source)
}

func TestModuleImportMultipleFromMath(t *testing.T) {
	source := `
import { PI, E, sqrt } from "math";
let pi = PI;
let e = E;
let s = sqrt(16);
`
	runVMTestNoError(t, source)
}

// ============================================================
// Performance Regression Tests
// ============================================================

func TestPerfRegressionLexer10kTokens(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance regression test in short mode")
	}
	input := generatePerfIdentifiers(10000)
	start := time.Now()
	l := frontend.NewLexer(input)
	for l.NextToken().Type != frontend.EOF {
	}
	elapsed := time.Since(start)
	if elapsed > 500*time.Millisecond {
		t.Errorf("lexing 10000 identifiers took %v, expected < 500ms", elapsed)
	}
}

func TestPerfRegressionParser5kStatements(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance regression test in short mode")
	}
	input := generatePerfLetStatements(5000)
	start := time.Now()
	l := frontend.NewLexer(input)
	p := frontend.NewParser(l)
	p.ParseProgram()
	elapsed := time.Since(start)
	if elapsed > 2*time.Second {
		t.Errorf("parsing 5000 let statements took %v, expected < 2s", elapsed)
	}
}

func TestPerfRegressionCompiler2kStatements(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance regression test in short mode")
	}
	input := generatePerfLetStatements(2000)
	start := time.Now()
	compileTestSource(t, input)
	elapsed := time.Since(start)
	if elapsed > 2*time.Second {
		t.Errorf("compiling 2000 let statements took %v, expected < 2s", elapsed)
	}
}

func TestPerfRegressionVM10kIterations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance regression test in short mode")
	}
	source := `
let sum = 0;
for (let i = 0; i < 10000; i = i + 1) {
  sum = sum + i;
}
`
	chunk := compileTestSource(t, source)
	runtime := rt.NewGoroutineRuntime()
	start := time.Now()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "perf_reg.ia", runtime)
	_ = vm.Run()
	elapsed := time.Since(start)
	if elapsed > 5*time.Second {
		t.Errorf("VM 10000 iterations took %v, expected < 5s", elapsed)
	}
}

// ============================================================
// Memory Usage Tests
// ============================================================

func TestMemoryLexerLargeInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory test in short mode")
	}
	input := generatePerfIdentifiers(50000)
	var memBefore goRuntime.MemStats
	goRuntime.GC()
	goRuntime.ReadMemStats(&memBefore)

	l := frontend.NewLexer(input)
	for l.NextToken().Type != frontend.EOF {
	}

	var memAfter goRuntime.MemStats
	goRuntime.ReadMemStats(&memAfter)
	alloc := memAfter.TotalAlloc - memBefore.TotalAlloc
	t.Logf("Lexer 50000 identifiers allocated %d bytes (%.2f MB)", alloc, float64(alloc)/1024/1024)
}

func TestMemoryParserLargeInput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory test in short mode")
	}
	input := generatePerfLetStatements(10000)
	var memBefore goRuntime.MemStats
	goRuntime.GC()
	goRuntime.ReadMemStats(&memBefore)

	l := frontend.NewLexer(input)
	p := frontend.NewParser(l)
	p.ParseProgram()

	var memAfter goRuntime.MemStats
	goRuntime.ReadMemStats(&memAfter)
	alloc := memAfter.TotalAlloc - memBefore.TotalAlloc
	t.Logf("Parser 10000 statements allocated %d bytes (%.2f MB)", alloc, float64(alloc)/1024/1024)
}

func TestMemoryVMLargeLoop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory test in short mode")
	}
	source := `
let sum = 0;
for (let i = 0; i < 50000; i = i + 1) {
  sum = sum + i;
}
`
	chunk := compileTestSource(t, source)
	vmRuntime := rt.NewGoroutineRuntime()

	var memBefore goRuntime.MemStats
	goRuntime.GC()
	goRuntime.ReadMemStats(&memBefore)

	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(vmRuntime), nil, "mem_test.ia", vmRuntime)
	_ = vm.Run()

	var memAfter goRuntime.MemStats
	goRuntime.ReadMemStats(&memAfter)
	alloc := memAfter.TotalAlloc - memBefore.TotalAlloc
	t.Logf("VM 50000 iterations allocated %d bytes (%.2f MB)", alloc, float64(alloc)/1024/1024)
}

// ============================================================
// Full Pipeline Integration Tests
// ============================================================

func TestFullPipelineSimpleProgram(t *testing.T) {
	source := `
let x = 1 + 2 * 3;
let y = x - 1;
`
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
	vmRuntime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(vmRuntime), nil, "pipeline_simple.ia", vmRuntime)
	if err := vm.Run(); err != nil {
		t.Fatalf("VM execution error: %v", err)
	}
}

func TestFullPipelineClassProgram(t *testing.T) {
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
let p = new Point(3, 4);
p.distance();
`
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
	vmRuntime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(vmRuntime), nil, "pipeline_class.ia", vmRuntime)
	if err := vm.Run(); err != nil {
		t.Fatalf("VM execution error: %v", err)
	}
}

func TestFullPipelineAsyncProgram(t *testing.T) {
	source := `
async function fetchData() {
  return 42;
}
let p = fetchData();
`
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
	vmRuntime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(vmRuntime), nil, "pipeline_async.ia", vmRuntime)
	if err := vm.Run(); err != nil {
		t.Fatalf("VM execution error: %v", err)
	}
}

func TestFullPipelineTryCatchProgram(t *testing.T) {
	source := `
try {
  let x = 1 + 2;
} catch (e) {
  let y = 0;
}
`
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
	vmRuntime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(vmRuntime), nil, "pipeline_try.ia", vmRuntime)
	if err := vm.Run(); err != nil {
		t.Fatalf("VM execution error: %v", err)
	}
}

// ============================================================
// Static Methods Integration Tests
// ============================================================

func TestStaticMethodCall(t *testing.T) {
	source := `
class MathUtils {
  static add(a, b) {
    return a + b;
  }
}
let result = MathUtils.add(3, 4);
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("static method call error: %v", err)
	}
	// Verify result is 7
	if val, ok := vm.GetEnv("result"); !ok {
		t.Fatal("result variable not found")
	} else if f, ok := val.(float64); !ok || f != 7.0 {
		t.Fatalf("expected result=7.0, got %v (type: %T)", val, val)
	}
}

func TestStaticMethodMultiple(t *testing.T) {
	source := `
class MathUtils {
  static add(a, b) {
    return a + b;
  }
  static multiply(a, b) {
    return a * b;
  }
}
let sum = MathUtils.add(3, 4);
let product = MathUtils.multiply(3, 4);
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("static methods error: %v", err)
	}
	if val, ok := vm.GetEnv("sum"); !ok || val.(float64) != 7.0 {
		t.Fatalf("expected sum=7.0, got %v", val)
	}
	if val, ok := vm.GetEnv("product"); !ok || val.(float64) != 12.0 {
		t.Fatalf("expected product=12.0, got %v", val)
	}
}

func TestStaticMethodWithInstanceMethod(t *testing.T) {
	source := `
class Calculator {
  static add(a, b) {
    return a + b;
  }
  subtract(a, b) {
    return a - b;
  }
}
let staticResult = Calculator.add(10, 5);
let calc = new Calculator();
let instanceResult = calc.subtract(10, 5);
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("static + instance methods error: %v", err)
	}
	if val, ok := vm.GetEnv("staticResult"); !ok || val.(float64) != 15.0 {
		t.Fatalf("expected staticResult=15.0, got %v", val)
	}
	if val, ok := vm.GetEnv("instanceResult"); !ok || val.(float64) != 5.0 {
		t.Fatalf("expected instanceResult=5.0, got %v", val)
	}
}

func TestStaticMethodInheritance(t *testing.T) {
	source := `
class Base {
  static baseMethod() {
    return 42;
  }
}
class Derived extends Base {
}
let result = Derived.baseMethod();
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("static method inheritance error: %v", err)
	}
	if val, ok := vm.GetEnv("result"); !ok || val.(float64) != 42.0 {
		t.Fatalf("expected result=42.0, got %v", val)
	}
}

// ============================================================
// Getters and Setters Integration Tests
// ============================================================

func TestGetterBasic(t *testing.T) {
	source := `
class Person {
  get name() {
    return this._name;
  }
}
let p = new Person();
let result = p.name;
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("getter basic error: %v", err)
	}
	// _name is not set, so getter should return null
	if val, ok := vm.GetEnv("result"); !ok {
		t.Fatal("result variable not found")
	} else if val != nil {
		t.Fatalf("expected result=null, got %v", val)
	}
}

func TestGetterWithConstructor(t *testing.T) {
	source := `
class Person {
  constructor(n) {
    this._name = n;
  }
  get name() {
    return this._name;
  }
}
let p = new Person("Alice");
let result = p.name;
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("getter with constructor error: %v", err)
	}
	if val, ok := vm.GetEnv("result"); !ok {
		t.Fatal("result variable not found")
	} else if s, ok := val.(string); !ok || s != "Alice" {
		t.Fatalf("expected result='Alice', got %v (type: %T)", val, val)
	}
}

func TestSetterBasic(t *testing.T) {
	source := `
class Person {
  set name(value) {
    this._name = value;
  }
}
let p = new Person();
p.name = "Bob";
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("setter basic error: %v", err)
	}
}

func TestGetterSetterCombined(t *testing.T) {
	source := `
class Person {
  get name() {
    return this._name;
  }
  set name(value) {
    this._name = value.toUpperCase();
  }
}
let p = new Person();
p.name = "john";
let result = p.name;
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("getter setter combined error: %v", err)
	}
	if val, ok := vm.GetEnv("result"); !ok {
		t.Fatal("result variable not found")
	} else if s, ok := val.(string); !ok || s != "JOHN" {
		t.Fatalf("expected result='JOHN', got %v (type: %T)", val, val)
	}
}

func TestGetterSetterWithTransformation(t *testing.T) {
	source := `
class Temperature {
  get fahrenheit() {
    return this._celsius * 9 / 5 + 32;
  }
  set fahrenheit(value) {
    this._celsius = (value - 32) * 5 / 9;
  }
}
let t = new Temperature();
t.fahrenheit = 212;
let celsius = t._celsius;
let fahrenheit = t.fahrenheit;
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("getter setter transformation error: %v", err)
	}
	if val, ok := vm.GetEnv("celsius"); !ok || val.(float64) != 100.0 {
		t.Fatalf("expected celsius=100.0, got %v", val)
	}
	if val, ok := vm.GetEnv("fahrenheit"); !ok || val.(float64) != 212.0 {
		t.Fatalf("expected fahrenheit=212.0, got %v", val)
	}
}

func TestGetterSetterMultiple(t *testing.T) {
	source := `
class Rectangle {
  get width() {
    return this._width;
  }
  set width(value) {
    this._width = value;
  }
  get height() {
    return this._height;
  }
  set height(value) {
    this._height = value;
  }
  get area() {
    return this._width * this._height;
  }
}
let r = new Rectangle();
r.width = 5;
r.height = 3;
let area = r.area;
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("getter setter multiple error: %v", err)
	}
	if val, ok := vm.GetEnv("area"); !ok || val.(float64) != 15.0 {
		t.Fatalf("expected area=15.0, got %v", val)
	}
}

func TestGetterSetterInheritance(t *testing.T) {
	source := `
class Base {
  get value() {
    return this._value;
  }
  set value(v) {
    this._value = v;
  }
}
class Derived extends Base {
}
let d = new Derived();
d.value = 99;
let result = d.value;
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("getter setter inheritance error: %v", err)
	}
	if val, ok := vm.GetEnv("result"); !ok || val.(float64) != 99.0 {
		t.Fatalf("expected result=99.0, got %v", val)
	}
}

// ============================================================
// Combined Static + Getters/Setters Tests
// ============================================================

func TestClassWithAllFeatures(t *testing.T) {
	source := `
class Utils {
  static PI() {
    return 3.14159;
  }
  static add(a, b) {
    return a + b;
  }
  constructor(val) {
    this._value = val;
  }
  get value() {
    return this._value;
  }
  set value(v) {
    this._value = v * 2;
  }
  instanceMethod() {
    return this._value + 1;
  }
}
let pi = Utils.PI();
let sum = Utils.add(10, 20);
let obj = new Utils(5);
let val1 = obj.value;
obj.value = 10;
let val2 = obj.value;
let inst = obj.instanceMethod();
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("class with all features error: %v", err)
	}
	if val, ok := vm.GetEnv("pi"); !ok || val.(float64) != 3.14159 {
		t.Fatalf("expected pi=3.14159, got %v", val)
	}
	if val, ok := vm.GetEnv("sum"); !ok || val.(float64) != 30.0 {
		t.Fatalf("expected sum=30.0, got %v", val)
	}
	if val, ok := vm.GetEnv("val1"); !ok || val.(float64) != 5.0 {
		t.Fatalf("expected val1=5.0, got %v", val)
	}
	if val, ok := vm.GetEnv("val2"); !ok || val.(float64) != 20.0 {
		t.Fatalf("expected val2=20.0, got %v", val)
	}
	if val, ok := vm.GetEnv("inst"); !ok || val.(float64) != 21.0 {
		t.Fatalf("expected inst=21.0, got %v", val)
	}
}

// ============================================================
// Helper functions
// ============================================================

func generatePerfIdentifiers(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += "var" + itoaTest(i) + " "
	}
	return result
}

func generatePerfLetStatements(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += "let x" + itoaTest(i) + " = " + itoaTest(i) + " + " + itoaTest(i*2) + ";\n"
	}
	return result
}
