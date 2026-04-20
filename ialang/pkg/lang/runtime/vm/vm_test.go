package vm

import (
	"strings"
	"testing"
)

func testChunk() *Chunk {
	return &Chunk{}
}

func addConstant(c *Chunk, v any) int {
	c.Constants = append(c.Constants, v)
	return len(c.Constants) - 1
}

func runChunkForTest(t *testing.T, chunk *Chunk) (*VM, error) {
	t.Helper()
	v := NewVM(chunk, nil, nil, "", nil)
	return v, v.Run()
}

func TestNewVMInitializesRuntimeState(t *testing.T) {
	chunk := testChunk()
	v := NewVM(chunk, map[string]Value{"mod": Object{"x": float64(1)}}, nil, "main.ia", nil)

	if v.chunk != chunk {
		t.Fatal("expected VM to keep provided chunk")
	}
	if v.env == nil {
		t.Fatal("expected lexical environment")
	}
	if v.asyncRuntime == nil {
		t.Fatal("expected default async runtime")
	}
	if v.exports == nil {
		t.Fatal("expected exports object")
	}
	if _, ok := v.globals["print"]; !ok {
		t.Fatal("expected print global")
	}
}

func TestStackPushPop(t *testing.T) {
	v := NewVM(testChunk(), nil, nil, "", nil)
	v.push(float64(1))
	v.push("two")

	val, err := v.pop()
	if err != nil {
		t.Fatalf("unexpected pop error: %v", err)
	}
	if val != "two" {
		t.Fatalf("expected LIFO value 'two', got %v", val)
	}

	val, err = v.pop()
	if err != nil {
		t.Fatalf("unexpected pop error: %v", err)
	}
	if val != float64(1) {
		t.Fatalf("expected 1, got %v", val)
	}

	if _, err := v.pop(); err == nil {
		t.Fatal("expected stack underflow error")
	}
}

func TestRunConstantAddReturn(t *testing.T) {
	chunk := testChunk()
	left := addConstant(chunk, float64(2))
	right := addConstant(chunk, float64(3))
	chunk.Emit(OpConstant, left, 0)
	chunk.Emit(OpConstant, right, 0)
	chunk.Emit(OpAdd, 0, 0)
	chunk.Emit(OpReturn, 1, 0)

	v := NewVM(chunk, nil, nil, "", nil)
	got, err := v.runChunk()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if got != float64(5) {
		t.Fatalf("expected 5, got %v", got)
	}
}

func TestDefineGetSetName(t *testing.T) {
	chunk := testChunk()
	name := addConstant(chunk, "x")
	one := addConstant(chunk, float64(1))
	two := addConstant(chunk, float64(2))
	chunk.Emit(OpConstant, one, 0)
	chunk.Emit(OpDefineName, name, 0)
	chunk.Emit(OpConstant, two, 0)
	chunk.Emit(OpSetName, name, 0)
	chunk.Emit(OpGetName, name, 0)
	chunk.Emit(OpReturn, 1, 0)

	v := NewVM(chunk, nil, nil, "", nil)
	got, err := v.runChunk()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if got != float64(2) {
		t.Fatalf("expected 2, got %v", got)
	}
	envVal, ok := v.GetEnv("x")
	if !ok || envVal != float64(2) {
		t.Fatalf("expected env x=2, got %v (ok=%v)", envVal, ok)
	}
}

func TestDefineAndGetGlobal(t *testing.T) {
	chunk := testChunk()
	name := addConstant(chunk, "g")
	value := addConstant(chunk, "global")
	chunk.Emit(OpConstant, value, 0)
	chunk.Emit(OpDefineGlobal, name, 0)
	chunk.Emit(OpGetGlobal, name, 0)
	chunk.Emit(OpReturn, 1, 0)

	v := NewVM(chunk, nil, nil, "", nil)
	got, err := v.runChunk()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if got != "global" {
		t.Fatalf("expected global value, got %v", got)
	}
	if v.Globals()["g"] != "global" {
		t.Fatalf("expected global map to contain g, got %v", v.Globals()["g"])
	}
}

func TestUnaryOps(t *testing.T) {
	t.Run("typeof", func(t *testing.T) {
		chunk := testChunk()
		value := addConstant(chunk, float64(42))
		chunk.Emit(OpConstant, value, 0)
		chunk.Emit(OpTypeof, 0, 0)
		chunk.Emit(OpReturn, 1, 0)

		v := NewVM(chunk, nil, nil, "", nil)
		got, err := v.runChunk()
		if err != nil {
			t.Fatalf("unexpected run error: %v", err)
		}
		if got != "number" {
			t.Fatalf("expected number, got %v", got)
		}
	})

	t.Run("not", func(t *testing.T) {
		chunk := testChunk()
		value := addConstant(chunk, false)
		chunk.Emit(OpConstant, value, 0)
		chunk.Emit(OpNot, 0, 0)
		chunk.Emit(OpReturn, 1, 0)

		v := NewVM(chunk, nil, nil, "", nil)
		got, err := v.runChunk()
		if err != nil {
			t.Fatalf("unexpected run error: %v", err)
		}
		if got != true {
			t.Fatalf("expected true, got %v", got)
		}
	})

	t.Run("negate", func(t *testing.T) {
		chunk := testChunk()
		value := addConstant(chunk, float64(7))
		chunk.Emit(OpConstant, value, 0)
		chunk.Emit(OpNeg, 0, 0)
		chunk.Emit(OpReturn, 1, 0)

		v := NewVM(chunk, nil, nil, "", nil)
		got, err := v.runChunk()
		if err != nil {
			t.Fatalf("unexpected run error: %v", err)
		}
		if got != float64(-7) {
			t.Fatalf("expected -7, got %v", got)
		}
	})
}

