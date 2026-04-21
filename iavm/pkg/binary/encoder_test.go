package binary

import (
	"testing"
	"iavm/pkg/core"
	"iavm/pkg/module"
)

func TestEncodeModule_Minimal(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
	}
	data, err := EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("encoded data is empty")
	}
	// Check magic header
	if string(data[:4]) != "IAVM" {
		t.Fatalf("expected magic 'IAVM', got %q", data[:4])
	}
}

func TestEncodeModule_WithFunctions(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types: []core.FuncType{
			{Params: []core.ValueKind{core.ValueI64}, Results: []core.ValueKind{core.ValueI64}},
		},
		Functions: []module.Function{
			{
				Name:      "add_one",
				TypeIndex: 0,
				Locals:    []core.ValueKind{core.ValueI64},
				Code: []core.Instruction{
					{Op: core.OpLoadLocal, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpAdd},
					{Op: core.OpReturn},
				},
			},
		},
	}
	data, err := EncodeModule(mod)
	if err != nil {
		t.Fatalf("EncodeModule failed: %v", err)
	}
	if len(data) < 10 {
		t.Fatalf("encoded data too short: %d bytes", len(data))
	}
}

func TestEncodeModule_NilModule(t *testing.T) {
	_, err := EncodeModule(nil)
	if err == nil {
		t.Fatal("expected error for nil module")
	}
}

func TestEncodeModule_InvalidMagic(t *testing.T) {
	mod := &module.Module{
		Magic:   "BAD!",
		Version: 1,
	}
	_, err := EncodeModule(mod)
	if err == nil {
		t.Fatal("expected error for invalid magic")
	}
}
