package runtime

import (
	"testing"
	"iavm/pkg/core"
	"iavm/pkg/module"
)

func TestStack_PushPop(t *testing.T) {
	stack := NewStack(64)

	val := core.Value{Kind: core.ValueI64, Raw: int64(42)}
	stack.Push(val)

	if stack.Size() != 1 {
		t.Fatalf("expected size 1, got %d", stack.Size())
	}

	popped := stack.Pop()
	if popped.Raw.(int64) != 42 {
		t.Fatalf("expected 42, got %v", popped.Raw)
	}
}

func TestStack_Peek(t *testing.T) {
	stack := NewStack(64)

	val := core.Value{Kind: core.ValueI64, Raw: int64(99)}
	stack.Push(val)

	peeked := stack.Peek(0)
	if peeked.Raw.(int64) != 99 {
		t.Fatalf("expected 99, got %v", peeked.Raw)
	}

	if stack.Size() != 1 {
		t.Fatal("peek removed element")
	}
}

func TestStack_PopEmpty(t *testing.T) {
	stack := NewStack(64)
	val := stack.Pop()
	if val.Kind != core.ValueNull {
		t.Fatalf("expected null from empty stack, got %v", val.Kind)
	}
}

func TestInterpret_ConstReturn(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(42)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if vm.stack.Size() != 1 {
		t.Fatalf("expected 1 item on stack, got %d", vm.stack.Size())
	}
	val := vm.stack.Pop()
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestInterpret_Add(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(5), int64(3)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpAdd},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val := vm.stack.Pop()
	if val.Kind != core.ValueI64 {
		t.Fatalf("expected I64, got %v", val.Kind)
	}
	if val.Raw.(int64) != 8 {
		t.Fatalf("expected 8, got %v", val.Raw)
	}
}

func TestInterpret_Arithmetic(t *testing.T) {
	tests := []struct {
		name     string
		op       core.OpCode
		a, b     int64
		expected int64
	}{
		{"Sub", core.OpSub, 10, 3, 7},
		{"Mul", core.OpMul, 4, 5, 20},
		{"Div", core.OpDiv, 20, 4, 5},
		{"Mod", core.OpMod, 17, 5, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := &module.Module{
				Magic:   "IAVM",
				Version: 1,
				Target:  "ialang",
				Types:   []core.FuncType{{}},
				Functions: []module.Function{
					{
						Name:      "entry",
						TypeIndex: 0,
						Constants: []any{tt.a, tt.b},
						Code: []core.Instruction{
							{Op: core.OpConst, A: 0},
							{Op: core.OpConst, A: 1},
							{Op: tt.op},
							{Op: core.OpReturn},
						},
					},
				},
			}

			vm, err := New(mod, Options{})
			if err != nil {
				t.Fatalf("New VM failed: %v", err)
			}

			err = vm.Run()
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}

			val := vm.stack.Pop()
			if val.Raw.(int64) != tt.expected {
				t.Fatalf("expected %d, got %v", tt.expected, val.Raw)
			}
		})
	}
}

func TestInterpret_Comparison(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(5), int64(3)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpGt},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val := vm.stack.Pop()
	if val.Kind != core.ValueBool {
		t.Fatalf("expected Bool, got %v", val.Kind)
	}
	if val.Raw.(bool) != true {
		t.Fatalf("expected true, got %v", val.Raw)
	}
}

func TestInterpret_JumpIfFalse(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(0), int64(99), int64(42)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},       // 0: push 0 (falsy)
					{Op: core.OpJumpIfFalse, A: 4}, // 1: jump to instruction 4
					{Op: core.OpConst, A: 1},       // 2: push 99 (skipped)
					{Op: core.OpReturn},            // 3: return (skipped)
					{Op: core.OpConst, A: 2},       // 4: push 42
					{Op: core.OpReturn},            // 5: return
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val := vm.stack.Pop()
	if val.Raw.(int64) != 42 {
		t.Fatalf("expected 42, got %v", val.Raw)
	}
}

