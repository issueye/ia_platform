package lang

import (
	comp "ialang/pkg/lang/compiler"
	"ialang/pkg/lang/frontend"
	rt "ialang/pkg/lang/runtime"
	rtbuiltin "ialang/pkg/lang/runtime/builtin"
	rvm "ialang/pkg/lang/runtime/vm"
	"testing"
	"time"
)

type immediateAwaitable struct {
	value rt.Value
	err   error
}

func (a *immediateAwaitable) Await() (rt.Value, error) {
	return a.value, a.err
}

func (a *immediateAwaitable) IsDone() bool {
	return true
}

type countingRuntime struct {
	spawnCalls int
	awaitCalls int
}

func (r *countingRuntime) Spawn(task rt.AsyncTask) rt.Awaitable {
	r.spawnCalls++
	val, err := task()
	return &immediateAwaitable{value: val, err: err}
}

func (r *countingRuntime) AwaitValue(v rt.Value) (rt.Value, error) {
	r.awaitCalls++
	awaitable, ok := v.(rt.Awaitable)
	if !ok {
		return v, nil
	}
	return awaitable.Await()
}

func (r *countingRuntime) Name() string {
	return "counting-runtime"
}

func TestVMUsesInjectedRuntimeForAsyncCallAndAwait(t *testing.T) {
	source := `
async function f() { return 1; }
let p = f();
let x = await p;
`
	chunk := compileChunkForTest(t, source)
	runtime := &countingRuntime{}
	vm := rvm.NewVM(chunk, map[string]rt.Value{}, nil, "test_async_call.ia", runtime)

	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
	if runtime.spawnCalls != 1 {
		t.Fatalf("spawnCalls = %d, want 1", runtime.spawnCalls)
	}
	if runtime.awaitCalls != 1 {
		t.Fatalf("awaitCalls = %d, want 1", runtime.awaitCalls)
	}
}

func TestVMAwaitNonAwaitableUsesRuntime(t *testing.T) {
	source := `let x = await 42;`
	chunk := compileChunkForTest(t, source)
	runtime := &countingRuntime{}
	vm := rvm.NewVM(chunk, map[string]rt.Value{}, nil, "test_await_non_awaitable.ia", runtime)

	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
	if runtime.spawnCalls != 0 {
		t.Fatalf("spawnCalls = %d, want 0", runtime.spawnCalls)
	}
	if runtime.awaitCalls != 1 {
		t.Fatalf("awaitCalls = %d, want 1", runtime.awaitCalls)
	}
}

func TestVMCatchAsyncAwaitTimeoutAsObject(t *testing.T) {
	source := `
import { llm } from "@agent/sdk";
try {
  let msg = await llm.chatAsync("x");
  throw "expected-timeout";
} catch (e) {
  if (e.code != "ASYNC_AWAIT_TIMEOUT") {
    throw "unexpected-await-timeout-code";
  }
  if (e.name != "AsyncAwaitTimeout") {
    throw "unexpected-await-timeout-name";
  }
  if (e.kind != "timeout") {
    throw "unexpected-await-timeout-kind";
  }
  if (e.retryable != true) {
    throw "unexpected-await-timeout-retryable";
  }
  if (e.runtime != "goroutine") {
    throw "unexpected-await-timeout-runtime";
  }
  if (e.module != "timeout_await_test.ia") {
    throw "unexpected-await-timeout-module";
  }
  if (e.ip < 0) {
    throw "unexpected-await-timeout-ip";
  }
  if (e.op < 0) {
    throw "unexpected-await-timeout-op";
  }
  if (e["stack_depth"] < 0) {
    throw "unexpected-await-timeout-stack-depth";
  }
}
`
	chunk := compileChunkForTest(t, source)
	runtime := rt.NewGoroutineRuntimeWithOptions(rt.GoroutineRuntimeOptions{
		AwaitTimeout: time.Millisecond,
	})
	modules := rtbuiltin.DefaultModules(runtime)
	vm := rvm.NewVMWithOptions(
		chunk,
		modules,
		nil,
		"timeout_await_test.ia",
		runtime,
		rvm.VMOptions{StructuredRuntimeErrors: true},
	)

	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
}

