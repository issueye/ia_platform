package runtime

import (
	"bytes"
	"io"
	"os"
	"testing"

	"iavm/pkg/core"
	"iavm/pkg/module"
)

func newTestVM(fns []module.Function, opts Options) (*VM, error) {
	mod := &module.Module{
		Functions: fns,
	}
	return New(mod, opts)
}

func TestBuiltinPrint(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	vm, err := newTestVM([]module.Function{{
		Name:      "main",
		Constants: []any{"print", "hello", 42},
		Code: []core.Instruction{
			{Op: core.OpConst, A: 1},          // push "hello"
			{Op: core.OpConst, A: 0},          // push "print"
			{Op: core.OpCall, A: 1, B: 0},     // call print("hello")
			{Op:  core.OpReturn},
		},
	}}, Options{})
	if err != nil {
		t.Fatalf("failed to create VM: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("VM run failed: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if output != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", output)
	}
}

func TestBuiltinPrintMultipleArgs(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	vm, err := newTestVM([]module.Function{{
		Name:      "main",
		Constants: []any{"print", "hello", int64(42)},
		Code: []core.Instruction{
			{Op: core.OpConst, A: 2},          // push 42
			{Op: core.OpConst, A: 1},          // push "hello"
			{Op: core.OpConst, A: 0},          // push "print"
			{Op: core.OpCall, A: 2, B: 0},     // call print("hello", 42)
			{Op:  core.OpReturn},
		},
	}}, Options{})
	if err != nil {
		t.Fatalf("failed to create VM: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("VM run failed: %v", err)
	}

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expected := "42 hello\n"
	if output != expected {
		t.Errorf("expected %q, got %q", expected, output)
	}
}

func TestBuiltinLen(t *testing.T) {
	tests := []struct {
		name     string
		consts   []any
		code     []core.Instruction
		expected int64
	}{
		{
			name:   "string length",
			consts: []any{"len", "hello"},
			code: []core.Instruction{
				{Op: core.OpConst, A: 1},          // push "hello"
				{Op: core.OpConst, A: 0},          // push "len"
				{Op: core.OpCall, A: 1, B: 0},     // call len("hello")
				{Op:  core.OpReturn},
			},
			expected: 5,
		},
		{
			name:   "empty string",
			consts: []any{"len", ""},
			code: []core.Instruction{
				{Op: core.OpConst, A: 1},
				{Op: core.OpConst, A: 0},
				{Op: core.OpCall, A: 1, B: 0},
				{Op:  core.OpReturn},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm, err := newTestVM([]module.Function{{
				Name:      "main",
				Constants: tt.consts,
				Code:      tt.code,
			}}, Options{})
			if err != nil {
				t.Fatalf("failed to create VM: %v", err)
			}

			err = vm.Run()
			if err != nil {
				t.Fatalf("VM run failed: %v", err)
			}

			if vm.stack.Size() != 1 {
				t.Fatalf("expected 1 value on stack, got %d", vm.stack.Size())
			}

			result, ok := vm.PopResult()
			if !ok {
				t.Fatal("failed to pop result")
			}

			if result.Kind != core.ValueI64 {
				t.Fatalf("expected i64 result, got %v", result.Kind)
			}

			if result.Raw.(int64) != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result.Raw.(int64))
			}
		})
	}
}

func TestBuiltinLenArray(t *testing.T) {
	vm, err := newTestVM([]module.Function{{
		Name:      "main",
		Constants: []any{"len"},
		Code: []core.Instruction{
			{Op: core.OpConst, A: 0},            // push 1
			{Op: core.OpConst, A: 0},            // push 2
			{Op: core.OpConst, A: 0},            // push 3
			{Op: core.OpMakeArray, A: 3},        // make array [1, 2, 3]
			{Op: core.OpConst, A: 0},            // push "len"
			{Op: core.OpCall, A: 1, B: 0},       // call len([1,2,3])
			{Op:  core.OpReturn},
		},
	}}, Options{})
	if err != nil {
		t.Fatalf("failed to create VM: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("VM run failed: %v", err)
	}

	if vm.stack.Size() != 1 {
		t.Fatalf("expected 1 value on stack, got %d", vm.stack.Size())
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("failed to pop result")
	}

	if result.Kind != core.ValueI64 {
		t.Fatalf("expected i64 result, got %v", result.Kind)
	}

	if result.Raw.(int64) != 3 {
		t.Errorf("expected 3, got %d", result.Raw.(int64))
	}
}

func TestBuiltinTypeof(t *testing.T) {
	tests := []struct {
		name     string
		consts   []any
		setup    []core.Instruction
		expected string
	}{
		{
			name:     "typeof string",
			consts:   []any{"typeof", "hello"},
			setup:    []core.Instruction{{Op: core.OpConst, A: 1}},
			expected: "string",
		},
		{
			name:     "typeof int",
			consts:   []any{"typeof", int64(42)},
			setup:    []core.Instruction{{Op: core.OpConst, A: 1}},
			expected: "number",
		},
		{
			name:     "typeof float",
			consts:   []any{"typeof", 3.14},
			setup:    []core.Instruction{{Op: core.OpConst, A: 1}},
			expected: "number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := append(tt.setup, core.Instruction{Op: core.OpConst, A: 0})
			code = append(code, core.Instruction{Op: core.OpCall, A: 1, B: 0})
			code = append(code, core.Instruction{Op: core.OpReturn})

			vm, err := newTestVM([]module.Function{{
				Name:      "main",
				Constants: tt.consts,
				Code:      code,
			}}, Options{})
			if err != nil {
				t.Fatalf("failed to create VM: %v", err)
			}

			err = vm.Run()
			if err != nil {
				t.Fatalf("VM run failed: %v", err)
			}

			if vm.stack.Size() != 1 {
				t.Fatalf("expected 1 value on stack, got %d", vm.stack.Size())
			}

			result, ok := vm.PopResult()
			if !ok {
				t.Fatal("failed to pop result")
			}

			if result.Kind != core.ValueString {
				t.Fatalf("expected string result, got %v", result.Kind)
			}

			if result.Raw.(string) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Raw.(string))
			}
		})
	}
}

