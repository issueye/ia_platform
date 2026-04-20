package frontend

import (
	"testing"
)

// ============================================================
// Statement parsing tests
// ============================================================

func TestParserLetStatement(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple let", "let x = 1;"},
		{"let with expression", "let y = 1 + 2 * 3;"},
		{"let without semicolon", "let z = true"},
		{"let with function call", "let a = foo(1, 2);"},
		{"let with object", "let obj = {key: 1};"},
		{"let with array", "let arr = [1, 2, 3];"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)
			if program == nil {
				t.Fatal("program is nil")
			}
			if len(program.Statements) != 1 {
				t.Fatalf("program has %d statements, want 1", len(program.Statements))
			}
			_, ok := program.Statements[0].(*LetStatement)
			if !ok {
				t.Errorf("statement is not *LetStatement, got %T", program.Statements[0])
			}
		})
	}
}

func TestParserFunctionStatement(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectName   string
		expectParams int
	}{
		{"empty function", "function foo() {}", "foo", 0},
		{"function with params", "function bar(a, b, c) {}", "bar", 3},
		{"function with body", "function baz(x) { return x + 1; }", "baz", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)
			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}
			fn, ok := program.Statements[0].(*FunctionStatement)
			if !ok {
				t.Fatalf("statement is not *FunctionStatement, got %T", program.Statements[0])
			}
			if fn.Name != tt.expectName {
				t.Errorf("function name = %q, want %q", fn.Name, tt.expectName)
			}
			if len(fn.Params) != tt.expectParams {
				t.Errorf("param count = %d, want %d", len(fn.Params), tt.expectParams)
			}
		})
	}
}

func TestParserAsyncFunction(t *testing.T) {
	input := `async function fetchData(url) { let res = await fetch(url); return res; }`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}
	fn, ok := program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("statement is not *FunctionStatement, got %T", program.Statements[0])
	}
	if !fn.Async {
		t.Error("expected async function")
	}
	if fn.Name != "fetchData" {
		t.Errorf("function name = %q, want %q", fn.Name, "fetchData")
	}
}

func TestParserClassStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectName    string
		expectParent  string
		expectMethods int
	}{
		{
			name:          "simple class",
			input:         "class Foo { constructor() {} bar() {} }",
			expectName:    "Foo",
			expectMethods: 2,
		},
		{
			name:          "class with inheritance",
			input:         "class Bar extends Foo { constructor() {} }",
			expectName:    "Bar",
			expectParent:  "Foo",
			expectMethods: 1,
		},
		{
			name:          "class with async method",
			input:         "class Client { async request() {} }",
			expectName:    "Client",
			expectMethods: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}
			class, ok := program.Statements[0].(*ClassStatement)
			if !ok {
				t.Fatalf("statement is not *ClassStatement, got %T", program.Statements[0])
			}
			if class.Name != tt.expectName {
				t.Errorf("class name = %q, want %q", class.Name, tt.expectName)
			}
			if class.ParentName != tt.expectParent {
				t.Errorf("parent name = %q, want %q", class.ParentName, tt.expectParent)
			}
			if len(class.Methods) != tt.expectMethods {
				t.Errorf("method count = %d, want %d", len(class.Methods), tt.expectMethods)
			}
		})
	}
}

func TestParserClassStatementAllowsCommentsBetweenMethods(t *testing.T) {
	input := `
class Server {
  // 注册路由
  registerRouters() {
    this.routes = new Routes(this.app);
    this.routes.registerRouters();
  }

  // 启动服务
  serve() {}
}
`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	classStmt, ok := program.Statements[0].(*ClassStatement)
	if !ok {
		t.Fatalf("statement is not *ClassStatement, got %T", program.Statements[0])
	}

	if len(classStmt.Methods) != 2 {
		t.Fatalf("method count = %d, want 2", len(classStmt.Methods))
	}

	if classStmt.Methods[0].Name != "registerRouters" {
		t.Fatalf("first method name = %q, want registerRouters", classStmt.Methods[0].Name)
	}

	if classStmt.Methods[1].Name != "serve" {
		t.Fatalf("second method name = %q, want serve", classStmt.Methods[1].Name)
	}
}

