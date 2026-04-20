package frontend

import (
	"testing"
)

func TestParserSwitchStatement(t *testing.T) {
	input := `
let result = 0;
switch (x) {
  case 1:
    result = 10;
    break;
  case 2:
    result = 20;
    break;
  default:
    result = 0;
}
`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	found := false
	for _, stmt := range program.Statements {
		if _, ok := stmt.(*SwitchStatement); ok {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected SwitchStatement in program")
	}
}

func TestParserTypeofExpression(t *testing.T) {
	tests := []string{
		`let x = typeof 42;`,
		`let x = typeof "hello";`,
		`let x = typeof true;`,
		`let x = typeof null;`,
		`let x = typeof y;`,
	}
	for _, src := range tests {
		p := newParser(t, src)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		letStmt := program.Statements[0].(*LetStatement)
		if _, ok := letStmt.Initializer.(*TypeofExpression); !ok {
			t.Fatalf("expected TypeofExpression, got %T", letStmt.Initializer)
		}
	}
}

func TestParserVoidExpression(t *testing.T) {
	input := `let x = void 42;`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	letStmt := program.Statements[0].(*LetStatement)
	if _, ok := letStmt.Initializer.(*VoidExpression); !ok {
		t.Fatalf("expected VoidExpression, got %T", letStmt.Initializer)
	}
}

func TestParserOptionalChainExpression(t *testing.T) {
	input := `let x = obj?.prop;`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	letStmt := program.Statements[0].(*LetStatement)
	if _, ok := letStmt.Initializer.(*OptionalChainExpression); !ok {
		t.Fatalf("expected OptionalChainExpression, got %T", letStmt.Initializer)
	}
}

func TestParserOptionalChainIndex(t *testing.T) {
	input := `let x = arr?.[0];`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	letStmt := program.Statements[0].(*LetStatement)
	if _, ok := letStmt.Initializer.(*OptionalChainExpression); !ok {
		t.Fatalf("expected OptionalChainExpression, got %T", letStmt.Initializer)
	}
}

func TestParserOptionalChainCall(t *testing.T) {
	input := `let x = obj?.method();`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	letStmt := program.Statements[0].(*LetStatement)
	if _, ok := letStmt.Initializer.(*OptionalChainExpression); !ok {
		t.Fatalf("expected OptionalChainExpression, got %T", letStmt.Initializer)
	}
}

func TestParserDoWhileStatement(t *testing.T) {
	input := `
let x = 0;
do {
  x = x + 1;
} while (x < 10);
`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	found := false
	for _, stmt := range program.Statements {
		if dw, ok := stmt.(*DoWhileStatement); ok {
			found = true
			if dw.Condition == nil {
				t.Error("DoWhileStatement condition is nil")
			}
			if dw.Body == nil {
				t.Error("DoWhileStatement body is nil")
			}
			break
		}
	}
	if !found {
		t.Fatal("expected DoWhileStatement in program")
	}
}

func TestParserAsyncFunctionExpression(t *testing.T) {
	input := `let f = async function(x) { return x; };`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	letStmt := program.Statements[0].(*LetStatement)
	fnExpr, ok := letStmt.Initializer.(*FunctionExpression)
	if !ok {
		t.Fatalf("expected FunctionExpression, got %T", letStmt.Initializer)
	}
	if !fnExpr.Async {
		t.Error("expected async function expression")
	}
}

func TestParserArrowFunctionNoBody(t *testing.T) {
	input := `let f = (x) => x + 1;`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	letStmt := program.Statements[0].(*LetStatement)
	arrow, ok := letStmt.Initializer.(*ArrowFunctionExpression)
	if !ok {
		t.Fatalf("expected ArrowFunctionExpression, got %T", letStmt.Initializer)
	}
	if len(arrow.Params) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(arrow.Params))
	}
}

func TestParserSuperExpression(t *testing.T) {
	input := `
class Animal {
  constructor(name) {
    this.name = name;
  }
}
class Dog extends Animal {
  constructor(name) {
    super(name);
  }
}
`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) < 2 {
		t.Fatalf("expected at least 2 statements, got %d", len(program.Statements))
	}
}

func TestParserGroupedExpression(t *testing.T) {
	input := `let x = (1 + 2) * 3;`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}
}

func TestParserExpressionListInArray(t *testing.T) {
	input := `let arr = [1, 2, 3];`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	letStmt := program.Statements[0].(*LetStatement)
	arrLit, ok := letStmt.Initializer.(*ArrayLiteral)
	if !ok {
		t.Fatalf("expected ArrayLiteral, got %T", letStmt.Initializer)
	}
	if len(arrLit.Elements) != 3 {
		t.Errorf("array length = %d, want 3", len(arrLit.Elements))
	}
}

func TestParserSpreadInArray(t *testing.T) {
	input := `let a = [1, 2]; let b = [...a, 3];`
	p := newParser(t, input)
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}
}

func TestParserCompoundAssignment(t *testing.T) {
	tests := []string{
		`let x = 1; x += 2;`,
		`let x = 1; x -= 2;`,
		`let x = 1; x *= 2;`,
		`let x = 1; x /= 2;`,
		`let x = 1; x %= 2;`,
	}
	for _, src := range tests {
		p := newParser(t, src)
		program := p.ParseProgram()
		checkParserErrors(t, p)
		if len(program.Statements) != 2 {
			t.Fatalf("expected 2 statements for %q, got %d", src, len(program.Statements))
		}
	}
}
