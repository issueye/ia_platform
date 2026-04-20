package compiler

import (
	"testing"

	"ialang/pkg/lang/frontend"
)

func compileSrc(t *testing.T, src string) *Chunk {
	t.Helper()
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	c := NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	return chunk
}

func hasOp(chunk *Chunk, op OpCode) bool {
	if opInCode(chunk.Code, op) {
		return true
	}
	for _, c := range chunk.Constants {
		if fn, ok := c.(*FunctionTemplate); ok && fn != nil && fn.Chunk != nil {
			if hasOp(fn.Chunk, op) {
				return true
			}
		}
	}
	return false
}

func opInCode(code []Instruction, op OpCode) bool {
	for _, ins := range code {
		if ins.Op == op {
			return true
		}
	}
	return false
}

func TestCompileClosures(t *testing.T) {
	src := `
let x = 10;
function outer() {
  let y = 20;
  function inner() {
    return x + y;
  }
  return inner();
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpClosure) {
		t.Fatal("expected OpClosure for nested function")
	}
}

func TestCompileArrowBlockBody(t *testing.T) {
	src := `let f = (x) => { return x + 1; };`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpClosure) {
		t.Fatal("expected OpClosure for arrow function")
	}
}

func TestCompileArrowConciseBody(t *testing.T) {
	src := `let f = (x) => x + 1;`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpClosure) {
		t.Fatal("expected OpClosure for concise arrow")
	}
}

func TestCompileClassInheritance(t *testing.T) {
	src := `
class Animal {
  constructor(name) {
    this.name = name;
  }
}
class Dog extends Animal {
  constructor(name) {
    super(name);
  }
  bark() { return "woof"; }
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpClass) {
		t.Fatal("expected OpClass")
	}
	if !hasOp(chunk, OpSuperCall) {
		t.Fatal("expected OpSuperCall in Dog constructor")
	}
}

func TestCompileClassStaticMethod(t *testing.T) {
	src := `
class Utils {
  static identity(x) { return x; }
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpClass) {
		t.Fatal("expected OpClass")
	}
}

func TestCompileClassGetterSetter(t *testing.T) {
	src := `
class Point {
  constructor(x) { this._x = x; }
  get x() { return this._x; }
  set x(v) { this._x = v; }
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpClass) {
		t.Fatal("expected OpClass")
	}
}

func TestCompileForInLoop(t *testing.T) {
	src := `
let obj = {a: 1, b: 2};
for (k in obj) {
  let v = obj[k];
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpObjectKeys) {
		t.Fatal("expected OpObjectKeys in for-in")
	}
}

func TestCompileForOfLoop(t *testing.T) {
	src := `
let arr = [1, 2, 3];
for (v of arr) {
  let x = v;
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpIndex) {
		t.Fatal("expected OpIndex in for-of")
	}
}

func TestCompileWhileLoop(t *testing.T) {
	src := `
let i = 0;
while (i < 10) {
  i = i + 1;
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpJumpIfFalse) {
		t.Fatal("expected OpJumpIfFalse in while")
	}
}

func TestCompileDoWhileLoop(t *testing.T) {
	src := `
let i = 0;
do {
  i = i + 1;
} while (i < 10);
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpJumpIfTrue) {
		t.Fatal("expected OpJumpIfTrue in do-while")
	}
}

func TestCompileForLoopWithPost(t *testing.T) {
	src := `
for (let i = 0; i < 10; i = i + 1) {
  let x = i;
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpJumpIfFalse) {
		t.Fatal("expected OpJumpIfFalse in for loop")
	}
}

func TestCompileBreakContinue(t *testing.T) {
	src := `
let i = 0;
while (i < 10) {
  i = i + 1;
  if (i == 5) { continue; }
  if (i == 8) { break; }
}
`
	compileSrc(t, src)
}

func TestCompileTryCatch(t *testing.T) {
	src := `
try {
  throw "error";
} catch(e) {
  let x = e;
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpPushTry) {
		t.Fatal("expected OpPushTry")
	}
	if !hasOp(chunk, OpThrow) {
		t.Fatal("expected OpThrow")
	}
}

