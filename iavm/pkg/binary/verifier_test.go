package binary

import (
	"iavm/pkg/core"
	"iavm/pkg/module"
	"testing"
)

func TestVerifyModule_ValidMinimal(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
	}
	result, err := VerifyModule(mod, VerifyOptions{})
	if err != nil {
		t.Fatalf("VerifyModule failed: %v", err)
	}
	if !result.Valid {
		t.Fatal("expected valid module")
	}
}

func TestVerifyModule_InvalidMagic(t *testing.T) {
	mod := &module.Module{
		Magic:   "BAD!",
		Version: 1,
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid magic")
	}
}

func TestVerifyModule_InvalidVersion(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 99,
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid version")
	}
}

func TestVerifyModule_InvalidTypeRef(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 5,
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid type reference")
	}
}

func TestVerifyModule_InvalidExportRef(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Exports: []module.Export{
			{Name: "main", Kind: module.ExportFunction, Index: 10},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid export reference")
	}
}

func TestVerifyModule_InvalidOpcode(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types: []core.FuncType{
			{},
		},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpCode(255)},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid opcode")
	}
}

func TestVerifyModule_EmptyImportModule(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Imports: []module.Import{
			{Module: "", Name: "test", Kind: module.ImportFunction},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for empty import module name")
	}
}

func TestVerifyModule_InvalidCapability(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Capabilities: []module.CapabilityDecl{
			{Kind: "invalid"},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid capability kind")
	}
}

func TestVerifyModule_RequireEntry_NoEntry(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types: []core.FuncType{
			{},
		},
		Functions: []module.Function{
			{
				Name:      "helper",
				TypeIndex: 0,
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{RequireEntry: true})
	if err == nil {
		t.Fatal("expected error for missing entry point")
	}
}

func TestVerifyModule_RequireEntry_WithMain(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types: []core.FuncType{
			{},
		},
		Functions: []module.Function{
			{
				Name:      "main",
				TypeIndex: 0,
			},
		},
	}
	result, err := VerifyModule(mod, VerifyOptions{RequireEntry: true})
	if err != nil {
		t.Fatalf("VerifyModule failed: %v", err)
	}
	if !result.Valid {
		t.Fatal("expected valid module with main function")
	}
}

func TestVerifyModule_InvalidJumpTarget(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpJump, A: 100},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid jump target")
	}
}

func TestVerifyModule_InvalidLocalIndex(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Locals:    []core.ValueKind{core.ValueNull},
				Code: []core.Instruction{
					{Op: core.OpLoadLocal, A: 5},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid local index")
	}
}

func TestVerifyModule_InvalidConstantIndex(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Constants: []any{int64(1)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 10},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid constant index")
	}
}

func TestVerifyModule_InvalidTryHandlerTarget(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpPushTry, A: 99},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for invalid try handler target")
	}
}

func TestVerifyModule_ValidControlFlow(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Locals:    []core.ValueKind{core.ValueI64},
				Constants: []any{int64(42)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpStoreLocal, A: 0},
					{Op: core.OpLoadLocal, A: 0},
					{Op: core.OpJump, A: 3},
					{Op: core.OpReturn},
				},
			},
		},
	}
	result, err := VerifyModule(mod, VerifyOptions{})
	if err != nil {
		t.Fatalf("VerifyModule failed: %v", err)
	}
	if !result.Valid {
		t.Fatalf("expected valid module, got errors: %v", result.Errors)
	}
}
<<<<<<< HEAD

func TestVerifyModule_StackUnderflow(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpAdd}, // pop 2, push 1, but stack is empty
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for stack underflow")
	}
}

func TestVerifyModule_StackOverflow(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: func() []core.Instruction {
					code := make([]core.Instruction, 1026)
					for i := range code {
						code[i] = core.Instruction{Op: core.OpConst, A: 0}
					}
					code[len(code)-1] = core.Instruction{Op: core.OpReturn}
					return code
				}(),
				Constants: []any{int64(1)},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for stack overflow")
	}
}
||||||| parent of 93cf715 (feat(iavm): 添加模块验证、资源限制和CLI命令支持)
=======

func TestVerifyModule_StackUnderflow(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpAdd},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for stack underflow")
	}
}

func TestVerifyModule_StackHeightMismatchAtJoin(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Constants: []any{false, int64(1)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpJumpIfFalse, A: 4},
					{Op: core.OpConst, A: 1},
					{Op: core.OpJump, A: 4},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for stack height mismatch")
	}
}

