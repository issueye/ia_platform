package chunkcodec

import (
	"encoding/json"
	"fmt"
	"math"

	bc "ialang/pkg/lang/bytecode"
	rttypes "ialang/pkg/lang/runtime/types"
)

const (
	FormatMagic   = "IALANG_CHUNK"
	FormatVersion = 1
)

type envelope struct {
	Magic   string       `json:"magic"`
	Version int          `json:"version"`
	Chunk   *serialChunk `json:"chunk"`
}

type serialChunk struct {
	Code      []serialInstruction `json:"code"`
	Constants []serialConstant    `json:"constants"`
}

type serialInstruction struct {
	Op uint8 `json:"op"`
	A  int   `json:"a"`
	B  int   `json:"b"`
}

type serialConstant struct {
	Type string          `json:"type"`
	Str  string          `json:"str,omitempty"`
	Num  float64         `json:"num,omitempty"`
	Bool *bool           `json:"bool,omitempty"`
	Fn   *serialFunction `json:"fn,omitempty"`
}

type serialFunction struct {
	Name      string       `json:"name"`
	Params    []string     `json:"params"`
	RestParam string       `json:"rest_param,omitempty"`
	Async     bool         `json:"async"`
	Chunk     *serialChunk `json:"chunk"`
}

func Serialize(chunk *bc.Chunk) ([]byte, error) {
	if chunk == nil {
		return nil, fmt.Errorf("chunk is nil")
	}
	encodedChunk, err := encodeChunk(chunk)
	if err != nil {
		return nil, err
	}

	out := envelope{
		Magic:   FormatMagic,
		Version: FormatVersion,
		Chunk:   encodedChunk,
	}
	data, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("encode chunk envelope error: %w", err)
	}
	return data, nil
}

func Deserialize(data []byte) (*bc.Chunk, error) {
	var in envelope
	if err := json.Unmarshal(data, &in); err != nil {
		return nil, fmt.Errorf("decode chunk envelope error: %w", err)
	}
	if in.Magic != FormatMagic {
		return nil, fmt.Errorf("invalid chunk magic: %q", in.Magic)
	}
	if in.Version != FormatVersion {
		return nil, fmt.Errorf("unsupported chunk format version: got %d, want %d", in.Version, FormatVersion)
	}
	if in.Chunk == nil {
		return nil, fmt.Errorf("chunk payload is empty")
	}
	return decodeChunk(in.Chunk)
}

func encodeChunk(chunk *bc.Chunk) (*serialChunk, error) {
	out := &serialChunk{
		Code:      make([]serialInstruction, len(chunk.Code)),
		Constants: make([]serialConstant, len(chunk.Constants)),
	}
	for i, ins := range chunk.Code {
		out.Code[i] = serialInstruction{
			Op: uint8(ins.Op),
			A:  ins.A,
			B:  ins.B,
		}
	}
	for i, c := range chunk.Constants {
		encoded, err := encodeConstant(c)
		if err != nil {
			return nil, fmt.Errorf("encode constant %d error: %w", i, err)
		}
		out.Constants[i] = encoded
	}
	return out, nil
}

func decodeChunk(encoded *serialChunk) (*bc.Chunk, error) {
	if encoded == nil {
		return nil, fmt.Errorf("encoded chunk is nil")
	}

	out := &bc.Chunk{
		Code:      make([]bc.Instruction, len(encoded.Code)),
		Constants: make([]any, len(encoded.Constants)),
	}
	for i, ins := range encoded.Code {
		out.Code[i] = bc.Instruction{
			Op: bc.OpCode(ins.Op),
			A:  ins.A,
			B:  ins.B,
		}
	}
	for i, c := range encoded.Constants {
		decoded, err := decodeConstant(c)
		if err != nil {
			return nil, fmt.Errorf("decode constant %d error: %w", i, err)
		}
		out.Constants[i] = decoded
	}
	return out, nil
}

func encodeConstant(v any) (serialConstant, error) {
	switch typed := v.(type) {
	case nil:
		return serialConstant{Type: "nil"}, nil
	case string:
		return serialConstant{Type: "string", Str: typed}, nil
	case bool:
		b := typed
		return serialConstant{Type: "bool", Bool: &b}, nil
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return serialConstant{}, fmt.Errorf("unsupported number constant: %v", typed)
		}
		return serialConstant{Type: "number", Num: typed}, nil
	case *bc.FunctionTemplate:
		if typed == nil {
			return serialConstant{Type: "nil"}, nil
		}
		if typed.Chunk == nil {
			return serialConstant{}, fmt.Errorf("function chunk is invalid: <nil>")
		}
		encodedChunk, err := encodeChunk(typed.Chunk)
		if err != nil {
			return serialConstant{}, fmt.Errorf("encode function chunk error: %w", err)
		}
		return serialConstant{
			Type: "function",
			Fn: &serialFunction{
				Name:      typed.Name,
				Params:    append([]string(nil), typed.Params...),
				RestParam: typed.RestParam,
				Async:     typed.Async,
				Chunk:     encodedChunk,
			},
		}, nil
	case *rttypes.UserFunction:
		if typed == nil {
			return serialConstant{Type: "nil"}, nil
		}
		if typed.Env != nil {
			return serialConstant{}, fmt.Errorf("cannot serialize function constant with bound env")
		}
		fnChunk, ok := typed.Chunk.(*bc.Chunk)
		if !ok || fnChunk == nil {
			return serialConstant{}, fmt.Errorf("function chunk is invalid: %T", typed.Chunk)
		}
		encodedChunk, err := encodeChunk(fnChunk)
		if err != nil {
			return serialConstant{}, fmt.Errorf("encode function chunk error: %w", err)
		}
		return serialConstant{
			Type: "function",
			Fn: &serialFunction{
				Name:      typed.Name,
				Params:    append([]string(nil), typed.Params...),
				RestParam: typed.RestParam,
				Async:     typed.Async,
				Chunk:     encodedChunk,
			},
		}, nil
	default:
		return serialConstant{}, fmt.Errorf("unsupported constant type: %T", v)
	}
}

func decodeConstant(encoded serialConstant) (any, error) {
	switch encoded.Type {
	case "nil":
		return nil, nil
	case "string":
		return encoded.Str, nil
	case "bool":
		if encoded.Bool == nil {
			return nil, fmt.Errorf("bool payload is empty")
		}
		return *encoded.Bool, nil
	case "number":
		return encoded.Num, nil
	case "function":
		if encoded.Fn == nil {
			return nil, fmt.Errorf("function payload is empty")
		}
		fnChunk, err := decodeChunk(encoded.Fn.Chunk)
		if err != nil {
			return nil, fmt.Errorf("decode function chunk error: %w", err)
		}
		return &bc.FunctionTemplate{
			Name:      encoded.Fn.Name,
			Params:    append([]string(nil), encoded.Fn.Params...),
			RestParam: encoded.Fn.RestParam,
			Async:     encoded.Fn.Async,
			Chunk:     fnChunk,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported constant type tag: %q", encoded.Type)
	}
}
