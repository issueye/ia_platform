package compiler

import "fmt"

func (c *Compiler) compileTryStatement(s *TryCatchStatement) {
	switch {
	case s.CatchBlock != nil && s.FinallyBlock == nil:
		c.compileTryCatchOnly(s.TryBlock, s.CatchName, s.CatchBlock)
	case s.CatchBlock != nil && s.FinallyBlock != nil:
		c.compileTryCatchFinally(s.TryBlock, s.CatchName, s.CatchBlock, s.FinallyBlock)
	case s.CatchBlock == nil && s.FinallyBlock != nil:
		c.compileTryFinallyOnly(s.TryBlock, s.FinallyBlock)
	default:
		c.addNodeError(s, "invalid try statement shape")
	}
}

func (c *Compiler) compileTryCatchOnly(tryBlock *BlockStatement, catchName string, catchBlock *BlockStatement) {
	catchNameIdx := c.chunk.AddConstant(catchName)
	pushTryPos := c.emit(OpPushTry, -1, catchNameIdx)
	c.compileBlock(tryBlock)
	c.chunk.Emit(OpPopTry, 0, 0)
	jumpEndPos := c.emit(OpJump, -1, 0)
	catchPos := len(c.chunk.Code)
	c.patchJump(pushTryPos, catchPos)
	c.compileBlock(catchBlock)
	c.patchJump(jumpEndPos, len(c.chunk.Code))
}

func (c *Compiler) compileTryCatchFinally(tryBlock *BlockStatement, catchName string, catchBlock *BlockStatement, finallyBlock *BlockStatement) {
	catchNameIdx := c.chunk.AddConstant(catchName)
	pushTryPos := c.emit(OpPushTry, -1, catchNameIdx)
	c.compileBlock(tryBlock)
	c.chunk.Emit(OpPopTry, 0, 0)
	jumpAfterCatch := c.emit(OpJump, -1, 0)
	catchPos := len(c.chunk.Code)
	c.patchJump(pushTryPos, catchPos)
	c.compileBlock(catchBlock)
	afterCatchPos := len(c.chunk.Code)
	c.patchJump(jumpAfterCatch, afterCatchPos)
	c.compileBlock(finallyBlock)
}

func (c *Compiler) compileTryFinallyOnly(tryBlock *BlockStatement, finallyBlock *BlockStatement) {
	errVar := c.newTempName("err")
	hasErrVar := c.newTempName("has_err")

	falseIdx := c.chunk.AddConstant(false)
	c.chunk.Emit(OpConstant, falseIdx, 0)
	hasErrNameIdx := c.chunk.AddConstant(hasErrVar)
	c.chunk.Emit(OpDefineName, hasErrNameIdx, 0)

	nilIdx := c.chunk.AddConstant(nil)
	c.chunk.Emit(OpConstant, nilIdx, 0)
	errNameIdx := c.chunk.AddConstant(errVar)
	c.chunk.Emit(OpDefineName, errNameIdx, 0)

	pushTryPos := c.emit(OpPushTry, -1, errNameIdx)
	c.compileBlock(tryBlock)
	c.chunk.Emit(OpPopTry, 0, 0)
	jumpAfterCatch := c.emit(OpJump, -1, 0)
	catchPos := len(c.chunk.Code)
	c.patchJump(pushTryPos, catchPos)

	trueIdx := c.chunk.AddConstant(true)
	c.chunk.Emit(OpConstant, trueIdx, 0)
	c.chunk.Emit(OpSetName, hasErrNameIdx, 0)

	afterCatchPos := len(c.chunk.Code)
	c.patchJump(jumpAfterCatch, afterCatchPos)
	c.compileBlock(finallyBlock)

	c.chunk.Emit(OpGetName, hasErrNameIdx, 0)
	jumpNoRethrow := c.emit(OpJumpIfFalse, -1, 0)
	c.chunk.Emit(OpGetName, errNameIdx, 0)
	c.chunk.Emit(OpThrow, 0, 0)
	c.patchJump(jumpNoRethrow, len(c.chunk.Code))
}

