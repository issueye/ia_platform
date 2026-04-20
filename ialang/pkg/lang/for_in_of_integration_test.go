package lang_test

import (
	"testing"

	rvm "ialang/pkg/lang/runtime/vm"
)

func mustEnvFloat64(t *testing.T, vmEnvGetter *rvm.VM, name string) float64 {
	t.Helper()
	val, ok := vmEnvGetter.GetEnv(name)
	if !ok {
		t.Fatalf("env %q not found", name)
	}
	num, ok := val.(float64)
	if !ok {
		t.Fatalf("env %q type = %T, want float64", name, val)
	}
	return num
}

func mustEnvBool(t *testing.T, vmEnvGetter *rvm.VM, name string) bool {
	t.Helper()
	val, ok := vmEnvGetter.GetEnv(name)
	if !ok {
		t.Fatalf("env %q not found", name)
	}
	b, ok := val.(bool)
	if !ok {
		t.Fatalf("env %q type = %T, want bool", name, val)
	}
	return b
}

func mustEnvString(t *testing.T, vmEnvGetter *rvm.VM, name string) string {
	t.Helper()
	val, ok := vmEnvGetter.GetEnv(name)
	if !ok {
		t.Fatalf("env %q not found", name)
	}
	s, ok := val.(string)
	if !ok {
		t.Fatalf("env %q type = %T, want string", name, val)
	}
	return s
}

func TestForOfIteratesNullElements(t *testing.T) {
	source := `
let arr = [1, null, 3];
let count = 0;
let seenNull = false;
for (value of arr) {
  count = count + 1;
  if (value == null) {
    seenNull = true;
  }
}
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
	if got := mustEnvFloat64(t, vm, "count"); got != 3 {
		t.Fatalf("count = %v, want 3", got)
	}
	if got := mustEnvBool(t, vm, "seenNull"); !got {
		t.Fatal("seenNull = false, want true")
	}
}

func TestForInDeterministicOrder(t *testing.T) {
	source := `
let obj = {"b": 2, "a": 1, "c": 3};
let keys = "";
for (k in obj) {
  keys = keys + k;
}
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
	if got := mustEnvString(t, vm, "keys"); got != "abc" {
		t.Fatalf("keys = %q, want %q", got, "abc")
	}
}

func TestForOfBreakAndContinue(t *testing.T) {
	source := `
let arr = [1, 2, 3, 4, 5];
let sum = 0;
for (x of arr) {
  if (x == 2) {
    continue;
  }
  if (x == 5) {
    break;
  }
  sum = sum + x;
}
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
	if got := mustEnvFloat64(t, vm, "sum"); got != 8 {
		t.Fatalf("sum = %v, want 8", got)
	}
}

func TestForOfString(t *testing.T) {
	source := `
let text = "abc";
let out = "";
for (ch of text) {
  out = out + ch;
}
`
	chunk := compileTestSource(t, source)
	vm := createTestVM(chunk)
	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
	if got := mustEnvString(t, vm, "out"); got != "abc" {
		t.Fatalf("out = %q, want %q", got, "abc")
	}
}
