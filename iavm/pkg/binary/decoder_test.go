package binary

import (
	"testing"
	"iavm/pkg/core"
	"iavm/pkg/module"
)

func TestDecodeModule_RoundTrip(t *testing.T) {
	original := &module.Module{
		Magic:      "IAVM",
		Version:    1,
		Target:     "ialang",
		ABIVersion: 1,
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
		Exports: []module.Export{
			{Name: "add_one", Kind: module.ExportFunction, Index: 0},
		},
	}

	data, err := EncodeModule(original)
	if err != nil {
		t.Fatalf("EncodeModule failed: %v", err)
	}

	decoded, err := DecodeModule(data)
	if err != nil {
		t.Fatalf("DecodeModule failed: %v", err)
	}

	if decoded.Magic != original.Magic {
		t.Errorf("magic mismatch: got %q, want %q", decoded.Magic, original.Magic)
	}
	if decoded.Version != original.Version {
		t.Errorf("version mismatch: got %d, want %d", decoded.Version, original.Version)
	}
	if len(decoded.Functions) != len(original.Functions) {
		t.Errorf("function count mismatch: got %d, want %d", len(decoded.Functions), len(original.Functions))
	}
	if decoded.Functions[0].Name != original.Functions[0].Name {
		t.Errorf("function name mismatch: got %q, want %q", decoded.Functions[0].Name, original.Functions[0].Name)
	}
	if len(decoded.Functions[0].Code) != len(original.Functions[0].Code) {
		t.Errorf("instruction count mismatch: got %d, want %d", len(decoded.Functions[0].Code), len(original.Functions[0].Code))
	}
}

func TestDecodeModule_InvalidMagic(t *testing.T) {
	data := []byte("BADX\x01\x00\x00\x00")
	_, err := DecodeModule(data)
	if err == nil {
		t.Fatal("expected error for invalid magic")
	}
}

func TestDecodeModule_TooShort(t *testing.T) {
	data := []byte("IA")
	_, err := DecodeModule(data)
	if err == nil {
		t.Fatal("expected error for too short data")
	}
}

func TestDecodeModule_WithConstants(t *testing.T) {
	original := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Functions: []module.Function{
			{
				Name:      "test",
				TypeIndex: 0,
				Constants: []any{int64(42), "hello", true, nil},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpReturn},
				},
			},
		},
	}

	data, err := EncodeModule(original)
	if err != nil {
		t.Fatalf("EncodeModule failed: %v", err)
	}

	decoded, err := DecodeModule(data)
	if err != nil {
		t.Fatalf("DecodeModule failed: %v", err)
	}

	if len(decoded.Functions[0].Constants) != 4 {
		t.Fatalf("expected 4 constants, got %d", len(decoded.Functions[0].Constants))
	}
	if decoded.Functions[0].Constants[0].(int64) != 42 {
		t.Errorf("expected const[0]=42, got %v", decoded.Functions[0].Constants[0])
	}
	if decoded.Functions[0].Constants[1].(string) != "hello" {
		t.Errorf("expected const[1]=hello, got %v", decoded.Functions[0].Constants[1])
	}
	if decoded.Functions[0].Constants[2].(bool) != true {
		t.Errorf("expected const[2]=true, got %v", decoded.Functions[0].Constants[2])
	}
}
