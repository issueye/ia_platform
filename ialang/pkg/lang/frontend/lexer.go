package frontend

import "strings"

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
	nextLine     int
	nextColumn   int
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:      input,
		nextLine:   1,
		nextColumn: 1,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.position = l.readPosition
		l.ch = 0
		l.line = l.nextLine
		l.column = l.nextColumn
		return
	}

	ch := l.input[l.readPosition]
	l.ch = ch
	l.position = l.readPosition
	l.line = l.nextLine
	l.column = l.nextColumn
	l.readPosition++

	if ch == '\n' {
		l.nextLine++
		l.nextColumn = 1
	} else {
		l.nextColumn++
	}
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	startLine := l.line
	startColumn := l.column
	var tok Token
	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: EQ, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: ARROW, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else {
			tok = Token{Type: ASSIGN, Literal: string(l.ch), Line: startLine, Column: startColumn}
		}
	case '+':
		if l.peekChar() == '+' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: PLUSPLUS, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: PLUSEQ, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else {
			tok = Token{Type: PLUS, Literal: string(l.ch), Line: startLine, Column: startColumn}
		}
	case '-':
		if l.peekChar() == '-' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: MINUSMINUS, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: MINUSEQ, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else {
			tok = Token{Type: MINUS, Literal: string(l.ch), Line: startLine, Column: startColumn}
		}
	case '*':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: MULTEQ, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else {
			tok = Token{Type: ASTERISK, Literal: string(l.ch), Line: startLine, Column: startColumn}
		}
	case '/':
		if l.peekChar() == '/' {
			// Collect // comment
			startLine := l.line
			startColumn := l.column
			var comment strings.Builder
			comment.WriteByte(l.ch) // first '/'
			l.readChar()            // move to second '/'
			comment.WriteByte(l.ch) // second '/'
			l.readChar()            // move past comment start
			for l.ch != '\n' && l.ch != 0 {
				comment.WriteByte(l.ch)
				l.readChar()
			}
			tok = Token{Type: COMMENT, Literal: comment.String(), Line: startLine, Column: startColumn}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: DIVEQ, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else {
			tok = Token{Type: SLASH, Literal: string(l.ch), Line: startLine, Column: startColumn}
		}
	case '%':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: MODEQ, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else {
			tok = Token{Type: MODULO, Literal: string(l.ch), Line: startLine, Column: startColumn}
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NEQ, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else {
			tok = Token{Type: BANG, Literal: string(l.ch), Line: startLine, Column: startColumn}
		}
	case '?':
		if l.peekChar() == '?' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: NULLISH, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else if l.peekChar() == '.' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: OPTCHAIN, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else {
			tok = Token{Type: QUESTION, Literal: string(l.ch), Line: startLine, Column: startColumn}
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: AND, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else {
			tok = Token{Type: BITAND, Literal: string(l.ch), Line: startLine, Column: startColumn}
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: OR, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else {
			tok = Token{Type: BITOR, Literal: string(l.ch), Line: startLine, Column: startColumn}
		}
	case '^':
		tok = Token{Type: BITXOR, Literal: string(l.ch), Line: startLine, Column: startColumn}
	case '<':
		if l.peekChar() == '<' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: SHL, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: LTE, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else {
			tok = Token{Type: LT, Literal: string(l.ch), Line: startLine, Column: startColumn}
		}
	case '>':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: SHR, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			tok = Token{Type: GTE, Literal: string(ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else {
			tok = Token{Type: GT, Literal: string(l.ch), Line: startLine, Column: startColumn}
		}
	case ',':
		tok = Token{Type: COMMA, Literal: string(l.ch), Line: startLine, Column: startColumn}
	case '#':
		if isLetter(l.peekChar()) {
			l.readChar() // move to first letter after '#'
			tok = Token{Type: PRIVATE, Literal: l.readIdentifier(), Line: startLine, Column: startColumn}
			return tok
		}
		tok = Token{Type: ILLEGAL, Literal: string(l.ch), Line: startLine, Column: startColumn}
	case ':':
		tok = Token{Type: COLON, Literal: string(l.ch), Line: startLine, Column: startColumn}
	case '.':
		if l.peekChar() == '.' && l.peekChar2() == '.' {
			ch := l.ch
			l.readChar()
			l.readChar()
			tok = Token{Type: SPREAD, Literal: string(ch) + string(l.ch) + string(l.ch), Line: startLine, Column: startColumn}
		} else {
			tok = Token{Type: DOT, Literal: string(l.ch), Line: startLine, Column: startColumn}
		}
	case ';':
		tok = Token{Type: SEMICOLON, Literal: string(l.ch), Line: startLine, Column: startColumn}
	case '(':
		tok = Token{Type: LPAREN, Literal: string(l.ch), Line: startLine, Column: startColumn}
	case ')':
		tok = Token{Type: RPAREN, Literal: string(l.ch), Line: startLine, Column: startColumn}
	case '{':
		tok = Token{Type: LBRACE, Literal: string(l.ch), Line: startLine, Column: startColumn}
	case '}':
		tok = Token{Type: RBRACE, Literal: string(l.ch), Line: startLine, Column: startColumn}
	case '[':
		tok = Token{Type: LBRACKET, Literal: string(l.ch), Line: startLine, Column: startColumn}
	case ']':
		tok = Token{Type: RBRACKET, Literal: string(l.ch), Line: startLine, Column: startColumn}
	case '"':
		tok.Type = STRING
		tok.Literal = l.readString()
		tok.Line = startLine
		tok.Column = startColumn
		return tok
	case '`':
		tok.Type = TEMPLATE
		tok.Literal = l.readRawString('`')
		tok.Line = startLine
		tok.Column = startColumn
		return tok
	case 0:
		tok = Token{Type: EOF, Literal: "", Line: startLine, Column: startColumn}
	default:
		if isLetter(l.ch) {
			lit := l.readIdentifier()
			return Token{Type: lookupIdent(lit), Literal: lit, Line: startLine, Column: startColumn}
		}
		if isDigit(l.ch) {
			return Token{Type: NUMBER, Literal: l.readNumber(), Line: startLine, Column: startColumn}
		}
		tok = Token{Type: ILLEGAL, Literal: string(l.ch), Line: startLine, Column: startColumn}
	}

	l.readChar()
	return tok
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) peekChar2() byte {
	if l.readPosition+1 >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition+1]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\n' || l.ch == '\r' || l.ch == '\t' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	start := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readNumber() string {
	start := l.position
	hasDot := false
	for isDigit(l.ch) || (!hasDot && l.ch == '.') {
		if l.ch == '.' {
			hasDot = true
		}
		l.readChar()
	}
	return l.input[start:l.position]
}

func (l *Lexer) readString() string {
	l.readChar()
	start := l.position
	for l.ch != '"' && l.ch != 0 {
		l.readChar()
	}
	lit := l.input[start:l.position]
	if l.ch == '"' {
		l.readChar()
	}
	return lit
}

func (l *Lexer) readRawString(quote byte) string {
	l.readChar()
	start := l.position
	for l.ch != quote && l.ch != 0 {
		l.readChar()
	}
	lit := l.input[start:l.position]
	if l.ch == quote {
		l.readChar()
	}
	return lit
}

func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}
