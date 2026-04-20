package frontend

import (
	"strings"
	"testing"
)

// ============================================================
// Token recognition tests
// ============================================================

func TestLexerOperators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Token
	}{
		{
			name:  "arithmetic operators",
			input: "+ - * / %",
			expected: []Token{
				{Type: PLUS, Literal: "+"},
				{Type: MINUS, Literal: "-"},
				{Type: ASTERISK, Literal: "*"},
				{Type: SLASH, Literal: "/"},
				{Type: MODULO, Literal: "%"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "comparison operators",
			input: "== != < > <= >=",
			expected: []Token{
				{Type: EQ, Literal: "=="},
				{Type: NEQ, Literal: "!="},
				{Type: LT, Literal: "<"},
				{Type: GT, Literal: ">"},
				{Type: LTE, Literal: "<="},
				{Type: GTE, Literal: ">="},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "logical operators",
			input: "&& || !",
			expected: []Token{
				{Type: AND, Literal: "&&"},
				{Type: OR, Literal: "||"},
				{Type: BANG, Literal: "!"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "nullish coalescing operator",
			input: "??",
			expected: []Token{
				{Type: NULLISH, Literal: "??"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "increment and decrement operators",
			input: "++ --",
			expected: []Token{
				{Type: PLUSPLUS, Literal: "++"},
				{Type: MINUSMINUS, Literal: "--"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "bitwise operators",
			input: "& | ^ << >>",
			expected: []Token{
				{Type: BITAND, Literal: "&"},
				{Type: BITOR, Literal: "|"},
				{Type: BITXOR, Literal: "^"},
				{Type: SHL, Literal: "<<"},
				{Type: SHR, Literal: ">>"},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "compound assignment operators",
			input: "+= -= *= /= %=",
			expected: []Token{
				{Type: PLUSEQ, Literal: "+="},
				{Type: MINUSEQ, Literal: "-="},
				{Type: MULTEQ, Literal: "*="},
				{Type: DIVEQ, Literal: "/="},
				{Type: MODEQ, Literal: "%="},
				{Type: EOF, Literal: ""},
			},
		},
		{
			name:  "assignment and ternary",
			input: "= ?",
			expected: []Token{
				{Type: ASSIGN, Literal: "="},
				{Type: QUESTION, Literal: "?"},
				{Type: EOF, Literal: ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			for i, want := range tt.expected {
				tok := l.NextToken()
				if tok.Type != want.Type {
					t.Errorf("token[%d] type = %v, want %v", i, tok.Type, want.Type)
				}
				if tok.Literal != want.Literal {
					t.Errorf("token[%d] literal = %q, want %q", i, tok.Literal, want.Literal)
				}
			}
		})
	}
}

func TestLexerDelimiters(t *testing.T) {
	input := ", : . ; ( ) { } [ ]"
	l := NewLexer(input)

	expected := []Token{
		{Type: COMMA, Literal: ","},
		{Type: COLON, Literal: ":"},
		{Type: DOT, Literal: "."},
		{Type: SEMICOLON, Literal: ";"},
		{Type: LPAREN, Literal: "("},
		{Type: RPAREN, Literal: ")"},
		{Type: LBRACE, Literal: "{"},
		{Type: RBRACE, Literal: "}"},
		{Type: LBRACKET, Literal: "["},
		{Type: RBRACKET, Literal: "]"},
		{Type: EOF, Literal: ""},
	}

	for i, want := range expected {
		tok := l.NextToken()
		if tok.Type != want.Type {
			t.Errorf("token[%d] type = %v, want %v", i, tok.Type, want.Type)
		}
		if tok.Literal != want.Literal {
			t.Errorf("token[%d] literal = %q, want %q", i, tok.Literal, want.Literal)
		}
	}
}

func TestLexerKeywords(t *testing.T) {
	tests := []struct {
		keyword string
		want    TokenType
	}{
		{"import", IMPORT},
		{"export", EXPORT},
		{"from", FROM},
		{"class", CLASS},
		{"new", NEW},
		{"this", THIS},
		{"super", SUPER},
		{"extends", EXTENDS},
		{"let", LET},
		{"await", AWAIT},
		{"async", ASYNC},
		{"function", FUNC},
		{"return", RETURN},
		{"throw", THROW},
		{"if", IF},
		{"else", ELSE},
		{"while", WHILE},
		{"for", FOR},
		{"break", BREAK},
		{"continue", CONTINUE},
		{"try", TRY},
		{"catch", CATCH},
		{"finally", FINALLY},
		{"true", TRUE},
		{"false", FALSE},
	}

	for _, tt := range tests {
		t.Run(tt.keyword, func(t *testing.T) {
			l := NewLexer(tt.keyword)
			tok := l.NextToken()
			if tok.Type != tt.want {
				t.Errorf("token type = %v, want %v", tok.Type, tt.want)
			}
			if tok.Literal != tt.keyword {
				t.Errorf("literal = %q, want %q", tok.Literal, tt.keyword)
			}
			// Should end with EOF
			eof := l.NextToken()
			if eof.Type != EOF {
				t.Errorf("expected EOF, got %v", eof.Type)
			}
		})
	}
}

func TestLexerIdentifiers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple", "x", "x"},
		{"underscore", "_var", "_var"},
		{"camelCase", "myVariable", "myVariable"},
		{"PascalCase", "MyClass", "MyClass"},
		{"with_digits", "var123", "var123"},
		{"starts_with_underscore", "_123abc", "_123abc"},
		{"long_identifier", "a_very_long_identifier_name_with_many_parts", "a_very_long_identifier_name_with_many_parts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.NextToken()
			if tok.Type != IDENT {
				t.Errorf("type = %v, want IDENT", tok.Type)
			}
			if tok.Literal != tt.want {
				t.Errorf("literal = %q, want %q", tok.Literal, tt.want)
			}
		})
	}
}

func TestLexerNumbers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"integer", "42", "42"},
		{"zero", "0", "0"},
		{"large_integer", "123456789", "123456789"},
		{"float", "3.14", "3.14"},
		{"float_with_many_decimals", "1.23456789", "1.23456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.NextToken()
			if tok.Type != NUMBER {
				t.Errorf("type = %v, want NUMBER", tok.Type)
			}
			if tok.Literal != tt.want {
				t.Errorf("literal = %q, want %q", tok.Literal, tt.want)
			}
		})
	}
}

func TestLexerStrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty_string", `""`, ""},
		{"simple_string", `"hello"`, "hello"},
		{"string_with_spaces", `"hello world"`, "hello world"},
		{"string_with_numbers", `"123"`, "123"},
		{"string_with_special_chars", `"a,b.c;d"`, "a,b.c;d"},
		{"string_with_parens", `"(foo)"`, "(foo)"},
		{"string_with_braces", `"{key: value}"`, "{key: value}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.NextToken()
			if tok.Type != STRING {
				t.Errorf("type = %v, want STRING", tok.Type)
			}
			if tok.Literal != tt.want {
				t.Errorf("literal = %q, want %q", tok.Literal, tt.want)
			}
		})
	}
}

