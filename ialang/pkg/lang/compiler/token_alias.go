package compiler

import tok "iacommon/pkg/ialang/token"

type TokenType = tok.TokenType

const (
	AND        = tok.AND
	OR         = tok.OR
	NULLISH    = tok.NULLISH
	PLUS       = tok.PLUS
	MINUS      = tok.MINUS
	ASTERISK   = tok.ASTERISK
	SLASH      = tok.SLASH
	MODULO     = tok.MODULO
	EQ         = tok.EQ
	NEQ        = tok.NEQ
	GT         = tok.GT
	LT         = tok.LT
	GTE        = tok.GTE
	LTE        = tok.LTE
	BITAND     = tok.BITAND
	BITOR      = tok.BITOR
	BITXOR     = tok.BITXOR
	SHL        = tok.SHL
	SHR        = tok.SHR
	SUPER      = tok.SUPER
	EXTENDS    = tok.EXTENDS
	BANG       = tok.BANG
	PLUSEQ     = tok.PLUSEQ
	MINUSEQ    = tok.MINUSEQ
	MULTEQ     = tok.MULTEQ
	DIVEQ      = tok.DIVEQ
	MODEQ      = tok.MODEQ
	PLUSPLUS   = tok.PLUSPLUS
	MINUSMINUS = tok.MINUSMINUS
)