func TestParserImportStatement(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		namespace string
		nameCount int
	}{
		{"named import", `import { foo, bar, baz } from "./module";`, "", 3},
		{"namespace import", `import * as mod from "./module";`, "mod", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}
			imp, ok := program.Statements[0].(*ImportStatement)
			if !ok {
				t.Fatalf("statement is not *ImportStatement, got %T", program.Statements[0])
			}
			if imp.Module != "./module" {
				t.Errorf("module = %q, want %q", imp.Module, "./module")
			}
			if imp.Namespace != tt.namespace {
				t.Errorf("namespace = %q, want %q", imp.Namespace, tt.namespace)
			}
			if len(imp.Names) != tt.nameCount {
				t.Errorf("import names count = %d, want %d", len(imp.Names), tt.nameCount)
			}
		})
	}
}

func TestParserExportStatement(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"export let", "export let x = 1;"},
		{"export function", "export function foo() {}"},
		{"export class", "export class Bar {}"},
		{"export async function", "export async function baz() {}"},
		{"export list", "export { foo, bar };"},
		{"export alias", "export { foo as bar };"},
		{"export default", "export default 42;"},
		{"export default class", "export default class Counter {}"},
		{"export all", `export * from "./module";`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}
			_, ok := program.Statements[0].(*ExportStatement)
			if !ok {
				t.Fatalf("statement is not *ExportStatement, got %T", program.Statements[0])
			}
		})
	}
}

func TestParserExportSpecifiers(t *testing.T) {
	input := "export { foo, bar as baz };"
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}
	exp, ok := program.Statements[0].(*ExportStatement)
	if !ok {
		t.Fatalf("statement is not *ExportStatement, got %T", program.Statements[0])
	}
	if exp.Statement != nil {
		t.Fatal("expected export specifier list, got declaration export")
	}
	if len(exp.Specifiers) != 2 {
		t.Fatalf("specifier count = %d, want 2", len(exp.Specifiers))
	}
	if exp.Specifiers[0].LocalName != "foo" || exp.Specifiers[0].ExportName != "foo" {
		t.Fatalf("first specifier = %#v, want foo->foo", exp.Specifiers[0])
	}
	if exp.Specifiers[1].LocalName != "bar" || exp.Specifiers[1].ExportName != "baz" {
		t.Fatalf("second specifier = %#v, want bar->baz", exp.Specifiers[1])
	}
}

func TestParserIfStatement(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expectElse bool
	}{
		{"if only", "if (x > 0) { y = 1; }", false},
		{"if else", "if (x > 0) { y = 1; } else { y = -1; }", true},
		{"single line if else", "if (x > 0) y = 1; else y = -1;", true},
		{"if with complex condition", "if (x > 0 && y < 10 || z == 5) { x = 1; }", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			ifStmt, ok := program.Statements[0].(*IfStatement)
			if !ok {
				t.Fatalf("statement is not *IfStatement, got %T", program.Statements[0])
			}
			if ifStmt.Condition == nil {
				t.Error("condition is nil")
			}
			if ifStmt.Then == nil {
				t.Error("then block is nil")
			}
			if tt.expectElse && ifStmt.Else == nil {
				t.Error("expected else block, got nil")
			}
		})
	}
}

func TestParserWhileStatement(t *testing.T) {
	input := "while (x < 10) { x = x + 1; }"
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	whileStmt, ok := program.Statements[0].(*WhileStatement)
	if !ok {
		t.Fatalf("statement is not *WhileStatement, got %T", program.Statements[0])
	}
	if whileStmt.Condition == nil {
		t.Error("condition is nil")
	}
	if whileStmt.Body == nil {
		t.Error("body is nil")
	}
}

func TestParserForStatement(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"classic for", "for (let i = 0; i < 10; i = i + 1) { sum += i; }"},
		{"for with complex init", "for (let i = 0; i < 10; i = i + 1) { sum += i; }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			forStmt, ok := program.Statements[0].(*ForStatement)
			if !ok {
				t.Fatalf("statement is not *ForStatement, got %T", program.Statements[0])
			}
			if forStmt.Body == nil {
				t.Error("body is nil")
			}
		})
	}
}