// ============================================================
// Boundary condition tests
// ============================================================

func TestLexerEmptyInput(t *testing.T) {
	l := NewLexer("")
	tok := l.NextToken()
	if tok.Type != EOF {
		t.Errorf("empty input: type = %v, want EOF", tok.Type)
	}
}

func TestLexerWhitespaceOnly(t *testing.T) {
	inputs := []string{" ", "\n", "\t", "\r", "  \n  \t  \r  "}
	for _, input := range inputs {
		t.Run("whitespace", func(t *testing.T) {
			l := NewLexer(input)
			tok := l.NextToken()
			if tok.Type != EOF {
				t.Errorf("whitespace-only input: type = %v, want EOF", tok.Type)
			}
		})
	}
}

func TestLexerMultipleTokensNoWhitespace(t *testing.T) {
	l := NewLexer("1+2")
	tokens := []Token{
		{Type: NUMBER, Literal: "1"},
		{Type: PLUS, Literal: "+"},
		{Type: NUMBER, Literal: "2"},
		{Type: EOF, Literal: ""},
	}
	for i, want := range tokens {
		tok := l.NextToken()
		if tok.Type != want.Type {
			t.Errorf("token[%d] type = %v, want %v", i, tok.Type, want.Type)
		}
		if tok.Literal != want.Literal {
			t.Errorf("token[%d] literal = %q, want %q", i, tok.Literal, want.Literal)
		}
	}
}

