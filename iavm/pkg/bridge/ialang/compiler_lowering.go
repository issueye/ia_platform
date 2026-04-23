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
	globalNames := collectGlobalNames(chunk)
	for i, c := range chunk.Constants {
		if _, ok := c.(*bytecode.FunctionTemplate); ok {
			funcIndexMap[i] = len(mod.Functions)
			mod.Functions = append(mod.Functions, module.Function{})
		}
	}

	// Second pass: lower each function with full funcIndexMap available
	for i, c := range chunk.Constants {
		if ft, ok := c.(*bytecode.FunctionTemplate); ok {
			mod.Functions[funcIndexMap[i]] = lowerFunction(ft, globalNames, funcIndexMap)
		}
	}

	// Collect global names from the chunk
	globalNames = collectGlobalNames(chunk)
	mod.Globals = buildGlobals(globalNames)

	// Create entry function from top-level chunk with global remapping
	entryFunc := lowerChunkAsFunctionWithGlobals(chunk, "entry", globalNames, funcIndexMap)
	entryFunc.IsEntryPoint = true
	mod.Functions = append(mod.Functions, entryFunc)

	// Add exports for named functions
	exportSet := make(map[string]struct{})
	for i, c := range chunk.Constants {
		if ft, ok := c.(*bytecode.FunctionTemplate); ok {
			if ft.Name != "" {
				mod.Exports = append(mod.Exports, module.Export{
					Name:  ft.Name,
					Kind:  module.ExportFunction,
					Index: uint32(funcIndexMap[i]),
				})
				exportSet[ft.Name] = struct{}{}
			}
		}
	}

	// Add exports for global variables (OpExportName/OpExportAs/OpExportDefault in top-level chunk)
	for _, inst := range chunk.Code {
		switch inst.Op {
		case bytecode.OpExportName:
			if int(inst.A) < len(chunk.Constants) {
				if name, ok := chunk.Constants[inst.A].(string); ok {
					if _, alreadyExported := exportSet[name]; !alreadyExported {
						if idx, exists := globalNames[name]; exists {
							mod.Exports = append(mod.Exports, module.Export{
								Name:  name,
								Kind:  module.ExportGlobal,
								Index: uint32(idx),
							})
							exportSet[name] = struct{}{}
						}
					}
				}
			}
		case bytecode.OpExportAs:
			if int(inst.A) < len(chunk.Constants) && int(inst.B) < len(chunk.Constants) {
				localName, localOK := chunk.Constants[inst.A].(string)
				exportName, exportOK := chunk.Constants[inst.B].(string)
				if localOK && exportOK {
					if _, alreadyExported := exportSet[exportName]; alreadyExported {
						continue
					}
					if idx, exists := globalNames[localName]; exists {
						mod.Exports = append(mod.Exports, module.Export{
							Name:  exportName,
							Kind:  module.ExportGlobal,
							Index: uint32(idx),
						})
						exportSet[exportName] = struct{}{}
					}
				}
			}
		case bytecode.OpExportDefault:
			if _, alreadyExported := exportSet["default"]; !alreadyExported {
				if idx, exists := globalNames["default"]; exists {
					mod.Exports = append(mod.Exports, module.Export{
						Name:  "default",
						Kind:  module.ExportGlobal,
						Index: uint32(idx),
					})
					exportSet["default"] = struct{}{}
				}
			}
		}
	}

	// Build module-level constant pool and remap per-function constant references
	buildModuleConstantPool(mod)

	return mod, nil
}

func buildGlobals(globalNames map[string]uint32) []module.Global {
	globals := make([]module.Global, len(globalNames))
	for name, idx := range globalNames {
		globals[idx] = module.Global{
			Name:    name,
			Mutable: true,
			Type:    core.ValueNull,
			Value:   nil,
		}
	}
	return globals
}

func buildModuleConstantPool(mod *module.Module) {
	constToIdx := make(map[string]int)
	var pool []any

	for fi, fn := range mod.Functions {
		remap := make([]uint32, len(fn.Constants))
		for ci, c := range fn.Constants {
			key := constantKey(c)
			if idx, ok := constToIdx[key]; ok {
				remap[ci] = uint32(idx)
			} else {
				idx := len(pool)
				pool = append(pool, c)
				constToIdx[key] = idx
				remap[ci] = uint32(idx)
			}
		}

		// Remap instructions that reference constants
		for ii, inst := range fn.Code {
			switch inst.Op {
			case core.OpConst, core.OpGetProp, core.OpSetProp, core.OpImportFunc:
				if int(inst.A) < len(remap) {
					mod.Functions[fi].Code[ii].A = remap[inst.A]
				}
			}
		}

		mod.Functions[fi].Constants = nil
	}

	mod.Constants = pool
}