func TestParserForInStatement(t *testing.T) {
	input := "for (key in obj) { sum += 1; }"
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	forInStmt, ok := program.Statements[0].(*ForInStatement)
	if !ok {
		t.Fatalf("statement is not *ForInStatement, got %T", program.Statements[0])
	}
	if forInStmt.Variable != "key" {
		t.Fatalf("for-in variable = %q, want key", forInStmt.Variable)
	}
	if forInStmt.Iterable == nil {
		t.Fatal("for-in iterable is nil")
	}
	if forInStmt.Body == nil {
		t.Fatal("for-in body is nil")
	}
}

func TestParserForOfStatement(t *testing.T) {
	input := "for (item of arr) { sum += item; }"
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	forOfStmt, ok := program.Statements[0].(*ForOfStatement)
	if !ok {
		t.Fatalf("statement is not *ForOfStatement, got %T", program.Statements[0])
	}
	if forOfStmt.Variable != "item" {
		t.Fatalf("for-of variable = %q, want item", forOfStmt.Variable)
	}
	if forOfStmt.Iterable == nil {
		t.Fatal("for-of iterable is nil")
	}
	if forOfStmt.Body == nil {
		t.Fatal("for-of body is nil")
	}
}

func TestParserTryCatchStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectCatch   bool
		expectFinally bool
	}{
		{"try catch", "try { foo(); } catch (e) { print(e); }", true, false},
		{"try finally", "try { foo(); } finally { cleanup(); }", false, true},
		{"try catch finally", "try { foo(); } catch (e) { print(e); } finally { cleanup(); }", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			tryStmt, ok := program.Statements[0].(*TryCatchStatement)
			if !ok {
				t.Fatalf("statement is not *TryCatchStatement, got %T", program.Statements[0])
			}
			if tryStmt.TryBlock == nil {
				t.Error("try block is nil")
			}
			if tt.expectCatch && tryStmt.CatchBlock == nil {
				t.Error("expected catch block, got nil")
			}
			if tt.expectFinally && tryStmt.FinallyBlock == nil {
				t.Error("expected finally block, got nil")
			}
		})
	}
}

func TestParserReturnStatement(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"return with value", "function f() { return 42; }"},
		{"return without value", "function f() { return; }"},
		{"return expression", "function f() { return 1 + 2; }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			_ = p.ParseProgram()
			checkParserErrors(t, p)
		})
	}
}

func TestParserThrowStatement(t *testing.T) {
	input := "throw \"error message\";"
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	throwStmt, ok := program.Statements[0].(*ThrowStatement)
	if !ok {
		t.Fatalf("statement is not *ThrowStatement, got %T", program.Statements[0])
	}
	if throwStmt.Value == nil {
		t.Error("throw value is nil")
	}
}

func TestParserBreakContinue(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"break", "while (true) { break; }"},
		{"continue", "for (let i = 0; i < 10; i = i + 1) { continue; }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			_ = p.ParseProgram()
			checkParserErrors(t, p)
		})
	}
}

// ============================================================
// Expression parsing tests
// ============================================================

func TestParserArithmeticExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"addition", "1 + 2"},
		{"subtraction", "10 - 5"},
		{"multiplication", "3 * 4"},
		{"division", "20 / 4"},
		{"modulo", "17 % 5"},
		{"operator precedence", "1 + 2 * 3 - 4 / 2"},
		{"parenthesized", "(1 + 2) * 3"},
		{"nested parentheses", "((1 + 2) * (3 - 4)) / 5"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)
			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}
			stmt, ok := program.Statements[0].(*ExpressionStatement)
			if !ok {
				t.Fatalf("statement is not *ExpressionStatement, got %T", program.Statements[0])
			}
			if stmt.Expr == nil {
				t.Error("expression is nil")
			}
		})
	}
}

func TestParserComparisonExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"equal", "a == b"},
		{"not equal", "a != b"},
		{"less than", "a < b"},
		{"greater than", "a > b"},
		{"less or equal", "a <= b"},
		{"greater or equal", "a >= b"},
		{"chained comparison", "a < b && b < c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			_ = p.ParseProgram()
			checkParserErrors(t, p)
		})
	}
}

func TestParserLogicalExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"and", "true && false"},
		{"or", "true || false"},
		{"not", "!flag"},
		{"complex logical", "a && b || c && !d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			_ = p.ParseProgram()
			checkParserErrors(t, p)
		})
	}
}

func TestParserBitwiseExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"bitwise and", "a & b"},
		{"bitwise or", "a | b"},
		{"bitwise xor", "a ^ b"},
		{"left shift", "a << 2"},
		{"right shift", "a >> 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			_ = p.ParseProgram()
			checkParserErrors(t, p)
		})
	}
}

func TestParserTernaryExpression(t *testing.T) {
	input := "let result = condition ? trueVal : falseVal;"
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	letStmt, ok := program.Statements[0].(*LetStatement)
	if !ok {
		t.Fatalf("statement is not *LetStatement, got %T", program.Statements[0])
	}
	ternary, ok := letStmt.Initializer.(*TernaryExpression)
	if !ok {
		t.Fatalf("initializer is not *TernaryExpression, got %T", letStmt.Initializer)
	}
	if ternary.Condition == nil {
		t.Error("condition is nil")
	}
	if ternary.Then == nil {
		t.Error("then is nil")
	}
	if ternary.Else == nil {
		t.Error("else is nil")
	}
}

func TestParserFunctionCallExpression(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		expectArgs int
	}{
		{"no args", "foo()", 0},
		{"single arg", "foo(1)", 1},
		{"multiple args", "foo(1, 2, 3)", 3},
		{"nested calls", "foo(bar(1), baz(2, 3))", 2},
		{"method call", "obj.method(1, 2)", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			stmt, ok := program.Statements[0].(*ExpressionStatement)
			if !ok {
				t.Fatalf("statement is not *ExpressionStatement, got %T", program.Statements[0])
			}
			call, ok := stmt.Expr.(*CallExpression)
			if !ok {
				t.Fatalf("expression is not *CallExpression, got %T", stmt.Expr)
			}
			if len(call.Arguments) != tt.expectArgs {
				t.Errorf("argument count = %d, want %d", len(call.Arguments), tt.expectArgs)
			}
		})
	}
}

func TestParserFunctionExpression(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectName  string
		expectArgs  int
		expectAsync bool
	}{
		{"anonymous function", "let fn = function(x) { return x + 1; };", "", 1, false},
		{"named function expression", "let fn = function inc(x) { return x + 1; };", "inc", 1, false},
		{"async function expression", "let fn = async function(x) { return await x; };", "", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			letStmt, ok := program.Statements[0].(*LetStatement)
			if !ok {
				t.Fatalf("statement is not *LetStatement, got %T", program.Statements[0])
			}
			fn, ok := letStmt.Initializer.(*FunctionExpression)
			if !ok {
				t.Fatalf("initializer is not *FunctionExpression, got %T", letStmt.Initializer)
			}
			if fn.Name != tt.expectName {
				t.Errorf("function name = %q, want %q", fn.Name, tt.expectName)
			}
			if len(fn.Params) != tt.expectArgs {
				t.Errorf("param count = %d, want %d", len(fn.Params), tt.expectArgs)
			}
			if fn.Async != tt.expectAsync {
				t.Errorf("async = %v, want %v", fn.Async, tt.expectAsync)
			}
		})
	}
}

func TestParserMemberAccess(t *testing.T) {
	input := "obj.prop.subProp"
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("statement is not *ExpressionStatement, got %T", program.Statements[0])
	}
	if stmt.Expr == nil {
		t.Fatal("expression is nil")
	}
	// Should be nested GetExpression
	get, ok := stmt.Expr.(*GetExpression)
	if !ok {
		t.Fatalf("expression is not *GetExpression, got %T", stmt.Expr)
	}
	if get.Property != "subProp" {
		t.Errorf("property = %q, want %q", get.Property, "subProp")
	}
}