func TestBuiltinStr(t *testing.T) {
	tests := []struct {
		name     string
		consts   []any
		setup    []core.Instruction
		expected string
	}{
		{
			name:     "str from int",
			consts:   []any{"str", int64(42)},
			setup:    []core.Instruction{{Op: core.OpConst, A: 1}},
			expected: "42",
		},
		{
			name:     "str from float",
			consts:   []any{"str", 3.14},
			setup:    []core.Instruction{{Op: core.OpConst, A: 1}},
			expected: "3.14",
		},
		{
			name:     "str from bool true",
			consts:   []any{"str", true},
			setup:    []core.Instruction{{Op: core.OpConst, A: 1}},
			expected: "true",
		},
		{
			name:     "str from null",
			consts:   []any{"str", nil},
			setup:    []core.Instruction{{Op: core.OpConst, A: 1}},
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := append(tt.setup, core.Instruction{Op: core.OpConst, A: 0})
			code = append(code, core.Instruction{Op: core.OpCall, A: 1, B: 0})
			code = append(code, core.Instruction{Op: core.OpReturn})

			vm, err := newTestVM([]module.Function{{
				Name:      "main",
				Constants: tt.consts,
				Code:      code,
			}}, Options{})
			if err != nil {
				t.Fatalf("failed to create VM: %v", err)
			}

			err = vm.Run()
			if err != nil {
				t.Fatalf("VM run failed: %v", err)
			}

			if vm.stack.Size() != 1 {
				t.Fatalf("expected 1 value on stack, got %d", vm.stack.Size())
			}

			result, ok := vm.PopResult()
			if !ok {
				t.Fatal("failed to pop result")
			}

			if result.Kind != core.ValueString {
				t.Fatalf("expected string result, got %v", result.Kind)
			}

			if result.Raw.(string) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result.Raw.(string))
			}
		})
	}
}

func TestBuiltinInt(t *testing.T) {
	tests := []struct {
		name     string
		consts   []any
		setup    []core.Instruction
		expected int64
		wantKind core.ValueKind
	}{
		{
			name:     "int from float",
			consts:   []any{"int", 3.9},
			setup:    []core.Instruction{{Op: core.OpConst, A: 1}},
			expected: 3,
			wantKind: core.ValueI64,
		},
		{
			name:     "int from string",
			consts:   []any{"int", "123"},
			setup:    []core.Instruction{{Op: core.OpConst, A: 1}},
			expected: 123,
			wantKind: core.ValueI64,
		},
		{
			name:     "int from int (identity)",
			consts:   []any{"int", int64(42)},
			setup:    []core.Instruction{{Op: core.OpConst, A: 1}},
			expected: 42,
			wantKind: core.ValueI64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := append(tt.setup, core.Instruction{Op: core.OpConst, A: 0})
			code = append(code, core.Instruction{Op: core.OpCall, A: 1, B: 0})
			code = append(code, core.Instruction{Op: core.OpReturn})

			vm, err := newTestVM([]module.Function{{
				Name:      "main",
				Constants: tt.consts,
				Code:      code,
			}}, Options{})
			if err != nil {
				t.Fatalf("failed to create VM: %v", err)
			}

			err = vm.Run()
			if err != nil {
				t.Fatalf("VM run failed: %v", err)
			}

			if vm.stack.Size() != 1 {
				t.Fatalf("expected 1 value on stack, got %d", vm.stack.Size())
			}

			result, ok := vm.PopResult()
			if !ok {
				t.Fatal("failed to pop result")
			}

			if result.Kind != tt.wantKind {
				t.Fatalf("expected %v result, got %v", tt.wantKind, result.Kind)
			}

			if result.Raw.(int64) != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result.Raw.(int64))
			}
		})
	}
}

