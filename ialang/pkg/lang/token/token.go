package token

import common "iacommon/pkg/ialang/token"

type TokenType = common.TokenType

type Token = common.Token

const (
	ILLEGAL = common.ILLEGAL
	EOF     = common.EOF
	COMMENT = common.COMMENT

	IDENT    = common.IDENT
	NUMBER   = common.NUMBER
	STRING   = common.STRING
	TEMPLATE = common.TEMPLATE
	PRIVATE  = common.PRIVATE

	ASSIGN     = common.ASSIGN
	BANG       = common.BANG
	EQ         = common.EQ
	NEQ        = common.NEQ
	AND        = common.AND
	OR         = common.OR
	NULLISH    = common.NULLISH
	PLUS       = common.PLUS
	MINUS      = common.MINUS
	ASTERISK   = common.ASTERISK
	SLASH      = common.SLASH
	MODULO     = common.MODULO
	LT         = common.LT
	GT         = common.GT
	LTE        = common.LTE
	GTE        = common.GTE
	BITAND     = common.BITAND
	BITOR      = common.BITOR
	BITXOR     = common.BITXOR
	SHL        = common.SHL
	SHR        = common.SHR
	COMMA      = common.COMMA
	COLON      = common.COLON
	DOT        = common.DOT
	SEMICOLON  = common.SEMICOLON
	QUESTION   = common.QUESTION
	OPTCHAIN   = common.OPTCHAIN
	SPREAD     = common.SPREAD
	PLUSEQ     = common.PLUSEQ
	MINUSEQ    = common.MINUSEQ
	MULTEQ     = common.MULTEQ
	DIVEQ      = common.DIVEQ
	MODEQ      = common.MODEQ
	PLUSPLUS   = common.PLUSPLUS
	MINUSMINUS = common.MINUSMINUS
	ARROW      = common.ARROW
	LPAREN     = common.LPAREN
	RPAREN     = common.RPAREN
	LBRACE     = common.LBRACE
	RBRACE     = common.RBRACE
	LBRACKET   = common.LBRACKET
	RBRACKET   = common.RBRACKET

	IMPORT   = common.IMPORT
	EXPORT   = common.EXPORT
	FROM     = common.FROM
	CLASS    = common.CLASS
	NEW      = common.NEW
	THIS     = common.THIS
	SUPER    = common.SUPER
	EXTENDS  = common.EXTENDS
	LET      = common.LET
	AWAIT    = common.AWAIT
	ASYNC    = common.ASYNC
	FUNC     = common.FUNC
	RETURN   = common.RETURN
	THROW    = common.THROW
	IF       = common.IF
	ELSE     = common.ELSE
	WHILE    = common.WHILE
	FOR      = common.FOR
	IN       = common.IN
	OF       = common.OF
	BREAK    = common.BREAK
	CONTINUE = common.CONTINUE
	TRY      = common.TRY
	CATCH    = common.CATCH
	FINALLY  = common.FINALLY
	DO       = common.DO
	TRUE     = common.TRUE
	FALSE    = common.FALSE
	NULL     = common.NULL
	SWITCH   = common.SWITCH
	CASE     = common.CASE
	DEFAULT  = common.DEFAULT
	TYPEOF   = common.TYPEOF
	VOID     = common.VOID
	STATIC   = common.STATIC
	GET      = common.GET
	SET      = common.SET
)

func LookupIdent(ident string) TokenType {
	return common.LookupIdent(ident)
}

func IsKeyword(tok TokenType) bool {
	return common.IsKeyword(tok)
}