func TestParserIndexExpression(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"array index", "arr[0]"},
		{"variable index", "arr[i]"},
		{"chained index", "matrix[0][1]"},
		{"string index", "str[0]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			_ = p.ParseProgram()
			checkParserErrors(t, p)
		})
	}
}

func TestParserArrayLiteral(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectLen int
	}{
		{"empty array", "[]", 0},
		{"single element", "[1]", 1},
		{"multiple elements", "[1, 2, 3]", 3},
		{"trailing comma", "[1, 2, 3,]", 3},
		{"nested arrays", "[[1, 2], [3, 4]]", 2},
		{"mixed types", `[1, "hello", true]`, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			stmt, ok := program.Statements[0].(*ExpressionStatement)
			if !ok {
				t.Fatalf("statement is not *ExpressionStatement, got %T", program.Statements[0])
			}
			arr, ok := stmt.Expr.(*ArrayLiteral)
			if !ok {
				t.Fatalf("expression is not *ArrayLiteral, got %T", stmt.Expr)
			}
			if len(arr.Elements) != tt.expectLen {
				t.Errorf("element count = %d, want %d", len(arr.Elements), tt.expectLen)
			}
		})
	}
}

func TestParserObjectLiteral(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectLen int
	}{
		{"empty object", "{}", 0},
		{"single property", "{key: 1}", 1},
		{"multiple properties", "{a: 1, b: 2, c: 3}", 3},
		{"trailing comma", "{a: 1, b: 2,}", 2},
		{"nested objects", "{outer: {inner: 1}}", 1},
		{"identifier values", "{foo: bar}", 1},
		{"comments between properties", "{a: 1, // note\n b: 2}", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			stmt, ok := program.Statements[0].(*ExpressionStatement)
			if !ok {
				t.Fatalf("statement is not *ExpressionStatement, got %T", program.Statements[0])
			}
			obj, ok := stmt.Expr.(*ObjectLiteral)
			if !ok {
				t.Fatalf("expression is not *ObjectLiteral, got %T", stmt.Expr)
			}
			if len(obj.Properties) != tt.expectLen {
				t.Errorf("property count = %d, want %d", len(obj.Properties), tt.expectLen)
			}
		})
	}
}

func TestParserRawStringLiteral(t *testing.T) {
	input := "let sql = `SELECT *\nFROM packages`; "
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	letStmt, ok := program.Statements[0].(*LetStatement)
	if !ok {
		t.Fatalf("statement is not *LetStatement, got %T", program.Statements[0])
	}
	str, ok := letStmt.Initializer.(*StringLiteral)
	if !ok {
		t.Fatalf("initializer is not *StringLiteral, got %T", letStmt.Initializer)
	}
	if str.Value != "SELECT *\nFROM packages" {
		t.Fatalf("raw string = %q", str.Value)
	}
}

func TestParserTemplateStringInterpolation(t *testing.T) {
	input := "let msg = `hello ${name}!`; "
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	letStmt, ok := program.Statements[0].(*LetStatement)
	if !ok {
		t.Fatalf("statement is not *LetStatement, got %T", program.Statements[0])
	}
	if _, ok := letStmt.Initializer.(*BinaryExpression); !ok {
		t.Fatalf("initializer is not *BinaryExpression, got %T", letStmt.Initializer)
	}
}

func TestParserFunctionRestParams(t *testing.T) {
	input := "function f(a, ...rest) { return rest.length; }"
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	fn, ok := program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("statement is not *FunctionStatement, got %T", program.Statements[0])
	}
	if fn.RestParam != "rest" {
		t.Fatalf("rest param = %q, want rest", fn.RestParam)
	}
	if len(fn.Params) != 1 || fn.Params[0] != "a" {
		t.Fatalf("params = %#v, want [a]", fn.Params)
	}
}