func TestVerifyModule_DirectCallArgumentMismatch(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types: []core.FuncType{
			{Params: []core.ValueKind{core.ValueI64, core.ValueI64}, Results: []core.ValueKind{core.ValueI64}},
			{},
		},
		Functions: []module.Function{
			{
				Name:      "add",
				TypeIndex: 0,
				Locals:    []core.ValueKind{core.ValueI64, core.ValueI64},
				Code: []core.Instruction{
					{Op: core.OpLoadLocal, A: 0},
					{Op: core.OpLoadLocal, A: 1},
					{Op: core.OpAdd},
					{Op: core.OpReturn},
				},
			},
			{
				Name:      "entry",
				TypeIndex: 1,
				Constants: []any{int64(5)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpCall, A: 0, B: 1},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for direct call argument mismatch")
	}
}

func TestVerifyModule_StackExceedsMaxStack(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Constants: []any{int64(1), int64(2)},
				MaxStack:  1,
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpReturn},
				},
			},
		},
	}
	_, err := VerifyModule(mod, VerifyOptions{})
	if err == nil {
		t.Fatal("expected error for max stack overflow")
	}
}

func TestVerifyModule_ResourceLimitsAllowValidModule(t *testing.T) {
	mod := resourceLimitTestModule()

	opts := VerifyOptions{
		MaxFunctions:           2,
		MaxConstants:           2,
		MaxCodeSizePerFunction: 4,
		MaxLocalsPerFunction:   2,
		MaxStackPerFunction:    3,
	}
	if _, err := VerifyModule(mod, opts); err != nil {
		t.Fatalf("expected resource limits to allow module, got %v", err)
	}
}

func TestVerifyModule_ResourceLimitFailures(t *testing.T) {
	cases := []struct {
		name string
		opts VerifyOptions
	}{
		{name: "max functions", opts: VerifyOptions{MaxFunctions: 1}},
		{name: "max constants", opts: VerifyOptions{MaxConstants: 1}},
		{name: "max code size", opts: VerifyOptions{MaxCodeSizePerFunction: 3}},
		{name: "max locals", opts: VerifyOptions{MaxLocalsPerFunction: 1}},
		{name: "max stack declaration", opts: VerifyOptions{MaxStackPerFunction: 2}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mod := resourceLimitTestModule()
			if tc.name == "max functions" {
				tc.opts.MaxFunctions = 0
				if _, err := VerifyModule(mod, tc.opts); err != nil {
					t.Fatalf("zero max functions should mean unlimited, got %v", err)
				}
				tc.opts.MaxFunctions = len(mod.Functions) - 1
			}
			if _, err := VerifyModule(mod, tc.opts); err == nil {
				t.Fatalf("expected %s limit error", tc.name)
			}
		})
	}
}

func TestVerifyModule_PerFunctionConstantLimit(t *testing.T) {
	mod := resourceLimitTestModule()
	mod.Constants = nil
	mod.Functions[0].Constants = []any{int64(1), int64(2)}

	if _, err := VerifyModule(mod, VerifyOptions{MaxConstants: 1}); err == nil {
		t.Fatal("expected per-function constant limit error")
	}
}

func TestVerifyModule_CapabilityAllowlist(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Capabilities: []module.CapabilityDecl{
			{Kind: module.CapabilityFS},
		},
	}

	if _, err := VerifyModule(mod, VerifyOptions{AllowedCapabilities: []module.CapabilityKind{module.CapabilityFS}}); err != nil {
		t.Fatalf("expected fs capability to be allowed, got %v", err)
	}
	if _, err := VerifyModule(mod, VerifyOptions{AllowedCapabilities: []module.CapabilityKind{module.CapabilityNetwork}}); err == nil {
		t.Fatal("expected fs capability to be denied by allowlist")
	}
}

func resourceLimitTestModule() *module.Module {
	return &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{int64(1), int64(2)},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Locals:    []core.ValueKind{core.ValueI64, core.ValueI64},
				MaxStack:  3,
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpAdd},
					{Op: core.OpReturn},
				},
			},
			{
				Name:      "helper",
				TypeIndex: 0,
				Code:      []core.Instruction{{Op: core.OpReturn}},
			},
		},
	}
}
>>>>>>> 93cf715 (feat(iavm): 添加模块验证、资源限制和CLI命令支持)
