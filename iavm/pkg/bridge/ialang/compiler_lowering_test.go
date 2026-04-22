package ialang

import (
	"iacommon/pkg/ialang/bytecode"
	"iavm/pkg/core"
	"iavm/pkg/module"
	"testing"
)

func TestLowerToModule_MinimalChunk(t *testing.T) {
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0, B: 0},
			{Op: bytecode.OpReturn},
		},
		Constants: []any{int64(42)},
	}

	mod, err := LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	if mod == nil {
		t.Fatal("LowerToModule returned nil")
	}
	if mod.Magic != "IAVM" {
		t.Errorf("expected magic 'IAVM', got %q", mod.Magic)
	}
	if len(mod.Functions) == 0 {
		t.Fatal("expected at least one function")
	}
}

func TestLowerToModule_ChunkWithFunction(t *testing.T) {
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0, B: 0},
			{Op: bytecode.OpCall, A: 0, B: 0},
			{Op: bytecode.OpReturn},
		},
		Constants: []any{
			&bytecode.FunctionTemplate{
				Name:   "main",
				Params: []string{},
				Chunk: &bytecode.Chunk{
					Code: []bytecode.Instruction{
						{Op: bytecode.OpConstant, A: 0, B: 0},
						{Op: bytecode.OpReturn},
					},
					Constants: []any{int64(1)},
				},
			},
		},
	}

	mod, err := LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	// Should have 2 functions: main + entry
	if len(mod.Functions) != 2 {
		t.Errorf("expected 2 functions, got %d", len(mod.Functions))
	}

	// Check that main is exported
	found := false
	for _, exp := range mod.Exports {
		if exp.Name == "main" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'main' to be exported")
	}
}

func TestLowerToModule_ExportedGlobalPopulatesGlobals(t *testing.T) {
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0},
			{Op: bytecode.OpDefineName, A: 1},
			{Op: bytecode.OpExportName, A: 1},
			{Op: bytecode.OpReturn},
		},
		Constants: []any{int64(42), "value"},
	}

	mod, err := LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	if len(mod.Globals) != 1 {
		t.Fatalf("expected 1 global, got %d", len(mod.Globals))
	}
	if mod.Globals[0].Name != "value" {
		t.Fatalf("expected global name value, got %q", mod.Globals[0].Name)
	}
	if len(mod.Exports) != 1 {
		t.Fatalf("expected 1 export, got %d", len(mod.Exports))
	}
	if mod.Exports[0].Name != "value" || mod.Exports[0].Kind != module.ExportGlobal || mod.Exports[0].Index != 0 {
		t.Fatalf("unexpected export: %+v", mod.Exports[0])
	}
}

func TestLowerToModule_ExportAsGlobalAlias(t *testing.T) {
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0},
			{Op: bytecode.OpDefineName, A: 1},
			{Op: bytecode.OpExportAs, A: 1, B: 2},
			{Op: bytecode.OpReturn},
		},
		Constants: []any{int64(42), "value", "answer"},
	}

	mod, err := LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	if len(mod.Exports) != 1 {
		t.Fatalf("expected 1 export, got %d", len(mod.Exports))
	}
	if mod.Exports[0].Name != "answer" {
		t.Fatalf("expected export alias answer, got %q", mod.Exports[0].Name)
	}
	if mod.Exports[0].Kind != module.ExportGlobal || mod.Exports[0].Index != 0 {
		t.Fatalf("unexpected export: %+v", mod.Exports[0])
	}
}

func TestLowerToModule_ExportDefaultExpression(t *testing.T) {
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0},
			{Op: bytecode.OpExportDefault},
			{Op: bytecode.OpReturn},
		},
		Constants: []any{int64(42)},
	}

	mod, err := LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	if len(mod.Exports) != 1 {
		t.Fatalf("expected 1 export, got %d", len(mod.Exports))
	}
	if mod.Exports[0].Name != "default" || mod.Exports[0].Kind != module.ExportGlobal {
		t.Fatalf("unexpected default export: %+v", mod.Exports[0])
	}
	defaultIdx := mod.Exports[0].Index
	if int(defaultIdx) >= len(mod.Globals) || mod.Globals[defaultIdx].Name != "default" {
		t.Fatalf("default export points at invalid global: idx=%d globals=%+v", defaultIdx, mod.Globals)
	}

	entryFn := mod.Functions[len(mod.Functions)-1]
	if entryFn.Code[1].Op != core.OpStoreGlobal || entryFn.Code[1].A != defaultIdx {
		t.Fatalf("expected OpExportDefault to lower to StoreGlobal default, got %+v", entryFn.Code[1])
	}
}