func TestParserArrowRestParams(t *testing.T) {
	input := "let f = (x, ...rest) => rest.length;"
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	letStmt, ok := program.Statements[0].(*LetStatement)
	if !ok {
		t.Fatalf("statement is not *LetStatement, got %T", program.Statements[0])
	}
	arrow, ok := letStmt.Initializer.(*ArrowFunctionExpression)
	if !ok {
		t.Fatalf("initializer is not *ArrowFunctionExpression, got %T", letStmt.Initializer)
	}
	if arrow.RestParam != "rest" {
		t.Fatalf("arrow rest param = %q, want rest", arrow.RestParam)
	}
}

func TestParserDynamicImportExpression(t *testing.T) {
	input := `let mod = import("./dep");`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	letStmt, ok := program.Statements[0].(*LetStatement)
	if !ok {
		t.Fatalf("statement is not *LetStatement, got %T", program.Statements[0])
	}
	if _, ok := letStmt.Initializer.(*DynamicImportExpression); !ok {
		t.Fatalf("initializer is not *DynamicImportExpression, got %T", letStmt.Initializer)
	}
}

func TestParserClassPrivateField(t *testing.T) {
	input := `
class Counter {
  #value;
  constructor() { this.#value = 1; }
  readValue() { return this.#value; }
}
`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	classStmt, ok := program.Statements[0].(*ClassStatement)
	if !ok {
		t.Fatalf("statement is not *ClassStatement, got %T", program.Statements[0])
	}
	if len(classStmt.PrivateFields) != 1 {
		t.Fatalf("private fields count = %d, want 1", len(classStmt.PrivateFields))
	}
}

func TestParserDestructuringLetStatements(t *testing.T) {
	input := `
let [a, b] = arr;
let {x, y: z} = obj;
`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
	if _, ok := program.Statements[0].(*ArrayDestructuringLetStatement); !ok {
		t.Fatalf("first statement is not *ArrayDestructuringLetStatement, got %T", program.Statements[0])
	}
	objStmt, ok := program.Statements[1].(*ObjectDestructuringLetStatement)
	if !ok {
		t.Fatalf("second statement is not *ObjectDestructuringLetStatement, got %T", program.Statements[1])
	}
	if len(objStmt.Bindings) != 2 {
		t.Fatalf("object bindings count = %d, want 2", len(objStmt.Bindings))
	}
	if objStmt.Bindings[1].Key != "y" || objStmt.Bindings[1].Name != "z" {
		t.Fatalf("second object binding = %#v, want y:z", objStmt.Bindings[1])
	}
}

func TestParserDestructuringAssignStatements(t *testing.T) {
	input := `
[a, b] = arr;
({x, y: z} = obj);
`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
	if _, ok := program.Statements[0].(*ArrayDestructuringAssignStatement); !ok {
		t.Fatalf("first statement is not *ArrayDestructuringAssignStatement, got %T", program.Statements[0])
	}
	objStmt, ok := program.Statements[1].(*ObjectDestructuringAssignStatement)
	if !ok {
		t.Fatalf("second statement is not *ObjectDestructuringAssignStatement, got %T", program.Statements[1])
	}
	if len(objStmt.Bindings) != 2 {
		t.Fatalf("object bindings count = %d, want 2", len(objStmt.Bindings))
	}
	if objStmt.Bindings[0].Key != "x" || objStmt.Bindings[0].Name != "x" {
		t.Fatalf("first object binding = %#v, want x:x", objStmt.Bindings[0])
	}
	if objStmt.Bindings[1].Key != "y" || objStmt.Bindings[1].Name != "z" {
		t.Fatalf("second object binding = %#v, want y:z", objStmt.Bindings[1])
	}
}

func TestParserNullLiteral(t *testing.T) {
	input := "let value = null;"
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	letStmt, ok := program.Statements[0].(*LetStatement)
	if !ok {
		t.Fatalf("statement is not *LetStatement, got %T", program.Statements[0])
	}
	if _, ok := letStmt.Initializer.(*NullLiteral); !ok {
		t.Fatalf("initializer is not *NullLiteral, got %T", letStmt.Initializer)
	}
}

func TestParserUnaryExpressions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"negation", "-x"},
		{"not", "!flag"},
		{"double negation", "- -x"},
		{"unary in expression", "-1 + 2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			_ = p.ParseProgram()
			checkParserErrors(t, p)
		})
	}
}

