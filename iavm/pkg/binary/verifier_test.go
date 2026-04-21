package binary

import (
	"testing"
	"iavm/pkg/core"
	"iavm/pkg/module"
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