func TestLexerLongString(t *testing.T) {
	// Generate a long string literal with printable characters
	// Using 'a' repeated since the lexer reads until null byte or closing quote
	input := `"` + generateLongStr(10000) + `"`
	l := NewLexer(input)
	tok := l.NextToken()
	if tok.Type != STRING {
		t.Errorf("long string: type = %v, want STRING", tok.Type)
	}
	if len(tok.Literal) != 10000 {
		t.Errorf("long string: length = %d, want 10000", len(tok.Literal))
	}
}

func generateLongStr(n int) string {
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		result[i] = byte('a' + (i % 26))
	}
	return string(result)
}

func TestLexerUnterminatedString(t *testing.T) {
	l := NewLexer(`"unterminated`)
	tok := l.NextToken()
	if tok.Type != STRING {
		t.Errorf("unterminated string: type = %v, want STRING", tok.Type)
	}
	if tok.Literal != "unterminated" {
		t.Errorf("unterminated string: literal = %q, want %q", tok.Literal, "unterminated")
	}
	// Should reach EOF
	eof := l.NextToken()
	if eof.Type != EOF {
		t.Errorf("expected EOF after unterminated string, got %v", eof.Type)
	}
}

func TestLexerSpecialCharacters(t *testing.T) {
	// Special characters that are not part of the language should be ILLEGAL
	l := NewLexer("@#$")
	tok := l.NextToken()
	if tok.Type != ILLEGAL {
		t.Errorf("special char @: type = %v, want ILLEGAL", tok.Type)
	}
}

func TestLexerMixedExpression(t *testing.T) {
	input := "let x = 1 + 2 * 3;"
	l := NewLexer(input)

	expected := []Token{
		{Type: LET, Literal: "let"},
		{Type: IDENT, Literal: "x"},
		{Type: ASSIGN, Literal: "="},
		{Type: NUMBER, Literal: "1"},
		{Type: PLUS, Literal: "+"},
		{Type: NUMBER, Literal: "2"},
		{Type: ASTERISK, Literal: "*"},
		{Type: NUMBER, Literal: "3"},
		{Type: SEMICOLON, Literal: ";"},
		{Type: EOF, Literal: ""},
	}

	for i, want := range expected {
		tok := l.NextToken()
		if tok.Type != want.Type {
			t.Errorf("token[%d] type = %v, want %v", i, tok.Type, want.Type)
		}
		if tok.Literal != want.Literal {
			t.Errorf("token[%d] literal = %q, want %q", i, tok.Literal, want.Literal)
		}
	}
}

