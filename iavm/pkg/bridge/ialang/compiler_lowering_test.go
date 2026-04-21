package ialang

import (
	"testing"
	"iacommon/pkg/ialang/bytecode"
	"iavm/pkg/core"
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