func TestInterpret_FunctionCall(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}, {}},
		Functions: []module.Function{
			{
				Name:      "double",
				TypeIndex: 0,
				Locals:    []core.ValueKind{core.ValueNull},
				Constants: []any{int64(2)},
				Code: []core.Instruction{
					{Op: core.OpLoadLocal, A: 0},
					{Op: core.OpConst, A: 0},
					{Op: core.OpMul},
					{Op: core.OpReturn},
				},
			},
			{
				Name:      "entry",
				TypeIndex: 1,
				Constants: []any{int64(21)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpCall, A: 0, B: 1},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val := vm.stack.Pop()
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestInterpret_FunctionCallStackBased(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}, {}},
		Functions: []module.Function{
			{
				Name:      "add",
				TypeIndex: 0,
				Locals:    []core.ValueKind{core.ValueNull, core.ValueNull, core.ValueNull},
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
				Constants: []any{int64(0), int64(3), int64(4)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},       // push function index 0
					{Op: core.OpConst, A: 1},       // push arg 3
					{Op: core.OpConst, A: 2},       // push arg 4
					{Op: core.OpCall, A: 2, B: 0},  // stack-based call add(3,4)
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val := vm.stack.Pop()
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 7 {
		t.Fatalf("expected 7, got %v", val)
	}
}

func TestInterpret_MakeArray(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(1), int64(2), int64(3)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpConst, A: 2},
					{Op: core.OpMakeArray, A: 3},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val := vm.stack.Pop()
	if val.Kind != core.ValueArrayRef {
		t.Fatalf("expected ArrayRef, got %v", val.Kind)
	}
	arr := val.Raw.([]core.Value)
	if len(arr) != 3 {
		t.Fatalf("expected array length 3, got %d", len(arr))
	}
}

func TestVM_MaxSteps(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(1)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 0},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{MaxSteps: 2})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != core.ErrResourceExhausted {
		t.Fatalf("expected ErrResourceExhausted, got %v", err)
	}
}

func TestInterpret_Dup(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(42)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpDup},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if vm.stack.Size() != 2 {
		t.Fatalf("expected 2 items on stack, got %d", vm.stack.Size())
	}
	b := vm.stack.Pop()
	a := vm.stack.Pop()
	if a.Raw.(int64) != 42 || b.Raw.(int64) != 42 {
		t.Fatalf("expected both 42, got %v and %v", a, b)
	}
}

func TestInterpret_Pop(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(1), int64(2)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpPop},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val := vm.stack.Pop()
	if val.Raw.(int64) != 1 {
		t.Fatalf("expected 1, got %v", val.Raw)
	}
}

func TestInterpret_BitOps(t *testing.T) {
	tests := []struct {
		name     string
		op       core.OpCode
		a, b     int64
		expected int64
	}{
		{"BitAnd", core.OpBitAnd, 0b1100, 0b1010, 0b1000},
		{"BitOr", core.OpBitOr, 0b1100, 0b1010, 0b1110},
		{"BitXor", core.OpBitXor, 0b1100, 0b1010, 0b0110},
		{"Shl", core.OpShl, 1, 4, 16},
		{"Shr", core.OpShr, 16, 2, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := &module.Module{
				Magic:   "IAVM",
				Version: 1,
				Target:  "ialang",
				Types:   []core.FuncType{{}},
				Functions: []module.Function{
					{
						Name:      "entry",
						TypeIndex: 0,
						Constants: []any{tt.a, tt.b},
						Code: []core.Instruction{
							{Op: core.OpConst, A: 0},
							{Op: core.OpConst, A: 1},
							{Op: tt.op},
							{Op: core.OpReturn},
						},
					},
				},
			}

			vm, err := New(mod, Options{})
			if err != nil {
				t.Fatalf("New VM failed: %v", err)
			}

			err = vm.Run()
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}

			val := vm.stack.Pop()
			if val.Kind != core.ValueI64 {
				t.Fatalf("expected I64, got %v", val.Kind)
			}
			if val.Raw.(int64) != tt.expected {
				t.Fatalf("expected %d, got %v", tt.expected, val.Raw)
			}
		})
	}
}