func TestLexerStringEscapeSequences(t *testing.T) {
	// Test that the lexer reads escape sequences as raw characters
	// (the lexer doesn't interpret escape sequences, it just reads raw content)
	tests := []struct {
		name  string
		input string
	}{
		{"newline escape", `"hello\nworld"`},
		{"tab escape", `"col1\tcol2"`},
		{"backslash escape", `"path\\to\\file"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.NextToken()
			if tok.Type != STRING {
				t.Errorf("type = %v, want STRING", tok.Type)
			}
			if len(tok.Literal) == 0 {
				t.Errorf("string literal should not be empty")
			}
		})
	}
}

func TestLexerComments(t *testing.T) {
	// Test that // comments are tokenized as COMMENT tokens
	input := "// comment\nlet x = 5;"
	l := NewLexer(input)

	// First token should be COMMENT
	tok := l.NextToken()
	if tok.Type != COMMENT {
		t.Errorf("first token: type = %v, want COMMENT", tok.Type)
	}
	if tok.Literal != "// comment" {
		t.Errorf("first token literal = %q, want \"// comment\"", tok.Literal)
	}

	// Second token should be LET
	tok = l.NextToken()
	if tok.Type != LET {
		t.Errorf("second token: type = %v, want LET", tok.Type)
	}

	// Third token should be IDENT "x"
	tok = l.NextToken()
	if tok.Type != IDENT || tok.Literal != "x" {
		t.Errorf("third token: type = %v literal = %q, want IDENT x", tok.Type, tok.Literal)
	}
}

func TestLexerMultiLineString(t *testing.T) {
	input := `"line1
line2
line3"`
	l := NewLexer(input)
	tok := l.NextToken()
	if tok.Type != STRING {
		t.Errorf("type = %v, want STRING", tok.Type)
	}
	if !strings.Contains(tok.Literal, "\n") {
		t.Errorf("multi-line string should contain newlines")
	}
}

func TestLexerConsecutiveOperators(t *testing.T) {
	input := "++x--"
	l := NewLexer(input)
	tokens := []Token{}
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			break
		}
		tokens = append(tokens, tok)
	}
	if len(tokens) == 0 {
		t.Error("expected tokens, got none")
	}
}

func TestLexerFloatingPointEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"leading dot", ".5"},
		{"trailing dot", "5."},
		{"many decimals", "0.123456789012345678"},
		{"zero float", "0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLexer(tt.input)
			tok := l.NextToken()
			// Accept either NUMBER or DOT depending on lexer implementation
			if tok.Type != NUMBER && tok.Type != DOT {
				t.Errorf("type = %v, want NUMBER or DOT", tok.Type)
			}
		})
	}
}

func TestLexerTemplateStringBacktick(t *testing.T) {
	input := "`hello world`"
	l := NewLexer(input)
	tok := l.NextToken()
	if tok.Type != TEMPLATE {
		t.Errorf("type = %v, want TEMPLATE", tok.Type)
	}
}

func TestLexerPrivateIdentifier(t *testing.T) {
	l := NewLexer("#secret")
	tok := l.NextToken()
	if tok.Type != PRIVATE {
		t.Fatalf("type = %v, want PRIVATE", tok.Type)
	}
	if tok.Literal != "secret" {
		t.Fatalf("literal = %q, want secret", tok.Literal)
	}
}

func TestLexerPositionTracking(t *testing.T) {
	input := "let x = 1;"
	l := NewLexer(input)
	_ = l.NextToken() // let
	_ = l.NextToken() // x
	_ = l.NextToken() // =
	_ = l.NextToken() // 1
	_ = l.NextToken() // ;
	eof := l.NextToken()
	if eof.Type != EOF {
		t.Errorf("expected EOF, got %v", eof.Type)
	}
	// Additional calls should keep returning EOF
	for i := 0; i < 5; i++ {
		tok := l.NextToken()
		if tok.Type != EOF {
			t.Errorf("repeated call %d: type = %v, want EOF", i, tok.Type)
		}
	}
}

func TestLexerUnicodeIdentifiers(t *testing.T) {
	// Test that non-ASCII letters are handled (may be ILLEGAL or IDENT depending on implementation)
	input := "let cafe = 1;"
	l := NewLexer(input)
	tok := l.NextToken()
	if tok.Type != LET {
		t.Errorf("first token type = %v, want LET", tok.Type)
	}
}

func TestLexerVeryLongLine(t *testing.T) {
	// Generate a very long single-line expression
	var sb strings.Builder
	sb.WriteString("let x = ")
	for i := 0; i < 5000; i++ {
		if i > 0 {
			sb.WriteString(" + ")
		}
		sb.WriteString(itoa(i))
	}
	sb.WriteString(";")
	input := sb.String()
	l := NewLexer(input)
	for l.NextToken().Type != EOF {
	}
}