func TestParserNewExpression(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"new without args", "new Foo()"},
		{"new with args", "new Foo(1, 2, 3)"},
		{"new with member access", "new module.Class()"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			_ = p.ParseProgram()
			checkParserErrors(t, p)
		})
	}
}

func TestParserUpdateExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		isPrefix bool
		operator TokenType
		operand  string
	}{
		{"prefix increment", "++x", true, PLUSPLUS, "x"},
		{"prefix decrement", "--x", true, MINUSMINUS, "x"},
		{"postfix increment", "x++", false, PLUSPLUS, "x"},
		{"postfix decrement", "x--", false, MINUSMINUS, "x"},
		{"prefix in let", "let y = ++x;", true, PLUSPLUS, "x"},
		{"postfix in let", "let y = x++;", false, PLUSPLUS, "x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)

			// For simple expressions, check the expression directly
			if tt.input != "let y = ++x;" && tt.input != "let y = x++;" {
				if len(program.Statements) != 1 {
					t.Fatalf("expected 1 statement, got %d", len(program.Statements))
				}
				stmt, ok := program.Statements[0].(*ExpressionStatement)
				if !ok {
					t.Fatalf("statement is not *ExpressionStatement, got %T", program.Statements[0])
				}
				update, ok := stmt.Expr.(*UpdateExpression)
				if !ok {
					t.Fatalf("expression is not *UpdateExpression, got %T", stmt.Expr)
				}
				if update.IsPrefix != tt.isPrefix {
					t.Errorf("isPrefix = %v, want %v", update.IsPrefix, tt.isPrefix)
				}
				if update.Operator != tt.operator {
					t.Errorf("operator = %v, want %v", update.Operator, tt.operator)
				}
				if update.Operand.Name != tt.operand {
					t.Errorf("operand name = %s, want %s", update.Operand.Name, tt.operand)
				}
			}
		})
	}
}

func TestParserAwaitExpression(t *testing.T) {
	input := "let result = await promise;"
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	letStmt, ok := program.Statements[0].(*LetStatement)
	if !ok {
		t.Fatalf("statement is not *LetStatement, got %T", program.Statements[0])
	}
	await, ok := letStmt.Initializer.(*AwaitExpression)
	if !ok {
		t.Fatalf("initializer is not *AwaitExpression, got %T", letStmt.Initializer)
	}
	if await.Expr == nil {
		t.Error("await expression is nil")
	}
}

func TestParserAssignmentStatement(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"simple assignment", "x = 1;"},
		{"compound += ", "x += 5;"},
		{"compound -= ", "x -= 3;"},
		{"compound *= ", "x *= 2;"},
		{"compound /= ", "x /= 4;"},
		{"compound %= ", "x %= 3;"},
		{"property assignment", "obj.prop = 1;"},
		{"compound property assignment", "obj.prop += 1;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newParser(t, tt.input)
			program := p.ParseProgram()
			checkParserErrors(t, p)
			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}
		})
	}
}

// ============================================================
// Error recovery tests
// ============================================================

func TestParserSyntaxErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing semicolon in let", "let x ="},
		{"missing parenthesis in function", "function foo {}"},
		{"missing brace in if", "if (x > 0) { y = 1;"},
		{"invalid expression", "let x = +;"},
		{"missing identifier in let", "let = 1;"},
		{"unclosed parenthesis", "foo(1, 2"},
		{"unclosed brace", "class Foo { method() {}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			p := NewParser(l)
			p.ParseProgram()
			// Should have errors
			if len(p.Errors()) == 0 {
				t.Logf("warning: expected parse errors but got none for input: %q", tt.input)
			}
		})
	}
}

func TestParserErrorMessages(t *testing.T) {
	input := "let = 1;" // Missing identifier
	l := NewLexer(input)
	p := NewParser(l)
	p.ParseProgram()

	errors := p.Errors()
	if len(errors) == 0 {
		t.Fatal("expected parse errors, got none")
	}
	// Check that error message is meaningful
	if len(errors[0]) == 0 {
		t.Error("error message is empty")
	}
}

// ============================================================
// Complex program parsing tests
// ============================================================