func TestInterpret_LogicalOps(t *testing.T) {
	tests := []struct {
		name     string
		op       core.OpCode
		a, b     core.Value
		expected core.Value
	}{
		{"And_true_true", core.OpAnd,
			core.Value{Kind: core.ValueBool, Raw: true},
			core.Value{Kind: core.ValueBool, Raw: true},
			core.Value{Kind: core.ValueBool, Raw: true}},
		{"And_true_false", core.OpAnd,
			core.Value{Kind: core.ValueBool, Raw: true},
			core.Value{Kind: core.ValueBool, Raw: false},
			core.Value{Kind: core.ValueBool, Raw: false}},
		{"And_false_true", core.OpAnd,
			core.Value{Kind: core.ValueBool, Raw: false},
			core.Value{Kind: core.ValueBool, Raw: true},
			core.Value{Kind: core.ValueBool, Raw: false}},
		{"Or_false_false", core.OpOr,
			core.Value{Kind: core.ValueBool, Raw: false},
			core.Value{Kind: core.ValueBool, Raw: false},
			core.Value{Kind: core.ValueBool, Raw: false}},
		{"Or_false_true", core.OpOr,
			core.Value{Kind: core.ValueBool, Raw: false},
			core.Value{Kind: core.ValueBool, Raw: true},
			core.Value{Kind: core.ValueBool, Raw: true}},
		{"Or_true_false", core.OpOr,
			core.Value{Kind: core.ValueBool, Raw: true},
			core.Value{Kind: core.ValueBool, Raw: false},
			core.Value{Kind: core.ValueBool, Raw: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := &module.Module{
				Magic:   "IAVM",
				Version: 1,
				Target:  "ialang",
				Types:   []core.FuncType{{}},
				Functions: []module.Function{
					{
						Name:      "entry",
						TypeIndex: 0,
						Constants: []any{tt.a, tt.b},
						Code: []core.Instruction{
							{Op: core.OpConst, A: 0},
							{Op: core.OpConst, A: 1},
							{Op: tt.op},
							{Op: core.OpReturn},
						},
					},
				},
			}

			vm, err := New(mod, Options{})
			if err != nil {
				t.Fatalf("New VM failed: %v", err)
			}

			err = vm.Run()
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}

			val := vm.stack.Pop()
			if val.Kind != core.ValueBool {
				t.Fatalf("expected Bool, got %v", val.Kind)
			}
			if val.Raw.(bool) != tt.expected.Raw.(bool) {
				t.Fatalf("expected %v, got %v", tt.expected.Raw.(bool), val.Raw.(bool))
			}
		})
	}
}

func TestInterpret_Typeof(t *testing.T) {
	tests := []struct {
		name     string
		val      any
		expected string
	}{
		{"Number", int64(42), "number"},
		{"String", "hello", "string"},
		{"Bool_true", true, "boolean"},
		{"Bool_false", false, "boolean"},
		{"Null", nil, "null"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := &module.Module{
				Magic:   "IAVM",
				Version: 1,
				Target:  "ialang",
				Types:   []core.FuncType{{}},
				Functions: []module.Function{
					{
						Name:      "entry",
						TypeIndex: 0,
						Constants: []any{tt.val},
						Code: []core.Instruction{
							{Op: core.OpConst, A: 0},
							{Op: core.OpTypeof},
							{Op: core.OpReturn},
						},
					},
				},
			}

			vm, err := New(mod, Options{})
			if err != nil {
				t.Fatalf("New VM failed: %v", err)
			}

			err = vm.Run()
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}

			val := vm.stack.Pop()
			if val.Kind != core.ValueString {
				t.Fatalf("expected String, got %v", val.Kind)
			}
			if val.Raw.(string) != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, val.Raw.(string))
			}
		})
	}
}