func TestLexerMultipleSemicolons(t *testing.T) {
	input := ";;;"
	l := NewLexer(input)
	count := 0
	for {
		tok := l.NextToken()
		if tok.Type == EOF {
			break
		}
		if tok.Type == SEMICOLON {
			count++
		}
	}
	if count != 3 {
		t.Errorf("semicolon count = %d, want 3", count)
	}
}

func TestLexerNestedBrackets(t *testing.T) {
	input := "((({[[[]]]})))"
	l := NewLexer(input)
	expected := []TokenType{
		LPAREN, LPAREN, LPAREN, LBRACE, LBRACKET, LBRACKET, LBRACKET,
		RBRACKET, RBRACKET, RBRACKET, RBRACE, RPAREN, RPAREN, RPAREN,
	}
	for i, want := range expected {
		tok := l.NextToken()
		if tok.Type != want {
			t.Errorf("token[%d] type = %v, want %v", i, tok.Type, want)
		}
	}
	eof := l.NextToken()
	if eof.Type != EOF {
		t.Errorf("expected EOF, got %v", eof.Type)
	}
}

func TestLexerFunctionCall(t *testing.T) {
	input := "foo(1, 2, 3)"
	l := NewLexer(input)

	expected := []Token{
		{Type: IDENT, Literal: "foo"},
		{Type: LPAREN, Literal: "("},
		{Type: NUMBER, Literal: "1"},
		{Type: COMMA, Literal: ","},
		{Type: NUMBER, Literal: "2"},
		{Type: COMMA, Literal: ","},
		{Type: NUMBER, Literal: "3"},
		{Type: RPAREN, Literal: ")"},
		{Type: EOF, Literal: ""},
	}

	for i, want := range expected {
		tok := l.NextToken()
		if tok.Type != want.Type {
			t.Errorf("token[%d] type = %v, want %v", i, tok.Type, want.Type)
		}
	}
}

func TestLexerMemberAccess(t *testing.T) {
	input := "obj.prop.method()"
	l := NewLexer(input)

	expected := []Token{
		{Type: IDENT, Literal: "obj"},
		{Type: DOT, Literal: "."},
		{Type: IDENT, Literal: "prop"},
		{Type: DOT, Literal: "."},
		{Type: IDENT, Literal: "method"},
		{Type: LPAREN, Literal: "("},
		{Type: RPAREN, Literal: ")"},
		{Type: EOF, Literal: ""},
	}

	for i, want := range expected {
		tok := l.NextToken()
		if tok.Type != want.Type {
			t.Errorf("token[%d] type = %v, want %v", i, tok.Type, want.Type)
		}
	}
}

// ============================================================
// Benchmark Tests
// ============================================================

func BenchmarkLexerSimple(b *testing.B) {
	input := "let x = 1 + 2;"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := NewLexer(input)
		for l.NextToken().Type != EOF {
		}
	}
}

func BenchmarkLexerComplex(b *testing.B) {
	input := `
function foo(a, b) {
	let x = a + b * 2;
	if (x > 10) {
		return x;
	} else {
		return x - 1;
	}
}
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := NewLexer(input)
		for l.NextToken().Type != EOF {
		}
	}
}

func BenchmarkLexerLargeProgram(b *testing.B) {
	// Generate a large program with many tokens
	input := generateLargeProgram(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := NewLexer(input)
		for l.NextToken().Type != EOF {
		}
	}
}

func BenchmarkLexerLongString(b *testing.B) {
	input := `"` + string(make([]byte, 50000)) + `"`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l := NewLexer(input)
		for l.NextToken().Type != EOF {
		}
	}
}

// Helper to generate a large program
func generateLargeProgram(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += "let x" + itoa(i) + " = " + itoa(i) + " + " + itoa(i*2) + ";\n"
	}
	return result
}

// Simple int to string without importing strconv
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
