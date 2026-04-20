package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

const (
	ILLEGAL TokenType = "ILLEGAL"
	EOF     TokenType = "EOF"
	COMMENT TokenType = "COMMENT"

	IDENT    TokenType = "IDENT"
	NUMBER   TokenType = "NUMBER"
	STRING   TokenType = "STRING"
	TEMPLATE TokenType = "TEMPLATE"
	PRIVATE  TokenType = "PRIVATE"

	ASSIGN     TokenType = "="
	BANG       TokenType = "!"
	EQ         TokenType = "=="
	NEQ        TokenType = "!="
	AND        TokenType = "&&"
	OR         TokenType = "||"
	NULLISH    TokenType = "??"
	PLUS       TokenType = "+"
	MINUS      TokenType = "-"
	ASTERISK   TokenType = "*"
	SLASH      TokenType = "/"
	MODULO     TokenType = "%"
	LT         TokenType = "<"
	GT         TokenType = ">"
	LTE        TokenType = "<="
	GTE        TokenType = ">="
	BITAND     TokenType = "&"
	BITOR      TokenType = "|"
	BITXOR     TokenType = "^"
	SHL        TokenType = "<<"
	SHR        TokenType = ">>"
	COMMA      TokenType = ","
	COLON      TokenType = ":"
	DOT        TokenType = "."
	SEMICOLON  TokenType = ";"
	QUESTION   TokenType = "?"
	OPTCHAIN   TokenType = "?."
	SPREAD     TokenType = "..."
	PLUSEQ     TokenType = "+="
	MINUSEQ    TokenType = "-="
	MULTEQ     TokenType = "*="
	DIVEQ      TokenType = "/="
	MODEQ      TokenType = "%="
	PLUSPLUS   TokenType = "++"
	MINUSMINUS TokenType = "--"
	ARROW      TokenType = "=>"
	LPAREN     TokenType = "("
	RPAREN     TokenType = ")"
	LBRACE     TokenType = "{"
	RBRACE     TokenType = "}"
	LBRACKET   TokenType = "["
	RBRACKET   TokenType = "]"

	IMPORT   TokenType = "IMPORT"
	EXPORT   TokenType = "EXPORT"
	FROM     TokenType = "FROM"
	CLASS    TokenType = "CLASS"
	NEW      TokenType = "NEW"
	THIS     TokenType = "THIS"
	SUPER    TokenType = "SUPER"
	EXTENDS  TokenType = "EXTENDS"
	LET      TokenType = "LET"
	AWAIT    TokenType = "AWAIT"
	ASYNC    TokenType = "ASYNC"
	FUNC     TokenType = "FUNCTION"
	RETURN   TokenType = "RETURN"
	THROW    TokenType = "THROW"
	IF       TokenType = "IF"
	ELSE     TokenType = "ELSE"
	WHILE    TokenType = "WHILE"
	FOR      TokenType = "FOR"
	IN       TokenType = "IN"
	OF       TokenType = "OF"
	BREAK    TokenType = "BREAK"
	CONTINUE TokenType = "CONTINUE"
	TRY      TokenType = "TRY"
	CATCH    TokenType = "CATCH"
	FINALLY  TokenType = "FINALLY"
	DO       TokenType = "DO"
	TRUE     TokenType = "TRUE"
	FALSE    TokenType = "FALSE"
	NULL     TokenType = "NULL"
	SWITCH   TokenType = "SWITCH"
	CASE     TokenType = "CASE"
	DEFAULT  TokenType = "DEFAULT"
	TYPEOF   TokenType = "TYPEOF"
	VOID     TokenType = "VOID"
	STATIC   TokenType = "STATIC"
	GET      TokenType = "GET"
	SET      TokenType = "SET"
)

var keywords = map[string]TokenType{
	"import":   IMPORT,
	"export":   EXPORT,
	"from":     FROM,
	"class":    CLASS,
	"new":      NEW,
	"this":     THIS,
	"super":    SUPER,
	"extends":  EXTENDS,
	"let":      LET,
	"await":    AWAIT,
	"async":    ASYNC,
	"function": FUNC,
	"return":   RETURN,
	"throw":    THROW,
	"if":       IF,
	"else":     ELSE,
	"while":    WHILE,
	"for":      FOR,
	"in":       IN,
	"of":       OF,
	"break":    BREAK,
	"continue": CONTINUE,
	"try":      TRY,
	"catch":    CATCH,
	"finally":  FINALLY,
	"do":       DO,
	"true":     TRUE,
	"false":    FALSE,
	"null":     NULL,
	"switch":   SWITCH,
	"case":     CASE,
	"default":  DEFAULT,
	"typeof":   TYPEOF,
	"void":     VOID,
	"static":   STATIC,
	"get":      GET,
	"set":      SET,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

func IsKeyword(tok TokenType) bool {
	_, ok := keywordsReverse[tok]
	return ok
}

var keywordsReverse map[TokenType]string

func init() {
	keywordsReverse = make(map[TokenType]string, len(keywords))
	for k, v := range keywords {
		keywordsReverse[v] = k
	}
}
