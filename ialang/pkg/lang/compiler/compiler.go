package compiler

import (
	"fmt"
)

type Compiler struct {
	chunk      *Chunk
	errors     []string
	loops      []loopContext
	tempVarSeq int
}

func NewCompiler() *Compiler {
	return &Compiler{chunk: NewChunk()}
}

type loopContext struct {
	breakJumps     []int
	continueJumps  []int
	continueTarget int
}

func (c *Compiler) addError(msg string) {
	c.errors = append(c.errors, msg)
}

func (c *Compiler) addNodeError(node Node, msg string) {
	if node != nil {
		pos := node.Pos()
		if pos.IsValid() {
			msg = fmt.Sprintf("%s (line %d, col %d)", msg, pos.Line, pos.Column)
		}
	}
	c.addError(msg)
}

func (c *Compiler) Compile(program *Program) (*Chunk, []string) {
	c.ensureChunkCapacity(program)
	for _, stmt := range program.Statements {
		c.compileStatement(stmt)
	}
	c.chunk.Emit(OpReturn, 0, 0)
	return c.chunk, c.errors
}

func (c *Compiler) ensureChunkCapacity(program *Program) {
	if program == nil {
		return
	}
	stmtCount := len(program.Statements)
	if stmtCount == 0 {
		return
	}
	codeCap := stmtCount*8 + 8
	constCap := stmtCount*4 + 8
	if len(c.chunk.Code) == 0 {
		c.chunk.Code = make([]Instruction, 0, codeCap)
	}
	if len(c.chunk.Constants) == 0 {
		c.chunk.Constants = make([]any, 0, constCap)
	}
}