func (c *Compiler) compileSwitchStatement(s *SwitchStatement) {
	c.compileExpression(s.Expression)
	c.pushLoop(-1)
	caseJumpEnds := []int{}

	for _, caseClause := range s.Cases {
		c.chunk.Emit(OpDup, 0, 0)
		c.compileExpression(caseClause.Value)
		c.chunk.Emit(OpEqual, 0, 0)
		jumpIfFalsePos := c.emit(OpJumpIfFalse, -1, 0)
		c.chunk.Emit(OpPop, 0, 0)
		for _, stmt := range caseClause.Statements {
			c.compileStatement(stmt)
		}
		jumpEndPos := c.emit(OpJump, -1, 0)
		caseJumpEnds = append(caseJumpEnds, jumpEndPos)
		c.patchJump(jumpIfFalsePos, len(c.chunk.Code))
	}

	c.chunk.Emit(OpPop, 0, 0)
	endPos := len(c.chunk.Code)

	if s.Default != nil {
		for _, stmt := range s.Default.Statements {
			c.compileStatement(stmt)
		}
	}

	loop := c.popLoop()
	for _, pos := range loop.breakJumps {
		c.patchJump(pos, endPos)
	}
	for _, pos := range caseJumpEnds {
		c.patchJump(pos, endPos)
	}
}

func (c *Compiler) compileForInStatement(s *ForInStatement) {
	c.compileExpression(s.Iterable)
	c.chunk.Emit(OpObjectKeys, 0, 0)

	keysTemp := c.newTempName("keys")
	keysIdx := c.chunk.AddConstant(keysTemp)
	c.chunk.Emit(OpDefineName, keysIdx, 0)

	zeroIdx := c.chunk.AddConstant(0.0)
	c.chunk.Emit(OpConstant, zeroIdx, 0)
	indexTemp := c.newTempName("idx")
	indexIdx := c.chunk.AddConstant(indexTemp)
	c.chunk.Emit(OpDefineName, indexIdx, 0)

	loopStart := len(c.chunk.Code)
	c.chunk.Emit(OpGetName, indexIdx, 0)
	c.chunk.Emit(OpGetName, keysIdx, 0)
	lengthPropIdx := c.chunk.AddConstant("length")
	c.chunk.Emit(OpGetProperty, lengthPropIdx, 0)
	c.chunk.Emit(OpLess, 0, 0)
	exitJumpPos := c.emit(OpJumpIfFalse, -1, 0)

	c.pushLoop(loopStart)

	c.chunk.Emit(OpGetName, keysIdx, 0)
	c.chunk.Emit(OpGetName, indexIdx, 0)
	c.chunk.Emit(OpIndex, 0, 0)
	varIdx := c.chunk.AddConstant(s.Variable)
	c.chunk.Emit(OpDefineName, varIdx, 0)

	c.compileBlock(s.Body)

	postStart := len(c.chunk.Code)
	c.setCurrentLoopContinueTarget(postStart)
	loop := c.popLoop()
	for _, j := range loop.continueJumps {
		c.patchJump(j, postStart)
	}

	oneIdx := c.chunk.AddConstant(1.0)
	c.chunk.Emit(OpGetName, indexIdx, 0)
	c.chunk.Emit(OpConstant, oneIdx, 0)
	c.chunk.Emit(OpAdd, 0, 0)
	c.chunk.Emit(OpSetName, indexIdx, 0)

	c.chunk.Emit(OpJump, loopStart, 0)

	exitPos := len(c.chunk.Code)
	c.patchJump(exitJumpPos, exitPos)
	for _, j := range loop.breakJumps {
		c.patchJump(j, exitPos)
	}
}

