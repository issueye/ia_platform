package chunkcodec

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"

	bc "iacommon/pkg/ialang/bytecode"
)

func TestSerializeDeserializeRoundTrip(t *testing.T) {
	fnChunk := &bc.Chunk{
		Code: []bc.Instruction{
			{Op: bc.OpConstant, A: 0, B: 0},
			{Op: bc.OpReturn, A: 1, B: 0},
		},
		Constants: []any{float64(7)},
	}
	chunk := &bc.Chunk{
		Code: []bc.Instruction{
			{Op: bc.OpConstant, A: 0, B: 0},
			{Op: bc.OpConstant, A: 1, B: 0},
			{Op: bc.OpConstant, A: 2, B: 0},
			{Op: bc.OpConstant, A: 3, B: 0},
			{Op: bc.OpClosure, A: 4, B: 0},
			{Op: bc.OpReturn, A: 0, B: 0},
		},
		Constants: []any{
			nil,
			"hello",
			true,
			float64(42),
			&bc.FunctionTemplate{
				Name:   "f",
				Params: []string{"x"},
				Async:  false,
				Chunk:  fnChunk,
			},
		},
	}

	data, err := Serialize(chunk)
	if err != nil {
		t.Fatalf("Serialize unexpected error: %v", err)
	}

	got, err := Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize unexpected error: %v", err)
	}

	if !reflect.DeepEqual(got, chunk) {
		t.Fatalf("roundtrip mismatch\n got: %#v\nwant: %#v", got, chunk)
	}
}

func TestDeserializeVersionMismatch(t *testing.T) {
	chunk := &bc.Chunk{
		Code:      []bc.Instruction{{Op: bc.OpReturn, A: 0, B: 0}},
		Constants: []any{},
	}
	data, err := Serialize(chunk)
	if err != nil {
		t.Fatalf("Serialize unexpected error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}
	payload["version"] = float64(FormatVersion + 1)
	badData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	_, err = Deserialize(badData)
	if err == nil {
		t.Fatal("Deserialize expected version error, got nil")
	}
}

func TestSerializeNilChunk(t *testing.T) {
	_, err := Serialize(nil)
	if err == nil {
		t.Fatal("Serialize(nil) expected error, got nil")
	}
}

func TestDeserializeInvalidMagic(t *testing.T) {
	badData, _ := json.Marshal(map[string]any{
		"magic":   "WRONG",
		"version": FormatVersion,
		"chunk":   map[string]any{},
	})
	_, err := Deserialize(badData)
	if err == nil {
		t.Fatal("expected invalid magic error")
	}
}

func TestDeserializeEmptyPayload(t *testing.T) {
	badData, _ := json.Marshal(map[string]any{
		"magic":   FormatMagic,
		"version": FormatVersion,
	})
	_, err := Deserialize(badData)
	if err == nil {
		t.Fatal("expected empty payload error")
	}
}

func TestDeserializeGarbage(t *testing.T) {
	_, err := Deserialize([]byte("not json"))
	if err == nil {
		t.Fatal("expected decode error")
	}
}

func TestSerializeNaN(t *testing.T) {
	chunk := &bc.Chunk{
		Code:      []bc.Instruction{{Op: bc.OpReturn, A: 0, B: 0}},
		Constants: []any{float64(math.NaN())},
	}
	_, err := Serialize(chunk)
	if err == nil {
		t.Fatal("expected NaN error")
	}
}

func TestSerializeInf(t *testing.T) {
	chunk := &bc.Chunk{
		Code:      []bc.Instruction{{Op: bc.OpReturn, A: 0, B: 0}},
		Constants: []any{math.Inf(1)},
	}
	_, err := Serialize(chunk)
	if err == nil {
		t.Fatal("expected Inf error")
	}
}

func TestSerializeNilFunctionTemplate(t *testing.T) {
	chunk := &bc.Chunk{
		Code: []bc.Instruction{
			{Op: bc.OpConstant, A: 0, B: 0},
			{Op: bc.OpReturn, A: 0, B: 0},
		},
		Constants: []any{(*bc.FunctionTemplate)(nil)},
	}
	data, err := Serialize(chunk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := Deserialize(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Constants[0] != nil {
		t.Fatalf("expected nil constant, got %v", got.Constants[0])
	}
}

func TestSerializeFunctionTemplateNilChunk(t *testing.T) {
	chunk := &bc.Chunk{
		Code: []bc.Instruction{
			{Op: bc.OpClosure, A: 0, B: 0},
			{Op: bc.OpReturn, A: 0, B: 0},
		},
		Constants: []any{&bc.FunctionTemplate{Name: "f"}},
	}
	_, err := Serialize(chunk)
	if err == nil {
		t.Fatal("expected nil chunk error")
	}
}

func TestSerializeUnsupportedType(t *testing.T) {
	chunk := &bc.Chunk{
		Code: []bc.Instruction{
			{Op: bc.OpConstant, A: 0, B: 0},
			{Op: bc.OpReturn, A: 0, B: 0},
		},
		Constants: []any{[]int{1, 2, 3}},
	}
	_, err := Serialize(chunk)
	if err == nil {
		t.Fatal("expected unsupported type error")
	}
}

func TestDeserializeNilEncodedChunk(t *testing.T) {
	_, err := decodeChunk(nil)
	if err == nil {
		t.Fatal("expected nil chunk error")
	}
}

func TestDeserializeUnsupportedTypeTag(t *testing.T) {
	_, err := decodeConstant(serialConstant{Type: "bogus"})
	if err == nil {
		t.Fatal("expected unsupported type tag error")
	}
}

func TestDeserializeBoolEmptyPayload(t *testing.T) {
	_, err := decodeConstant(serialConstant{Type: "bool"})
	if err == nil {
		t.Fatal("expected empty bool payload error")
	}
}

func TestDeserializeFunctionEmptyPayload(t *testing.T) {
	_, err := decodeConstant(serialConstant{Type: "function"})
	if err == nil {
		t.Fatal("expected empty function payload error")
	}
}

func TestDeserializeInvalidInstructionOp(t *testing.T) {
	encoded := &serialChunk{
		Code: []serialInstruction{{Op: 255, A: 0, B: 0}},
		Constants: []serialConstant{{Type: "nil"}},
	}
	got, err := decodeChunk(encoded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Code[0].Op != bc.OpCode(255) {
		t.Fatalf("unexpected opcode: %v", got.Code[0].Op)
	}
}
