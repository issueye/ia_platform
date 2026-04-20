package frontend

import (
	"fmt"
	"ialang/pkg/lang/ast"
	"ialang/pkg/lang/token"
	"strings"
)

const (
	_ int = iota
	LOWEST
	TERTIARY // ?: 三元运算符
	NULLISHPRECEDENCE
	LOGICALOR
	LOGICALAND
	EQUALS
	COMPARE
	SUM
	PRODUCT // * / %
	PREFIX
	CALL
	MEMBER
	INDEX
)

var precedences = map[TokenType]int{
	QUESTION: TERTIARY,
	NULLISH:  NULLISHPRECEDENCE,
	OR:       LOGICALOR,
	AND:      LOGICALAND,
	BITOR:    SUM,
	BITXOR:   SUM,
	BITAND:   SUM,
	SHL:      PRODUCT,
	SHR:      PRODUCT,
	EQ:       EQUALS,
	NEQ:      EQUALS,
	LT:       COMPARE,
	GT:       COMPARE,
	LTE:      COMPARE,
	GTE:      COMPARE,
	PLUS:     SUM,
	MINUS:    SUM,
	ASTERISK: PRODUCT,
	SLASH:    PRODUCT,
	MODULO:   PRODUCT,
	LPAREN:   CALL,
	DOT:      MEMBER,
	OPTCHAIN: MEMBER,
	LBRACKET: INDEX,
}

type (
	prefixParseFn func() Expression
	infixParseFn  func(Expression) Expression
)

type Parser struct {
	l      *Lexer
	errors []string

	curToken  Token
	peekToken Token

	prefixParseFns map[TokenType]prefixParseFn
	infixParseFns  map[TokenType]infixParseFn
}

func nodeInfoFromToken(tok Token) NodeInfo {
	return NodeInfo{
		Start: Position{
			Line:   tok.Line,
			Column: tok.Column,
		},
	}
}

func nodeInfoFromNode(n Node) NodeInfo {
	if n == nil {
		return NodeInfo{}
	}
	return NodeInfo{Start: n.Pos()}
}

func (p *Parser) addErrorWithToken(tok Token, msg string) {
	if tok.Line > 0 && tok.Column > 0 {
		msg = fmt.Sprintf("%s (line %d, col %d)", msg, tok.Line, tok.Column)
	}
	p.errors = append(p.errors, msg)
}

func (p *Parser) addErrorAtPosition(pos Position, msg string) {
	if pos.IsValid() {
		msg = fmt.Sprintf("%s (line %d, col %d)", msg, pos.Line, pos.Column)
	}
	p.errors = append(p.errors, msg)
}

func (p *Parser) addError(msg string) {
	p.addErrorWithToken(p.curToken, msg)
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}

	p.prefixParseFns = map[TokenType]prefixParseFn{
		IDENT:      p.parseIdentifier,
		THIS:       p.parseIdentifier,
		SUPER:      p.parseSuperExpression,
		NUMBER:     p.parseNumberLiteral,
		STRING:     p.parseStringLiteral,
		TEMPLATE:   p.parseTemplateLiteral,
		LPAREN:     p.parseParenExpression,
		NEW:        p.parseNewExpression,
		AWAIT:      p.parseAwaitExpression,
		IMPORT:     p.parseDynamicImportExpression,
		FUNC:       p.parseFunctionExpression,
		ASYNC:      p.parseAsyncExpression,
		BANG:       p.parseUnaryExpression,
		MINUS:      p.parseUnaryExpression,
		PLUSPLUS:   p.parseUpdateExpression,
		MINUSMINUS: p.parseUpdateExpression,
		TRUE:       p.parseBoolLiteral,
		FALSE:      p.parseBoolLiteral,
		NULL:       p.parseNullLiteral,
		LBRACKET:   p.parseArrayLiteral,
		LBRACE:     p.parseObjectLiteral,
		TYPEOF:     p.parseTypeofExpression,
		VOID:       p.parseVoidExpression,
	}

	p.infixParseFns = map[TokenType]infixParseFn{
		PLUS:     p.parseBinaryExpression,
		MINUS:    p.parseBinaryExpression,
		ASTERISK: p.parseBinaryExpression,
		SLASH:    p.parseBinaryExpression,
		MODULO:   p.parseBinaryExpression,
		BITAND:   p.parseBinaryExpression,
		BITOR:    p.parseBinaryExpression,
		BITXOR:   p.parseBinaryExpression,
		SHL:      p.parseBinaryExpression,
		SHR:      p.parseBinaryExpression,
		EQ:       p.parseBinaryExpression,
		NEQ:      p.parseBinaryExpression,
		AND:      p.parseBinaryExpression,
		OR:       p.parseBinaryExpression,
		NULLISH:  p.parseBinaryExpression,
		LT:       p.parseBinaryExpression,
		GT:       p.parseBinaryExpression,
		LTE:      p.parseBinaryExpression,
		GTE:      p.parseBinaryExpression,
		QUESTION: p.parseTernaryExpression,
		LPAREN:   p.parseCallExpression,
		DOT:      p.parseGetExpression,
		LBRACKET: p.parseIndexExpression,
		OPTCHAIN: p.parseOptionalChainExpression,
	}

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() *Program {
	program := &Program{NodeInfo: nodeInfoFromToken(p.curToken)}
	var pendingComments []*ast.Comment

	for p.curToken.Type != EOF {
		// Collect leading comments
		for p.curToken.Type == COMMENT {
			pendingComments = append(pendingComments, &ast.Comment{
				NodeInfo: nodeInfoFromToken(p.curToken),
				Text:     p.curToken.Literal,
			})
			p.nextToken()
		}
		if p.curToken.Type == EOF {
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			// Attach pending comments to this statement
			stmt.SetLeadingComments(pendingComments)
			pendingComments = nil
			program.Statements = append(program.Statements, stmt)
		} else {
			// If no statement was parsed but we have comments, discard them
			// (they were likely trailing comments at end of file)
			pendingComments = nil
		}
		p.nextToken()
	}
	return program
}

