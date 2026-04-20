package frontend

import tok "ialang/pkg/lang/token"

type TokenType = tok.TokenType
type Token = tok.Token

const (
	ILLEGAL = tok.ILLEGAL
	EOF     = tok.EOF
	COMMENT = tok.COMMENT

	IDENT    = tok.IDENT
	NUMBER   = tok.NUMBER
	STRING   = tok.STRING
	TEMPLATE = tok.TEMPLATE
	PRIVATE  = tok.PRIVATE

	ASSIGN     = tok.ASSIGN
	BANG       = tok.BANG
	EQ         = tok.EQ
	NEQ        = tok.NEQ
	AND        = tok.AND
	OR         = tok.OR
	NULLISH    = tok.NULLISH
	PLUS       = tok.PLUS
	MINUS      = tok.MINUS
	ASTERISK   = tok.ASTERISK
	SLASH      = tok.SLASH
	MODULO     = tok.MODULO
	QUESTION   = tok.QUESTION
	OPTCHAIN   = tok.OPTCHAIN
	SPREAD     = tok.SPREAD
	PLUSEQ     = tok.PLUSEQ
	MINUSEQ    = tok.MINUSEQ
	MULTEQ     = tok.MULTEQ
	DIVEQ      = tok.DIVEQ
	MODEQ      = tok.MODEQ
	PLUSPLUS   = tok.PLUSPLUS
	MINUSMINUS = tok.MINUSMINUS
	ARROW      = tok.ARROW
	LT         = tok.LT
	GT         = tok.GT
	LTE        = tok.LTE
	GTE        = tok.GTE
	BITAND     = tok.BITAND
	BITOR      = tok.BITOR
	BITXOR     = tok.BITXOR
	SHL        = tok.SHL
	SHR        = tok.SHR
	COMMA      = tok.COMMA
	COLON      = tok.COLON
	DOT        = tok.DOT
	SEMICOLON  = tok.SEMICOLON
	LPAREN     = tok.LPAREN
	RPAREN     = tok.RPAREN
	LBRACE     = tok.LBRACE
	RBRACE     = tok.RBRACE
	LBRACKET   = tok.LBRACKET
	RBRACKET   = tok.RBRACKET

	IMPORT   = tok.IMPORT
	EXPORT   = tok.EXPORT
	FROM     = tok.FROM
	CLASS    = tok.CLASS
	NEW      = tok.NEW
	THIS     = tok.THIS
	SUPER    = tok.SUPER
	EXTENDS  = tok.EXTENDS
	LET      = tok.LET
	AWAIT    = tok.AWAIT
	ASYNC    = tok.ASYNC
	FUNC     = tok.FUNC
	RETURN   = tok.RETURN
	THROW    = tok.THROW
	IF       = tok.IF
	ELSE     = tok.ELSE
	WHILE    = tok.WHILE
	FOR      = tok.FOR
	IN       = tok.IN
	OF       = tok.OF
	BREAK    = tok.BREAK
	CONTINUE = tok.CONTINUE
	TRY      = tok.TRY
	CATCH    = tok.CATCH
	FINALLY  = tok.FINALLY
	DO       = tok.DO
	TRUE     = tok.TRUE
	FALSE    = tok.FALSE
	NULL     = tok.NULL
	SWITCH   = tok.SWITCH
	CASE     = tok.CASE
	DEFAULT  = tok.DEFAULT
	TYPEOF   = tok.TYPEOF
	VOID     = tok.VOID
	STATIC   = tok.STATIC
	GET      = tok.GET
	SET      = tok.SET
)

func lookupIdent(ident string) TokenType {
	return tok.LookupIdent(ident)
}
