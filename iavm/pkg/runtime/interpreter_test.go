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
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 8 {
		t.Fatalf("expected 8, got %v", val)
	}
}

func TestInterpret_Sub(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(10), int64(3)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpSub},
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

func TestInterpret_Mul(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(6), int64(7)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpMul},
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

func TestInterpret_Div(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(21), int64(3)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpDiv},
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

func TestInterpret_Mod(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(17), int64(5)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpMod},
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

func TestInterpret_Neg(t *testing.T) {
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
					{Op: core.OpNeg},
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
	if val.Kind != core.ValueI64 || val.Raw.(int64) != -42 {
		t.Fatalf("expected -42, got %v", val)
	}
}

func TestInterpret_Not(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{true},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpNot},
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
	if val.Kind != core.ValueBool || val.Raw.(bool) != false {
		t.Fatalf("expected false, got %v", val)
	}
}

func TestInterpret_Eq(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(42), int64(42)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpEq},
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
	if val.Kind != core.ValueBool || val.Raw.(bool) != true {
		t.Fatalf("expected true, got %v", val)
	}
}

func TestInterpret_Gt(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(10), int64(5)},
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
	if val.Kind != core.ValueBool || val.Raw.(bool) != true {
		t.Fatalf("expected true, got %v", val)
	}
}

func TestInterpret_Lt(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(3), int64(7)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpLt},
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
	if val.Kind != core.ValueBool || val.Raw.(bool) != true {
		t.Fatalf("expected true, got %v", val)
	}
}

func TestInterpret_Jump(t *testing.T) {
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
					{Op: core.OpJump, A: 2},
					{Op: core.OpConst, A: 0}, // skipped
					{Op: core.OpConst, A: 1},
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
				Constants: []any{false, int64(42), int64(99)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpJumpIfFalse, A: 4},
					{Op: core.OpConst, A: 1},
					{Op: core.OpReturn},
					{Op: core.OpConst, A: 2},
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
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 99 {
		t.Fatalf("expected 99, got %v", val)
	}
}

func TestInterpret_Call(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}, {}},
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
				Constants: []any{int64(3), int64(4)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpCall, A: 0, B: 2},
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

func TestInterpret_LoadStoreLocal(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Locals:    []core.ValueKind{core.ValueI64},
				Constants: []any{int64(42)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpStoreLocal, A: 0},
					{Op: core.OpLoadLocal, A: 0},
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

func TestInterpret_LoadStoreGlobal(t *testing.T) {
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
					{Op: core.OpStoreGlobal, A: 0},
					{Op: core.OpLoadGlobal, A: 0},
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
		t.Fatalf("expected array ref, got %v", val.Kind)
	}
	arr := val.Raw.([]core.Value)
	if len(arr) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(arr))
	}
}

func TestInterpret_MakeObject(t *testing.T) {
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
					{Op: core.OpMakeObject},
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
	if val.Kind != core.ValueObjectRef {
		t.Fatalf("expected object ref, got %v", val.Kind)
	}
}

func TestInterpret_GetSetProp(t *testing.T) {
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
					{Op: core.OpDup},
					{Op: core.OpGetProp, A: 0},
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

func TestInterpret_DupPop(t *testing.T) {
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
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 42 {
		t.Fatalf("expected 42, got %v", val)
	}
}

func TestInterpret_BitAnd(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(0b1100), int64(0b1010)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpBitAnd},
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
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 0b1000 {
		t.Fatalf("expected 8, got %v", val)
	}
}

func TestInterpret_BitOr(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(0b1100), int64(0b1010)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpBitOr},
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
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 0b1110 {
		t.Fatalf("expected 14, got %v", val)
	}
}

func TestInterpret_Shl(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(1), int64(3)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpShl},
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
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 8 {
		t.Fatalf("expected 8, got %v", val)
	}
}

func TestInterpret_Shr(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(8), int64(2)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpShr},
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

func TestInterpret_And(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{true, false},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpAnd},
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
	if val.Kind != core.ValueBool || val.Raw.(bool) != false {
		t.Fatalf("expected false, got %v", val)
	}
}

func TestInterpret_Or(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{true, false},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpOr},
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
	if val.Kind != core.ValueBool || val.Raw.(bool) != true {
		t.Fatalf("expected true, got %v", val)
	}
}

func TestInterpret_Typeof(t *testing.T) {
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
	if val.Kind != core.ValueString || val.Raw.(string) != "number" {
		t.Fatalf("expected 'number', got %v", val)
	}
}

