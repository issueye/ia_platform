package lang_test

import (
	"fmt"
	"testing"

	"ialang/pkg/lang/compiler"
	"ialang/pkg/lang/frontend"
	"ialang/pkg/lang/runtime"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	rvm "ialang/pkg/lang/runtime/vm"
)

func runSource(source string) error {
	l := frontend.NewLexer(source)
	p := frontend.NewParser(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		return fmt.Errorf("parse errors: %v", p.Errors())
	}

	c := compiler.NewCompiler()
	chunk, compileErrs := c.Compile(program)
	if len(compileErrs) > 0 {
		return fmt.Errorf("compile errors: %v", compileErrs)
	}

	modules := rtbuiltin.DefaultModules(nil)
	asyncRuntime := runtime.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, modules, nil, "test", asyncRuntime)

	return vm.Run()
}

func TestLexerOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "arithmetic operators",
			input:    "3 * 4 + 10 / 2 - 1 % 3",
			expected: "no errors",
		},
		{
			name:     "comparison operators",
			input:    "10 >= 5 && 5 <= 10 && 3 < 5 && 7 > 2",
			expected: "no errors",
		},
		{
			name:     "compound assignment",
			input:    "let x = 10; x += 5; x -= 3; x *= 2; x /= 4; x %= 3",
			expected: "no errors",
		},
		{
			name:     "ternary operator",
			input:    "let x = 10 > 5 ? \"yes\" : \"no\"",
			expected: "no errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := frontend.NewLexer(tt.input)
			parser := frontend.NewParser(lexer)
			_ = parser.ParseProgram()

			errors := parser.Errors()
			if len(errors) > 0 && tt.expected == "no errors" {
				t.Errorf("unexpected parse errors: %v", errors)
			}
		})
	}
}

func TestVMOperators(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "multiplication",
			source: "let x = 6 * 7; print(x)",
		},
		{
			name:   "division",
			source: "let x = 20 / 4; print(x)",
		},
		{
			name:   "modulo",
			source: "let x = 17 % 5; print(x)",
		},
		{
			name:   "compound add",
			source: "let x = 10; x += 5; print(x)",
		},
		{
			name:   "compound subtract",
			source: "let x = 10; x -= 3; print(x)",
		},
		{
			name:   "compound multiply",
			source: "let x = 10; x *= 2; print(x)",
		},
		{
			name:   "compound divide",
			source: "let x = 24; x /= 4; print(x)",
		},
		{
			name:   "compound modulo",
			source: "let x = 17; x %= 5; print(x)",
		},
		{
			name:   "prefix increment",
			source: "let x = 5; let y = ++x; print(x); print(y)",
		},
		{
			name:   "postfix increment",
			source: "let x = 5; let y = x++; print(x); print(y)",
		},
		{
			name:   "prefix decrement",
			source: "let x = 5; let y = --x; print(x); print(y)",
		},
		{
			name:   "postfix decrement",
			source: "let x = 5; let y = x--; print(x); print(y)",
		},
		{
			name:   "increment in expression",
			source: "let x = 10; print(++x)",
		},
		{
			name:   "decrement in expression",
			source: "let x = 10; print(x--)",
		},
		{
			name:   "multiple increments",
			source: "let x = 0; ++x; ++x; ++x; print(x)",
		},
		{
			name:   "increment and compound add",
			source: "let x = 5; ++x; x += 3; print(x)",
		},
		{
			name:   "for loop with postfix increment",
			source: "let sum = 0; for (let i = 0; i < 3; i++) { sum += i; } print(sum)",
		},
		{
			name:   "for loop with prefix increment",
			source: "let sum = 0; for (let i = 0; i < 3; ++i) { sum += i; } print(sum)",
		},
		{
			name:   "greater or equal true",
			source: "let x = 10 >= 5; print(x)",
		},
		{
			name:   "greater or equal false",
			source: "let x = 3 >= 5; print(x)",
		},
		{
			name:   "less or equal true",
			source: "let x = 5 <= 10; print(x)",
		},
		{
			name:   "less or equal false",
			source: "let x = 5 <= 3; print(x)",
		},
		{
			name:   "ternary true branch",
			source: "let x = 10 > 5 ? \"yes\" : \"no\"; print(x)",
		},
		{
			name:   "ternary false branch",
			source: "let x = 3 > 5 ? \"yes\" : \"no\"; print(x)",
		},
		{
			name:   "nested ternary",
			source: "let score = 85; let grade = score >= 90 ? \"A\" : score >= 80 ? \"B\" : \"C\"; print(grade)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runSource(tt.source)
			if err != nil {
				t.Errorf("execution error: %v", err)
			}
		})
	}
}