func TestLowerToModule_NilInput(t *testing.T) {
	_, err := LowerToModule(nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestLowerToModule_InvalidInput(t *testing.T) {
	_, err := LowerToModule("not a chunk")
	if err == nil {
		t.Fatal("expected error for invalid input type")
	}
}

func TestLowerToModule_OpcodeMapping(t *testing.T) {
	chunk := &bytecode.Chunk{
		Code: []bytecode.Instruction{
			{Op: bytecode.OpConstant, A: 0, B: 0},
			{Op: bytecode.OpConstant, A: 1, B: 0},
			{Op: bytecode.OpAdd},
			{Op: bytecode.OpSub},
			{Op: bytecode.OpMul},
			{Op: bytecode.OpDiv},
			{Op: bytecode.OpEqual},
			{Op: bytecode.OpGreater},
			{Op: bytecode.OpLess},
			{Op: bytecode.OpReturn},
		},
		Constants: []any{int64(10), int64(20)},
	}

	mod, err := LowerToModule(chunk)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}

	// Entry function should have 10 instructions
	entryFn := mod.Functions[len(mod.Functions)-1]
	if len(entryFn.Code) != 10 {
		t.Errorf("expected 10 instructions, got %d", len(entryFn.Code))
	}

	// Check opcode mapping
	expected := []core.OpCode{core.OpConst, core.OpConst, core.OpAdd, core.OpSub, core.OpMul, core.OpDiv, core.OpEq, core.OpGt, core.OpLt, core.OpReturn}
	for i, inst := range entryFn.Code {
		if inst.Op != expected[i] {
			t.Errorf("instruction[%d]: expected %v, got %v", i, expected[i], inst.Op)
		}
	}
}

func TestLowerToModule_NewOpcodeMapping(t *testing.T) {
	tests := []struct {
		name     string
		ialang   bytecode.OpCode
		expected core.OpCode
	}{
		{"Dup", bytecode.OpDup, core.OpDup},
		{"Pop", bytecode.OpPop, core.OpPop},
		{"BitAnd", bytecode.OpBitAnd, core.OpBitAnd},
		{"BitOr", bytecode.OpBitOr, core.OpBitOr},
		{"BitXor", bytecode.OpBitXor, core.OpBitXor},
		{"Shl", bytecode.OpShl, core.OpShl},
		{"Shr", bytecode.OpShr, core.OpShr},
		{"And", bytecode.OpAnd, core.OpAnd},
		{"Or", bytecode.OpOr, core.OpOr},
		{"Typeof", bytecode.OpTypeof, core.OpTypeof},
		{"PushTry", bytecode.OpPushTry, core.OpPushTry},
		{"PopTry", bytecode.OpPopTry, core.OpPopTry},
		{"Throw", bytecode.OpThrow, core.OpThrow},
		{"Index", bytecode.OpIndex, core.OpIndex},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := &bytecode.Chunk{
				Code: []bytecode.Instruction{
					{Op: bytecode.OpConstant, A: 0, B: 0},
					{Op: bytecode.OpConstant, A: 0, B: 0},
					{Op: tt.ialang},
					{Op: bytecode.OpReturn},
				},
				Constants: []any{int64(42)},
			}

			mod, err := LowerToModule(chunk)
			if err != nil {
				t.Fatalf("LowerToModule failed: %v", err)
			}

			entryFn := mod.Functions[len(mod.Functions)-1]
			if len(entryFn.Code) < 3 {
				t.Fatalf("expected at least 3 instructions, got %d", len(entryFn.Code))
			}

			inst := entryFn.Code[2]
			if inst.Op != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, inst.Op)
			}
		})
	}
}