func constantKey(c any) string {
	switch v := c.(type) {
	case nil:
		return "nil"
	case bool:
		return fmt.Sprintf("bool:%v", v)
	case int:
		return fmt.Sprintf("int:%d", v)
	case int64:
		return fmt.Sprintf("int64:%d", v)
	case float64:
		return fmt.Sprintf("float64:%v", v)
	case string:
		return fmt.Sprintf("str:%s", v)
	default:
		return fmt.Sprintf("type:%T val:%v", c, c)
	}
}

func lowerFunction(ft *bytecode.FunctionTemplate, globalNames map[string]uint32, funcMap map[int]int) module.Function {
	fn := module.Function{
		Name:      ft.Name,
		TypeIndex: 0,
	}

	// Build local variable map for function parameters and locals
	localMap := make(map[string]uint32)
	var nextLocalIdx uint32 = 0 // parameters start at local 0 (no implicit self for regular functions)

	// Map parameters to local indices
	for _, param := range ft.Params {
		localMap[param] = nextLocalIdx
		nextLocalIdx++
	}
	if ft.RestParam != "" {
		localMap[ft.RestParam] = nextLocalIdx
		nextLocalIdx++
	}

	// Scan function code to find local variable definitions (OpDefineName inside function)
	if ft.Chunk != nil {
		for _, inst := range ft.Chunk.Code {
			switch inst.Op {
			case bytecode.OpDefineName:
				if int(inst.A) < len(ft.Chunk.Constants) {
					if name, ok := ft.Chunk.Constants[inst.A].(string); ok {
						if _, exists := localMap[name]; !exists {
							localMap[name] = nextLocalIdx
							nextLocalIdx++
						}
					}
				}
			case bytecode.OpPushTry:
				if inst.B > 0 && int(inst.B) < len(ft.Chunk.Constants) {
					if name, ok := ft.Chunk.Constants[inst.B].(string); ok {
						if _, exists := localMap[name]; !exists {
							localMap[name] = nextLocalIdx
							nextLocalIdx++
						}
					}
				}
			}
		}
	}

	// Allocate locals
	totalLocals := int(nextLocalIdx)
	for i := 0; i < totalLocals; i++ {
		fn.Locals = append(fn.Locals, core.ValueNull)
	}

	// Lower constants from chunk
	if ft.Chunk != nil {
		constMap := make(map[int]int)
		fn.Constants, constMap = lowerConstants(ft.Chunk.Constants, ft.Chunk.Code)
		fn.Code = lowerInstructions(ft.Chunk.Code, funcMap)

		// Remap constant indices and names
		for i, inst := range fn.Code {
			switch inst.Op {
			case core.OpConst, core.OpLoadGlobal, core.OpStoreGlobal, core.OpGetProp, core.OpSetProp, core.OpImportFunc:
				if int(inst.A) < len(ft.Chunk.Constants) {
					if newIdx, ok := constMap[int(inst.A)]; ok && newIdx >= 0 {
						fn.Code[i].A = uint32(newIdx)
					}
				}
			}

			// Check if this is a local variable access
			remappedA := fn.Code[i].A
			if inst.Op == core.OpLoadGlobal || inst.Op == core.OpStoreGlobal {
				if int(remappedA) < len(fn.Constants) {
					if name, ok := fn.Constants[remappedA].(string); ok {
						if localIdx, isLocal := localMap[name]; isLocal {
							if inst.Op == core.OpLoadGlobal {
								fn.Code[i].Op = core.OpLoadLocal
							} else {
								fn.Code[i].Op = core.OpStoreLocal
							}
							fn.Code[i].A = localIdx
						} else if idx, exists := globalNames[name]; exists {
							fn.Code[i].A = idx
						} else if inst.Op == core.OpLoadGlobal {
							fn.Code[i].Op = core.OpConst
						}
					}
				}
			}

			if inst.Op == core.OpPushTry && inst.B > 0 {
				origB := int(inst.B)
				if origB < len(ft.Chunk.Constants) {
					if name, ok := ft.Chunk.Constants[origB].(string); ok {
						if localIdx, isLocal := localMap[name]; isLocal {
							fn.Code[i].B = localIdx + 1
						} else {
							fn.Code[i].B = 0
						}
					}
				}
			}
		}
	}

	return fn
}