func (c *Compiler) compileForOfStatement(s *ForOfStatement) {
	iterTemp := c.newTempName("iter")
	iterIdx := c.chunk.AddConstant(iterTemp)
	c.compileExpression(s.Iterable)
	c.chunk.Emit(OpDefineName, iterIdx, 0)

	zeroIdx := c.chunk.AddConstant(0.0)
	c.chunk.Emit(OpConstant, zeroIdx, 0)
	indexTemp := c.newTempName("idx")
	indexIdx := c.chunk.AddConstant(indexTemp)
	c.chunk.Emit(OpDefineName, indexIdx, 0)

	loopStart := len(c.chunk.Code)
	c.chunk.Emit(OpGetName, indexIdx, 0)
	c.chunk.Emit(OpGetName, iterIdx, 0)
	lengthPropIdx := c.chunk.AddConstant("length")
	c.chunk.Emit(OpGetProperty, lengthPropIdx, 0)
	c.chunk.Emit(OpLess, 0, 0)
	exitJumpPos := c.emit(OpJumpIfFalse, -1, 0)

	c.pushLoop(loopStart)

	c.chunk.Emit(OpGetName, iterIdx, 0)
	c.chunk.Emit(OpGetName, indexIdx, 0)
	c.chunk.Emit(OpIndex, 0, 0)
	varIdx := c.chunk.AddConstant(s.Variable)
	c.chunk.Emit(OpDefineName, varIdx, 0)

	c.compileBlock(s.Body)

	postStart := len(c.chunk.Code)
	c.setCurrentLoopContinueTarget(postStart)
	loop := c.popLoop()
	for _, j := range loop.continueJumps {
		c.patchJump(j, postStart)
	}

	oneIdx := c.chunk.AddConstant(1.0)
	c.chunk.Emit(OpGetName, indexIdx, 0)
	c.chunk.Emit(OpConstant, oneIdx, 0)
	c.chunk.Emit(OpAdd, 0, 0)
	c.chunk.Emit(OpSetName, indexIdx, 0)

	c.chunk.Emit(OpJump, loopStart, 0)

	exitPos := len(c.chunk.Code)
	c.patchJump(exitJumpPos, exitPos)
	for _, j := range loop.breakJumps {
		c.patchJump(j, exitPos)
	}
}

func (c *Compiler) compileDoWhileStatement(s *DoWhileStatement) {
	loopStart := len(c.chunk.Code)
	c.pushLoop(loopStart)
	c.compileBlock(s.Body)
	loop := c.popLoop()
	for _, j := range loop.continueJumps {
		c.patchJump(j, loopStart)
	}
	c.compileExpression(s.Condition)
	c.chunk.Emit(OpJumpIfTrue, loopStart, 0)
	exitPos := len(c.chunk.Code)
	for _, j := range loop.breakJumps {
		c.patchJump(j, exitPos)
	}
}

func (c *Compiler) newTempName(suffix string) string {
	name := fmt.Sprintf("__ialang_tmp_%s_%d", suffix, c.tempVarSeq)
	c.tempVarSeq++
	return name
}

func (c *Compiler) compileArrayDestructuringLet(s *ArrayDestructuringLetStatement) {
	tempName := c.newTempName("arr_destruct")
	tempIdx := c.chunk.AddConstant(tempName)

	c.compileExpression(s.Initializer)
	c.chunk.Emit(OpDefineName, tempIdx, 0)

	for i, name := range s.Names {
		c.chunk.Emit(OpGetName, tempIdx, 0)
		indexIdx := c.chunk.AddConstant(float64(i))
		c.chunk.Emit(OpConstant, indexIdx, 0)
		c.chunk.Emit(OpIndex, 0, 0)

		nameIdx := c.chunk.AddConstant(name)
		c.chunk.Emit(OpDefineName, nameIdx, 0)
	}
}

