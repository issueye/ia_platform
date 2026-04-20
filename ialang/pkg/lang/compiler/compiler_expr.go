package compiler

import (
	"fmt"
	"strconv"
)

func (c *Compiler) compileExpression(expr Expression) {
	switch e := expr.(type) {
	case *NumberLiteral:
		n, err := strconv.ParseFloat(e.Value, 64)
		if err != nil {
			c.addNodeError(e, fmt.Sprintf("invalid number: %s", e.Value))
			return
		}
		idx := c.chunk.AddConstant(n)
		c.chunk.Emit(OpConstant, idx, 0)
	case *StringLiteral:
		idx := c.chunk.AddConstant(e.Value)
		c.chunk.Emit(OpConstant, idx, 0)
	case *BoolLiteral:
		idx := c.chunk.AddConstant(e.Value)
		c.chunk.Emit(OpConstant, idx, 0)
	case *NullLiteral:
		idx := c.chunk.AddConstant(nil)
		c.chunk.Emit(OpConstant, idx, 0)
	case *Identifier:
		nameIdx := c.chunk.AddConstant(e.Name)
		c.chunk.Emit(OpGetName, nameIdx, 0)
	case *ArrayLiteral:
		c.compileArrayLiteral(e)
	case *ObjectLiteral:
		c.compileObjectLiteral(e)
	case *BinaryExpression:
		switch e.Operator {
		case AND:
			c.compileLogicalAnd(e)
		case OR:
			c.compileLogicalOr(e)
		case NULLISH:
			c.compileNullishCoalescing(e)
		case PLUS:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpAdd, 0, 0)
		case MINUS:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpSub, 0, 0)
		case ASTERISK:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpMul, 0, 0)
		case SLASH:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpDiv, 0, 0)
		case MODULO:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpMod, 0, 0)
		case EQ:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpEqual, 0, 0)
		case NEQ:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpNotEqual, 0, 0)
		case GT:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpGreater, 0, 0)
		case LT:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpLess, 0, 0)
		case GTE:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpGreaterEqual, 0, 0)
		case LTE:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpLessEqual, 0, 0)
		case BITAND:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpBitAnd, 0, 0)
		case BITOR:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpBitOr, 0, 0)
		case BITXOR:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpBitXor, 0, 0)
		case SHL:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpShl, 0, 0)
		case SHR:
			c.compileExpression(e.Left)
			c.compileExpression(e.Right)
			c.chunk.Emit(OpShr, 0, 0)
		default:
			c.addNodeError(e, fmt.Sprintf("unsupported binary operator: %s", e.Operator))
		}
	case *CallExpression:
		c.compileCallExpression(e)
	case *FunctionExpression:
		fn := c.compileFunctionTemplate(e.Name, e.Params, e.RestParam, e.ParamDefaults, e.Async, e.Body)
		if fn == nil {
			return
		}
		fnIdx := c.chunk.AddConstant(fn)
		c.chunk.Emit(OpClosure, fnIdx, 0)
	case *ArrowFunctionExpression:
		fn := c.compileArrowFunction(e)
		if fn == nil {
			return
		}
		fnIdx := c.chunk.AddConstant(fn)
		c.chunk.Emit(OpClosure, fnIdx, 0)
	case *NewExpression:
		c.compileExpression(e.Callee)
		for _, arg := range e.Arguments {
			c.compileExpression(arg)
		}
		c.chunk.Emit(OpNew, len(e.Arguments), 0)
	case *GetExpression:
		c.compileExpression(e.Object)
		nameIdx := c.chunk.AddConstant(e.Property)
		c.chunk.Emit(OpGetProperty, nameIdx, 0)
	case *IndexExpression:
		c.compileExpression(e.Object)
		c.compileExpression(e.Index)
		c.chunk.Emit(OpIndex, 0, 0)
	case *AwaitExpression:
		c.compileExpression(e.Expr)
		c.chunk.Emit(OpAwait, 0, 0)
	case *DynamicImportExpression:
		c.compileExpression(e.Module)
		c.chunk.Emit(OpImportDynamic, 0, 0)
	case *TernaryExpression:
		c.compileExpression(e.Condition)
		jumpFalsePos := c.emit(OpJumpIfFalse, -1, 0)
		c.compileExpression(e.Then)
		jumpEndPos := c.emit(OpJump, -1, 0)
		c.patchJump(jumpFalsePos, len(c.chunk.Code))
		c.compileExpression(e.Else)
		c.patchJump(jumpEndPos, len(c.chunk.Code))
	case *SuperExpression:
		propIdx := c.chunk.AddConstant(e.Property)
		c.chunk.Emit(OpSuper, propIdx, 0)
	case *SuperCallExpression:
		for _, arg := range e.Arguments {
			c.compileExpression(arg)
		}
		c.chunk.Emit(OpSuperCall, len(e.Arguments), 0)
	case *OptionalChainExpression:
		c.compileOptionalChainExpression(e)
	case *UnaryExpression:
		c.compileExpression(e.Right)
		switch e.Operator {
		case BANG:
			c.chunk.Emit(OpNot, 0, 0)
		case MINUS:
			c.chunk.Emit(OpNeg, 0, 0)
		default:
			c.addNodeError(e, fmt.Sprintf("unsupported unary operator: %s", e.Operator))
		}
	case *UpdateExpression:
		c.compileUpdateExpression(e)
	case *TypeofExpression:
		c.compileExpression(e.Expr)
		c.chunk.Emit(OpTypeof, 0, 0)
	case *VoidExpression:
		c.compileExpression(e.Expr)
		c.chunk.Emit(OpPop, 0, 0)
		nilIdx := c.chunk.AddConstant(nil)
		c.chunk.Emit(OpConstant, nilIdx, 0)
	default:
		c.addNodeError(expr, fmt.Sprintf("unsupported expression: %T", expr))
	}
}