func lowerChunkAsFunction(chunk *bytecode.Chunk, name string) module.Function {
	fn := module.Function{
		Name:      name,
		TypeIndex: 0,
	}
	fn.Constants, _ = lowerConstants(chunk.Constants, chunk.Code)
	fn.Code = lowerInstructions(chunk.Code, nil)
	return fn
}

func lowerChunkAsFunctionWithGlobals(chunk *bytecode.Chunk, name string, globalNames map[string]uint32, funcIndexMap map[int]int) module.Function {
	fn := module.Function{
		Name:      name,
		TypeIndex: 0,
	}

	entryLocalMap := make(map[string]uint32)
	var nextEntryLocal uint32
	for _, inst := range chunk.Code {
		if inst.Op == bytecode.OpPushTry && inst.B > 0 && int(inst.B) < len(chunk.Constants) {
			if catchName, ok := chunk.Constants[inst.B].(string); ok {
				if _, exists := entryLocalMap[catchName]; !exists {
					entryLocalMap[catchName] = nextEntryLocal
					nextEntryLocal++
				}
			}
		}
	}
	for i := 0; i < int(nextEntryLocal); i++ {
		fn.Locals = append(fn.Locals, core.ValueNull)
	}

	constMap := make(map[int]int)
	fn.Constants, constMap = lowerConstants(chunk.Constants, chunk.Code)
	fn.Code = lowerInstructions(chunk.Code, funcIndexMap)

	// Build reverse map: function name -> function index in iavm module
	nameToFuncIdx := make(map[string]uint32)
	for constIdx, funcIdx := range funcIndexMap {
		if constIdx < len(chunk.Constants) {
			if ft, ok := chunk.Constants[constIdx].(*bytecode.FunctionTemplate); ok && ft.Name != "" {
				nameToFuncIdx[ft.Name] = uint32(funcIdx)
			}
		}
	}

	// Remap instructions: fix constant indices and handle global names
	for i, inst := range fn.Code {
		if i < len(chunk.Code) && chunk.Code[i].Op == bytecode.OpExportDefault {
			if idx, exists := globalNames["default"]; exists {
				fn.Code[i].Op = core.OpStoreGlobal
				fn.Code[i].A = idx
			}
			continue
		}

		switch inst.Op {
		case core.OpConst, core.OpLoadGlobal, core.OpStoreGlobal, core.OpGetProp, core.OpSetProp, core.OpImportFunc:
			if int(inst.A) < len(chunk.Constants) {
				if newIdx, ok := constMap[int(inst.A)]; ok && newIdx >= 0 {
					fn.Code[i].A = uint32(newIdx)
				}
			}
		}

		// Handle global name remapping and function references
		// Use fn.Code[i].A which has been remapped by the first pass
		remappedA := fn.Code[i].A
		if inst.Op == core.OpLoadGlobal {
			if int(remappedA) < len(fn.Constants) {
				if globalName, ok := fn.Constants[remappedA].(string); ok {
					if localIdx, isLocal := entryLocalMap[globalName]; isLocal {
						fn.Code[i].Op = core.OpLoadLocal
						fn.Code[i].A = localIdx
					} else if funcIdx, isFunc := nameToFuncIdx[globalName]; isFunc {
						fn.Code[i].Op = core.OpConst
						fn.Constants = append(fn.Constants, int64(funcIdx))
						fn.Code[i].A = uint32(len(fn.Constants) - 1)
					} else if _, isGlobal := globalNames[globalName]; !isGlobal {
						fn.Code[i].Op = core.OpConst
					} else if idx, exists := globalNames[globalName]; exists {
						fn.Code[i].A = idx
					}
				}
			}
		}
		if inst.Op == core.OpStoreGlobal {
			if int(remappedA) < len(fn.Constants) {
				if globalName, ok := fn.Constants[remappedA].(string); ok {
					if localIdx, isLocal := entryLocalMap[globalName]; isLocal {
						fn.Code[i].Op = core.OpStoreLocal
						fn.Code[i].A = localIdx
					} else if idx, exists := globalNames[globalName]; exists {
						fn.Code[i].A = idx
					}
				}
			}
		}
		if inst.Op == core.OpPushTry && inst.B > 0 {
			origB := int(inst.B)
			if origB < len(chunk.Constants) {
				if catchName, ok := chunk.Constants[origB].(string); ok {
					if localIdx, isLocal := entryLocalMap[catchName]; isLocal {
						fn.Code[i].B = localIdx + 1
					} else {
						fn.Code[i].B = 0
					}
				}
			}
		}
	}

	return fn
}