func (c *Compiler) compileStatement(stmt Statement) {
	switch s := stmt.(type) {
	case *ImportStatement:
		if s.Namespace != "" {
			moduleIdx := c.chunk.AddConstant(s.Module)
			nameIdx := c.chunk.AddConstant(s.Namespace)
			c.chunk.Emit(OpImportNamespace, moduleIdx, nameIdx)
			return
		}
		for _, name := range s.Names {
			moduleIdx := c.chunk.AddConstant(s.Module)
			nameIdx := c.chunk.AddConstant(name)
			c.chunk.Emit(OpImportName, moduleIdx, nameIdx)
		}
	case *ExportStatement:
		switch {
		case s.DefaultName != "" && s.Statement != nil:
			c.compileStatement(s.Statement)
			nameIdx := c.chunk.AddConstant(s.DefaultName)
			c.chunk.Emit(OpGetName, nameIdx, 0)
			c.chunk.Emit(OpExportDefault, 0, 0)
		case s.Statement != nil:
			c.compileStatement(s.Statement)
			for _, name := range c.exportNames(s.Statement) {
				nameIdx := c.chunk.AddConstant(name)
				c.chunk.Emit(OpExportName, nameIdx, 0)
			}
		case s.ExportAllModule != "":
			moduleIdx := c.chunk.AddConstant(s.ExportAllModule)
			c.chunk.Emit(OpExportAll, moduleIdx, 0)
		case len(s.Specifiers) > 0:
			for _, spec := range s.Specifiers {
				localIdx := c.chunk.AddConstant(spec.LocalName)
				exportIdx := c.chunk.AddConstant(spec.ExportName)
				if spec.LocalName == spec.ExportName {
					c.chunk.Emit(OpExportName, localIdx, 0)
					continue
				}
				c.chunk.Emit(OpExportAs, localIdx, exportIdx)
			}
		case s.Default != nil:
			c.compileExpression(s.Default)
			c.chunk.Emit(OpExportDefault, 0, 0)
		default:
			c.addNodeError(s, "empty export statement")
			return
		}
	case *LetStatement:
		c.compileExpression(s.Initializer)
		nameIdx := c.chunk.AddConstant(s.Name)
		c.chunk.Emit(OpDefineName, nameIdx, 0)
	case *ArrayDestructuringLetStatement:
		c.compileArrayDestructuringLet(s)
	case *ObjectDestructuringLetStatement:
		c.compileObjectDestructuringLet(s)
	case *ArrayDestructuringAssignStatement:
		c.compileArrayDestructuringAssign(s)
	case *ObjectDestructuringAssignStatement:
		c.compileObjectDestructuringAssign(s)
	case *AssignStatement:
		c.compileExpression(s.Value)
		nameIdx := c.chunk.AddConstant(s.Name)
		c.chunk.Emit(OpSetName, nameIdx, 0)
	case *CompoundAssignStatement:
		c.compileCompoundAssign(s)
	case *SetPropertyStatement:
		c.compileExpression(s.Object)
		c.compileExpression(s.Value)
		nameIdx := c.chunk.AddConstant(s.Property)
		c.chunk.Emit(OpSetProperty, nameIdx, 0)
	case *CompoundSetPropertyStatement:
		c.compileCompoundSetProperty(s)
	case *FunctionStatement:
		fn := c.compileFunctionTemplate(s.Name, s.Params, s.RestParam, s.ParamDefaults, s.Async, s.Body)
		if fn == nil {
			return
		}
		fnIdx := c.chunk.AddConstant(fn)
		c.chunk.Emit(OpClosure, fnIdx, 0)
		nameIdx := c.chunk.AddConstant(s.Name)
		c.chunk.Emit(OpDefineName, nameIdx, 0)
	case *ClassStatement:
		c.compileClassStatement(s)
	case *ReturnStatement:
		if s.Value != nil {
			c.compileExpression(s.Value)
			c.chunk.Emit(OpReturn, 1, 0)
		} else {
			c.chunk.Emit(OpReturn, 0, 0)
		}
	case *ThrowStatement:
		if s.Value != nil {
			c.compileExpression(s.Value)
		} else {
			nilIdx := c.chunk.AddConstant(nil)
			c.chunk.Emit(OpConstant, nilIdx, 0)
		}
		c.chunk.Emit(OpThrow, 0, 0)
	case *TryCatchStatement:
		c.compileTryStatement(s)
	case *IfStatement:
		c.compileExpression(s.Condition)
		jumpIfFalsePos := c.emit(OpJumpIfFalse, -1, 0)
		c.compileBlock(s.Then)
		if s.Else != nil {
			jumpEndPos := c.emit(OpJump, -1, 0)
			c.patchJump(jumpIfFalsePos, len(c.chunk.Code))
			c.compileBlock(s.Else)
			c.patchJump(jumpEndPos, len(c.chunk.Code))
		} else {
			c.patchJump(jumpIfFalsePos, len(c.chunk.Code))
		}
	case *WhileStatement:
		loopStart := len(c.chunk.Code)
		c.compileExpression(s.Condition)
		exitJumpPos := c.emit(OpJumpIfFalse, -1, 0)
		c.pushLoop(loopStart)
		c.compileBlock(s.Body)
		loop := c.popLoop()
		for _, j := range loop.continueJumps {
			c.patchJump(j, loopStart)
		}
		c.chunk.Emit(OpJump, loopStart, 0)
		exitPos := len(c.chunk.Code)
		c.patchJump(exitJumpPos, exitPos)
		for _, j := range loop.breakJumps {
			c.patchJump(j, exitPos)
		}
	case *DoWhileStatement:
		c.compileDoWhileStatement(s)
	case *ForStatement:
		if s.Init != nil {
			c.compileStatement(s.Init)
		}
		loopStart := len(c.chunk.Code)
		if s.Condition != nil {
			c.compileExpression(s.Condition)
		} else {
			trueIdx := c.chunk.AddConstant(true)
			c.chunk.Emit(OpConstant, trueIdx, 0)
		}
		exitJumpPos := c.emit(OpJumpIfFalse, -1, 0)
		c.pushLoop(-1)
		c.compileBlock(s.Body)
		postStart := len(c.chunk.Code)
		c.setCurrentLoopContinueTarget(postStart)
		loop := c.popLoop()
		for _, j := range loop.continueJumps {
			c.patchJump(j, postStart)
		}
		if s.Post != nil {
			c.compileStatement(s.Post)
		}
		c.chunk.Emit(OpJump, loopStart, 0)
		exitPos := len(c.chunk.Code)
		c.patchJump(exitJumpPos, exitPos)
		for _, j := range loop.breakJumps {
			c.patchJump(j, exitPos)
		}
	case *ForInStatement:
		c.compileForInStatement(s)
	case *ForOfStatement:
		c.compileForOfStatement(s)
	case *BreakStatement:
		j := c.emit(OpJump, -1, 0)
		c.addLoopBreakJump(s, j)
	case *ContinueStatement:
		j := c.emit(OpJump, -1, 0)
		c.addLoopContinueJump(s, j)
	case *ExpressionStatement:
		c.compileExpression(s.Expr)
		c.chunk.Emit(OpPop, 0, 0)
	case *SwitchStatement:
		c.compileSwitchStatement(s)
	default:
		c.addNodeError(stmt, fmt.Sprintf("unsupported statement: %T", stmt))
	}
}