func TestJumpIfFalseSelectsElseBranch(t *testing.T) {
	chunk := testChunk()
	cond := addConstant(chunk, false)
	thenVal := addConstant(chunk, "then")
	elseVal := addConstant(chunk, "else")
	chunk.Emit(OpConstant, cond, 0)
	chunk.Emit(OpJumpIfFalse, 4, 0)
	chunk.Emit(OpConstant, thenVal, 0)
	chunk.Emit(OpJump, 5, 0)
	chunk.Emit(OpConstant, elseVal, 0)
	chunk.Emit(OpReturn, 1, 0)

	v := NewVM(chunk, nil, nil, "", nil)
	got, err := v.runChunk()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if got != "else" {
		t.Fatalf("expected else branch, got %v", got)
	}
}

func TestJumpLoopAccumulatesValue(t *testing.T) {
	chunk := testChunk()
	name := addConstant(chunk, "i")
	zero := addConstant(chunk, float64(0))
	one := addConstant(chunk, float64(1))
	three := addConstant(chunk, float64(3))
	chunk.Emit(OpConstant, zero, 0)
	chunk.Emit(OpDefineName, name, 0)
	loopStart := len(chunk.Code)
	chunk.Emit(OpGetName, name, 0)
	chunk.Emit(OpConstant, three, 0)
	chunk.Emit(OpLess, 0, 0)
	chunk.Emit(OpJumpIfFalse, 11, 0)
	chunk.Emit(OpGetName, name, 0)
	chunk.Emit(OpConstant, one, 0)
	chunk.Emit(OpAdd, 0, 0)
	chunk.Emit(OpSetName, name, 0)
	chunk.Emit(OpJump, loopStart, 0)
	chunk.Emit(OpGetName, name, 0)
	chunk.Emit(OpReturn, 1, 0)

	v := NewVM(chunk, nil, nil, "", nil)
	got, err := v.runChunk()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if got != float64(3) {
		t.Fatalf("expected loop result 3, got %v", got)
	}
}

func TestNativeFunctionCall(t *testing.T) {
	chunk := testChunk()
	fnIdx := addConstant(chunk, NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, nil
		}
		return args[0].(float64) + args[1].(float64), nil
	}))
	left := addConstant(chunk, float64(4))
	right := addConstant(chunk, float64(5))
	chunk.Emit(OpConstant, fnIdx, 0)
	chunk.Emit(OpConstant, left, 0)
	chunk.Emit(OpConstant, right, 0)
	chunk.Emit(OpCall, 2, 0)
	chunk.Emit(OpReturn, 1, 0)

	v := NewVM(chunk, nil, nil, "", nil)
	got, err := v.runChunk()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if got != float64(9) {
		t.Fatalf("expected 9, got %v", got)
	}
}

func TestUserFunctionCall(t *testing.T) {
	fnChunk := testChunk()
	name := addConstant(fnChunk, "x")
	one := addConstant(fnChunk, float64(1))
	fnChunk.Emit(OpGetName, name, 0)
	fnChunk.Emit(OpConstant, one, 0)
	fnChunk.Emit(OpAdd, 0, 0)
	fnChunk.Emit(OpReturn, 1, 0)

	chunk := testChunk()
	fnIdx := addConstant(chunk, &UserFunction{Name: "inc", Params: []string{"x"}, Chunk: fnChunk})
	arg := addConstant(chunk, float64(6))
	chunk.Emit(OpConstant, fnIdx, 0)
	chunk.Emit(OpConstant, arg, 0)
	chunk.Emit(OpCall, 1, 0)
	chunk.Emit(OpReturn, 1, 0)

	v := NewVM(chunk, nil, nil, "", nil)
	got, err := v.runChunk()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if got != float64(7) {
		t.Fatalf("expected 7, got %v", got)
	}
}

func TestRuntimeErrorPropagates(t *testing.T) {
	chunk := testChunk()
	left := addConstant(chunk, "not-number")
	right := addConstant(chunk, float64(2))
	chunk.Emit(OpConstant, left, 0)
	chunk.Emit(OpConstant, right, 0)
	chunk.Emit(OpSub, 0, 0)
	chunk.Emit(OpReturn, 1, 0)

	_, err := runChunkForTest(t, chunk)
	if err == nil {
		t.Fatal("expected runtime error")
	}
	if !strings.Contains(err.Error(), "operator - expects numbers") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestThrowCaughtByTryFrame(t *testing.T) {
	chunk := testChunk()
	catchName := addConstant(chunk, "err")
	thrown := addConstant(chunk, "boom")
	caught := addConstant(chunk, "caught")
	chunk.Emit(OpPushTry, 4, catchName)
	chunk.Emit(OpConstant, thrown, 0)
	chunk.Emit(OpThrow, 0, 0)
	chunk.Emit(OpPopTry, 0, 0)
	chunk.Emit(OpConstant, caught, 0)
	chunk.Emit(OpReturn, 1, 0)

	v := NewVM(chunk, nil, nil, "", nil)
	got, err := v.runChunk()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if got != "caught" {
		t.Fatalf("expected caught value, got %v", got)
	}
	if envVal, ok := v.GetEnv("err"); !ok || envVal != "boom" {
		t.Fatalf("expected caught error in env, got %v (ok=%v)", envVal, ok)
	}
}