func TestInterpret_Index(t *testing.T) {
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
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 2 {
		t.Fatalf("expected 2, got %v", val)
	}
}

func TestInterpret_Index_String(t *testing.T) {
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

// TestInterpret_Truthy verifies that OpTruthy correctly preserves the isTruthy
// semantic for various value kinds.
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
							{Op: core.OpTruthy},
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

// TestInterpret_JumpIfTrue verifies that OpJumpIfTrue correctly jumps when the
// value is truthy.
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
				Constants: []any{int64(42), int64(99)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},       // push 42 (truthy value)
					{Op: core.OpJumpIfTrue, A: 4},   // if truthy -> jump to instruction 4
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

func TestInterpret_ObjectKeys(t *testing.T) {
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"a", int64(1), "b", int64(2)},
				Code: []core.Instruction{
					{Op: core.OpMakeObject},
					{Op: core.OpDup},
					{Op: core.OpConst, A: 1},
					{Op: core.OpSetProp, A: 0},
					{Op: core.OpDup},
					{Op: core.OpConst, A: 3},
					{Op: core.OpSetProp, A: 2},
					{Op: core.OpObjectKeys},
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
		t.Fatalf("expected array ref, got %v", val.Kind)
	}
	arr := val.Raw.([]core.Value)
	if len(arr) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(arr))
	}
}

func TestInterpret_JumpIfNullish_NonNull(t *testing.T) {
	// Stack has non-null value, should NOT jump, return 42.
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(42), int64(99)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},          // push 42 (non-null)
					{Op: core.OpJumpIfNullish, A: 4},   // should NOT jump
					{Op: core.OpConst, A: 0},           // push 42
					{Op: core.OpReturn},                // return 42
					{Op: core.OpConst, A: 1},           // push 99 (jump target, skipped)
					{Op: core.OpReturn},                // return 99
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

func TestInterpret_JumpIfNullish_Null(t *testing.T) {
	// Stack has null, should jump to target and return 99.
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(42), int64(99), nil},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 2},          // push null
					{Op: core.OpJumpIfNullish, A: 4},   // SHOULD jump (null is nullish)
					{Op: core.OpConst, A: 0},           // push 42 (skipped)
					{Op: core.OpReturn},                // return 42 (skipped)
					{Op: core.OpConst, A: 1},           // push 99 (jump target)
					{Op: core.OpReturn},                // return 99
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
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 99 {
		t.Fatalf("expected 99, got %v", val)
	}
}

func TestInterpret_JumpIfNotNullish(t *testing.T) {
	// Stack has non-null value, should jump to target and return 99.
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(42), int64(99)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},            // push 42 (non-null)
					{Op: core.OpJumpIfNotNullish, A: 4},  // SHOULD jump (42 is not null)
					{Op: core.OpConst, A: 0},             // push 42 (skipped)
					{Op: core.OpReturn},                  // return 42 (skipped)
					{Op: core.OpConst, A: 1},             // push 99 (jump target)
					{Op: core.OpReturn},                  // return 99
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
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 99 {
		t.Fatalf("expected 99, got %v", val)
	}
}

func TestInterpret_JumpIfNotNullish_Null(t *testing.T) {
	// Stack has null, should NOT jump, return 42.
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{int64(42), int64(99), nil},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 2},            // push null
					{Op: core.OpJumpIfNotNullish, A: 4},  // should NOT jump (null is nullish)
					{Op: core.OpConst, A: 0},             // push 42
					{Op: core.OpReturn},                  // return 42
					{Op: core.OpConst, A: 1},             // push 99 (jump target, skipped)
					{Op: core.OpReturn},                  // return 99
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

