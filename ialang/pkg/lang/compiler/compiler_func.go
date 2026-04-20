package compiler

func (c *Compiler) compileFunctionTemplate(name string, params []string, restParam string, paramDefaults []DefaultParam, async bool, body *BlockStatement) *FunctionTemplate {
	bodyStmtCount := 0
	if body != nil {
		bodyStmtCount = len(body.Statements)
	}
	fnChunk := NewChunkSized(bodyStmtCount*8+4+len(params)*10, bodyStmtCount*4+4)
	fnCompiler := &Compiler{chunk: fnChunk}

	fnCompiler.compileDefaultParams(params, paramDefaults)

	fnCompiler.compileBlock(body)
	fnChunk.Emit(OpReturn, 0, 0)
	if len(fnCompiler.errors) > 0 {
		c.errors = append(c.errors, fnCompiler.errors...)
		return nil
	}

	defaultConsts := make([]any, len(params))
	for i, dp := range paramDefaults {
		if dp.Value != nil {
			defaultConsts[i] = dp.Value
		}
	}

	return &FunctionTemplate{
		Name:          name,
		Params:        params,
		RestParam:     restParam,
		ParamDefaults: defaultConsts,
		Async:         async,
		Chunk:         fnChunk,
	}
}

func (c *Compiler) compileDefaultParams(params []string, paramDefaults []DefaultParam) {
	for i, dp := range paramDefaults {
		if dp.Value == nil {
			continue
		}

		nameIdx := c.chunk.AddConstant(params[i])
		c.chunk.Emit(OpGetName, nameIdx, 0)
		notNullJumpPos := c.emit(OpJumpIfNotNullish, -1, 0)
		c.chunk.Emit(OpPop, 0, 0)
		c.compileExpression(dp.Value)
		c.chunk.Emit(OpSetName, nameIdx, 0)
		c.patchJump(notNullJumpPos, len(c.chunk.Code))
	}
}

func (c *Compiler) compileArrowFunction(e *ArrowFunctionExpression) *FunctionTemplate {
	bodyStmtCount := 0
	if e.Body != nil {
		bodyStmtCount = len(e.Body.Statements)
	}
	fnChunk := NewChunkSized(bodyStmtCount*8+8+len(e.Params)*10, bodyStmtCount*4+4)
	fnCompiler := &Compiler{chunk: fnChunk}

	fnCompiler.compileDefaultParams(e.Params, e.ParamDefaults)

	if e.Concise {
		fnCompiler.compileExpression(e.Expr)
		fnCompiler.chunk.Emit(OpReturn, 1, 0)
	} else {
		fnCompiler.compileBlock(e.Body)
		fnCompiler.chunk.Emit(OpReturn, 0, 0)
	}

	if len(fnCompiler.errors) > 0 {
		c.errors = append(c.errors, fnCompiler.errors...)
		return nil
	}

	defaultConsts := make([]any, len(e.Params))
	for i, dp := range e.ParamDefaults {
		if dp.Value != nil {
			defaultConsts[i] = dp.Value
		}
	}

	return &FunctionTemplate{
		Name:          "",
		Params:        e.Params,
		RestParam:     e.RestParam,
		ParamDefaults: defaultConsts,
		Async:         e.Async,
		Chunk:         fnChunk,
	}
}

func (c *Compiler) compileClassStatement(s *ClassStatement) {
	if s.ParentName != "" {
		parentIdx := c.chunk.AddConstant(s.ParentName)
		c.chunk.Emit(OpConstant, parentIdx, 0)
	}
	classNameConst := c.chunk.AddConstant(s.Name)
	c.chunk.Emit(OpConstant, classNameConst, 0)

	instanceMethods := []ClassMethod{}
	staticMethods := []ClassMethod{}
	getters := []ClassMethod{}
	setters := []ClassMethod{}

	for _, m := range s.Methods {
		switch {
		case m.IsGetter:
			getters = append(getters, m)
		case m.IsSetter:
			setters = append(setters, m)
		case m.Static:
			staticMethods = append(staticMethods, m)
		default:
			instanceMethods = append(instanceMethods, m)
		}
	}

	for _, m := range instanceMethods {
		methodNameConst := c.chunk.AddConstant(m.Name)
		c.chunk.Emit(OpConstant, methodNameConst, 0)
		fn := c.compileFunctionTemplate(m.Name, m.Params, m.RestParam, m.ParamDefaults, m.Async, m.Body)
		if fn == nil {
			return
		}
		fnIdx := c.chunk.AddConstant(fn)
		c.chunk.Emit(OpClosure, fnIdx, 0)
	}

	for _, m := range staticMethods {
		methodNameConst := c.chunk.AddConstant(m.Name)
		c.chunk.Emit(OpConstant, methodNameConst, 0)
		fn := c.compileFunctionTemplate(m.Name, m.Params, m.RestParam, m.ParamDefaults, m.Async, m.Body)
		if fn == nil {
			return
		}
		fnIdx := c.chunk.AddConstant(fn)
		c.chunk.Emit(OpClosure, fnIdx, 0)
	}

	for _, m := range getters {
		methodNameConst := c.chunk.AddConstant(m.Name)
		c.chunk.Emit(OpConstant, methodNameConst, 0)
		fn := c.compileFunctionTemplate(m.Name, m.Params, m.RestParam, m.ParamDefaults, m.Async, m.Body)
		if fn == nil {
			return
		}
		fnIdx := c.chunk.AddConstant(fn)
		c.chunk.Emit(OpClosure, fnIdx, 0)
	}

	for _, m := range setters {
		methodNameConst := c.chunk.AddConstant(m.Name)
		c.chunk.Emit(OpConstant, methodNameConst, 0)
		fn := c.compileFunctionTemplate(m.Name, m.Params, m.RestParam, m.ParamDefaults, m.Async, m.Body)
		if fn == nil {
			return
		}
		fnIdx := c.chunk.AddConstant(fn)
		c.chunk.Emit(OpClosure, fnIdx, 0)
	}

	for _, f := range s.PrivateFields {
		fieldNameConst := c.chunk.AddConstant(f.Name)
		c.chunk.Emit(OpConstant, fieldNameConst, 0)
	}

	hasParent := 0
	if s.ParentName != "" {
		hasParent = 1
	}

	operand := len(instanceMethods) |
		(len(staticMethods) << 4) |
		(len(getters) << 8) |
		(len(setters) << 12) |
		(len(s.PrivateFields) << 16) |
		(hasParent << 20)
	c.chunk.Emit(OpClass, operand, 0)
	nameIdx := c.chunk.AddConstant(s.Name)
	c.chunk.Emit(OpDefineName, nameIdx, 0)
}