func (c *Compiler) compileLogicalAnd(e *BinaryExpression) {
	c.compileExpression(e.Left)
	c.compileExpression(e.Right)
	c.chunk.Emit(OpAnd, 0, 0)
}

func (c *Compiler) compileLogicalOr(e *BinaryExpression) {
	c.compileExpression(e.Left)
	c.compileExpression(e.Right)
	c.chunk.Emit(OpOr, 0, 0)
}

func (c *Compiler) compileNullishCoalescing(e *BinaryExpression) {
	c.compileExpression(e.Left)
	notNullJumpPos := c.emit(OpJumpIfNotNullish, -1, 0)
	c.chunk.Emit(OpPop, 0, 0)
	c.compileExpression(e.Right)
	c.patchJump(notNullJumpPos, len(c.chunk.Code))
}

func (c *Compiler) compileOptionalChainExpression(e *OptionalChainExpression) {
	c.compileExpression(e.Base)
	notNullJumpPos := c.emit(OpJumpIfNotNullish, -1, 0)
	c.chunk.Emit(OpPop, 0, 0)
	nullIdx := c.chunk.AddConstant(nil)
	c.chunk.Emit(OpConstant, nullIdx, 0)
	jumpToEndPos := c.emit(OpJump, -1, 0)
	c.patchJump(notNullJumpPos, len(c.chunk.Code))
	switch access := e.Access.(type) {
	case *GetExpression:
		nameIdx := c.chunk.AddConstant(access.Property)
		c.chunk.Emit(OpGetProperty, nameIdx, 0)
	case *IndexExpression:
		c.compileExpression(access.Index)
		c.chunk.Emit(OpIndex, 0, 0)
	case *CallExpression:
		if getExpr, ok := access.Callee.(*GetExpression); ok {
			nameIdx := c.chunk.AddConstant(getExpr.Property)
			c.chunk.Emit(OpGetProperty, nameIdx, 0)
			for _, arg := range access.Arguments {
				c.compileExpression(arg)
			}
			c.chunk.Emit(OpCall, len(access.Arguments), 0)
		} else {
			for _, arg := range access.Arguments {
				c.compileExpression(arg)
			}
			c.chunk.Emit(OpCall, len(access.Arguments), 0)
		}
	default:
		c.addNodeError(e, fmt.Sprintf("unsupported optional chain access: %T", access))
	}
	c.patchJump(jumpToEndPos, len(c.chunk.Code))
}

