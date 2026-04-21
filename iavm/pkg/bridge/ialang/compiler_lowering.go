package ialang

import (
	"fmt"
	"iacommon/pkg/ialang/bytecode"
	"iavm/pkg/core"
	"iavm/pkg/module"
)

func LowerToModule(input any) (*module.Module, error) {
	chunk, ok := input.(*bytecode.Chunk)
	if !ok {
		return nil, fmt.Errorf("expected *bytecode.Chunk, got %T", input)
	}
	if chunk == nil {
		return nil, fmt.Errorf("nil chunk")
	}

	mod := &module.Module{
		Magic:      "IAVM",
		Version:    1,
		Target:     "ialang",
		ABIVersion: 1,
	}

	// Build default type: () -> null
	mod.Types = append(mod.Types, core.FuncType{
		Params:  []core.ValueKind{},
		Results: []core.ValueKind{core.ValueNull},
	})

	// First pass: collect all function templates and build index map
	funcIndexMap := make(map[int]int)
	for i, c := range chunk.Constants {
		if ft, ok := c.(*bytecode.FunctionTemplate); ok {
			funcIndexMap[i] = len(mod.Functions)
			mod.Functions = append(mod.Functions, lowerFunction(ft))
		}
	}

	// Create entry function from top-level chunk
	entryFunc := lowerChunkAsFunction(chunk, "entry")
	entryFunc.IsEntryPoint = true
	mod.Functions = append(mod.Functions, entryFunc)

	// Add exports for named functions
	for i, c := range chunk.Constants {
		if ft, ok := c.(*bytecode.FunctionTemplate); ok {
			if ft.Name != "" {
				mod.Exports = append(mod.Exports, module.Export{
					Name:  ft.Name,
					Kind:  module.ExportFunction,
					Index: uint32(funcIndexMap[i]),
				})
			}
		}
	}

	return mod, nil
}

func lowerFunction(ft *bytecode.FunctionTemplate) module.Function {
	fn := module.Function{
		Name:      ft.Name,
		TypeIndex: 0,
	}

	// Lower locals: params + implicit self slot
	totalLocals := len(ft.Params) + 1
	if ft.RestParam != "" {
		totalLocals++
	}
	for i := 0; i < totalLocals; i++ {
		fn.Locals = append(fn.Locals, core.ValueNull)
	}

	// Lower constants from chunk
	if ft.Chunk != nil {
		fn.Constants = lowerConstants(ft.Chunk.Constants, ft.Chunk.Code)
		fn.Code = lowerInstructions(ft.Chunk.Code)
	}

	return fn
}

func lowerChunkAsFunction(chunk *bytecode.Chunk, name string) module.Function {
	fn := module.Function{
		Name:      name,
		TypeIndex: 0,
	}
	fn.Constants = lowerConstants(chunk.Constants, chunk.Code)
	fn.Code = lowerInstructions(chunk.Code)
	return fn
}

func lowerConstants(constants []any, code []bytecode.Instruction) []any {
	var result []any
	for _, c := range constants {
		switch v := c.(type) {
		case *bytecode.FunctionTemplate:
			// Skip function templates in constants - they become separate functions
			continue
		case nil, bool, int, int64, float64, string:
			result = append(result, c)
		default:
			_ = v
			result = append(result, nil)
		}
	}
	return result
}

