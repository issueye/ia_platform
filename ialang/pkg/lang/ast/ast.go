package ast

import "ialang/pkg/lang/token"

type TokenType = token.TokenType

type Position struct {
	Line   int
	Column int
}

func (p Position) IsValid() bool {
	return p.Line > 0 && p.Column > 0
}

func PositionFromToken(tok token.Token) Position {
	return Position{Line: tok.Line, Column: tok.Column}
}

type Node interface {
	node()
	Pos() Position
}

type NodeInfo struct {
	Start Position
}

func (n NodeInfo) Pos() Position {
	return n.Start
}

type Statement interface {
	Node
	stmtNode()
	SetLeadingComments(comments []*Comment)
	GetLeadingComments() []*Comment
}

type Expression interface {
	Node
	exprNode()
}

type Program struct {
	NodeInfo
	Statements []Statement
}

func (*Program) node() {}

type Comment struct {
	NodeInfo
	Text string
}

func (*Comment) node()                          {}
func (*Comment) stmtNode()                      {}
func (*Comment) SetLeadingComments([]*Comment)  {}
func (*Comment) GetLeadingComments() []*Comment { return nil }

type StatementComments struct {
	leadingComments []*Comment
}

func (sc *StatementComments) SetLeadingComments(comments []*Comment) {
	sc.leadingComments = comments
}

func (sc *StatementComments) GetLeadingComments() []*Comment {
	return sc.leadingComments
}

type BlockStatement struct {
	StatementComments
	NodeInfo
	Statements []Statement
}

func (*BlockStatement) node() {}

type ImportStatement struct {
	StatementComments
	NodeInfo
	Names     []string
	Namespace string
	Module    string
}

func (*ImportStatement) node()     {}
func (*ImportStatement) stmtNode() {}

type ExportStatement struct {
	StatementComments
	NodeInfo
	Statement       Statement
	Specifiers      []ExportSpecifier
	Default         Expression
	DefaultName     string
	ExportAllModule string
}

func (*ExportStatement) node()     {}
func (*ExportStatement) stmtNode() {}

type ExportSpecifier struct {
	Pos        Position
	LocalName  string
	ExportName string
}

type LetStatement struct {
	StatementComments
	NodeInfo
	Name        string
	Initializer Expression
}

func (*LetStatement) node()     {}
func (*LetStatement) stmtNode() {}

type ArrayDestructuringLetStatement struct {
	StatementComments
	NodeInfo
	Names       []string
	Initializer Expression
}

func (*ArrayDestructuringLetStatement) node()     {}
func (*ArrayDestructuringLetStatement) stmtNode() {}

type ObjectDestructureBinding struct {
	Pos  Position
	Key  string
	Name string
}

type ObjectDestructuringLetStatement struct {
	StatementComments
	NodeInfo
	Bindings    []ObjectDestructureBinding
	Initializer Expression
}

func (*ObjectDestructuringLetStatement) node()     {}
func (*ObjectDestructuringLetStatement) stmtNode() {}

type ArrayDestructuringAssignStatement struct {
	StatementComments
	NodeInfo
	Names []string
	Value Expression
}

func (*ArrayDestructuringAssignStatement) node()     {}
func (*ArrayDestructuringAssignStatement) stmtNode() {}

type ObjectDestructuringAssignStatement struct {
	StatementComments
	NodeInfo
	Bindings []ObjectDestructureBinding
	Value    Expression
}

func (*ObjectDestructuringAssignStatement) node()     {}
func (*ObjectDestructuringAssignStatement) stmtNode() {}

type AssignStatement struct {
	StatementComments
	NodeInfo
	Name  string
	Value Expression
}

func (*AssignStatement) node()     {}
func (*AssignStatement) stmtNode() {}

type CompoundAssignStatement struct {
	StatementComments
	NodeInfo
	Name     string
	Operator TokenType
	Value    Expression
}

func (*CompoundAssignStatement) node()     {}
func (*CompoundAssignStatement) stmtNode() {}

type SetPropertyStatement struct {
	StatementComments
	NodeInfo
	Object   Expression
	Property string
	Value    Expression
}

func (*SetPropertyStatement) node()     {}
func (*SetPropertyStatement) stmtNode() {}

type CompoundSetPropertyStatement struct {
	StatementComments
	NodeInfo
	Object   Expression
	Property string
	Operator TokenType
	Value    Expression
}

func (*CompoundSetPropertyStatement) node()     {}
func (*CompoundSetPropertyStatement) stmtNode() {}

type ExpressionStatement struct {
	StatementComments
	NodeInfo
	Expr Expression
}

func (*ExpressionStatement) node()     {}
func (*ExpressionStatement) stmtNode() {}