func TestCompileTryCatchFinally(t *testing.T) {
	src := `
try {
  throw "err";
} catch(e) {
  let x = e;
} finally {
  let y = 1;
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpPushTry) {
		t.Fatal("expected OpPushTry")
	}
}

func TestCompileTryFinally(t *testing.T) {
	src := `
try {
  let x = 1;
} finally {
  let y = 2;
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpPushTry) {
		t.Fatal("expected OpPushTry in try-finally")
	}
}

func TestCompileIfElse(t *testing.T) {
	src := `
let x = 1;
if (x > 0) {
  let a = 1;
} else {
  let b = 2;
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpJumpIfFalse) {
		t.Fatal("expected OpJumpIfFalse in if-else")
	}
}

func TestCompileIfOnly(t *testing.T) {
	src := `
let x = 1;
if (x > 0) {
  let a = 1;
}
`
	compileSrc(t, src)
}

func TestCompileTernaryExpression(t *testing.T) {
	src := `let x = true ? 1 : 2;`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpJumpIfFalse) {
		t.Fatal("expected OpJumpIfFalse in ternary")
	}
}

func TestCompileNullishCoalescing(t *testing.T) {
	src := `let x = null ?? 42;`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpJumpIfNotNullish) {
		t.Fatal("expected OpJumpIfNotNullish in nullish coalescing")
	}
}

func TestCompileLogicalAnd(t *testing.T) {
	src := `let x = true && false;`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpAnd) {
		t.Fatal("expected OpAnd")
	}
}

func TestCompileLogicalOr(t *testing.T) {
	src := `let x = true || false;`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpOr) {
		t.Fatal("expected OpOr")
	}
}

func TestCompileBinaryOperators(t *testing.T) {
	cases := []struct {
		src string
		op  OpCode
	}{
		{"let x = 1 + 2;", OpAdd},
		{"let x = 3 - 1;", OpSub},
		{"let x = 2 * 3;", OpMul},
		{"let x = 6 / 2;", OpDiv},
		{"let x = 7 % 3;", OpMod},
		{"let x = 1 == 1;", OpEqual},
		{"let x = 1 != 2;", OpNotEqual},
		{"let x = 1 > 0;", OpGreater},
		{"let x = 0 < 1;", OpLess},
		{"let x = 1 >= 1;", OpGreaterEqual},
		{"let x = 1 <= 2;", OpLessEqual},
		{"let x = 1 & 2;", OpBitAnd},
		{"let x = 1 | 2;", OpBitOr},
		{"let x = 1 ^ 2;", OpBitXor},
		{"let x = 1 << 2;", OpShl},
		{"let x = 4 >> 1;", OpShr},
	}
	for _, tc := range cases {
		t.Run(tc.src, func(t *testing.T) {
			chunk := compileSrc(t, tc.src)
			if !hasOp(chunk, tc.op) {
				t.Fatalf("expected op %d in compiled output", tc.op)
			}
		})
	}
}

func TestCompileUnaryOperators(t *testing.T) {
	src := `
let a = !true;
let b = -5;
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpNot) {
		t.Fatal("expected OpNot")
	}
	if !hasOp(chunk, OpNeg) {
		t.Fatal("expected OpNeg")
	}
}