func lowerInstructions(ialangInsts []bytecode.Instruction) []core.Instruction {
	var iavmInsts []core.Instruction

	for _, inst := range ialangInsts {
		iavmInst := core.Instruction{}

		switch inst.Op {
		case bytecode.OpConstant:
			iavmInst.Op = core.OpConst
			iavmInst.A = uint32(inst.A)

		case bytecode.OpAdd:
			iavmInst.Op = core.OpAdd

		case bytecode.OpSub:
			iavmInst.Op = core.OpSub

		case bytecode.OpMul:
			iavmInst.Op = core.OpMul

		case bytecode.OpDiv:
			iavmInst.Op = core.OpDiv

		case bytecode.OpMod:
			iavmInst.Op = core.OpMod

		case bytecode.OpNeg:
			iavmInst.Op = core.OpNeg

		case bytecode.OpNot:
			iavmInst.Op = core.OpNot

		case bytecode.OpEqual:
			iavmInst.Op = core.OpEq

		case bytecode.OpNotEqual:
			iavmInst.Op = core.OpNe

		case bytecode.OpGreater:
			iavmInst.Op = core.OpGt

		case bytecode.OpLess:
			iavmInst.Op = core.OpLt

		case bytecode.OpGreaterEqual:
			iavmInst.Op = core.OpGe

		case bytecode.OpLessEqual:
			iavmInst.Op = core.OpLe

		case bytecode.OpJump:
			iavmInst.Op = core.OpJump
			iavmInst.A = uint32(inst.A)

		case bytecode.OpJumpIfFalse:
			iavmInst.Op = core.OpJumpIfFalse
			iavmInst.A = uint32(inst.A)

		case bytecode.OpJumpIfTrue:
			// Not: jump_if_true = jump_if_false(not(val))
			iavmInst.Op = core.OpNot
			iavmInsts = append(iavmInsts, iavmInst)
			iavmInst = core.Instruction{Op: core.OpJumpIfFalse, A: uint32(inst.A)}

		case bytecode.OpCall:
			// ialang's OpCall pops function from stack and calls it
			// iavm's OpCall with A=argCount, B=0 means stack-based call
			iavmInst.Op = core.OpCall
			iavmInst.A = uint32(inst.A) // arg count hint
			iavmInst.B = 0

		case bytecode.OpReturn:
			iavmInst.Op = core.OpReturn

		case bytecode.OpGetGlobal:
			iavmInst.Op = core.OpLoadGlobal
			iavmInst.A = uint32(inst.A)

		case bytecode.OpDefineGlobal:
			iavmInst.Op = core.OpStoreGlobal
			iavmInst.A = uint32(inst.A)

		case bytecode.OpGetName:
			iavmInst.Op = core.OpLoadGlobal
			iavmInst.A = uint32(inst.A)

		case bytecode.OpSetName:
			iavmInst.Op = core.OpStoreGlobal
			iavmInst.A = uint32(inst.A)

		case bytecode.OpDefineName:
			iavmInst.Op = core.OpStoreGlobal
			iavmInst.A = uint32(inst.A)

		case bytecode.OpArray:
			iavmInst.Op = core.OpMakeArray
			iavmInst.A = uint32(inst.A)

		case bytecode.OpObject:
			iavmInst.Op = core.OpMakeObject

		case bytecode.OpGetProperty:
			iavmInst.Op = core.OpGetProp
			iavmInst.A = uint32(inst.A)

		case bytecode.OpSetProperty:
			iavmInst.Op = core.OpSetProp
			iavmInst.A = uint32(inst.A)

		case bytecode.OpImportName:
			iavmInst.Op = core.OpImportFunc
			iavmInst.A = uint32(inst.A)

		case bytecode.OpPop:
			iavmInst.Op = core.OpPop

		case bytecode.OpDup:
			iavmInst.Op = core.OpDup

		case bytecode.OpTruthy:
			iavmInst.Op = core.OpNot
			iavmInsts = append(iavmInsts, iavmInst)
			iavmInst = core.Instruction{Op: core.OpNot}

		case bytecode.OpAnd:
			iavmInst.Op = core.OpAnd

		case bytecode.OpOr:
			iavmInst.Op = core.OpOr

		case bytecode.OpClosure:
			// Closures become function references
			iavmInst.Op = core.OpConst
			iavmInst.A = uint32(inst.A)

		case bytecode.OpClass:
			iavmInst.Op = core.OpMakeObject

		case bytecode.OpNew:
			iavmInst.Op = core.OpCall
			iavmInst.A = uint32(inst.A)
			iavmInst.B = uint32(inst.B)

		case bytecode.OpIndex:
			// Index access maps to property get
			iavmInst.Op = core.OpGetProp
			iavmInst.A = uint32(inst.A)

		case bytecode.OpBitAnd:
			iavmInst.Op = core.OpBitAnd

		case bytecode.OpBitOr:
			iavmInst.Op = core.OpBitOr

		case bytecode.OpBitXor:
			iavmInst.Op = core.OpBitXor

		case bytecode.OpShl:
			iavmInst.Op = core.OpShl

		case bytecode.OpShr:
			iavmInst.Op = core.OpShr

		case bytecode.OpTypeof:
			iavmInst.Op = core.OpTypeof

		case bytecode.OpPushTry:
			iavmInst.Op = core.OpPushTry
			iavmInst.A = uint32(inst.A)

		case bytecode.OpPopTry:
			iavmInst.Op = core.OpPopTry

		case bytecode.OpThrow:
			iavmInst.Op = core.OpThrow

		case bytecode.OpImportNamespace, bytecode.OpImportDynamic,
			bytecode.OpExportName, bytecode.OpExportAs, bytecode.OpExportDefault,
			bytecode.OpExportAll, bytecode.OpSuper, bytecode.OpSuperCall,
			bytecode.OpObjectKeys, bytecode.OpSpreadArray,
			bytecode.OpSpreadObject, bytecode.OpSpreadCall, bytecode.OpAwait,
			bytecode.OpJumpIfNullish, bytecode.OpJumpIfNotNullish:
			// Unsupported in minimal iavm, use Nop
			iavmInst.Op = core.OpNop

		default:
			iavmInst.Op = core.OpNop
		}

		iavmInsts = append(iavmInsts, iavmInst)
	}

	return iavmInsts
}
