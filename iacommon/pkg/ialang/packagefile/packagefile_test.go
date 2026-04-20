package packagefile

import (
	"encoding/json"
	"reflect"
	"testing"

	bc "iacommon/pkg/ialang/bytecode"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	pkg := &Package{
		Entry: "/entry.ia",
		Modules: map[string]*bc.Chunk{
			"/entry.ia": {
				Code: []bc.Instruction{
					{Op: bc.OpImportName, A: 0, B: 1},
					{Op: bc.OpReturn, A: 0, B: 0},
				},
				Constants: []any{"./helper", "answer"},
			},
			"/helper.ia": {
				Code: []bc.Instruction{
					{Op: bc.OpConstant, A: 0, B: 0},
					{Op: bc.OpDefineName, A: 1, B: 0},
					{Op: bc.OpExportName, A: 1, B: 0},
					{Op: bc.OpReturn, A: 0, B: 0},
				},
				Constants: []any{float64(42), "answer"},
			},
		},
		Imports: map[string]map[string]string{
			"/entry.ia": {
				"./helper": "/helper.ia",
			},
		},
	}

	data, err := Encode(pkg)
	if err != nil {
		t.Fatalf("Encode unexpected error: %v", err)
	}

	got, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, pkg) {
		t.Fatalf("roundtrip mismatch\n got: %#v\nwant: %#v", got, pkg)
	}
}

func TestDecodeVersionMismatch(t *testing.T) {
	pkg := &Package{
		Entry: "/entry.ia",
		Modules: map[string]*bc.Chunk{
			"/entry.ia": {
				Code:      []bc.Instruction{{Op: bc.OpReturn, A: 0, B: 0}},
				Constants: []any{},
			},
		},
	}
	data, err := Encode(pkg)
	if err != nil {
		t.Fatalf("Encode unexpected error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	payload["version"] = float64(PackageFormatVersion + 1)
	badData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	_, err = Decode(badData)
	if err == nil {
		t.Fatal("Decode expected version error, got nil")
	}
}

func TestResolveImport(t *testing.T) {
	pkg := &Package{
		Entry: "/entry.ia",
		Modules: map[string]*bc.Chunk{
			"/entry.ia":  {Code: []bc.Instruction{}, Constants: []any{}},
			"/helper.ia": {Code: []bc.Instruction{}, Constants: []any{}},
		},
		Imports: map[string]map[string]string{
			"/entry.ia": {"./helper": "/helper.ia"},
		},
	}

	target, ok := pkg.ResolveImport("/entry.ia", "./helper")
	if !ok {
		t.Fatal("ResolveImport expected success, got false")
	}
	if target != "/helper.ia" {
		t.Fatalf("ResolveImport target = %q, want %q", target, "/helper.ia")
	}

	if _, ok := pkg.ResolveImport("/entry.ia", "./missing"); ok {
		t.Fatal("ResolveImport missing import expected false")
	}
	if _, ok := pkg.ResolveImport("/missing.ia", "./helper"); ok {
		t.Fatal("ResolveImport missing module expected false")
	}
}