// DefaultParam represents a function parameter with a default value.
type DefaultParam struct {
	Pos   Position
	Name  string
	Value Expression // default value expression
}

type FunctionStatement struct {
	StatementComments
	NodeInfo
	Name          string
	Params        []string
	RestParam     string
	ParamDefaults []DefaultParam // parallel to Params, zero-value entries for params without defaults
	Async         bool
	Body          *BlockStatement
}

func (*FunctionStatement) node()     {}
func (*FunctionStatement) stmtNode() {}

type FunctionExpression struct {
	NodeInfo
	Name          string
	Params        []string
	RestParam     string
	ParamDefaults []DefaultParam // parallel to Params, zero-value entries for params without defaults
	Async         bool
	Body          *BlockStatement
}

func (*FunctionExpression) node()     {}
func (*FunctionExpression) exprNode() {}

type ArrowFunctionExpression struct {
	NodeInfo
	Params        []string
	RestParam     string
	ParamDefaults []DefaultParam // parallel to Params, zero-value entries for params without defaults
	Async         bool
	Body          *BlockStatement // nil for concise body (single expression)
	Concise       bool            // true for concise body (=> expr)
	Expr          Expression      // used when Concise is true
}

func (*ArrowFunctionExpression) node()     {}
func (*ArrowFunctionExpression) exprNode() {}

type ClassMethod struct {
	Pos           Position
	Name          string
	Params        []string
	RestParam     string
	ParamDefaults []DefaultParam // parallel to Params
	Async         bool
	Body          *BlockStatement
	Static        bool // true for static methods
	IsGetter      bool // true for getter methods
	IsSetter      bool // true for setter methods
}

type ClassStatement struct {
	StatementComments
	NodeInfo
	Name          string
	ParentName    string
	Methods       []ClassMethod
	PrivateFields []ClassPrivateField
}

func (*ClassStatement) node()     {}
func (*ClassStatement) stmtNode() {}

type ClassPrivateField struct {
	Pos  Position
	Name string
}

type ReturnStatement struct {
	StatementComments
	NodeInfo
	Value Expression
}

func (*ReturnStatement) node()     {}
func (*ReturnStatement) stmtNode() {}

type ThrowStatement struct {
	StatementComments
	NodeInfo
	Value Expression
}

func (*ThrowStatement) node()     {}
func (*ThrowStatement) stmtNode() {}

type BreakStatement struct {
	StatementComments
	NodeInfo
}

func (*BreakStatement) node()     {}
func (*BreakStatement) stmtNode() {}

type ContinueStatement struct {
	StatementComments
	NodeInfo
}

func (*ContinueStatement) node()     {}
func (*ContinueStatement) stmtNode() {}

type TryCatchStatement struct {
	StatementComments
	NodeInfo
	TryBlock     *BlockStatement
	CatchName    string
	CatchBlock   *BlockStatement
	FinallyBlock *BlockStatement
}

func (*TryCatchStatement) node()     {}
func (*TryCatchStatement) stmtNode() {}

type IfStatement struct {
	StatementComments
	NodeInfo
	Condition Expression
	Then      *BlockStatement
	Else      *BlockStatement
}

func (*IfStatement) node()     {}
func (*IfStatement) stmtNode() {}

type WhileStatement struct {
	StatementComments
	NodeInfo
	Condition Expression
	Body      *BlockStatement
}

func (*WhileStatement) node()     {}
func (*WhileStatement) stmtNode() {}

type DoWhileStatement struct {
	StatementComments
	NodeInfo
	Condition Expression
	Body      *BlockStatement
}

func (*DoWhileStatement) node()     {}
func (*DoWhileStatement) stmtNode() {}

type ForStatement struct {
	StatementComments
	NodeInfo
	Init      Statement
	Condition Expression
	Post      Statement
	Body      *BlockStatement
}

func (*ForStatement) node()     {}
func (*ForStatement) stmtNode() {}

type ForInStatement struct {
	StatementComments
	NodeInfo
	Variable string     // loop variable name
	Iterable Expression // object to iterate over
	Body     *BlockStatement
}

func (*ForInStatement) node()     {}
func (*ForInStatement) stmtNode() {}

type ForOfStatement struct {
	StatementComments
	NodeInfo
	Variable string     // loop variable name
	Iterable Expression // iterable to loop over
	Body     *BlockStatement
}

func (*ForOfStatement) node()     {}
func (*ForOfStatement) stmtNode() {}

type Identifier struct {
	NodeInfo
	Name string
}

func (*Identifier) node()     {}
func (*Identifier) exprNode() {}

type NumberLiteral struct {
	NodeInfo
	Value string
}

func (*NumberLiteral) node()     {}
func (*NumberLiteral) exprNode() {}

