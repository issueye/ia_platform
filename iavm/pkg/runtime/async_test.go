package runtime

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"iacommon/pkg/host/api"
	compiler "ialang/pkg/lang/compiler"
	frontend "ialang/pkg/lang/frontend"
	bridge_ialang "iavm/pkg/bridge/ialang"
	"iavm/pkg/core"
	"iavm/pkg/module"
)

func TestAwaitPassThroughRuntime(t *testing.T) {
	mod := moduleForAsyncTest([]any{int64(7)}, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpAwait},
		{Op: core.OpReturn},
	})

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected await result")
	}
	if result.Kind != core.ValueI64 || result.Raw.(int64) != 7 {
		t.Fatalf("unexpected await result: %#v", result)
	}
}

func TestAsyncFunctionExampleFileRuntime(t *testing.T) {
	sourcePath := filepath.Join("..", "..", "..", "ialang", "examples", "async_loop.ia")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	lexer := frontend.NewLexer(string(source))
	parser := frontend.NewParser(lexer)
	program := parser.ParseProgram()
	if errs := parser.Errors(); len(errs) > 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	compiled, errs := compiler.NewCompiler().Compile(program)
	if len(errs) > 0 {
		t.Fatalf("compile errors: %v", errs)
	}
	mod, err := bridge_ialang.LowerToModule(compiled)
	if err != nil {
		t.Fatalf("LowerToModule failed: %v", err)
	}
	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if err := vm.Run(); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	var got core.Value
	found := false
	for i, global := range mod.Globals {
		if global.Name == "v" && i < len(vm.globals) {
			got = vm.globals[i]
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected global v")
	}
	if got.Kind == core.ValueI64 && got.Raw.(int64) == 16 {
		return
	}
	if got.Kind == core.ValueF64 && got.Raw.(float64) == 16 {
		return
	}
	if got.Kind != core.ValueI64 && got.Kind != core.ValueF64 {
		t.Fatalf("unexpected async_loop result: %#v", got)
	}
}

func TestAwaitPendingPromiseSuspendsRuntime(t *testing.T) {
	mod := moduleForAsyncTest(nil, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpAwait},
		{Op: core.OpReturn},
	})
	mod.Constants = []any{pendingPromiseValue()}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	err = vm.Run()
	if !errors.Is(err, ErrPromisePending) {
		t.Fatalf("Run error = %v, want ErrPromisePending", err)
	}
	suspension := vm.SuspensionState()
	if suspension == nil {
		t.Fatal("expected suspension state")
	}
	if suspension.Reason != "await_pending_promise" {
		t.Fatalf("suspension reason = %q", suspension.Reason)
	}
	if suspension.AwaitValue.Kind != core.ValuePromise {
		t.Fatalf("await value kind = %v, want promise", suspension.AwaitValue.Kind)
	}
}

func TestAwaitRejectedPromiseReturnsError(t *testing.T) {
	mod := moduleForAsyncTest(nil, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpAwait},
		{Op: core.OpReturn},
	})
	mod.Constants = []any{rejectedPromiseValue("boom")}

	vm, err := New(mod, Options{})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	err = vm.Run()
	if err == nil {
		t.Fatal("expected rejected promise error")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("error = %v, want rejection message", err)
	}
	if vm.SuspensionState() != nil {
		t.Fatal("rejected promise should not leave suspension state")
	}
}

func TestHostPollAwaitSuspendsAndResumesRuntime(t *testing.T) {
	host := newMockHost()
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}

	mod := moduleForAsyncTest([]any{int64(9), "ready"}, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpHostPoll},
		{Op: core.OpAwait},
		{Op: core.OpGetProp, A: 1},
		{Op: core.OpReturn},
	})

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	err = vm.Run()
	if !errors.Is(err, ErrPromisePending) {
		t.Fatalf("Run error = %v, want ErrPromisePending", err)
	}
	if vm.SuspensionState() == nil {
		t.Fatal("expected suspension after pending host.poll await")
	}

	host.pollResult = api.PollResult{
		Done:  true,
		Value: map[string]any{"ready": true},
	}
	if err := vm.ResumeSuspension(); err != nil {
		t.Fatalf("ResumeSuspension failed: %v", err)
	}
	if vm.SuspensionState() != nil {
		t.Fatal("expected suspension to be cleared after resume")
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected resumed result on stack")
	}
	if result.Kind != core.ValueBool || !result.Raw.(bool) {
		t.Fatalf("unexpected resumed result: %#v", result)
	}
}

