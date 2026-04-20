package lang_test

import (
	"testing"

	rt "ialang/pkg/lang/runtime"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	rvm "ialang/pkg/lang/runtime/vm"
)

// ============================================================
// Spread Operator Integration Tests
// ============================================================

func TestSpreadArrayBasic(t *testing.T) {
	source := `
let arr1 = [1, 2, 3];
let arr2 = [...arr1, 4, 5];
`
	runVMTestNoError(t, source)
}

func TestSpreadArrayAtEnd(t *testing.T) {
	source := `
let arr1 = [1, 2];
let arr2 = [0, ...arr1];
`
	runVMTestNoError(t, source)
}

func TestSpreadArrayInMiddle(t *testing.T) {
	source := `
let arr1 = [2, 3];
let arr2 = [1, ...arr1, 4];
`
	runVMTestNoError(t, source)
}

func TestSpreadMultipleArrays(t *testing.T) {
	source := `
let arr1 = [1];
let arr2 = [2];
let arr3 = [...arr1, ...arr2];
`
	runVMTestNoError(t, source)
}

func TestSpreadObjectBasic(t *testing.T) {
	source := `
let obj1 = { a: 1, b: 2 };
let obj2 = { ...obj1, c: 3 };
`
	runVMTestNoError(t, source)
}

func TestSpreadObjectOverride(t *testing.T) {
	source := `
let obj1 = { a: 1, b: 2 };
let obj2 = { ...obj1, b: 3 };
`
	runVMTestNoError(t, source)
}

func TestSpreadMultipleObjects(t *testing.T) {
	source := `
let obj1 = { a: 1 };
let obj2 = { b: 2 };
let obj3 = { ...obj1, ...obj2 };
`
	runVMTestNoError(t, source)
}

func TestSpreadFunctionCallBasic(t *testing.T) {
	source := `
function sum(a, b, c) {
    return a + b + c;
}
let args = [1, 2, 3];
sum(...args);
`
	runVMTestNoError(t, source)
}

func TestSpreadFunctionCallWithRegularArgs(t *testing.T) {
	source := `
function concat(a, b, c, d) {
    return a + b + c + d;
}
let args = [2, 3];
concat(1, ...args, 4);
`
	runVMTestNoError(t, source)
}

func TestSpreadMultipleFunctionCalls(t *testing.T) {
	source := `
function sum(a, b, c, d, e) {
    return a + b + c + d + e;
}
let args1 = [1, 2];
let args2 = [4, 5];
sum(...args1, 3, ...args2);
`
	runVMTestNoError(t, source)
}

// Tests to verify actual values
func TestSpreadArrayValues(t *testing.T) {
	source := `
let arr1 = [1, 2, 3];
let arr2 = [...arr1, 4, 5];
arr2[0] == 1 && arr2[1] == 2 && arr2[2] == 3 && arr2[3] == 4 && arr2[4] == 5
`
	chunk := compileTestSource(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "spread_array_values.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("spread array values error: %v", err)
	}
}

func TestSpreadObjectValues(t *testing.T) {
	source := `
let obj1 = { a: 1, b: 2 };
let obj2 = { ...obj1, c: 3 };
obj2["a"] == 1 && obj2["b"] == 2 && obj2["c"] == 3
`
	chunk := compileTestSource(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "spread_object_values.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("spread object values error: %v", err)
	}
}

func TestSpreadFunctionCallValues(t *testing.T) {
	source := `
function sum(a, b, c) {
    return a + b + c;
}
let args = [1, 2, 3];
sum(...args) == 6
`
	chunk := compileTestSource(t, source)
	runtime := rt.NewGoroutineRuntime()
	vm := rvm.NewVM(chunk, rtbuiltin.DefaultModules(runtime), nil, "spread_call_values.ia", runtime)
	if err := vm.Run(); err != nil {
		t.Fatalf("spread call values error: %v", err)
	}
}