type StringLiteral struct {
	NodeInfo
	Value string
}

func (*StringLiteral) node()     {}
func (*StringLiteral) exprNode() {}

type BoolLiteral struct {
	NodeInfo
	Value bool
}

func (*BoolLiteral) node()     {}
func (*BoolLiteral) exprNode() {}

type NullLiteral struct {
	NodeInfo
}

func (*NullLiteral) node()     {}
func (*NullLiteral) exprNode() {}

type ArrayLiteral struct {
	NodeInfo
	Elements []Expression
}

func (*ArrayLiteral) node()     {}
func (*ArrayLiteral) exprNode() {}

// SpreadElement represents a spread element in array or object literals.
// For arrays: [...arr, x]
// For objects: { ...obj, x: 1 }
type SpreadElement struct {
	NodeInfo
	Expr Expression
}

func (*SpreadElement) node()     {}
func (*SpreadElement) exprNode() {}

type ObjectProperty struct {
	Pos   Position
	Key   string
	Value Expression
}

// ObjectSpreadProperty represents a spread property in object literals.
type ObjectSpreadProperty struct {
	Pos  Position
	Expr Expression
}

type ObjectLiteral struct {
	NodeInfo
	Properties  []ObjectProperty
	SpreadProps []ObjectSpreadProperty
}

func (*ObjectLiteral) node()     {}
func (*ObjectLiteral) exprNode() {}

type BinaryExpression struct {
	NodeInfo
	Left     Expression
	Operator token.TokenType
	Right    Expression
}

func (*BinaryExpression) node()     {}
func (*BinaryExpression) exprNode() {}

type UnaryExpression struct {
	NodeInfo
	Operator token.TokenType
	Right    Expression
}

func (*UnaryExpression) node()     {}
func (*UnaryExpression) exprNode() {}

type UpdateExpression struct {
	NodeInfo
	Operator token.TokenType
	Operand  *Identifier
	IsPrefix bool
}

func (*UpdateExpression) node()     {}
func (*UpdateExpression) exprNode() {}

type CallExpression struct {
	NodeInfo
	Callee        Expression
	Arguments     []Expression
	SpreadArgs    []int // indices of arguments that are spread
	HasSpreadCall bool  // true if any argument uses spread
}

func (*CallExpression) node()     {}
func (*CallExpression) exprNode() {}

type NewExpression struct {
	NodeInfo
	Callee    Expression
	Arguments []Expression
}

func (*NewExpression) node()     {}
func (*NewExpression) exprNode() {}

type GetExpression struct {
	NodeInfo
	Object   Expression
	Property string
}

func (*GetExpression) node()     {}
func (*GetExpression) exprNode() {}

type IndexExpression struct {
	NodeInfo
	Object Expression
	Index  Expression
}

func (*IndexExpression) node()     {}
func (*IndexExpression) exprNode() {}

type AwaitExpression struct {
	NodeInfo
	Expr Expression
}

func (*AwaitExpression) node()     {}
func (*AwaitExpression) exprNode() {}

type DynamicImportExpression struct {
	NodeInfo
	Module Expression
}

func (*DynamicImportExpression) node()     {}
func (*DynamicImportExpression) exprNode() {}

type TernaryExpression struct {
	NodeInfo
	Condition Expression
	Then      Expression
	Else      Expression
}

func (*TernaryExpression) node()     {}
func (*TernaryExpression) exprNode() {}

type SuperExpression struct {
	NodeInfo
	Property string
}

func (*SuperExpression) node()     {}
func (*SuperExpression) exprNode() {}

type SuperCallExpression struct {
	NodeInfo
	Arguments []Expression
}

func (*SuperCallExpression) node()     {}
func (*SuperCallExpression) exprNode() {}

type OptionalChainExpression struct {
	NodeInfo
	Base   Expression
	Access Expression // GetExpression, IndexExpression, or CallExpression
}

func (*OptionalChainExpression) node()     {}
func (*OptionalChainExpression) exprNode() {}

type CaseClause struct {
	NodeInfo
	Value      Expression
	Statements []Statement
}

func (*CaseClause) node() {}

type SwitchStatement struct {
	StatementComments
	NodeInfo
	Expression Expression
	Cases      []*CaseClause
	Default    *BlockStatement
}

func (*SwitchStatement) node()     {}
func (*SwitchStatement) stmtNode() {}

type TypeofExpression struct {
	NodeInfo
	Expr Expression
}

func (*TypeofExpression) node()     {}
func (*TypeofExpression) exprNode() {}

type VoidExpression struct {
	NodeInfo
	Expr Expression
}

func (*VoidExpression) node()     {}
func (*VoidExpression) exprNode() {}
