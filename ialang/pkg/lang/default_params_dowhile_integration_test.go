package lang_test

import (
	"testing"

	comp "ialang/pkg/lang/compiler"
	"ialang/pkg/lang/frontend"
	rt "ialang/pkg/lang/runtime"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	rvm "ialang/pkg/lang/runtime/vm"
)

func compileChunkForDefaultParamsDowhile(t *testing.T, source string) *rt.Chunk {
	t.Helper()
	l := frontend.NewLexer(source)
	p := frontend.NewParser(l)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("parse errors: %v", p.Errors())
	}
	c := comp.NewCompiler()
	chunk, errs := c.Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	return chunk
}

type vmHelper struct {
	vm *rvm.VM
}

func newVMHelper(t *testing.T, source string) *vmHelper {
	t.Helper()
	chunk := compileChunkForDefaultParamsDowhile(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "test.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
	return &vmHelper{vm: vm}
}

func (h *vmHelper) get(name string) (rvm.Value, bool) {
	// Try env first, then globals
	if val, ok := h.vm.GetEnv(name); ok {
		return val, true
	}
	if val, ok := h.vm.Globals()[name]; ok {
		return val, true
	}
	return nil, false
}

func expectFloat64(t *testing.T, h *vmHelper, name string, expected float64) {
	t.Helper()
	val, ok := h.get(name)
	if !ok {
		t.Fatalf("variable '%s' not found", name)
	}
	if v, ok := val.(float64); ok {
		if v != expected {
			t.Errorf("expected %s=%v, got %v", name, expected, v)
		}
	} else {
		t.Errorf("expected %s to be float64, got %T", name, val)
	}
}

func expectBool(t *testing.T, h *vmHelper, name string, expected bool) {
	t.Helper()
	val, ok := h.get(name)
	if !ok {
		t.Fatalf("variable '%s' not found", name)
	}
	if v, ok := val.(bool); ok {
		if v != expected {
			t.Errorf("expected %s=%v, got %v", name, expected, v)
		}
	} else {
		t.Errorf("expected %s to be bool, got %T", name, val)
	}
}

// ============================================================
// Default Parameters Integration Tests
// ============================================================

func TestRuntime_DefaultParams_Basic(t *testing.T) {
	src := `
function foo(x = 10) {
  return x;
}
let a = foo();
let b = foo(5);
`
	h := newVMHelper(t, src)
	expectFloat64(t, h, "a", 10)
	expectFloat64(t, h, "b", 5)
}

func TestRuntime_DefaultParams_Multiple(t *testing.T) {
	src := `
function add(x = 1, y = 2, z = 3) {
  return x + y + z;
}
let a = add();
let b = add(10);
let c = add(10, 20);
let d = add(10, 20, 30);
`
	h := newVMHelper(t, src)
	expectFloat64(t, h, "a", 6)   // 1+2+3
	expectFloat64(t, h, "b", 15)  // 10+2+3
	expectFloat64(t, h, "c", 33)  // 10+20+3
	expectFloat64(t, h, "d", 60)  // 10+20+30
}

func TestRuntime_DefaultParams_ArrowFunction(t *testing.T) {
	src := `
let add = (x = 10, y = 20) => x + y;
let a = add();
let b = add(5);
let c = add(5, 15);
`
	h := newVMHelper(t, src)
	expectFloat64(t, h, "a", 30) // 10+20
	expectFloat64(t, h, "b", 25) // 5+20
	expectFloat64(t, h, "c", 20) // 5+15
}

// ============================================================
// Do-While Integration Tests
// ============================================================

func TestRuntime_DoWhile_Basic(t *testing.T) {
	src := `
let x = 0;
do {
  x = x + 1;
} while (x < 3);
`
	h := newVMHelper(t, src)
	expectFloat64(t, h, "x", 3)
}

func TestRuntime_DoWhile_AlwaysExecutesOnce(t *testing.T) {
	src := `
let executed = false;
do {
  executed = true;
} while (false);
`
	h := newVMHelper(t, src)
	expectBool(t, h, "executed", true)
}

func TestRuntime_DoWhile_WithBreak(t *testing.T) {
	src := `
let x = 0;
do {
  x = x + 1;
  if (x > 5) {
    break;
  }
} while (x < 10);
`
	h := newVMHelper(t, src)
	expectFloat64(t, h, "x", 6)
}

func TestRuntime_DoWhile_Counting(t *testing.T) {
	src := `
let sum = 0;
let i = 1;
do {
  sum = sum + i;
  i = i + 1;
} while (i <= 5);
`
	h := newVMHelper(t, src)
	expectFloat64(t, h, "sum", 15) // 1+2+3+4+5
}

// ============================================================
// Combined Tests
// ============================================================

func TestRuntime_DefaultParams_And_DoWhile_Combined(t *testing.T) {
	src := `
function countUp(max = 5) {
  let sum = 0;
  let i = 1;
  do {
    sum = sum + i;
    i = i + 1;
  } while (i <= max);
  return sum;
}

let a = countUp();
let b = countUp(3);
`
	h := newVMHelper(t, src)
	expectFloat64(t, h, "a", 15) // 1+2+3+4+5
	expectFloat64(t, h, "b", 6)  // 1+2+3
}
