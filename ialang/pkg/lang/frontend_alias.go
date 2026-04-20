package lang

import "ialang/pkg/lang/frontend"

type Lexer = frontend.Lexer
type Parser = frontend.Parser

func NewLexer(input string) *Lexer {
	return frontend.NewLexer(input)
}

func NewParser(l *Lexer) *Parser {
	return frontend.NewParser(l)
}