func TestVMCatchAsyncTaskTimeoutAsObject(t *testing.T) {
	source := `
import { llm } from "@agent/sdk";
try {
  let msg = await llm.chatAsync("x");
  throw "expected-timeout";
} catch (e) {
  if (e.code != "ASYNC_TASK_TIMEOUT") {
    throw "unexpected-task-timeout-code";
  }
  if (e.name != "AsyncTaskTimeout") {
    throw "unexpected-task-timeout-name";
  }
  if (e.kind != "timeout") {
    throw "unexpected-task-timeout-kind";
  }
  if (e.retryable != true) {
    throw "unexpected-task-timeout-retryable";
  }
  if (e.runtime != "goroutine") {
    throw "unexpected-task-timeout-runtime";
  }
  if (e.module != "timeout_task_test.ia") {
    throw "unexpected-task-timeout-module";
  }
  if (e.ip < 0) {
    throw "unexpected-task-timeout-ip";
  }
  if (e.op < 0) {
    throw "unexpected-task-timeout-op";
  }
  if (e["stack_depth"] < 0) {
    throw "unexpected-task-timeout-stack-depth";
  }
}
`
	chunk := compileChunkForTest(t, source)
	runtime := rt.NewGoroutineRuntimeWithOptions(rt.GoroutineRuntimeOptions{
		TaskTimeout: time.Millisecond,
	})
	modules := rtbuiltin.DefaultModules(runtime)
	vm := rvm.NewVMWithOptions(
		chunk,
		modules,
		nil,
		"timeout_task_test.ia",
		runtime,
		rvm.VMOptions{StructuredRuntimeErrors: true},
	)

	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
}

func TestVMLegacyRuntimeErrorCatchValue(t *testing.T) {
	source := `
try {
  not_exists = 1;
} catch (e) {
  if (e != "assignment to undefined variable: not_exists") {
    throw "unexpected-legacy-runtime-error-value";
  }
}
`
	chunk := compileChunkForTest(t, source)
	vm := rvm.NewVM(chunk, map[string]rt.Value{}, nil, "legacy_runtime_error_test.ia", rt.NewGoroutineRuntime())

	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
}

func TestVMStructuredRuntimeErrorCatchValue(t *testing.T) {
	source := `
try {
  not_exists = 1;
} catch (e) {
  if (e.code != "RUNTIME_ERROR") {
    throw "unexpected-structured-runtime-error-code";
  }
  if (e.name != "RuntimeError") {
    throw "unexpected-structured-runtime-error-name";
  }
  if (e.kind != "runtime") {
    throw "unexpected-structured-runtime-error-kind";
  }
  if (e.retryable != false) {
    throw "unexpected-structured-runtime-error-retryable";
  }
  if (e.runtime != "goroutine") {
    throw "unexpected-structured-runtime-error-runtime";
  }
  if (e.module != "structured_runtime_error_test.ia") {
    throw "unexpected-structured-runtime-error-module";
  }
  if (e.ip < 0) {
    throw "unexpected-structured-runtime-error-ip";
  }
  if (e.op < 0) {
    throw "unexpected-structured-runtime-error-op";
  }
  if (e["stack_depth"] < 0) {
    throw "unexpected-structured-runtime-error-stack-depth";
  }
}
`
	chunk := compileChunkForTest(t, source)
	vm := rvm.NewVMWithOptions(
		chunk,
		map[string]rt.Value{},
		nil,
		"structured_runtime_error_test.ia",
		rt.NewGoroutineRuntime(),
		rvm.VMOptions{StructuredRuntimeErrors: true},
	)

	if err := vm.Run(); err != nil {
		t.Fatalf("vm.Run() unexpected error: %v", err)
	}
}

func compileChunkForTest(t *testing.T, source string) *rt.Chunk {
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