func TestInterpret_TryCatch(t *testing.T) {
	// try { throw "error"; } catch (e) { return 42; }
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"error", int64(42)},
				Code: []core.Instruction{
					{Op: core.OpPushTry, A: 3},  // catch handler at instruction 3
					{Op: core.OpConst, A: 0},     // push "error"
					{Op: core.OpThrow},           // throw "error"
					{Op: core.OpPopTry},          // pop try (unreachable)
					{Op: core.OpConst, A: 1},     // push 42 (catch handler)
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

func TestInterpret_TryCatch_Rethrow(t *testing.T) {
	// try { throw "inner"; } catch (e) { throw e; }
	// This should propagate the exception since there's no outer handler.
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"inner"},
				Code: []core.Instruction{
					{Op: core.OpPushTry, A: 3},
					{Op: core.OpConst, A: 0},
					{Op: core.OpThrow},
					{Op: core.OpPopTry},
					{Op: core.OpThrow}, // rethrow the caught exception
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
	if err == nil {
		t.Fatal("expected uncaught exception error")
	}
}

func TestInterpret_NestedTryCatch(t *testing.T) {
	// try {
	//   try { throw "err"; } catch (e) { return 21; }
	// } catch (e) { return 99; }
	// inner catch should handle it, return 21.
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"err", int64(21), int64(99)},
				Code: []core.Instruction{
					{Op: core.OpPushTry, A: 7},   // outer catch at 7
					{Op: core.OpPushTry, A: 4},   // inner catch at 4
					{Op: core.OpConst, A: 0},
					{Op: core.OpThrow},
					{Op: core.OpPopTry},
					{Op: core.OpConst, A: 1},     // return 21
					{Op: core.OpReturn},
					{Op: core.OpPopTry},
					{Op: core.OpConst, A: 2},     // return 99 (outer catch)
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
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 21 {
		t.Fatalf("expected 21, got %v", val)
	}
}

func TestInterpret_TryCatch_Unhandled(t *testing.T) {
	// throw without any try-catch => error
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Constants: []any{"boom"},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpThrow},
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
	if err == nil {
		t.Fatal("expected error for unhandled exception")
	}
}

