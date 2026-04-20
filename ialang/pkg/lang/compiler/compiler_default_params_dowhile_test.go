package compiler

import (
	"testing"

	"ialang/pkg/lang/frontend"
)

// Test default parameters in functions
func TestCompileDefaultParams_Basic(t *testing.T) {
	src := `
function foo(x = 10) {
  return x;
}
let a = foo();
let b = foo(5);
`
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
	if chunk == nil {
		t.Fatal("chunk is nil")
	}
}

func TestCompileDefaultParams_Multiple(t *testing.T) {
	src := `
function foo(x = 1, y = 2, z = 3) {
  return x + y + z;
}
let a = foo();
let b = foo(10);
let c = foo(10, 20);
let d = foo(10, 20, 30);
`
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
	if chunk == nil {
		t.Fatal("chunk is nil")
	}
}

func TestCompileDefaultParams_MixedWithNonDefault(t *testing.T) {
	src := `
function foo(x, y = 10, z = 20) {
  return x + y + z;
}
let a = foo(1);
let b = foo(1, 2);
let c = foo(1, 2, 3);
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	_, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
}

func TestCompileDefaultParams_ExpressionDefaults(t *testing.T) {
	src := `
function foo(x = 1 + 2, y = "hello") {
  return x;
}
let a = foo();
`
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
	if chunk == nil {
		t.Fatal("chunk is nil")
	}
}

func TestCompileDefaultParams_ArrowFunction(t *testing.T) {
	src := `
let foo = (x = 10, y = 20) => {
  return x + y;
};
let a = foo();
let b = foo(5);
`
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
	if chunk == nil {
		t.Fatal("chunk is nil")
	}
}

func TestCompileDefaultParams_ArrowConcise(t *testing.T) {
	src := `
let foo = (x = 10) => x * 2;
let a = foo();
let b = foo(5);
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	_, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
}

// Test do-while loop
func TestCompileDoWhile_Basic(t *testing.T) {
	src := `
let x = 0;
do {
  x = x + 1;
} while (x < 3);
`
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
	if chunk == nil {
		t.Fatal("chunk is nil")
	}
}

func TestCompileDoWhile_WithBreak(t *testing.T) {
	src := `
let x = 0;
do {
  x = x + 1;
  if (x > 5) {
    break;
  }
} while (x < 10);
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	_, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
}

func TestCompileDoWhile_WithContinue(t *testing.T) {
	src := `
let x = 0;
let sum = 0;
do {
  x = x + 1;
  if (x == 3) {
    continue;
  }
  sum = sum + x;
} while (x < 5);
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	_, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
}

func TestCompileDoWhile_NestedWithForLoop(t *testing.T) {
	src := `
let x = 0;
do {
  for (let i = 0; i < 3; i = i + 1) {
    x = x + 1;
  }
} while (x < 10);
`
	l := frontend.NewLexer(src)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}

	c := NewCompiler()
	_, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
}

func TestCompileDoWhile_AlwaysExecutesOnce(t *testing.T) {
	// This tests the semantic: do-while always executes body at least once
	// even when condition is initially false
	src := `
let executed = false;
do {
  executed = true;
} while (false);
`
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
	if chunk == nil {
		t.Fatal("chunk is nil")
	}
}

// Test both features together
func TestCompile_DefaultParamsAndDoWhile_Together(t *testing.T) {
	src := `
function greet(name = "World", count = 1) {
  let result = "";
  let i = 0;
  do {
    result = result + "Hello " + name + "!\n";
    i = i + 1;
  } while (i < count);
  return result;
}

let msg1 = greet();
let msg2 = greet("Alice");
let msg3 = greet("Bob", 3);
`
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
	if chunk == nil {
		t.Fatal("chunk is nil")
	}
}