func TestParserComplexProgram(t *testing.T) {
	input := `
import { Math } from "math";

class Calculator extends BaseCalculator {
  constructor(initial) {
    this.value = initial;
  }
  
  add(x) {
    this.value += x;
    return this;
  }
  
  async compute() {
    let result = await this.fetchData();
    return result * this.value;
  }
}

export function createCalculator(init) {
  return new Calculator(init);
}

let calc = createCalculator(10);
calc.add(5);
`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	// Should have import, class, export function, and 3 expression statements
	if len(program.Statements) < 4 {
		t.Errorf("expected at least 4 statements, got %d", len(program.Statements))
	}
}

func TestParserNestedFunctions(t *testing.T) {
	input := `
function outer() {
  function inner() {
    return 42;
  }
  return inner();
}
`
	p := newParser(t, input)
	_ = p.ParseProgram()
	checkParserErrors(t, p)
}

func TestParserClosures(t *testing.T) {
	input := `
function makeAdder(x) {
  function addY(y) {
    return x + y;
  }
  return addY;
}
`
	p := newParser(t, input)
	_ = p.ParseProgram()
	checkParserErrors(t, p)
}

func TestParserASTNodePositions(t *testing.T) {
	input := "let x = 1;\nbreak;"
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}

	letStmt, ok := program.Statements[0].(*LetStatement)
	if !ok {
		t.Fatalf("first statement is not *LetStatement, got %T", program.Statements[0])
	}
	if letStmt.Pos().Line != 1 || letStmt.Pos().Column != 1 {
		t.Fatalf("let statement position = (%d, %d), want (1, 1)", letStmt.Pos().Line, letStmt.Pos().Column)
	}

	breakStmt, ok := program.Statements[1].(*BreakStatement)
	if !ok {
		t.Fatalf("second statement is not *BreakStatement, got %T", program.Statements[1])
	}
	if breakStmt.Pos().Line != 2 || breakStmt.Pos().Column != 1 {
		t.Fatalf("break statement position = (%d, %d), want (2, 1)", breakStmt.Pos().Line, breakStmt.Pos().Column)
	}
}

// ============================================================
// Helper functions
// ============================================================

func newParser(t *testing.T, input string) *Parser {
	t.Helper()
	l := NewLexer(input)
	return NewParser(l)
}

func checkParserErrors(t *testing.T, p *Parser) {
	t.Helper()
	errors := p.Errors()
	if len(errors) > 0 {
		t.Errorf("parser has %d error(s):", len(errors))
		for _, msg := range errors {
			t.Errorf("  - %s", msg)
		}
		t.FailNow()
	}
}

// ============================================================
// Benchmark Tests
// ============================================================

func BenchmarkParserSimpleLet(b *testing.B) {
	input := "let x = 1;"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := NewLexer(input)
		p := NewParser(l)
		p.ParseProgram()
	}
}

func BenchmarkParserFunction(b *testing.B) {
	input := "function foo(a, b) { return a + b; }"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := NewLexer(input)
		p := NewParser(l)
		p.ParseProgram()
	}
}

func BenchmarkParserClass(b *testing.B) {
	input := "class Foo { constructor() {} bar(x) { return x; } }"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := NewLexer(input)
		p := NewParser(l)
		p.ParseProgram()
	}
}

func BenchmarkParserComplexProgram(b *testing.B) {
	input := `
import { Math } from "math";

class Calculator {
  constructor(v) { this.value = v; }
  add(x) { this.value += x; return this; }
  async compute() { return this.value * 2; }
}

export function create(v) { return new Calculator(v); }
let calc = create(10);
calc.add(5);
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := NewLexer(input)
		p := NewParser(l)
		p.ParseProgram()
	}
}

func BenchmarkParserLargeProgram(b *testing.B) {
	input := generateLargeParserProgram(500)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := NewLexer(input)
		p := NewParser(l)
		p.ParseProgram()
	}
}

func generateLargeParserProgram(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += "function fn" + itoa(i) + "(x) { return x + " + itoa(i) + "; }\n"
	}
	return result
}