func (c *Compiler) compileUpdateExpression(e *UpdateExpression) {
	nameIdx := c.chunk.AddConstant(e.Operand.Name)

	var op OpCode
	switch e.Operator {
	case PLUSPLUS:
		op = OpAdd
	case MINUSMINUS:
		op = OpSub
	default:
		c.addNodeError(e, fmt.Sprintf("unsupported update operator: %s", e.Operator))
		return
	}

	if e.IsPrefix {
		c.chunk.Emit(OpGetName, nameIdx, 0)
		oneIdx := c.chunk.AddConstant(1.0)
		c.chunk.Emit(OpConstant, oneIdx, 0)
		c.chunk.Emit(op, 0, 0)
		c.chunk.Emit(OpDup, 0, 0)
		c.chunk.Emit(OpSetName, nameIdx, 0)
	} else {
		c.chunk.Emit(OpGetName, nameIdx, 0)
		c.chunk.Emit(OpDup, 0, 0)
		oneIdx := c.chunk.AddConstant(1.0)
		c.chunk.Emit(OpConstant, oneIdx, 0)
		c.chunk.Emit(op, 0, 0)
		c.chunk.Emit(OpSetName, nameIdx, 0)
	}
}

func (c *Compiler) compileArrayLiteral(e *ArrayLiteral) {
	hasSpread := false
	for _, elem := range e.Elements {
		if _, isSpread := elem.(*SpreadElement); isSpread {
			hasSpread = true
			break
		}
	}

	if hasSpread {
		spreadCount := 0
		for i, elem := range e.Elements {
			if spreadElem, isSpread := elem.(*SpreadElement); isSpread {
				c.compileExpression(spreadElem.Expr)
				spreadCount++
			} else {
				c.compileExpression(elem)
			}
			_ = i
		}
		c.chunk.Emit(OpSpreadArray, len(e.Elements), spreadCount)
	} else {
		for _, elem := range e.Elements {
			c.compileExpression(elem)
		}
		c.chunk.Emit(OpArray, len(e.Elements), 0)
	}
}

func (c *Compiler) compileObjectLiteral(e *ObjectLiteral) {
	hasSpread := len(e.SpreadProps) > 0

	if hasSpread {
		for _, p := range e.Properties {
			keyIdx := c.chunk.AddConstant(p.Key)
			c.chunk.Emit(OpConstant, keyIdx, 0)
			c.compileExpression(p.Value)
		}
		for _, sp := range e.SpreadProps {
			c.compileExpression(sp.Expr)
		}
		c.chunk.Emit(OpSpreadObject, len(e.Properties), len(e.SpreadProps))
	} else {
		for _, p := range e.Properties {
			keyIdx := c.chunk.AddConstant(p.Key)
			c.chunk.Emit(OpConstant, keyIdx, 0)
			c.compileExpression(p.Value)
		}
		c.chunk.Emit(OpObject, len(e.Properties), 0)
	}
}

func (c *Compiler) compileCallExpression(e *CallExpression) {
	c.compileExpression(e.Callee)

	if e.HasSpreadCall {
		for i, arg := range e.Arguments {
			isSpread := false
			for _, idx := range e.SpreadArgs {
				if idx == i {
					isSpread = true
					break
				}
			}
			c.compileExpression(arg)
			_ = isSpread
		}
		c.chunk.Emit(OpSpreadCall, len(e.Arguments), len(e.SpreadArgs))
	} else {
		for _, arg := range e.Arguments {
			c.compileExpression(arg)
		}
		c.chunk.Emit(OpCall, len(e.Arguments), 0)
	}
}