func TestInterpret_Closure(t *testing.T) {
	// OpClosure loads a function reference; OpCall invokes it.
	mod := &module.Module{
		Magic:   "IAVM",
		Version: 1,
		Target:  "ialang",
		Types:   []core.FuncType{{}, {}},
		Functions: []module.Function{
			{
				Name:      "target",
				TypeIndex: 0,
				Constants: []any{int64(42)},
				Code: []core.Instruction{
					{Op: core.OpConst, A: 0},
					{Op: core.OpReturn},
				},
			},
			{
				Name:      "entry",
				TypeIndex: 1,
				Code: []core.Instruction{
					{Op: core.OpClosure, A: 0},
					{Op: core.OpCall, A: 0, B: 0},
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

func TestInterpret_AddStringConcat(t *testing.T) {
	mod := &module.Module{
		Magic: "IAVM", Version: 1, Target: "ialang",
		Types: []core.FuncType{{}},
		Functions: []module.Function{{
			Name: "entry", TypeIndex: 0,
			Constants: []any{"hello ", "world"},
			Code: []core.Instruction{
				{Op: core.OpConst, A: 0}, {Op: core.OpConst, A: 1},
				{Op: core.OpAdd}, {Op: core.OpReturn},
			},
		}},
	}
	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := vm.Run(); err != nil {
		t.Fatal(err)
	}
	val := vm.stack.Pop()
	if val.Kind != core.ValueString || val.Raw.(string) != "hello world" {
		t.Fatalf("expected 'hello world', got %v", val)
	}
}

func TestInterpret_AddStringNumber(t *testing.T) {
	mod := &module.Module{
		Magic: "IAVM", Version: 1, Target: "ialang",
		Types: []core.FuncType{{}},
		Functions: []module.Function{{
			Name: "entry", TypeIndex: 0,
			Constants: []any{"count: ", int64(42)},
			Code: []core.Instruction{
				{Op: core.OpConst, A: 0}, {Op: core.OpConst, A: 1},
				{Op: core.OpAdd}, {Op: core.OpReturn},
			},
		}},
	}
	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := vm.Run(); err != nil {
		t.Fatal(err)
	}
	val := vm.stack.Pop()
	if val.Kind != core.ValueString || val.Raw.(string) != "count: 42" {
		t.Fatalf("expected 'count: 42', got %v", val)
	}
}

func TestInterpret_AddI64F64(t *testing.T) {
	mod := &module.Module{
		Magic: "IAVM", Version: 1, Target: "ialang",
		Types: []core.FuncType{{}},
		Functions: []module.Function{{
			Name: "entry", TypeIndex: 0,
			Constants: []any{int64(3), float64(2.5)},
			Code: []core.Instruction{
				{Op: core.OpConst, A: 0}, {Op: core.OpConst, A: 1},
				{Op: core.OpAdd}, {Op: core.OpReturn},
			},
		}},
	}
	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := vm.Run(); err != nil {
		t.Fatal(err)
	}
	val := vm.stack.Pop()
	if val.Kind != core.ValueF64 || val.Raw.(float64) != 5.5 {
		t.Fatalf("expected F64(5.5), got %v", val)
	}
}

func TestInterpret_CompareI64F64(t *testing.T) {
	mod := &module.Module{
		Magic: "IAVM", Version: 1, Target: "ialang",
		Types: []core.FuncType{{}},
		Functions: []module.Function{{
			Name: "entry", TypeIndex: 0,
			Constants: []any{int64(3), float64(5.0)},
			Code: []core.Instruction{
				{Op: core.OpConst, A: 0}, {Op: core.OpConst, A: 1},
				{Op: core.OpLt}, {Op: core.OpReturn},
			},
		}},
	}
	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := vm.Run(); err != nil {
		t.Fatal(err)
	}
	val := vm.stack.Pop()
	if val.Kind != core.ValueBool || !val.Raw.(bool) {
		t.Fatalf("expected true, got %v", val)
	}
}

func TestInterpret_EqualI64F64(t *testing.T) {
	mod := &module.Module{
		Magic: "IAVM", Version: 1, Target: "ialang",
		Types: []core.FuncType{{}},
		Functions: []module.Function{{
			Name: "entry", TypeIndex: 0,
			Constants: []any{int64(5), float64(5.0)},
			Code: []core.Instruction{
				{Op: core.OpConst, A: 0}, {Op: core.OpConst, A: 1},
				{Op: core.OpEq}, {Op: core.OpReturn},
			},
		}},
	}
	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := vm.Run(); err != nil {
		t.Fatal(err)
	}
	val := vm.stack.Pop()
	if val.Kind != core.ValueBool || !val.Raw.(bool) {
		t.Fatalf("expected true, got %v", val)
	}
}

func TestInterpret_F64ArrayIndex(t *testing.T) {
	mod := &module.Module{
		Magic: "IAVM", Version: 1, Target: "ialang",
		Types: []core.FuncType{{}},
		Functions: []module.Function{{
			Name: "entry", TypeIndex: 0,
			Constants: []any{int64(10), int64(20), int64(30), float64(1)},
			Code: []core.Instruction{
				{Op: core.OpConst, A: 0}, {Op: core.OpConst, A: 1}, {Op: core.OpConst, A: 2},
				{Op: core.OpMakeArray, A: 3},
				{Op: core.OpConst, A: 3}, {Op: core.OpIndex}, {Op: core.OpReturn},
			},
		}},
	}
	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := vm.Run(); err != nil {
		t.Fatal(err)
	}
	val := vm.stack.Pop()
	if val.Kind != core.ValueI64 || val.Raw.(int64) != 20 {
		t.Fatalf("expected 20, got %v", val)
	}
}

func TestInterpret_TryCatch_BindCatchVar(t *testing.T) {
	mod := &module.Module{
		Magic: "IAVM", Version: 1, Target: "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{"error_value"},
		Functions: []module.Function{{
			Name: "entry", TypeIndex: 0,
			Locals:    []core.ValueKind{core.ValueNull},
			Constants: []any{"error_value"},
			Code: []core.Instruction{
				{Op: core.OpPushTry, A: 3, B: 1},
				{Op: core.OpConst, A: 0},
				{Op: core.OpThrow},
				{Op: core.OpPopTry},
				{Op: core.OpLoadLocal, A: 0},
				{Op: core.OpReturn},
			},
		}},
	}
	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	val := vm.stack.Pop()
	if val.Kind != core.ValueString || val.Raw.(string) != "error_value" {
		t.Fatalf("expected 'error_value', got %v", val)
	}
}

func TestInterpret_F64StringIndex(t *testing.T) {
	mod := &module.Module{
		Magic: "IAVM", Version: 1, Target: "ialang",
		Types: []core.FuncType{{}},
		Functions: []module.Function{{
			Name: "entry", TypeIndex: 0,
			Constants: []any{"hello", float64(1)},
			Code: []core.Instruction{
				{Op: core.OpConst, A: 0}, {Op: core.OpConst, A: 1},
				{Op: core.OpIndex}, {Op: core.OpReturn},
			},
		}},
	}
	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := vm.Run(); err != nil {
		t.Fatal(err)
	}
	val := vm.stack.Pop()
	if val.Kind != core.ValueString || val.Raw.(string) != "e" {
		t.Fatalf("expected 'e', got %v", val)
	}
}