func collectGlobalNames(chunk *bytecode.Chunk) map[string]uint32 {
	globalNames := make(map[string]uint32)
	var nextIdx uint32

	for _, inst := range chunk.Code {
		switch inst.Op {
		case bytecode.OpSetName, bytecode.OpDefineName:
			if int(inst.A) < len(chunk.Constants) {
				if name, ok := chunk.Constants[inst.A].(string); ok {
					if _, exists := globalNames[name]; !exists {
						globalNames[name] = nextIdx
						nextIdx++
					}
				}
			}
		case bytecode.OpExportDefault:
			if _, exists := globalNames["default"]; !exists {
				globalNames["default"] = nextIdx
				nextIdx++
			}
		}
	}

	// Also collect from function chunks, but only OpSetName (not OpDefineName)
	// OpDefineName inside functions declares locals, not globals.
	// OpSetName inside functions may reference implicit globals.
	for _, c := range chunk.Constants {
		if ft, ok := c.(*bytecode.FunctionTemplate); ok && ft.Chunk != nil {
			for _, inst := range ft.Chunk.Code {
				switch inst.Op {
				case bytecode.OpSetName:
					if int(inst.A) < len(ft.Chunk.Constants) {
						if name, ok := ft.Chunk.Constants[inst.A].(string); ok {
							if _, exists := globalNames[name]; !exists {
								globalNames[name] = nextIdx
								nextIdx++
							}
						}
					}
				}
			}
		}
	}

	return globalNames
}

func lowerConstants(constants []any, code []bytecode.Instruction) ([]any, map[int]int) {
	result := []any{}
	indexMap := make(map[int]int)

	for i, c := range constants {
		switch v := c.(type) {
		case *bytecode.FunctionTemplate:
			// Skip function templates in constants - they become separate functions
			indexMap[i] = -1
			continue
		case nil, bool, int, int64, float64, string:
			indexMap[i] = len(result)
			result = append(result, c)
		default:
			_ = v
			indexMap[i] = len(result)
			result = append(result, nil)
		}
	}
	return result, indexMap
}

func lowerInstructions(ialangInsts []bytecode.Instruction, funcMap map[int]int) []core.Instruction {
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
			iavmInst.Op = core.OpJumpIfTrue
			iavmInst.A = uint32(inst.A)

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
			iavmInst.Op = core.OpTruthy

		case bytecode.OpAnd:
			iavmInst.Op = core.OpAnd

		case bytecode.OpOr:
			iavmInst.Op = core.OpOr

		case bytecode.OpClosure:
			if funcMap != nil {
				if funcIdx, ok := funcMap[inst.A]; ok {
					iavmInst.Op = core.OpClosure
					iavmInst.A = uint32(funcIdx)
				} else {
					iavmInst.Op = core.OpNop
				}
			} else {
				iavmInst.Op = core.OpNop
			}

		case bytecode.OpClass:
			iavmInst.Op = core.OpMakeObject

		case bytecode.OpNew:
			iavmInst.Op = core.OpCall
			iavmInst.A = uint32(inst.A)
			iavmInst.B = uint32(inst.B)

		case bytecode.OpIndex:
			iavmInst.Op = core.OpIndex

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
			iavmInst.B = uint32(inst.B)

		case bytecode.OpPopTry:
			iavmInst.Op = core.OpPopTry

		case bytecode.OpThrow:
			iavmInst.Op = core.OpThrow

		case bytecode.OpObjectKeys:
			iavmInst.Op = core.OpObjectKeys

		case bytecode.OpJumpIfNullish:
			iavmInst.Op = core.OpJumpIfNullish
			iavmInst.A = uint32(inst.A)

		case bytecode.OpJumpIfNotNullish:
			iavmInst.Op = core.OpJumpIfNotNullish
			iavmInst.A = uint32(inst.A)

		case bytecode.OpImportNamespace, bytecode.OpImportDynamic,
			bytecode.OpExportName, bytecode.OpExportAs, bytecode.OpExportDefault,
			bytecode.OpExportAll, bytecode.OpSuper, bytecode.OpSuperCall,
			bytecode.OpSpreadArray, bytecode.OpSpreadObject, bytecode.OpSpreadCall,
			bytecode.OpAwait:
			// Unsupported in minimal iavm, use Nop
			iavmInst.Op = core.OpNop

		default:
			iavmInst.Op = core.OpNop
		}

		iavmInsts = append(iavmInsts, iavmInst)
	}

	return iavmInsts
}