func TestBuiltinFloat(t *testing.T) {
	tests := []struct {
		name     string
		consts   []any
		setup    []core.Instruction
		expected float64
		wantKind core.ValueKind
	}{
		{
			name:     "float from int",
			consts:   []any{"float", int64(42)},
			setup:    []core.Instruction{{Op: core.OpConst, A: 1}},
			expected: 42.0,
			wantKind: core.ValueF64,
		},
		{
			name:     "float from string",
			consts:   []any{"float", "3.14"},
			setup:    []core.Instruction{{Op: core.OpConst, A: 1}},
			expected: 3.14,
			wantKind: core.ValueF64,
		},
		{
			name:     "float from float (identity)",
			consts:   []any{"float", 2.71},
			setup:    []core.Instruction{{Op: core.OpConst, A: 1}},
			expected: 2.71,
			wantKind: core.ValueF64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := append(tt.setup, core.Instruction{Op: core.OpConst, A: 0})
			code = append(code, core.Instruction{Op: core.OpCall, A: 1, B: 0})
			code = append(code, core.Instruction{Op: core.OpReturn})

			vm, err := newTestVM([]module.Function{{
				Name:      "main",
				Constants: tt.consts,
				Code:      code,
			}}, Options{})
			if err != nil {
				t.Fatalf("failed to create VM: %v", err)
			}

			err = vm.Run()
			if err != nil {
				t.Fatalf("VM run failed: %v", err)
			}

			if vm.stack.Size() != 1 {
				t.Fatalf("expected 1 value on stack, got %d", vm.stack.Size())
			}

			result, ok := vm.PopResult()
			if !ok {
				t.Fatal("failed to pop result")
			}

			if result.Kind != tt.wantKind {
				t.Fatalf("expected %v result, got %v", tt.wantKind, result.Kind)
			}

			if result.Raw.(float64) != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result.Raw.(float64))
			}
		})
	}
}

func TestBuiltinNotFound(t *testing.T) {
	vm, err := newTestVM([]module.Function{{
		Name:      "main",
		Constants: []any{"nonexistent"},
		Code: []core.Instruction{
			{Op: core.OpConst, A: 0},          // push "nonexistent"
			{Op: core.OpCall, A: 0, B: 0},     // call nonexistent()
			{Op:  core.OpReturn},
		},
	}}, Options{})
	if err != nil {
		t.Fatalf("failed to create VM: %v", err)
	}

	err = vm.Run()
	if err == nil {
		t.Fatal("expected error for nonexistent builtin, got nil")
	}
}

func TestBuiltinRegistration(t *testing.T) {
	vm, err := newTestVM([]module.Function{{
		Name: "main",
		Code: []core.Instruction{
			{Op: core.OpReturn},
		},
	}}, Options{})
	if err != nil {
		t.Fatalf("failed to create VM: %v", err)
	}

	expectedBuiltins := []string{"print", "len", "typeof", "str", "int", "float"}
	for _, name := range expectedBuiltins {
		fn, ok := vm.GetBuiltin(name)
		if !ok {
			t.Errorf("builtin %q not registered", name)
		}
		if fn == nil {
			t.Errorf("builtin %q is nil", name)
		}
	}
}

func TestBuiltinChainedCalls(t *testing.T) {
	// Test: typeof(str(42)) should return "string"
	vm, err := newTestVM([]module.Function{{
		Name:      "main",
		Constants: []any{"typeof", "str", int64(42)},
		Code: []core.Instruction{
			{Op: core.OpConst, A: 2},          // push 42
			{Op: core.OpConst, A: 1},          // push "str"
			{Op: core.OpCall, A: 1, B: 0},     // call str(42) -> "42"
			{Op: core.OpConst, A: 0},          // push "typeof"
			{Op: core.OpCall, A: 1, B: 0},     // call typeof("42") -> "string"
			{Op:  core.OpReturn},
		},
	}}, Options{})
	if err != nil {
		t.Fatalf("failed to create VM: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("VM run failed: %v", err)
	}

	if vm.stack.Size() != 1 {
		t.Fatalf("expected 1 value on stack, got %d", vm.stack.Size())
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("failed to pop result")
	}

	if result.Kind != core.ValueString {
		t.Fatalf("expected string result, got %v", result.Kind)
	}

	if result.Raw.(string) != "string" {
		t.Errorf("expected 'string', got %q", result.Raw.(string))
	}
}