func (p *Parser) parseStatement() Statement {
	switch p.curToken.Type {
	case IMPORT:
		if p.peekToken.Type == LPAREN {
			return p.parseExpressionStatement()
		}
		return p.parseImportStatement()
	case EXPORT:
		return p.parseExportStatement()
	case CLASS:
		return p.parseClassStatement()
	case LET:
		return p.parseLetStatement()
	case FUNC:
		return p.parseFunctionStatement(false)
	case ASYNC:
		return p.parseAsyncFunctionStatement()
	case RETURN:
		return p.parseReturnStatement()
	case THROW:
		return p.parseThrowStatement()
	case IF:
		return p.parseIfStatement()
	case WHILE:
		return p.parseWhileStatement()
	case DO:
		return p.parseDoWhileStatement()
	case FOR:
		return p.parseForStatement()
	case TRY:
		return p.parseTryCatchStatement()
	case BREAK:
		return p.parseBreakStatement()
	case CONTINUE:
		return p.parseContinueStatement()
	case SWITCH:
		return p.parseSwitchStatement()
	case IDENT:
		return p.parseAssignmentOrExpressionStatement()
	case THIS:
		return p.parseAssignmentOrExpressionStatement()
	case LPAREN:
		return p.parseAssignmentOrExpressionStatement()
	case LBRACKET:
		return p.parseAssignmentOrExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseExportStatement() Statement {
	startTok := p.curToken
	p.nextToken()
	stmt := &ExportStatement{NodeInfo: nodeInfoFromToken(startTok)}
	switch p.curToken.Type {
	case LET:
		stmt.Statement = p.parseLetStatement()
		return stmt
	case FUNC:
		stmt.Statement = p.parseFunctionStatement(false)
		return stmt
	case ASYNC:
		stmt.Statement = p.parseAsyncFunctionStatement()
		return stmt
	case CLASS:
		stmt.Statement = p.parseClassStatement()
		return stmt
	case DEFAULT:
		p.nextToken()
		switch p.curToken.Type {
		case CLASS:
			classStmt := p.parseClassStatement()
			cs, ok := classStmt.(*ClassStatement)
			if !ok || cs == nil || cs.Name == "" {
				p.addError("export default class requires named class declaration")
				return nil
			}
			stmt.Statement = cs
			stmt.DefaultName = cs.Name
			return stmt
		case FUNC:
			fnStmt := p.parseFunctionStatement(false)
			fs, ok := fnStmt.(*FunctionStatement)
			if !ok || fs == nil || fs.Name == "" {
				p.addError("export default function requires named function declaration")
				return nil
			}
			stmt.Statement = fs
			stmt.DefaultName = fs.Name
			return stmt
		case ASYNC:
			fnStmt := p.parseAsyncFunctionStatement()
			fs, ok := fnStmt.(*FunctionStatement)
			if !ok || fs == nil || fs.Name == "" {
				p.addError("export default async function requires named function declaration")
				return nil
			}
			stmt.Statement = fs
			stmt.DefaultName = fs.Name
			return stmt
		}
		stmt.Default = p.parseExpression(LOWEST)
		if stmt.Default == nil {
			return nil
		}
		if p.peekToken.Type == SEMICOLON {
			p.nextToken()
		}
		return stmt
	case ASTERISK:
		if !p.expectPeek(FROM) {
			return nil
		}
		if !p.expectPeek(STRING) {
			return nil
		}
		stmt.ExportAllModule = p.curToken.Literal
		if p.peekToken.Type == SEMICOLON {
			p.nextToken()
		}
		return stmt
	case LBRACE:
		stmt.Specifiers = p.parseExportSpecifiers()
		if stmt.Specifiers == nil {
			return nil
		}
		if p.peekToken.Type == SEMICOLON {
			p.nextToken()
		}
		return stmt
	default:
		p.addError(fmt.Sprintf("unsupported export target: %s", p.curToken.Type))
		return nil
	}
}

func (p *Parser) parseExportSpecifiers() []ExportSpecifier {
	specifiers := []ExportSpecifier{}
	if p.peekToken.Type == RBRACE {
		p.nextToken()
		return specifiers
	}

	for {
		p.nextToken()
		if p.curToken.Type != IDENT {
			p.addError(fmt.Sprintf("expected identifier in export list, got %s", p.curToken.Type))
			return nil
		}
		localName := p.curToken.Literal
		exportName := localName

		if p.peekToken.Type == IDENT && p.peekToken.Literal == "as" {
			p.nextToken() // consume "as"
			if !p.expectPeek(IDENT) {
				return nil
			}
			exportName = p.curToken.Literal
		}

		specifiers = append(specifiers, ExportSpecifier{
			Pos:        Position{Line: p.curToken.Line, Column: p.curToken.Column},
			LocalName:  localName,
			ExportName: exportName,
		})

		if p.peekToken.Type == COMMA {
			p.nextToken()
			if p.peekToken.Type == RBRACE {
				p.nextToken()
				return specifiers
			}
			continue
		}
		break
	}

	if !p.expectPeek(RBRACE) {
		return nil
	}
	return specifiers
}

func (p *Parser) parseClassStatement() Statement {
	stmt := &ClassStatement{NodeInfo: nodeInfoFromToken(p.curToken)}
	if !p.expectPeek(IDENT) {
		return nil
	}
	stmt.Name = p.curToken.Literal

	if p.peekToken.Type == EXTENDS {
		p.nextToken()
		if !p.expectPeek(IDENT) {
			return nil
		}
		stmt.ParentName = p.curToken.Literal
	}

	if !p.expectPeek(LBRACE) {
		return nil
	}

	p.nextToken()
	for p.curToken.Type != RBRACE && p.curToken.Type != EOF {
		if p.curToken.Type == COMMENT {
			p.nextToken()
			continue
		}

		if p.curToken.Type == PRIVATE {
			stmt.PrivateFields = append(stmt.PrivateFields, ClassPrivateField{
				Pos:  Position{Line: p.curToken.Line, Column: p.curToken.Column},
				Name: manglePrivateName(p.curToken.Literal),
			})
			if p.peekToken.Type == ASSIGN {
				p.addError("private field initializer is not supported yet")
				return nil
			}
			if p.peekToken.Type == SEMICOLON {
				p.nextToken()
			}
			p.nextToken()
			continue
		}

		method := ClassMethod{Pos: Position{Line: p.curToken.Line, Column: p.curToken.Column}}

		// Check for static/get/set modifiers
		switch p.curToken.Type {
		case STATIC:
			method.Static = true
			if !p.expectPeek(IDENT) {
				return nil
			}
		case GET:
			method.IsGetter = true
			if !p.expectPeek(IDENT) {
				return nil
			}
		case SET:
			method.IsSetter = true
			if !p.expectPeek(IDENT) {
				return nil
			}
		case ASYNC:
			method.Async = true
			if !p.expectPeek(IDENT) {
				return nil
			}
		}

		if p.curToken.Type != IDENT {
			p.addError(fmt.Sprintf("expected class method name, got %s", p.curToken.Type))
			return nil
		}
		method.Name = p.curToken.Literal

		// Getters have () with no parameters
		if method.IsGetter {
			if !p.expectPeek(LPAREN) {
				return nil
			}
			if !p.expectPeek(RPAREN) {
				return nil
			}
			if !p.expectPeek(LBRACE) {
				return nil
			}
			method.Body = p.parseBlockStatement()
			stmt.Methods = append(stmt.Methods, method)
			p.nextToken()
			continue
		}

		// Setters have one parameter
		if method.IsSetter {
			if !p.expectPeek(LPAREN) {
				return nil
			}
			// Parse single parameter
			if p.peekToken.Type != IDENT {
				p.addError("setter must have exactly one parameter")
				return nil
			}
			p.nextToken()
			method.Params = append(method.Params, p.curToken.Literal)
			method.ParamDefaults = append(method.ParamDefaults, DefaultParam{})
			if !p.expectPeek(RPAREN) {
				return nil
			}
			if !p.expectPeek(LBRACE) {
				return nil
			}
			method.Body = p.parseBlockStatement()
			stmt.Methods = append(stmt.Methods, method)
			p.nextToken()
			continue
		}

		// Regular methods (including static methods)
		if !p.expectPeek(LPAREN) {
			return nil
		}
		method.Params, method.ParamDefaults, method.RestParam = p.parseFunctionParams()
		if !p.expectPeek(LBRACE) {
			return nil
		}
		method.Body = p.parseBlockStatement()
		stmt.Methods = append(stmt.Methods, method)
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseImportStatement() Statement {
	stmt := &ImportStatement{NodeInfo: nodeInfoFromToken(p.curToken)}

	if !p.expectAnyPeek(LBRACE, ASTERISK) {
		return nil
	}

	if p.curToken.Type == ASTERISK {
		if p.peekToken.Type != IDENT || p.peekToken.Literal != "as" {
			p.addError("expected 'as' in namespace import")
			return nil
		}
		p.nextToken() // consume as
		if !p.expectPeek(IDENT) {
			return nil
		}
		stmt.Namespace = p.curToken.Literal
	} else {
		p.nextToken()
		for p.curToken.Type != RBRACE && p.curToken.Type != EOF {
			if p.curToken.Type != IDENT {
				p.addError(fmt.Sprintf("expected identifier in import, got %s", p.curToken.Type))
				return nil
			}
			stmt.Names = append(stmt.Names, p.curToken.Literal)

			if p.peekToken.Type == COMMA {
				p.nextToken()
				p.nextToken()
				continue
			}
			if p.peekToken.Type == RBRACE {
				p.nextToken()
			}
		}
	}

	if !p.expectPeek(FROM) {
		return nil
	}
	if !p.expectPeek(STRING) {
		return nil
	}
	stmt.Module = p.curToken.Literal

	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseLetStatement() Statement {
	startTok := p.curToken
	if !p.expectAnyPeek(IDENT, LBRACKET, LBRACE) {
		return nil
	}

	switch p.curToken.Type {
	case IDENT:
		stmt := &LetStatement{
			NodeInfo: nodeInfoFromToken(startTok),
			Name:     p.curToken.Literal,
		}
		if !p.expectPeek(ASSIGN) {
			return nil
		}
		p.nextToken()
		stmt.Initializer = p.parseExpression(LOWEST)
		if p.peekToken.Type == SEMICOLON {
			p.nextToken()
		}
		return stmt
	case LBRACKET:
		return p.parseArrayDestructuringLetStatement(startTok)
	case LBRACE:
		return p.parseObjectDestructuringLetStatement(startTok)
	default:
		p.addError(fmt.Sprintf("unsupported let target: %s", p.curToken.Type))
		return nil
	}
}

func (p *Parser) parseArrayDestructuringLetStatement(startTok Token) Statement {
	stmt := &ArrayDestructuringLetStatement{NodeInfo: nodeInfoFromToken(startTok)}
	if p.peekToken.Type == RBRACKET {
		p.addError("array destructuring requires at least one binding")
		return nil
	}

	for {
		p.nextToken()
		if p.curToken.Type != IDENT {
			p.addError(fmt.Sprintf("expected identifier in array destructuring, got %s", p.curToken.Type))
			return nil
		}
		stmt.Names = append(stmt.Names, p.curToken.Literal)
		if p.peekToken.Type != COMMA {
			break
		}
		p.nextToken()
		if p.peekToken.Type == RBRACKET {
			p.nextToken()
			break
		}
	}

	if p.curToken.Type != RBRACKET {
		if !p.expectPeek(RBRACKET) {
			return nil
		}
	}
	if !p.expectPeek(ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Initializer = p.parseExpression(LOWEST)
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseObjectDestructuringLetStatement(startTok Token) Statement {
	stmt := &ObjectDestructuringLetStatement{NodeInfo: nodeInfoFromToken(startTok)}
	if p.peekToken.Type == RBRACE {
		p.addError("object destructuring requires at least one binding")
		return nil
	}

	for {
		p.nextToken()
		if p.curToken.Type != IDENT && p.curToken.Type != STRING {
			p.addError(fmt.Sprintf("expected identifier or string in object destructuring, got %s", p.curToken.Type))
			return nil
		}
		key := p.curToken.Literal
		name := key
		pos := Position{Line: p.curToken.Line, Column: p.curToken.Column}

		if p.peekToken.Type == COLON {
			p.nextToken()
			if !p.expectPeek(IDENT) {
				return nil
			}
			name = p.curToken.Literal
			pos = Position{Line: p.curToken.Line, Column: p.curToken.Column}
		}
		stmt.Bindings = append(stmt.Bindings, ObjectDestructureBinding{
			Pos:  pos,
			Key:  key,
			Name: name,
		})

		if p.peekToken.Type != COMMA {
			break
		}
		p.nextToken()
		if p.peekToken.Type == RBRACE {
			p.nextToken()
			break
		}
	}

	if p.curToken.Type != RBRACE {
		if !p.expectPeek(RBRACE) {
			return nil
		}
	}
	if !p.expectPeek(ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Initializer = p.parseExpression(LOWEST)
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseExpressionStatement() Statement {
	stmt := &ExpressionStatement{
		NodeInfo: nodeInfoFromToken(p.curToken),
		Expr:     p.parseExpression(LOWEST),
	}
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseFunctionStatement(async bool) Statement {
	stmt := &FunctionStatement{NodeInfo: nodeInfoFromToken(p.curToken)}
	if !p.expectPeek(IDENT) {
		return nil
	}
	stmt.Name = p.curToken.Literal
	stmt.Async = async

	if !p.expectPeek(LPAREN) {
		return nil
	}
	stmt.Params, stmt.ParamDefaults, stmt.RestParam = p.parseFunctionParams()

	if !p.expectPeek(LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseAsyncFunctionStatement() Statement {
	startTok := p.curToken
	if !p.expectPeek(FUNC) {
		return nil
	}
	stmt := p.parseFunctionStatement(true)
	if fn, ok := stmt.(*FunctionStatement); ok {
		fn.NodeInfo = nodeInfoFromToken(startTok)
	}
	return stmt
}

func (p *Parser) parseFunctionExpression() Expression {
	startTok := p.curToken
	exp := &FunctionExpression{NodeInfo: nodeInfoFromToken(startTok)}

	if p.peekToken.Type == IDENT {
		p.nextToken()
		exp.Name = p.curToken.Literal
	}

	if !p.expectPeek(LPAREN) {
		return nil
	}
	exp.Params, exp.ParamDefaults, exp.RestParam = p.parseFunctionParams()

	if !p.expectPeek(LBRACE) {
		return nil
	}
	exp.Body = p.parseBlockStatement()
	return exp
}

func (p *Parser) parseAsyncFunctionExpression() Expression {
	startTok := p.curToken
	if !p.expectPeek(FUNC) {
		return nil
	}
	exp := p.parseFunctionExpression()
	if fn, ok := exp.(*FunctionExpression); ok {
		fn.NodeInfo = nodeInfoFromToken(startTok)
		fn.Async = true
	}
	return exp
}

// parseParenExpression handles both grouped expressions and arrow functions.
func (p *Parser) parseParenExpression() Expression {
	startTok := p.curToken

	// Parse content inside ()
	p.nextToken()

	// Empty parens: ()
	if p.curToken.Type == RPAREN {
		p.nextToken()
		if p.curToken.Type == ARROW {
			return p.finishArrowFunction(startTok, []string{}, nil, "", false)
		}
		p.addError("expected expression inside parentheses")
		return nil
	}

	// Try parsing as arrow params first (comma-separated identifiers)
	params, defaults, restParam := p.tryParseArrowParamsWithDefaults()

	if params != nil {
		// Successfully parsed as params
		if p.curToken.Type != RPAREN {
			if !p.expectPeek(RPAREN) {
				return nil
			}
		}

		if p.peekToken.Type == ARROW {
			p.nextToken()
			return p.finishArrowFunction(startTok, params, defaults, restParam, false)
		}

		// Not followed by => - if single param, treat as grouped expression
		if len(params) == 1 {
			return &Identifier{
				NodeInfo: nodeInfoFromToken(startTok),
				Name:     params[0],
			}
		}
		p.addError("expected '=>' after parameter list")
		return nil
	}

	// Not arrow params - parse as regular grouped expression
	innerExp := p.parseExpression(LOWEST)
	if innerExp == nil {
		return nil
	}
	if _, ok := innerExp.(*ObjectLiteral); ok && p.peekToken.Type == ASSIGN {
		// Allow parenthesized object destructuring assignment target:
		// ({x, y: z} = obj)
		// Keep ')' for parseAssignmentOrExpressionStatement to consume.
		return innerExp
	}

	if !p.expectPeek(RPAREN) {
		return nil
	}

	return innerExp
}

// tryParseArrowParams attempts to parse comma-separated identifiers as arrow function params.
// Returns params if successful, nil otherwise.
// After this call, p.curToken is at the last token consumed (either IDENT or RPAREN).
func (p *Parser) tryParseArrowParams() []string {
	params, _, _ := p.tryParseArrowParamsWithDefaults()
	return params
}

// tryParseArrowParamsWithDefaults is like tryParseArrowParams but also returns default values.
func (p *Parser) tryParseArrowParamsWithDefaults() ([]string, []DefaultParam, string) {
	params := []string{}
	defaults := []DefaultParam{}
	restParam := ""

	if p.curToken.Type == SPREAD {
		if !p.expectPeek(IDENT) {
			return nil, nil, ""
		}
		restParam = p.curToken.Literal
		if p.peekToken.Type != RPAREN {
			return nil, nil, ""
		}
		p.nextToken()
		return params, defaults, restParam
	}

	if p.curToken.Type != IDENT {
		return nil, nil, ""
	}

	for {
		if p.curToken.Type != IDENT {
			return nil, nil, ""
		}
		paramName := p.curToken.Literal
		params = append(params, paramName)
		defaults = append(defaults, DefaultParam{})

		if p.peekToken.Type == ASSIGN {
			p.nextToken()
			p.nextToken()
			defaults[len(defaults)-1] = DefaultParam{
				Pos:   Position{Line: p.curToken.Line, Column: p.curToken.Column},
				Name:  paramName,
				Value: p.parseExpression(LOWEST),
			}
			if p.curToken.Type == RPAREN {
				return params, defaults, restParam
			}
		}

		if p.peekToken.Type != COMMA {
			break
		}

		p.nextToken()
		p.nextToken()
		if p.curToken.Type == SPREAD {
			if !p.expectPeek(IDENT) {
				return nil, nil, ""
			}
			restParam = p.curToken.Literal
			if p.peekToken.Type != RPAREN {
				return nil, nil, ""
			}
			p.nextToken()
			return params, defaults, restParam
		}
	}

	if p.peekToken.Type != RPAREN {
		return nil, nil, ""
	}
	p.nextToken()
	return params, defaults, restParam
}

// parseAsyncExpression handles both async function expressions and async arrow functions.
func (p *Parser) parseAsyncExpression() Expression {
	startTok := p.curToken

	if p.peekToken.Type == FUNC {
		// async function ...
		p.nextToken()
		exp := p.parseFunctionExpression()
		if fn, ok := exp.(*FunctionExpression); ok {
			fn.NodeInfo = nodeInfoFromToken(startTok)
			fn.Async = true
		}
		return exp
	}

	// async (...) => ... - async arrow function
	if p.peekToken.Type == LPAREN {
		p.nextToken() // move to (
		p.nextToken() // move inside

		// Check for empty params: async ()
		if p.curToken.Type == RPAREN {
			p.nextToken()
			if p.curToken.Type == ARROW {
				return p.finishArrowFunction(startTok, []string{}, nil, "", true)
			}
			p.addError("expected '=>' after 'async ()'")
			return nil
		}

		// Try parsing as arrow params
		params, defaults, restParam := p.tryParseArrowParamsWithDefaults()
		if params == nil {
			p.addError("expected parameter list in async arrow function")
			return nil
		}

		if p.curToken.Type != ARROW {
			p.addError("expected '=>' after async parameter list")
			return nil
		}
		p.nextToken()

		return p.finishArrowFunction(startTok, params, defaults, restParam, true)
	}

	p.addError(fmt.Sprintf("expected 'function' or '(' after 'async', got %s", p.peekToken.Type))
	return nil
}

// finishArrowFunction creates an ArrowFunctionExpression after seeing =>.
// p.curToken is the ARROW token. It advances to parse the body.
func (p *Parser) finishArrowFunction(startTok Token, params []string, defaults []DefaultParam, restParam string, async bool) Expression {
	p.nextToken() // move past =>

	arrowExp := &ArrowFunctionExpression{
		NodeInfo:      nodeInfoFromToken(startTok),
		Params:        params,
		RestParam:     restParam,
		ParamDefaults: defaults,
		Async:         async,
	}

	// Check if concise body (single expression) or block body
	if p.curToken.Type == LBRACE {
		// Block body
		arrowExp.Body = p.parseBlockStatement()
		arrowExp.Concise = false
	} else {
		// Concise body - single expression
		arrowExp.Expr = p.parseExpression(LOWEST)
		arrowExp.Concise = true
	}

	return arrowExp
}

func (p *Parser) parseAssignStatement() Statement {
	stmt := &AssignStatement{
		NodeInfo: nodeInfoFromToken(p.curToken),
		Name:     p.curToken.Literal,
	}
	if !p.expectPeek(ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseAssignmentOrExpressionStatement() Statement {
	startedWithParen := p.curToken.Type == LPAREN
	left := p.parseExpression(LOWEST)
	leftNodeInfo := nodeInfoFromNode(left)

	peekType := p.peekToken.Type
	isCompoundAssign := peekType == PLUSEQ || peekType == MINUSEQ || peekType == MULTEQ || peekType == DIVEQ || peekType == MODEQ

	if p.peekToken.Type == ASSIGN || isCompoundAssign {
		p.nextToken()
		op := p.curToken.Type
		p.nextToken()
		value := p.parseExpression(LOWEST)
		if startedWithParen {
			if _, ok := left.(*ObjectLiteral); ok {
				if !p.expectPeek(RPAREN) {
					return nil
				}
			}
		}
		if p.peekToken.Type == SEMICOLON {
			p.nextToken()
		}

		switch l := left.(type) {
		case *Identifier:
			if isCompoundAssign {
				return &CompoundAssignStatement{
					NodeInfo: leftNodeInfo,
					Name:     l.Name,
					Operator: op,
					Value:    value,
				}
			}
			return &AssignStatement{NodeInfo: leftNodeInfo, Name: l.Name, Value: value}
		case *GetExpression:
			if isCompoundAssign {
				return &CompoundSetPropertyStatement{
					NodeInfo: leftNodeInfo,
					Object:   l.Object,
					Property: l.Property,
					Operator: op,
					Value:    value,
				}
			}
			return &SetPropertyStatement{
				NodeInfo: leftNodeInfo,
				Object:   l.Object,
				Property: l.Property,
				Value:    value,
			}
		case *ArrayLiteral:
			if isCompoundAssign {
				p.addErrorAtPosition(leftNodeInfo.Start, fmt.Sprintf("invalid assignment target: %T", left))
				return nil
			}
			names, ok := p.parseArrayDestructuringAssignNames(l)
			if !ok {
				return nil
			}
			return &ArrayDestructuringAssignStatement{
				NodeInfo: leftNodeInfo,
				Names:    names,
				Value:    value,
			}
		case *ObjectLiteral:
			if isCompoundAssign {
				p.addErrorAtPosition(leftNodeInfo.Start, fmt.Sprintf("invalid assignment target: %T", left))
				return nil
			}
			bindings, ok := p.parseObjectDestructuringAssignBindings(l)
			if !ok {
				return nil
			}
			return &ObjectDestructuringAssignStatement{
				NodeInfo: leftNodeInfo,
				Bindings: bindings,
				Value:    value,
			}
		default:
			p.addErrorAtPosition(leftNodeInfo.Start, fmt.Sprintf("invalid assignment target: %T", left))
			return nil
		}
	}
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}
	return &ExpressionStatement{NodeInfo: leftNodeInfo, Expr: left}
}

func (p *Parser) parseArrayDestructuringAssignNames(arr *ArrayLiteral) ([]string, bool) {
	if len(arr.Elements) == 0 {
		p.addErrorAtPosition(nodeInfoFromNode(arr).Start, "array destructuring assignment requires at least one binding")
		return nil, false
	}

	names := make([]string, 0, len(arr.Elements))
	for i, elem := range arr.Elements {
		ident, ok := elem.(*Identifier)
		if !ok {
			p.addErrorAtPosition(
				nodeInfoFromNode(elem).Start,
				fmt.Sprintf("expected identifier in array destructuring assignment at index %d, got %T", i, elem),
			)
			return nil, false
		}
		names = append(names, ident.Name)
	}

	return names, true
}

func (p *Parser) parseObjectDestructuringAssignBindings(obj *ObjectLiteral) ([]ObjectDestructureBinding, bool) {
	if len(obj.SpreadProps) > 0 {
		p.addErrorAtPosition(nodeInfoFromNode(obj).Start, "object destructuring assignment does not support spread properties")
		return nil, false
	}
	if len(obj.Properties) == 0 {
		p.addErrorAtPosition(nodeInfoFromNode(obj).Start, "object destructuring assignment requires at least one binding")
		return nil, false
	}

	bindings := make([]ObjectDestructureBinding, 0, len(obj.Properties))
	for _, prop := range obj.Properties {
		ident, ok := prop.Value.(*Identifier)
		if !ok {
			p.addErrorAtPosition(
				prop.Pos,
				fmt.Sprintf("expected identifier in object destructuring assignment for key %q, got %T", prop.Key, prop.Value),
			)
			return nil, false
		}
		bindings = append(bindings, ObjectDestructureBinding{
			Pos:  nodeInfoFromNode(ident).Start,
			Key:  prop.Key,
			Name: ident.Name,
		})
	}

	return bindings, true
}

func (p *Parser) parseReturnStatement() Statement {
	stmt := &ReturnStatement{NodeInfo: nodeInfoFromToken(p.curToken)}
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
		return stmt
	}
	if p.peekToken.Type == RBRACE {
		return stmt
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseThrowStatement() Statement {
	stmt := &ThrowStatement{NodeInfo: nodeInfoFromToken(p.curToken)}
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
		return stmt
	}
	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseBreakStatement() Statement {
	startTok := p.curToken
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}
	return &BreakStatement{NodeInfo: nodeInfoFromToken(startTok)}
}

func (p *Parser) parseContinueStatement() Statement {
	startTok := p.curToken
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}
	return &ContinueStatement{NodeInfo: nodeInfoFromToken(startTok)}
}

func (p *Parser) parseTryCatchStatement() Statement {
	stmt := &TryCatchStatement{NodeInfo: nodeInfoFromToken(p.curToken)}
	if !p.expectPeek(LBRACE) {
		return nil
	}
	stmt.TryBlock = p.parseBlockStatement()

	if p.peekToken.Type == CATCH {
		p.nextToken()
		if !p.expectPeek(LPAREN) {
			return nil
		}
		if !p.expectPeek(IDENT) {
			return nil
		}
		stmt.CatchName = p.curToken.Literal
		if !p.expectPeek(RPAREN) {
			return nil
		}
		if !p.expectPeek(LBRACE) {
			return nil
		}
		stmt.CatchBlock = p.parseBlockStatement()
	}

	if p.peekToken.Type == FINALLY {
		p.nextToken()
		if !p.expectPeek(LBRACE) {
			return nil
		}
		stmt.FinallyBlock = p.parseBlockStatement()
	}

	if stmt.CatchBlock == nil && stmt.FinallyBlock == nil {
		p.addError("try statement requires catch or finally")
		return nil
	}
	return stmt
}

func (p *Parser) parseIfStatement() Statement {
	stmt := &IfStatement{NodeInfo: nodeInfoFromToken(p.curToken)}
	if !p.expectPeek(LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(RPAREN) {
		return nil
	}
	stmt.Then = p.parseBlockOrSingleStatement()
	if stmt.Then == nil {
		return nil
	}

	if p.peekToken.Type == ELSE {
		p.nextToken()
		stmt.Else = p.parseBlockOrSingleStatement()
		if stmt.Else == nil {
			return nil
		}
	}
	return stmt
}

func (p *Parser) parseWhileStatement() Statement {
	stmt := &WhileStatement{NodeInfo: nodeInfoFromToken(p.curToken)}
	if !p.expectPeek(LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(RPAREN) {
		return nil
	}
	if !p.expectPeek(LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseDoWhileStatement() Statement {
	stmt := &DoWhileStatement{NodeInfo: nodeInfoFromToken(p.curToken)}
	if !p.expectPeek(LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	if !p.expectPeek(WHILE) {
		return nil
	}
	if !p.expectPeek(LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if !p.expectPeek(RPAREN) {
		return nil
	}
	if p.peekToken.Type == SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *Parser) parseForStatement() Statement {
	startTok := p.curToken
	if !p.expectPeek(LPAREN) {
		return nil
	}
	p.nextToken()

	// Check for for-in: for (key in obj) { }
	// Check for for-of: for (value of arr) { }

	// First, check if current token is IDENT and next is IN/OF
	if p.curToken.Type == IDENT && (p.peekToken.Type == IN || p.peekToken.Type == OF) {
		// Simple for-in/for-of: for (key in obj)
		varName := p.curToken.Literal
		loopType := p.peekToken.Type

		p.nextToken() // Move to IN/OF
		p.nextToken() // Move to iterable

		iterable := p.parseExpression(LOWEST)

		if !p.expectPeek(RPAREN) {
			return nil
		}
		if !p.expectPeek(LBRACE) {
			return nil
		}
		body := p.parseBlockStatement()

		if loopType == IN {
			return &ForInStatement{
				NodeInfo: nodeInfoFromToken(startTok),
				Variable: varName,
				Iterable: iterable,
				Body:     body,
			}
		}

		return &ForOfStatement{
			NodeInfo: nodeInfoFromToken(startTok),
			Variable: varName,
			Iterable: iterable,
			Body:     body,
		}
	}

	// Not for-in/for-of, parse as C-style for loop
	return p.parseCStyleForStatementFromCurrent(startTok)
}

func (p *Parser) parseCStyleForStatementFromCurrent(startTok Token) Statement {
	stmt := &ForStatement{NodeInfo: nodeInfoFromToken(startTok)}
	// LPAREN already consumed, p.curToken is first token inside parens
	// DO NOT call p.nextToken() here - p.curToken is already at the right position

	if p.curToken.Type != SEMICOLON {
		switch {
		case p.curToken.Type == LET:
			stmt.Init = p.parseLetStatement()
		case p.curToken.Type == IDENT && p.peekToken.Type == ASSIGN:
			stmt.Init = p.parseAssignStatement()
		default:
			stmt.Init = &ExpressionStatement{NodeInfo: nodeInfoFromToken(p.curToken), Expr: p.parseExpression(LOWEST)}
			if p.peekToken.Type == SEMICOLON {
				p.nextToken()
			}
		}
	}
	if p.curToken.Type != SEMICOLON {
		if !p.expectPeek(SEMICOLON) {
			return nil
		}
	}

	p.nextToken()
	if p.curToken.Type != SEMICOLON {
		stmt.Condition = p.parseExpression(LOWEST)
	}
	if p.curToken.Type != SEMICOLON {
		if !p.expectPeek(SEMICOLON) {
			return nil
		}
	}

	p.nextToken()
	if p.curToken.Type != RPAREN {
		stmt.Post = p.parseForPostStatement()
	}
	if p.curToken.Type != RPAREN {
		if !p.expectPeek(RPAREN) {
			return nil
		}
	}
	if !p.expectPeek(LBRACE) {
		return nil
	}
	stmt.Body = p.parseBlockStatement()
	return stmt
}

func (p *Parser) parseExpression(precedence int) Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.addError(fmt.Sprintf("no prefix parse function for %s", p.curToken.Type))
		return nil
	}
	leftExp := prefix()

	// Check for postfix ++/-- (only on identifiers)
	if p.peekToken.Type == PLUSPLUS || p.peekToken.Type == MINUSMINUS {
		if ident, ok := leftExp.(*Identifier); ok {
			p.nextToken()
			leftExp = &UpdateExpression{
				NodeInfo: nodeInfoFromNode(ident),
				Operator: p.curToken.Type,
				Operand:  ident,
				IsPrefix: false,
			}
		} else {
			p.addError("postfix ++/-- can only be applied to identifiers")
			return nil
		}
	}

	for p.peekToken.Type != SEMICOLON &&
		p.peekToken.Type != RBRACE &&
		p.peekToken.Type != RBRACKET &&
		precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() Expression {
	return &Identifier{
		NodeInfo: nodeInfoFromToken(p.curToken),
		Name:     p.curToken.Literal,
	}
}

func (p *Parser) parseSuperExpression() Expression {
	startTok := p.curToken
	if p.peekToken.Type == DOT {
		p.nextToken()
		if p.peekToken.Type != IDENT && !token.IsKeyword(p.peekToken.Type) {
			p.addError(fmt.Sprintf("expected identifier after 'super.', got %s", p.peekToken.Type))
			return nil
		}
		p.nextToken()
		return &SuperExpression{
			NodeInfo: nodeInfoFromToken(startTok),
			Property: p.curToken.Literal,
		}
	}
	if p.peekToken.Type == LPAREN {
		p.nextToken()
		return &SuperCallExpression{
			NodeInfo:  nodeInfoFromToken(startTok),
			Arguments: p.parseExpressionList(RPAREN),
		}
	}
	p.addError("expected '.' or '(' after 'super'")
	return nil
}

func (p *Parser) parseNumberLiteral() Expression {
	return &NumberLiteral{
		NodeInfo: nodeInfoFromToken(p.curToken),
		Value:    p.curToken.Literal,
	}
}

func (p *Parser) parseStringLiteral() Expression {
	return &StringLiteral{
		NodeInfo: nodeInfoFromToken(p.curToken),
		Value:    p.curToken.Literal,
	}
}

func (p *Parser) parseTemplateLiteral() Expression {
	return p.buildTemplateExpression(p.curToken)
}

func (p *Parser) buildTemplateExpression(tok Token) Expression {
	raw := tok.Literal
	type templatePart struct {
		text   string
		isExpr bool
	}
	parts := []templatePart{}

	start := 0
	for i := 0; i < len(raw); i++ {
		if raw[i] == '$' && i+1 < len(raw) && raw[i+1] == '{' && (i == 0 || raw[i-1] != '\\') {
			parts = append(parts, templatePart{text: raw[start:i]})
			i += 2
			exprStart := i
			depth := 1
			for i < len(raw) && depth > 0 {
				switch raw[i] {
				case '{':
					depth++
				case '}':
					depth--
				}
				if depth == 0 {
					break
				}
				i++
			}
			if depth != 0 || i >= len(raw) {
				p.addErrorWithToken(tok, "unterminated template interpolation")
				return nil
			}

			exprSource := strings.TrimSpace(raw[exprStart:i])
			if exprSource == "" {
				p.addErrorWithToken(tok, "empty template interpolation")
				return nil
			}
			parts = append(parts, templatePart{text: exprSource, isExpr: true})
			start = i + 1
		}
	}
	parts = append(parts, templatePart{text: raw[start:]})

	var out Expression
	for _, part := range parts {
		var piece Expression
		if part.isExpr {
			parsed := parseTemplateInterpolationExpression(part.text)
			if parsed.err != "" {
				p.addErrorWithToken(tok, parsed.err)
				return nil
			}
			piece = parsed.expr
		} else {
			piece = &StringLiteral{
				NodeInfo: nodeInfoFromToken(tok),
				Value:    part.text,
			}
		}

		if out == nil {
			out = piece
			continue
		}
		out = &BinaryExpression{
			NodeInfo: nodeInfoFromToken(tok),
			Left:     out,
			Operator: PLUS,
			Right:    piece,
		}
	}

	if out == nil {
		return &StringLiteral{
			NodeInfo: nodeInfoFromToken(tok),
			Value:    "",
		}
	}
	return out
}

type templateInterpolationParseResult struct {
	expr Expression
	err  string
}

func parseTemplateInterpolationExpression(source string) templateInterpolationParseResult {
	l := NewLexer(source)
	p := NewParser(l)
	program := p.ParseProgram()
	if errs := p.Errors(); len(errs) > 0 {
		return templateInterpolationParseResult{err: fmt.Sprintf("invalid template interpolation: %s", strings.Join(errs, "; "))}
	}
	if len(program.Statements) != 1 {
		return templateInterpolationParseResult{err: "template interpolation must contain a single expression"}
	}
	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok || stmt.Expr == nil {
		return templateInterpolationParseResult{err: "template interpolation must be an expression"}
	}
	return templateInterpolationParseResult{expr: stmt.Expr}
}

func (p *Parser) parseBoolLiteral() Expression {
	return &BoolLiteral{
		NodeInfo: nodeInfoFromToken(p.curToken),
		Value:    p.curToken.Type == TRUE,
	}
}

func (p *Parser) parseNullLiteral() Expression {
	return &NullLiteral{NodeInfo: nodeInfoFromToken(p.curToken)}
}

func (p *Parser) parseArrayLiteral() Expression {
	arr := &ArrayLiteral{
		NodeInfo: nodeInfoFromToken(p.curToken),
	}
	arr.Elements = p.parseArrayElements()
	return arr
}

// parseArrayElements parses array elements including spread elements.
func (p *Parser) parseArrayElements() []Expression {
	elements := []Expression{}
	if p.peekToken.Type == RBRACKET {
		p.nextToken()
		return elements
	}

	p.nextToken()
	for p.curToken.Type != RBRACKET && p.curToken.Type != EOF {
		// Check for spread element
		if p.curToken.Type == SPREAD {
			p.nextToken()
			spreadExpr := p.parseExpression(LOWEST)
			if spreadExpr == nil {
				return nil
			}
			elements = append(elements, &SpreadElement{
				NodeInfo: nodeInfoFromToken(p.curToken),
				Expr:     spreadExpr,
			})
		} else {
			elem := p.parseExpression(LOWEST)
			if elem == nil {
				return nil
			}
			elements = append(elements, elem)
		}

		if p.peekToken.Type == COMMA {
			p.nextToken()
			if p.peekToken.Type == RBRACKET {
				p.nextToken()
				return elements
			}
			p.nextToken()
			continue
		}
		break
	}
	if !p.expectPeek(RBRACKET) {
		return nil
	}
	return elements
}

func (p *Parser) parseObjectLiteral() Expression {
	obj := &ObjectLiteral{NodeInfo: nodeInfoFromToken(p.curToken)}
	if p.peekToken.Type == RBRACE {
		p.nextToken()
		return obj
	}

	for {
		p.nextToken()
		for p.curToken.Type == COMMENT {
			if p.peekToken.Type == RBRACE {
				p.nextToken()
				return obj
			}
			p.nextToken()
		}

		// Check for spread property: { ...obj }
		if p.curToken.Type == SPREAD {
			p.nextToken()
			spreadExpr := p.parseExpression(LOWEST)
			if spreadExpr == nil {
				return nil
			}
			obj.SpreadProps = append(obj.SpreadProps, ObjectSpreadProperty{
				Pos:  Position{Line: p.curToken.Line, Column: p.curToken.Column},
				Expr: spreadExpr,
			})
			if p.peekToken.Type == COMMA {
				p.nextToken()
				if p.peekToken.Type == RBRACE {
					p.nextToken()
					return obj
				}
				continue
			}
			break
		}

		var key string
		keyPos := Position{Line: p.curToken.Line, Column: p.curToken.Column}
		switch p.curToken.Type {
		case IDENT, STRING:
			key = p.curToken.Literal
		default:
			p.addError(fmt.Sprintf("expected object key identifier or string, got %s", p.curToken.Type))
			return nil
		}
		var value Expression
		valuePos := keyPos
		if p.peekToken.Type == COLON {
			p.nextToken()
			p.nextToken()
			value = p.parseExpression(LOWEST)
			valuePos = Position{Line: p.curToken.Line, Column: p.curToken.Column}
		} else {
			// Shorthand object property: {x} -> {x: x}
			if p.curToken.Type != IDENT {
				p.addError("object shorthand property requires identifier key")
				return nil
			}
			value = &Identifier{
				NodeInfo: nodeInfoFromToken(p.curToken),
				Name:     key,
			}
		}
		obj.Properties = append(obj.Properties, ObjectProperty{
			Pos:   valuePos,
			Key:   key,
			Value: value,
		})
		if p.peekToken.Type == COMMA {
			p.nextToken()
			if p.peekToken.Type == RBRACE {
				p.nextToken()
				return obj
			}
			continue
		}
		break
	}
	if !p.expectPeek(RBRACE) {
		return nil
	}
	return obj
}

func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken()
	exp := p.parseExpression(LOWEST)
	if _, ok := exp.(*ObjectLiteral); ok && p.peekToken.Type == ASSIGN {
		// Allow parenthesized object destructuring assignment target:
		// ({x, y: z} = obj)
		// Keep ')' for parseAssignmentOrExpressionStatement to consume.
		return exp
	}
	if !p.expectPeek(RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseAwaitExpression() Expression {
	startTok := p.curToken
	p.nextToken()
	return &AwaitExpression{
		NodeInfo: nodeInfoFromToken(startTok),
		// Parse member/call chain as await operand, but stop before low-precedence binary ops.
		Expr: p.parseExpression(SUM),
	}
}

func (p *Parser) parseDynamicImportExpression() Expression {
	startTok := p.curToken
	if !p.expectPeek(LPAREN) {
		return nil
	}
	p.nextToken()
	moduleExpr := p.parseExpression(LOWEST)
	if moduleExpr == nil {
		p.addError("dynamic import expects module expression")
		return nil
	}
	if !p.expectPeek(RPAREN) {
		return nil
	}
	return &DynamicImportExpression{
		NodeInfo: nodeInfoFromToken(startTok),
		Module:   moduleExpr,
	}
}

func (p *Parser) parseNewExpression() Expression {
	startTok := p.curToken
	p.nextToken()
	target := p.parseExpression(PREFIX)
	if call, ok := target.(*CallExpression); ok {
		return &NewExpression{
			NodeInfo:  nodeInfoFromToken(startTok),
			Callee:    call.Callee,
			Arguments: call.Arguments,
		}
	}
	return &NewExpression{
		NodeInfo:  nodeInfoFromToken(startTok),
		Callee:    target,
		Arguments: []Expression{},
	}
}

func (p *Parser) parseUnaryExpression() Expression {
	exp := &UnaryExpression{
		NodeInfo: nodeInfoFromToken(p.curToken),
		Operator: p.curToken.Type,
	}
	p.nextToken()
	exp.Right = p.parseExpression(PREFIX)
	return exp
}

func (p *Parser) parseUpdateExpression() Expression {
	startTok := p.curToken
	if !p.expectPeek(IDENT) {
		return nil
	}
	return &UpdateExpression{
		NodeInfo: nodeInfoFromToken(startTok),
		Operator: startTok.Type,
		Operand: &Identifier{
			NodeInfo: nodeInfoFromToken(p.curToken),
			Name:     p.curToken.Literal,
		},
		IsPrefix: true,
	}
}

func (p *Parser) parseBinaryExpression(left Expression) Expression {
	exp := &BinaryExpression{
		NodeInfo: nodeInfoFromToken(p.curToken),
		Left:     left,
		Operator: p.curToken.Type,
	}
	precedence := p.curPrecedence()
	p.nextToken()
	exp.Right = p.parseExpression(precedence)
	return exp
}

func (p *Parser) parseTernaryExpression(condition Expression) Expression {
	exp := &TernaryExpression{
		NodeInfo:  nodeInfoFromNode(condition),
		Condition: condition,
	}
	p.nextToken() // 跳过 ?

	precedence := p.curPrecedence()
	exp.Then = p.parseExpression(precedence)

	if !p.expectPeek(COLON) {
		return nil
	}

	p.nextToken()
	exp.Else = p.parseExpression(precedence)
	return exp
}

func (p *Parser) parseCallExpression(callee Expression) Expression {
	args, spreadArgs, hasSpread := p.parseCallArguments()
	return &CallExpression{
		NodeInfo:      nodeInfoFromNode(callee),
		Callee:        callee,
		Arguments:     args,
		SpreadArgs:    spreadArgs,
		HasSpreadCall: hasSpread,
	}
}

// parseCallArguments parses function call arguments including spread arguments.
// Returns the arguments list, indices of spread arguments, and whether spread is used.
func (p *Parser) parseCallArguments() ([]Expression, []int, bool) {
	args := []Expression{}
	spreadArgs := []int{}
	hasSpread := false

	if p.peekToken.Type == RPAREN {
		p.nextToken()
		return args, spreadArgs, hasSpread
	}

	p.nextToken()
	idx := 0
	for p.curToken.Type != RPAREN && p.curToken.Type != EOF {
		// Check for spread argument
		if p.curToken.Type == SPREAD {
			p.nextToken()
			spreadExpr := p.parseExpression(LOWEST)
			if spreadExpr == nil {
				return nil, nil, false
			}
			spreadArgs = append(spreadArgs, idx)
			args = append(args, spreadExpr)
			hasSpread = true
		} else {
			arg := p.parseExpression(LOWEST)
			if arg == nil {
				return nil, nil, false
			}
			args = append(args, arg)
		}

		idx++
		if p.peekToken.Type == COMMA {
			p.nextToken()
			if p.peekToken.Type == RPAREN {
				p.nextToken()
				return args, spreadArgs, hasSpread
			}
			p.nextToken()
			continue
		}
		break
	}
	if !p.expectPeek(RPAREN) {
		return nil, nil, false
	}
	return args, spreadArgs, hasSpread
}

func (p *Parser) parseGetExpression(left Expression) Expression {
	if p.peekToken.Type != IDENT && p.peekToken.Type != PRIVATE && !token.IsKeyword(p.peekToken.Type) {
		p.addError(fmt.Sprintf("expected identifier after '.', got %s", p.peekToken.Type))
		return nil
	}
	p.nextToken()
	property := p.curToken.Literal
	if p.curToken.Type == PRIVATE {
		property = manglePrivateName(property)
	}
	return &GetExpression{
		NodeInfo: nodeInfoFromNode(left),
		Object:   left,
		Property: property,
	}
}

func (p *Parser) parseIndexExpression(left Expression) Expression {
	p.nextToken()
	index := p.parseExpression(LOWEST)
	if !p.expectPeek(RBRACKET) {
		return nil
	}
	return &IndexExpression{
		NodeInfo: nodeInfoFromNode(left),
		Object:   left,
		Index:    index,
	}
}

func (p *Parser) parseOptionalChainExpression(left Expression) Expression {
	p.nextToken() // consume ?.

	// After ?., we can have:
	// - IDENT: optional property access (obj?.prop)
	// - [: optional index access (arr?.[index])
	// - (: optional call (fn?.())
	switch p.curToken.Type {
	case LBRACKET:
		p.nextToken()
		index := p.parseExpression(LOWEST)
		if !p.expectPeek(RBRACKET) {
			return nil
		}
		return &OptionalChainExpression{
			NodeInfo: nodeInfoFromNode(left),
			Base:     left,
			Access: &IndexExpression{
				NodeInfo: nodeInfoFromToken(p.curToken),
				Object:   nil, // Will be resolved at compile time
				Index:    index,
			},
		}
	case LPAREN:
		return &OptionalChainExpression{
			NodeInfo: nodeInfoFromNode(left),
			Base:     left,
			Access: &CallExpression{
				NodeInfo:  nodeInfoFromToken(p.curToken),
				Callee:    nil, // Will be resolved at compile time
				Arguments: p.parseExpressionList(RPAREN),
			},
		}
	default:
		if p.curToken.Type == IDENT || token.IsKeyword(p.curToken.Type) {
			if p.peekToken.Type == LPAREN {
				methodName := p.curToken.Literal
				p.nextToken()

				return &OptionalChainExpression{
					NodeInfo: nodeInfoFromNode(left),
					Base:     left,
					Access: &CallExpression{
						NodeInfo: nodeInfoFromToken(p.curToken),
						Callee: &GetExpression{
							NodeInfo: nodeInfoFromToken(p.curToken),
							Object:   nil,
							Property: methodName,
						},
						Arguments: p.parseExpressionList(RPAREN),
					},
				}
			}

			return &OptionalChainExpression{
				NodeInfo: nodeInfoFromNode(left),
				Base:     left,
				Access: &GetExpression{
					NodeInfo: nodeInfoFromToken(p.curToken),
					Object:   nil,
					Property: p.curToken.Literal,
				},
			}
		}
		p.addError(fmt.Sprintf("expected identifier, '[' or '(' after '?.', got %s", p.curToken.Type))
		return nil
	}
}

func (p *Parser) parseExpressionList(end TokenType) []Expression {
	list := []Expression{}
	if p.peekToken.Type == end {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))
	for p.peekToken.Type == COMMA {
		p.nextToken()
		if p.peekToken.Type == end {
			p.nextToken()
			return list
		}
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(end) {
		return nil
	}
	return list
}

func (p *Parser) parseForPostStatement() Statement {
	if p.curToken.Type == IDENT && p.peekToken.Type == ASSIGN {
		return p.parseAssignStatement()
	}
	return &ExpressionStatement{
		NodeInfo: nodeInfoFromToken(p.curToken),
		Expr:     p.parseExpression(LOWEST),
	}
}

func (p *Parser) parseFunctionParams() ([]string, []DefaultParam, string) {
	params := []string{}
	defaults := []DefaultParam{}
	restParam := ""
	if p.peekToken.Type == RPAREN {
		p.nextToken()
		return params, defaults, restParam
	}

	p.nextToken()
	for {
		if p.curToken.Type == SPREAD {
			if !p.expectPeek(IDENT) {
				return nil, nil, ""
			}
			restParam = p.curToken.Literal
			if p.peekToken.Type == COMMA {
				p.addError("rest parameter must be the last parameter")
				return nil, nil, ""
			}
			break
		}

		if p.curToken.Type != IDENT {
			p.addError(fmt.Sprintf("expected function parameter identifier, got %s", p.curToken.Type))
			return nil, nil, ""
		}
		paramName := p.curToken.Literal
		params = append(params, paramName)
		defaults = append(defaults, DefaultParam{})

		if p.peekToken.Type == ASSIGN {
			p.nextToken()
			p.nextToken()
			defaults[len(defaults)-1] = DefaultParam{
				Pos:   Position{Line: p.curToken.Line, Column: p.curToken.Column},
				Name:  paramName,
				Value: p.parseExpression(LOWEST),
			}
		}

		if p.peekToken.Type != COMMA {
			break
		}
		p.nextToken()
		p.nextToken()
	}
	if !p.expectPeek(RPAREN) {
		return nil, nil, ""
	}
	return params, defaults, restParam
}

func (p *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{NodeInfo: nodeInfoFromToken(p.curToken)}
	p.nextToken()
	var pendingComments []*ast.Comment

	for p.curToken.Type != RBRACE && p.curToken.Type != EOF {
		// Collect leading comments
		for p.curToken.Type == COMMENT {
			pendingComments = append(pendingComments, &ast.Comment{
				NodeInfo: nodeInfoFromToken(p.curToken),
				Text:     p.curToken.Literal,
			})
			p.nextToken()
		}
		if p.curToken.Type == RBRACE || p.curToken.Type == EOF {
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			stmt.SetLeadingComments(pendingComments)
			pendingComments = nil
			block.Statements = append(block.Statements, stmt)
		} else {
			pendingComments = nil
		}
		p.nextToken()
	}
	return block
}

func (p *Parser) parseBlockOrSingleStatement() *BlockStatement {
	if p.peekToken.Type == LBRACE {
		p.nextToken()
		return p.parseBlockStatement()
	}

	startTok := p.peekToken
	p.nextToken()
	stmt := p.parseStatement()
	if stmt == nil {
		return nil
	}
	return &BlockStatement{
		NodeInfo:   nodeInfoFromToken(startTok),
		Statements: []Statement{stmt},
	}
}

func (p *Parser) expectPeek(t TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// expectAnyPeek checks if peekToken matches any of the given token types.
// If it matches, advances to that token and returns true.
// Otherwise, reports an error and returns false.
func (p *Parser) expectAnyPeek(types ...TokenType) bool {
	for _, t := range types {
		if p.peekToken.Type == t {
			p.nextToken()
			return true
		}
	}
	// Report error with first expected type
	p.peekError(types[0])
	return false
}

func (p *Parser) peekError(t TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s", t, p.peekToken.Type)
	p.addErrorWithToken(p.peekToken, msg)
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseSwitchStatement() Statement {
	stmt := &SwitchStatement{NodeInfo: nodeInfoFromToken(p.curToken)}
	if !p.expectPeek(LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Expression = p.parseExpression(LOWEST)
	if !p.expectPeek(RPAREN) {
		return nil
	}
	if !p.expectPeek(LBRACE) {
		return nil
	}

	p.nextToken()
	for p.curToken.Type != RBRACE && p.curToken.Type != EOF {
		switch p.curToken.Type {
		case CASE:
			caseClause := &CaseClause{NodeInfo: nodeInfoFromToken(p.curToken)}
			p.nextToken()
			caseClause.Value = p.parseExpression(LOWEST)
			if !p.expectPeek(COLON) {
				return nil
			}

			p.nextToken()
			for p.curToken.Type != CASE && p.curToken.Type != DEFAULT && p.curToken.Type != RBRACE && p.curToken.Type != EOF {
				stmt := p.parseStatement()
				if stmt != nil {
					caseClause.Statements = append(caseClause.Statements, stmt)
				}
				p.nextToken()
			}
			stmt.Cases = append(stmt.Cases, caseClause)
		case DEFAULT:
			if !p.expectPeek(COLON) {
				return nil
			}
			stmt.Default = p.parseBlockStatement()
		default:
			p.addError(fmt.Sprintf("expected 'case' or 'default' in switch, got %s", p.curToken.Type))
			return nil
		}
	}
	return stmt
}

func (p *Parser) parseTypeofExpression() Expression {
	startTok := p.curToken
	p.nextToken()
	return &TypeofExpression{
		NodeInfo: nodeInfoFromToken(startTok),
		Expr:     p.parseExpression(PREFIX),
	}
}

func (p *Parser) parseVoidExpression() Expression {
	startTok := p.curToken
	p.nextToken()
	return &VoidExpression{
		NodeInfo: nodeInfoFromToken(startTok),
		Expr:     p.parseExpression(PREFIX),
	}
}

func manglePrivateName(name string) string {
	return "__private_" + name
}