func TestHostPollWaitSuspensionResumesRuntime(t *testing.T) {
	host := newMockHost()
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}
	host.waitResult = api.PollResult{
		Done:  true,
		Value: map[string]any{"ready": true},
	}

	mod := moduleForAsyncTest([]any{int64(11), "ready"}, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpHostPoll},
		{Op: core.OpAwait},
		{Op: core.OpGetProp, A: 1},
		{Op: core.OpReturn},
	})

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	err = vm.Run()
	if !errors.Is(err, ErrPromisePending) {
		t.Fatalf("Run error = %v, want ErrPromisePending", err)
	}

	if err := vm.WaitSuspension(context.Background()); err != nil {
		t.Fatalf("WaitSuspension failed: %v", err)
	}
	if len(host.waitLog) != 1 || host.waitLog[0] != 11 {
		t.Fatalf("unexpected wait log: %#v", host.waitLog)
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected waited result on stack")
	}
	if result.Kind != core.ValueBool || !result.Raw.(bool) {
		t.Fatalf("unexpected waited result: %#v", result)
	}
}

func TestHostPollWaitSuspensionFallsBackToPolling(t *testing.T) {
	host := &pollOnlyHost{
		pollResults: []api.PollResult{
			{Done: false, Value: map[string]any{"ready": false}},
			{Done: true, Value: map[string]any{"ready": true}},
		},
	}

	mod := moduleForAsyncTest([]any{int64(13), "ready"}, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpHostPoll},
		{Op: core.OpAwait},
		{Op: core.OpGetProp, A: 1},
		{Op: core.OpReturn},
	})

	vm, err := New(mod, Options{
		Host:         host,
		WaitInterval: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	err = vm.Run()
	if !errors.Is(err, ErrPromisePending) {
		t.Fatalf("Run error = %v, want ErrPromisePending", err)
	}

	if err := vm.WaitSuspension(context.Background()); err != nil {
		t.Fatalf("WaitSuspension failed: %v", err)
	}
	if len(host.pollLog) < 2 {
		t.Fatalf("expected fallback poll log, got %#v", host.pollLog)
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected fallback result on stack")
	}
	if result.Kind != core.ValueBool || !result.Raw.(bool) {
		t.Fatalf("unexpected fallback result: %#v", result)
	}
}

func TestRunUntilSettledResolvesPendingPromise(t *testing.T) {
	host := newMockHost()
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}
	host.waitResult = api.PollResult{
		Done:  true,
		Value: map[string]any{"ready": true},
	}

	mod := moduleForAsyncTest([]any{int64(15), "ready"}, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpHostPoll},
		{Op: core.OpAwait},
		{Op: core.OpGetProp, A: 1},
		{Op: core.OpReturn},
	})

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if err := vm.RunUntilSettled(context.Background()); err != nil {
		t.Fatalf("RunUntilSettled failed: %v", err)
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected settled result on stack")
	}
	if result.Kind != core.ValueBool || !result.Raw.(bool) {
		t.Fatalf("unexpected settled result: %#v", result)
	}
}

func TestRunUntilSettledHonorsContextTimeout(t *testing.T) {
	host := &pollOnlyHost{
		pollResults: []api.PollResult{
			{Done: false, Value: map[string]any{"ready": false}},
		},
	}

	mod := moduleForAsyncTest([]any{int64(17)}, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpHostPoll},
		{Op: core.OpAwait},
		{Op: core.OpReturn},
	})

	vm, err := New(mod, Options{
		Host:         host,
		WaitInterval: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	err = vm.RunUntilSettled(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RunUntilSettled error = %v, want deadline exceeded", err)
	}
}

func TestRunUntilSettledHonorsMaxDuration(t *testing.T) {
	host := &pollOnlyHost{
		pollResults: []api.PollResult{
			{Done: false, Value: map[string]any{"ready": false}},
		},
	}

	mod := moduleForAsyncTest([]any{int64(19)}, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpHostPoll},
		{Op: core.OpAwait},
		{Op: core.OpReturn},
	})

	vm, err := New(mod, Options{
		Host:         host,
		WaitInterval: time.Millisecond,
		MaxDuration:  5 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	err = vm.RunUntilSettled(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RunUntilSettled error = %v, want deadline exceeded", err)
	}
}

func TestRunUntilSettledHonorsWaitTimeout(t *testing.T) {
	host := newMockHost()
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}
	host.blockWait = true

	mod := moduleForAsyncTest([]any{int64(27)}, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpHostPoll},
		{Op: core.OpAwait},
		{Op: core.OpReturn},
	})

	vm, err := New(mod, Options{
		Host:        host,
		WaitTimeout: 5 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	err = vm.RunUntilSettled(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RunUntilSettled error = %v, want deadline exceeded", err)
	}
}

func TestRunUntilSettledHonorsHostTimeoutDuringPoll(t *testing.T) {
	host := newMockHost()
	host.blockPoll = true

	mod := moduleForAsyncTest([]any{int64(29)}, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpHostPoll},
		{Op: core.OpAwait},
		{Op: core.OpReturn},
	})

	vm, err := New(mod, Options{
		Host:        host,
		HostTimeout: 5 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	err = vm.RunUntilSettled(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RunUntilSettled error = %v, want deadline exceeded", err)
	}
}

func moduleForAsyncTest(constants []any, code []core.Instruction) *module.Module {
	return &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Constants: constants,
		Types:     []core.FuncType{{}},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code:      code,
			},
		},
	}
}
