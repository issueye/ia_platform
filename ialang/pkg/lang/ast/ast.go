package ast

import (
	common "iacommon/pkg/ialang/ast"
	"iacommon/pkg/ialang/token"
)

type TokenType = common.TokenType

type Position = common.Position

type Node = common.Node

type NodeInfo = common.NodeInfo

type Statement = common.Statement

type Expression = common.Expression

type Program = common.Program

type Comment = common.Comment

type StatementComments = common.StatementComments

type BlockStatement = common.BlockStatement

type ImportStatement = common.ImportStatement

type ExportStatement = common.ExportStatement

type ExportSpecifier = common.ExportSpecifier

type LetStatement = common.LetStatement

type ArrayDestructuringLetStatement = common.ArrayDestructuringLetStatement

type ObjectDestructureBinding = common.ObjectDestructureBinding

type ObjectDestructuringLetStatement = common.ObjectDestructuringLetStatement

type ArrayDestructuringAssignStatement = common.ArrayDestructuringAssignStatement

type ObjectDestructuringAssignStatement = common.ObjectDestructuringAssignStatement

type AssignStatement = common.AssignStatement

type CompoundAssignStatement = common.CompoundAssignStatement

type SetPropertyStatement = common.SetPropertyStatement

type CompoundSetPropertyStatement = common.CompoundSetPropertyStatement

type ExpressionStatement = common.ExpressionStatement

type DefaultParam = common.DefaultParam

type FunctionStatement = common.FunctionStatement

type FunctionExpression = common.FunctionExpression

type ArrowFunctionExpression = common.ArrowFunctionExpression

type ClassMethod = common.ClassMethod

type ClassStatement = common.ClassStatement

type ClassPrivateField = common.ClassPrivateField

type ReturnStatement = common.ReturnStatement

type ThrowStatement = common.ThrowStatement

type BreakStatement = common.BreakStatement

type ContinueStatement = common.ContinueStatement

type TryCatchStatement = common.TryCatchStatement

type IfStatement = common.IfStatement

type WhileStatement = common.WhileStatement

type DoWhileStatement = common.DoWhileStatement

type ForStatement = common.ForStatement

type ForInStatement = common.ForInStatement

type ForOfStatement = common.ForOfStatement

type Identifier = common.Identifier

type NumberLiteral = common.NumberLiteral

type StringLiteral = common.StringLiteral

type BoolLiteral = common.BoolLiteral

type NullLiteral = common.NullLiteral

type ArrayLiteral = common.ArrayLiteral

type SpreadElement = common.SpreadElement

type ObjectProperty = common.ObjectProperty

type ObjectSpreadProperty = common.ObjectSpreadProperty

type ObjectLiteral = common.ObjectLiteral

type BinaryExpression = common.BinaryExpression

type UnaryExpression = common.UnaryExpression

type UpdateExpression = common.UpdateExpression

type CallExpression = common.CallExpression

type NewExpression = common.NewExpression

type GetExpression = common.GetExpression

type IndexExpression = common.IndexExpression

type AwaitExpression = common.AwaitExpression

type DynamicImportExpression = common.DynamicImportExpression

type TernaryExpression = common.TernaryExpression

type SuperExpression = common.SuperExpression

type SuperCallExpression = common.SuperCallExpression

type OptionalChainExpression = common.OptionalChainExpression

type CaseClause = common.CaseClause

type SwitchStatement = common.SwitchStatement

type TypeofExpression = common.TypeofExpression

type VoidExpression = common.VoidExpression

func PositionFromToken(tok token.Token) Position {
	return common.PositionFromToken(tok)
}