func (c *Compiler) compileBlock(block *BlockStatement) {
	if block == nil {
		return
	}
	for _, stmt := range block.Statements {
		c.compileStatement(stmt)
	}
}

func (c *Compiler) emit(op OpCode, a, b int) int {
	pos := len(c.chunk.Code)
	c.chunk.Emit(op, a, b)
	return pos
}

func (c *Compiler) patchJump(pos, target int) {
	if pos < 0 || pos >= len(c.chunk.Code) {
		c.addError(fmt.Sprintf("invalid jump patch position: %d", pos))
		return
	}
	c.chunk.Code[pos].A = target
}

func (c *Compiler) pushLoop(continueTarget int) {
	c.loops = append(c.loops, loopContext{continueTarget: continueTarget})
}

func (c *Compiler) popLoop() loopContext {
	if len(c.loops) == 0 {
		c.addError("internal error: popLoop with empty stack")
		return loopContext{}
	}
	last := len(c.loops) - 1
	loop := c.loops[last]
	c.loops = c.loops[:last]
	return loop
}

func (c *Compiler) setCurrentLoopContinueTarget(target int) {
	if len(c.loops) == 0 {
		c.addError("internal error: set continue target outside loop")
		return
	}
	c.loops[len(c.loops)-1].continueTarget = target
}

func (c *Compiler) addLoopBreakJump(stmt *BreakStatement, pos int) {
	if len(c.loops) == 0 {
		c.addNodeError(stmt, "break used outside of loop")
		return
	}
	i := len(c.loops) - 1
	c.loops[i].breakJumps = append(c.loops[i].breakJumps, pos)
}

func (c *Compiler) addLoopContinueJump(stmt *ContinueStatement, pos int) {
	if len(c.loops) == 0 {
		c.addNodeError(stmt, "continue used outside of loop")
		return
	}
	i := len(c.loops) - 1
	c.loops[i].continueJumps = append(c.loops[i].continueJumps, pos)
}

func (c *Compiler) exportNames(stmt Statement) []string {
	switch s := stmt.(type) {
	case *LetStatement:
		return []string{s.Name}
	case *ArrayDestructuringLetStatement:
		return append([]string(nil), s.Names...)
	case *ObjectDestructuringLetStatement:
		names := make([]string, 0, len(s.Bindings))
		for _, b := range s.Bindings {
			names = append(names, b.Name)
		}
		return names
	case *FunctionStatement:
		return []string{s.Name}
	case *ClassStatement:
		return []string{s.Name}
	case *CompoundAssignStatement:
		return []string{}
	case *CompoundSetPropertyStatement:
		return []string{}
	default:
		return []string{}
	}
}