func TestCompileUpdateExpressions(t *testing.T) {
	src := `
let x = 0;
let a = x++;
let b = ++x;
let c = x--;
let d = --x;
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpDup) {
		t.Fatal("expected OpDup in update expression")
	}
}

func TestCompileCompoundAssign(t *testing.T) {
	src := `
let x = 10;
x += 5;
x -= 3;
x *= 2;
x /= 4;
x %= 3;
`
	compileSrc(t, src)
}

func TestCompileCompoundSetProperty(t *testing.T) {
	src := `
let obj = {x: 10};
obj.x += 5;
`
	compileSrc(t, src)
}

func TestCompileOptionalChainProperty(t *testing.T) {
	src := `let x = obj?.name;`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpJumpIfNotNullish) {
		t.Fatal("expected OpJumpIfNotNullish in optional chain")
	}
}

func TestCompileOptionalChainIndex(t *testing.T) {
	src := `let x = arr?.[0];`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpJumpIfNotNullish) {
		t.Fatal("expected OpJumpIfNotNullish in optional chain index")
	}
}

func TestCompileOptionalChainCall(t *testing.T) {
	src := `let x = fn?.();`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpJumpIfNotNullish) {
		t.Fatal("expected OpJumpIfNotNullish in optional chain call")
	}
}

func TestCompileOptionalChainMethod(t *testing.T) {
	src := `let x = obj?.method();`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpJumpIfNotNullish) {
		t.Fatal("expected OpJumpIfNotNullish in optional chain method")
	}
}

func TestCompileNewExpression(t *testing.T) {
	src := `
class Foo { constructor() { this.x = 1; } }
let f = new Foo();
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpNew) {
		t.Fatal("expected OpNew")
	}
}

func TestCompileAwaitExpression(t *testing.T) {
	src := `let v = await promise;`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpAwait) {
		t.Fatal("expected OpAwait")
	}
}

func TestCompileDynamicImport(t *testing.T) {
	src := `let mod = import("./mod");`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpImportDynamic) {
		t.Fatal("expected OpImportDynamic")
	}
}

func TestCompileSpreadArray(t *testing.T) {
	src := `
let a = [1, 2];
let b = [0, ...a, 3];
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpSpreadArray) {
		t.Fatal("expected OpSpreadArray")
	}
}

func TestCompileSpreadObject(t *testing.T) {
	src := `
let a = {x: 1};
let b = {y: 2, ...a};
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpSpreadObject) {
		t.Fatal("expected OpSpreadObject")
	}
}

func TestCompileSpreadCall(t *testing.T) {
	src := `
function sum(a, b) { return a + b; }
let args = [1, 2];
sum(...args);
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpSpreadCall) {
		t.Fatal("expected OpSpreadCall")
	}
}

func TestCompileArrayLiteral(t *testing.T) {
	src := `let a = [1, 2, 3];`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpArray) {
		t.Fatal("expected OpArray")
	}
}

func TestCompileObjectLiteral(t *testing.T) {
	src := `let o = {x: 1, y: 2};`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpObject) {
		t.Fatal("expected OpObject")
	}
}

func TestCompileSuperCall(t *testing.T) {
	src := `
class Base {
  constructor(x) { this.x = x; }
}
class Child extends Base {
  constructor(x) {
    super(x);
  }
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpSuperCall) {
		t.Fatal("expected OpSuperCall")
	}
}

func TestCompileSuperMethod(t *testing.T) {
	src := `
class Base {
  greet() { return "hi"; }
}
class Child extends Base {
  greet() { return super.greet(); }
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpSuper) {
		t.Fatal("expected OpSuper")
	}
}

func TestCompileImportNamed(t *testing.T) {
	src := `import { foo, bar } from "mod";`
	chunk := compileSrc(t, src)
	count := 0
	for _, ins := range chunk.Code {
		if ins.Op == OpImportName {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected 2 OpImportName, got %d", count)
	}
}

func TestCompileExportNames(t *testing.T) {
	src := `export let x = 1;`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpExportName) {
		t.Fatal("expected OpExportName")
	}
}

func TestCompileSetProperty(t *testing.T) {
	src := `
let obj = {};
obj.x = 42;
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpSetProperty) {
		t.Fatal("expected OpSetProperty")
	}
}

func TestCompileThrowWithValue(t *testing.T) {
	src := `throw "error";`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpThrow) {
		t.Fatal("expected OpThrow")
	}
}

