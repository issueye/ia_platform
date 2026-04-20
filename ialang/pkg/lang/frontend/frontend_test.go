package frontend

import "testing"

func TestLexerAndParserBasicProgram(t *testing.T) {
	src := `
let x = 1 + 2;
if (x > 1) { x = x + 1; } else { x = x - 1; }
`
	l := NewLexer(src)
	p := NewParser(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if program == nil {
		t.Fatal("program is nil")
	}
	if len(program.Statements) == 0 {
		t.Fatal("program has no statements")
	}
}

func TestParserAsyncAwaitProgram(t *testing.T) {
	src := `
async function f() { return 1; }
let p = f();
let v = await p;
`
	l := NewLexer(src)
	p := NewParser(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	if program == nil || len(program.Statements) < 3 {
		t.Fatalf("unexpected statement count: %d", len(program.Statements))
	}
}