func TestInterpret_PushTryPopTry(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(42)},
				Code: []core.Instruction{
					{Op: core.OpPushTry, A: 2},
					{Op: core.OpConst, A: 0},
					{Op: core.OpPopTry},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val := vm.stack.Pop()
	if val.Raw.(int64) != 42 {
		t.Fatalf("expected 42, got %v", val.Raw)
	}
}

func TestInterpret_Throw(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"test error"},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpThrow},
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err == nil {
		t.Fatal("expected error from throw, got nil")
	}
}

func TestInterpret_IndexArray(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(1), int64(2), int64(3), int64(1)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpConst, A: 2},
					{Op: core.OpMakeArray, A: 3},
					{Op: core.OpConst, A: 3},
					{Op: core.OpIndex},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val := vm.stack.Pop()
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 2 {
		t.Fatalf("expected 2, got %v", val)
	}
}

func TestInterpret_IndexObject(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"x", int64(42)},
				Code: []core.Instruction{
					{Op: core.OpMakeObject},
					{Op: core.OpDup},
					{Op: core.OpConst, A: 1},
					{Op: core.OpSetProp, A: 0},
					{Op: core.OpConst, A: 0},
					{Op: core.OpIndex},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val := vm.stack.Pop()
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestInterpret_IndexString(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"hello", int64(1)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpIndex},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val := vm.stack.Pop()
	if val.Kind != core.ValueString || val.Raw.(string) != "e" {
		t.Fatalf("expected 'e', got %v", val)
	}
}

// TestInterpret_Truthy verifies that OpNot; OpNot (the lowering of OpTruthy)
// correctly preserves the isTruthy semantic for various value kinds.
func TestInterpret_Truthy(t *testing.T) {
	tests := []struct {
		name     string
		consts   []any
		expected bool
	}{
		{"truthy_int", []any{int64(42)}, true},
		{"falsy_int", []any{int64(0)}, false},
		{"truthy_float", []any{float64(3.14)}, true},
		{"falsy_float", []any{float64(0.0)}, false},
		{"truthy_string", []any{"hello"}, true},
		{"falsy_string", []any{""}, false},
		{"truthy_bool", []any{true}, true},
		{"falsy_bool", []any{false}, false},
		{"null", []any{nil}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod := &module.Module{
				Magic:   "IAVM",
				Version: 1,
				Target:  "ialang",
				Types:   []core.FuncType{{}},
				Functions: []module.Function{
					{
						Name:      "entry",
						TypeIndex: 0,
						Constants: tt.consts,
						Code: []core.Instruction{
							{Op: core.OpConst, A: 0},
							{Op: core.OpNot},    // !isTruthy(val)
							{Op: core.OpNot},    // !!isTruthy(val) == isTruthy(val)
							{Op: core.OpReturn},
						},
					},
				},
			}

			vm, err := New(mod, Options{})
			if err != nil {
				t.Fatalf("New VM failed: %v", err)
			}

			err = vm.Run()
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}

			val := vm.stack.Pop()
			if val.Kind != core.ValueBool || val.Raw.(bool) != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, val)
			}
		})
	}
}

// TestInterpret_JumpIfTrue verifies that OpNot; OpJumpIfFalse (the lowering of OpJumpIfTrue)
// correctly jumps when the value is truthy.
func TestInterpret_JumpIfTrue(t *testing.T) {
	// If truthy, jump over the falsy branch and return 42; else return 99.
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(42), int64(99), int64(0)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},       // push 42 (truthy value)
					{Op: core.OpNot},                // !isTruthy(42) = false
					{Op: core.OpJumpIfFalse, A: 5},  // if !false -> jump to instruction 5
					{Op: core.OpConst, A: 1},        // push 99 (falsy branch, skipped)
					{Op: core.OpReturn},             // return 99 (skipped)
					{Op: core.OpConst, A: 0},        // push 42 (truthy branch)
					{Op: core.OpReturn},             // return 42
				},
			},
		},
	}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New VM failed: %v", err)
	}

	err = vm.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	val := vm.stack.Pop()
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}