func TestCompileThrowWithoutValue(t *testing.T) {
	src := `throw;`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpThrow) {
		t.Fatal("expected OpThrow")
	}
}

func TestCompileReturnWithValue(t *testing.T) {
	src := `
function f() { return 42; }
`
	compileSrc(t, src)
}

func TestCompileReturnWithoutValue(t *testing.T) {
	src := `
function f() { return; }
`
	compileSrc(t, src)
}

func TestCompileSwitchWithBreak(t *testing.T) {
	src := `
let x = 1;
switch (x) {
  case 1:
    let a = 10;
    break;
  case 2:
    let b = 20;
    break;
  default:
    let c = 30;
}
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpEqual) {
		t.Fatal("expected OpEqual in switch comparison")
	}
}

func TestCompileExpressionStatement(t *testing.T) {
	src := `1 + 2;`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpPop) {
		t.Fatal("expected OpPop after expression statement")
	}
}

func TestCompileTypeofInFunction(t *testing.T) {
	src := `
function f(x) { return typeof x; }
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpTypeof) {
		t.Fatal("expected OpTypeof")
	}
}

func TestCompileVoidInFunction(t *testing.T) {
	src := `
function f() { return void 0; }
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpPop) {
		t.Fatal("expected OpPop from void")
	}
}

func TestCompileGetExpression(t *testing.T) {
	src := `
let obj = {x: 1};
let y = obj.x;
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpGetProperty) {
		t.Fatal("expected OpGetProperty")
	}
}

func TestCompileIndexExpression(t *testing.T) {
	src := `
let arr = [1, 2, 3];
let y = arr[0];
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpIndex) {
		t.Fatal("expected OpIndex")
	}
}

func TestCompileLiterals(t *testing.T) {
	src := `
let a = 42;
let b = 3.14;
let c = "hello";
let d = true;
let e = false;
let f = null;
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpConstant) {
		t.Fatal("expected OpConstant for literals")
	}
}

func TestCompileAssignStatement(t *testing.T) {
	src := `
let x = 0;
x = 42;
`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpSetName) {
		t.Fatal("expected OpSetName")
	}
}

func TestCompileExportAll(t *testing.T) {
	src := `export * from "./dep";`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpExportAll) {
		t.Fatal("expected OpExportAll")
	}
}

func TestCompileFunctionExpression(t *testing.T) {
	src := `let f = function(x) { return x; };`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpClosure) {
		t.Fatal("expected OpClosure for function expression")
	}
}

func TestCompileFunctionWithRestParam(t *testing.T) {
	src := `function f(a, ...rest) { return rest.length; }`
	compileSrc(t, src)
}

func TestCompileFunctionWithDefaults(t *testing.T) {
	src := `function f(x = 10, y = 20) { return x + y; }`
	compileSrc(t, src)
}

func TestCompileNestedFunctions(t *testing.T) {
	src := `
function outer() {
  let a = 1;
  function middle() {
    let b = 2;
    function inner() {
      return a + b;
    }
    return inner();
  }
  return middle();
}
`
	compileSrc(t, src)
}

func TestCompileClassWithPrivateField(t *testing.T) {
	src := `
class Box {
  #value;
  constructor(v) { this.#value = v; }
  readValue() { return this.#value; }
}
`
	compileSrc(t, src)
}

func TestCompileExportDefaultExpression(t *testing.T) {
	src := `export default 42;`
	chunk := compileSrc(t, src)
	if !hasOp(chunk, OpExportDefault) {
		t.Fatal("expected OpExportDefault")
	}
}

func TestCompileForLoopNoCondition(t *testing.T) {
	src := `
let i = 0;
for (;;) {
  i = i + 1;
  if (i > 10) { break; }
}
`
	compileSrc(t, src)
}

func TestCompileForLoopNoInit(t *testing.T) {
	src := `
let i = 0;
for (; i < 10; i = i + 1) {
  let x = i;
}
`
	compileSrc(t, src)
}