func (c *Compiler) compileObjectDestructuringLet(s *ObjectDestructuringLetStatement) {
	tempName := c.newTempName("obj_destruct")
	tempIdx := c.chunk.AddConstant(tempName)

	c.compileExpression(s.Initializer)
	c.chunk.Emit(OpDefineName, tempIdx, 0)

	for _, binding := range s.Bindings {
		c.chunk.Emit(OpGetName, tempIdx, 0)
		keyIdx := c.chunk.AddConstant(binding.Key)
		c.chunk.Emit(OpGetProperty, keyIdx, 0)

		nameIdx := c.chunk.AddConstant(binding.Name)
		c.chunk.Emit(OpDefineName, nameIdx, 0)
	}
}

func (c *Compiler) compileArrayDestructuringAssign(s *ArrayDestructuringAssignStatement) {
	tempName := c.newTempName("arr_assign_destruct")
	tempIdx := c.chunk.AddConstant(tempName)

	c.compileExpression(s.Value)
	c.chunk.Emit(OpDefineName, tempIdx, 0)

	for i, name := range s.Names {
		c.chunk.Emit(OpGetName, tempIdx, 0)
		indexIdx := c.chunk.AddConstant(float64(i))
		c.chunk.Emit(OpConstant, indexIdx, 0)
		c.chunk.Emit(OpIndex, 0, 0)

		nameIdx := c.chunk.AddConstant(name)
		c.chunk.Emit(OpSetName, nameIdx, 0)
	}
}

func (c *Compiler) compileObjectDestructuringAssign(s *ObjectDestructuringAssignStatement) {
	tempName := c.newTempName("obj_assign_destruct")
	tempIdx := c.chunk.AddConstant(tempName)

	c.compileExpression(s.Value)
	c.chunk.Emit(OpDefineName, tempIdx, 0)

	for _, binding := range s.Bindings {
		c.chunk.Emit(OpGetName, tempIdx, 0)
		keyIdx := c.chunk.AddConstant(binding.Key)
		c.chunk.Emit(OpGetProperty, keyIdx, 0)

		nameIdx := c.chunk.AddConstant(binding.Name)
		c.chunk.Emit(OpSetName, nameIdx, 0)
	}
}

func (c *Compiler) compileCompoundAssign(s *CompoundAssignStatement) {
	nameIdx := c.chunk.AddConstant(s.Name)
	c.chunk.Emit(OpGetName, nameIdx, 0)
	c.compileExpression(s.Value)

	var op OpCode
	switch s.Operator {
	case PLUSEQ:
		op = OpAdd
	case MINUSEQ:
		op = OpSub
	case MULTEQ:
		op = OpMul
	case DIVEQ:
		op = OpDiv
	case MODEQ:
		op = OpMod
	default:
		c.addNodeError(s, fmt.Sprintf("unsupported compound assignment operator: %s", s.Operator))
		return
	}
	c.chunk.Emit(op, 0, 0)
	c.chunk.Emit(OpSetName, nameIdx, 0)
}

func (c *Compiler) compileCompoundSetProperty(s *CompoundSetPropertyStatement) {
	tempName := c.newTempName("compound")

	c.compileExpression(s.Object)
	propIdx := c.chunk.AddConstant(s.Property)
	c.chunk.Emit(OpGetProperty, propIdx, 0)
	c.compileExpression(s.Value)

	var op OpCode
	switch s.Operator {
	case PLUSEQ:
		op = OpAdd
	case MINUSEQ:
		op = OpSub
	case MULTEQ:
		op = OpMul
	case DIVEQ:
		op = OpDiv
	case MODEQ:
		op = OpMod
	default:
		c.addNodeError(s, fmt.Sprintf("unsupported compound property assignment operator: %s", s.Operator))
		return
	}
	c.chunk.Emit(op, 0, 0)

	tempNameIdx := c.chunk.AddConstant(tempName)
	c.chunk.Emit(OpDefineName, tempNameIdx, 0)

	c.compileExpression(s.Object)
	c.chunk.Emit(OpGetName, tempNameIdx, 0)
	c.chunk.Emit(OpSetProperty, propIdx, 0)
}
