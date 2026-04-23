package runtime

import (
	"context"
	"errors"
	"math/rand"
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

func TestRunUntilSettledUsesCapabilityWaitTimeoutProfile(t *testing.T) {
	host := newMockHost()
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}
	host.blockWait = true

	mod := &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{"network", int64(31)},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"wait_timeout_ms": int64(5),
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpHostPoll},
					{Op: core.OpAwait},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{
		Host:        host,
		WaitTimeout: time.Hour,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	err = vm.RunUntilSettled(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RunUntilSettled error = %v, want deadline exceeded", err)
	}
}

func TestRunUntilSettledRetriesPollTimeout(t *testing.T) {
	host := newMockHost()
	host.pollDeadlineFailures = 1
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}
	host.waitResult = api.PollResult{
		Done:  true,
		Value: map[string]any{"ready": true},
	}

	mod := moduleForAsyncTest([]any{int64(33), "ready"}, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpHostPoll},
		{Op: core.OpAwait},
		{Op: core.OpGetProp, A: 1},
		{Op: core.OpReturn},
	})

	vm, err := New(mod, Options{
		Host:         host,
		HostTimeout:  5 * time.Millisecond,
		RetryCount:   1,
		RetryBackoff: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := vm.RunUntilSettled(context.Background()); err != nil {
		t.Fatalf("RunUntilSettled failed: %v", err)
	}
	if len(host.pollLog) < 2 {
		t.Fatalf("expected poll retry attempts, got %#v", host.pollLog)
	}
}

func TestRunUntilSettledRetriesWaitTimeoutWithCapabilityProfile(t *testing.T) {
	host := newMockHost()
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}
	host.waitDeadlineFailures = 1
	host.waitResult = api.PollResult{
		Done:  true,
		Value: map[string]any{"ready": true},
	}

	mod := &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{"network", int64(35), "ready"},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"wait_timeout_ms":  int64(5),
					"retry_count":      int64(1),
					"retry_backoff_ms": int64(1),
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpHostPoll},
					{Op: core.OpAwait},
					{Op: core.OpGetProp, A: 2},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := vm.RunUntilSettled(context.Background()); err != nil {
		t.Fatalf("RunUntilSettled failed: %v", err)
	}
	if len(host.waitLog) < 2 {
		t.Fatalf("expected wait retry attempts, got %#v", host.waitLog)
	}
}

func TestRunUntilSettledRetriesExplicitRetryablePollError(t *testing.T) {
	host := newMockHost()
	host.pollRetryableFailures = 1
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}
	host.waitResult = api.PollResult{
		Done:  true,
		Value: map[string]any{"ready": true},
	}

	mod := moduleForAsyncTest([]any{int64(41), "ready"}, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpHostPoll},
		{Op: core.OpAwait},
		{Op: core.OpGetProp, A: 1},
		{Op: core.OpReturn},
	})

	vm, err := New(mod, Options{
		Host:         host,
		RetryCount:   1,
		RetryBackoff: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := vm.RunUntilSettled(context.Background()); err != nil {
		t.Fatalf("RunUntilSettled failed: %v", err)
	}
	if len(host.pollLog) < 2 {
		t.Fatalf("expected retryable poll attempts, got %#v", host.pollLog)
	}
}

func TestRunUntilSettledRetriesExplicitRetryableWaitError(t *testing.T) {
	host := newMockHost()
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}
	host.waitRetryableFailures = 1
	host.waitResult = api.PollResult{
		Done:  true,
		Value: map[string]any{"ready": true},
	}

	mod := moduleForAsyncTest([]any{int64(43), "ready"}, []core.Instruction{
		{Op: core.OpConst, A: 0},
		{Op: core.OpHostPoll},
		{Op: core.OpAwait},
		{Op: core.OpGetProp, A: 1},
		{Op: core.OpReturn},
	})

	vm, err := New(mod, Options{
		Host:         host,
		RetryCount:   1,
		RetryBackoff: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := vm.RunUntilSettled(context.Background()); err != nil {
		t.Fatalf("RunUntilSettled failed: %v", err)
	}
	if len(host.waitLog) < 2 {
		t.Fatalf("expected retryable wait attempts, got %#v", host.waitLog)
	}
}

func TestComputeRetryBackoffUsesMultiplierAndCap(t *testing.T) {
	if got := computeRetryBackoff(0, 10*time.Millisecond, 2, 50*time.Millisecond); got != 10*time.Millisecond {
		t.Fatalf("attempt0 backoff = %v, want 10ms", got)
	}
	if got := computeRetryBackoff(1, 10*time.Millisecond, 2, 50*time.Millisecond); got != 20*time.Millisecond {
		t.Fatalf("attempt1 backoff = %v, want 20ms", got)
	}
	if got := computeRetryBackoff(3, 10*time.Millisecond, 2, 50*time.Millisecond); got != 50*time.Millisecond {
		t.Fatalf("attempt3 backoff = %v, want 50ms cap", got)
	}
}

func TestApplyRetryJitterUsesDeterministicSeed(t *testing.T) {
	base := 10 * time.Millisecond

	vmA := &VM{retryRand: rand.New(rand.NewSource(1))}
	vmB := &VM{retryRand: rand.New(rand.NewSource(1))}

	gotA := vmA.applyRetryJitter(base, 0.5, 0)
	gotB := vmB.applyRetryJitter(base, 0.5, 0)
	if gotA != gotB {
		t.Fatalf("jittered backoff mismatch for same seed: %v vs %v", gotA, gotB)
	}
	if gotA == base {
		t.Fatalf("expected jitter to change backoff, got %v", gotA)
	}
	if gotA < 5*time.Millisecond || gotA > 15*time.Millisecond {
		t.Fatalf("jittered backoff = %v, want within [5ms,15ms]", gotA)
	}

	if got := vmA.applyRetryJitter(base, 0, 0); got != base {
		t.Fatalf("zero jitter backoff = %v, want %v", got, base)
	}

	vmC := &VM{retryRand: rand.New(rand.NewSource(1))}
	if got := vmC.applyRetryJitter(base, 1, 5*time.Millisecond); got != 5*time.Millisecond {
		t.Fatalf("capped jitter backoff = %v, want 5ms", got)
	}
}

func TestNextRetryBackoffPrefersRetryHintOverComputedBackoff(t *testing.T) {
	vm := &VM{retryRand: rand.New(rand.NewSource(1))}

	err := api.MarkRetryableAfter(errors.New("retry later"), 3*time.Second)
	if got := vm.nextRetryBackoff(err, 0, 10*time.Millisecond, 5*time.Second, 2, 0.5); got != 3*time.Second {
		t.Fatalf("hinted backoff = %v, want 3s", got)
	}

	err = api.MarkRetryableAfter(errors.New("retry later"), 10*time.Second)
	if got := vm.nextRetryBackoff(err, 0, 10*time.Millisecond, 5*time.Second, 2, 0.5); got != 5*time.Second {
		t.Fatalf("capped hinted backoff = %v, want 5s", got)
	}
}

func TestHostPollPendingPromisePreservesCapabilityRetryProfile(t *testing.T) {
	host := newMockHost()
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}

	mod := &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{"network", int64(39), "ready"},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"host_timeout_ms":      int64(5),
					"retry_count":          int64(1),
					"retry_backoff_ms":     int64(1),
					"retry_max_elapsed_ms": int64(7),
					"retry_jitter":         float64(0.5),
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpHostPoll},
					{Op: core.OpAwait},
					{Op: core.OpGetProp, A: 2},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	vm.retryRand = rand.New(rand.NewSource(1))

	err = vm.Run()
	if !errors.Is(err, ErrPromisePending) {
		t.Fatalf("Run error = %v, want ErrPromisePending", err)
	}

	suspension := vm.SuspensionState()
	if suspension == nil {
		t.Fatal("expected suspension state")
	}
	state, ok := suspension.AwaitValue.Raw.(*promiseState)
	if !ok || state == nil {
		t.Fatal("expected pending promise state")
	}
	if state.RetryJitter != 0.5 {
		t.Fatalf("retry jitter = %v, want 0.5", state.RetryJitter)
	}
	if state.RetryMaxElapsed != 7*time.Millisecond {
		t.Fatalf("retry max elapsed = %v, want 7ms", state.RetryMaxElapsed)
	}

	host.pollDeadlineFailures = 1
	host.pollResult = api.PollResult{
		Done:  true,
		Value: map[string]any{"ready": true},
	}
	if err := vm.ResumeSuspension(); err != nil {
		t.Fatalf("ResumeSuspension failed: %v", err)
	}
	if len(host.pollLog) < 3 {
		t.Fatalf("expected resume retry attempts, got %#v", host.pollLog)
	}

	result, ok := vm.PopResult()
	if !ok {
		t.Fatal("expected resumed result on stack")
	}
	if result.Kind != core.ValueBool || !result.Raw.(bool) {
		t.Fatalf("unexpected resumed result: %#v", result)
	}
}

func TestRunUntilSettledUsesCapabilityRetryBackoffProfile(t *testing.T) {
	host := newMockHost()
	host.pollDeadlineFailures = 2
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}
	host.waitResult = api.PollResult{
		Done:  true,
		Value: map[string]any{"ready": true},
	}

	mod := &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{"network", int64(37), "ready"},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"host_timeout_ms":      int64(5),
					"retry_count":          int64(2),
					"retry_backoff_ms":     int64(1),
					"retry_backoff_max_ms": int64(2),
					"retry_multiplier":     float64(2),
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpHostPoll},
					{Op: core.OpAwait},
					{Op: core.OpGetProp, A: 2},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := vm.RunUntilSettled(context.Background()); err != nil {
		t.Fatalf("RunUntilSettled failed: %v", err)
	}
	if len(host.pollLog) < 3 {
		t.Fatalf("expected capability retry attempts, got %#v", host.pollLog)
	}
}

func TestRunUntilSettledCanDisableCapabilityRetryProfile(t *testing.T) {
	host := newMockHost()
	host.pollDeadlineFailures = 1
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}

	mod := &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{"network", int64(37), "ready"},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"retry_enabled": false,
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpHostPoll},
					{Op: core.OpAwait},
					{Op: core.OpGetProp, A: 2},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{
		Host:         host,
		HostTimeout:  5 * time.Millisecond,
		RetryCount:   2,
		RetryBackoff: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	err = vm.RunUntilSettled(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RunUntilSettled error = %v, want deadline exceeded", err)
	}
	if len(host.pollLog) != 1 {
		t.Fatalf("expected retry-disabled capability to poll once, got %#v", host.pollLog)
	}
}

func TestRunUntilSettledCanDisableCapabilityPollRetryProfile(t *testing.T) {
	host := newMockHost()
	host.pollDeadlineFailures = 1
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}

	mod := &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{"network", int64(37), "ready"},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"retry_poll_enabled": false,
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpHostPoll},
					{Op: core.OpAwait},
					{Op: core.OpGetProp, A: 2},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{
		Host:         host,
		HostTimeout:  5 * time.Millisecond,
		RetryCount:   2,
		RetryBackoff: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	err = vm.RunUntilSettled(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RunUntilSettled error = %v, want deadline exceeded", err)
	}
	if len(host.pollLog) != 1 {
		t.Fatalf("expected poll-retry-disabled capability to poll once, got %#v", host.pollLog)
	}
}

func TestRunUntilSettledDisablesWaitRetryWhenPollRetryDisabledByDefault(t *testing.T) {
	host := newMockHost()
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}
	host.waitDeadlineFailures = 1
	host.waitResult = api.PollResult{
		Done:  true,
		Value: map[string]any{"ready": true},
	}

	mod := &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{"network", int64(35), "ready"},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"wait_timeout_ms":    int64(5),
					"retry_count":        int64(1),
					"retry_backoff_ms":   int64(1),
					"retry_poll_enabled": false,
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpHostPoll},
					{Op: core.OpAwait},
					{Op: core.OpGetProp, A: 2},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	err = vm.RunUntilSettled(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RunUntilSettled error = %v, want deadline exceeded", err)
	}
	if len(host.waitLog) != 1 {
		t.Fatalf("expected wait retry to remain disabled by default compatibility, got %#v", host.waitLog)
	}
}

func TestRunUntilSettledCanDisableOnlyCapabilityWaitRetryProfile(t *testing.T) {
	host := newMockHost()
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}
	host.waitDeadlineFailures = 1
	host.waitResult = api.PollResult{
		Done:  true,
		Value: map[string]any{"ready": true},
	}

	mod := &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{"network", int64(35), "ready"},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"wait_timeout_ms":    int64(5),
					"retry_count":        int64(1),
					"retry_backoff_ms":   int64(1),
					"retry_wait_enabled": false,
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpHostPoll},
					{Op: core.OpAwait},
					{Op: core.OpGetProp, A: 2},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	err = vm.RunUntilSettled(context.Background())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("RunUntilSettled error = %v, want deadline exceeded", err)
	}
	if len(host.waitLog) != 1 {
		t.Fatalf("expected wait retry-disabled capability to wait once, got %#v", host.waitLog)
	}
}

func TestRunUntilSettledKeepsWaitRetryWhenPollRetryDisabledExplicitlyEnabled(t *testing.T) {
	host := newMockHost()
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}
	host.waitDeadlineFailures = 1
	host.waitResult = api.PollResult{
		Done:  true,
		Value: map[string]any{"ready": true},
	}

	mod := &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{"network", int64(35), "ready"},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"wait_timeout_ms":    int64(5),
					"retry_count":        int64(1),
					"retry_backoff_ms":   int64(1),
					"retry_poll_enabled": false,
					"retry_wait_enabled": true,
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpHostPoll},
					{Op: core.OpAwait},
					{Op: core.OpGetProp, A: 2},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{Host: host})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := vm.RunUntilSettled(context.Background()); err != nil {
		t.Fatalf("RunUntilSettled failed: %v", err)
	}
	if len(host.waitLog) != 2 {
		t.Fatalf("expected explicit wait retry to remain enabled, got %#v", host.waitLog)
	}
}

func TestRunUntilSettledKeepsPollRetryWhenCallRetryIsDisabled(t *testing.T) {
	host := newMockHost()
	host.pollDeadlineFailures = 2
	host.pollResult = api.PollResult{
		Done:  false,
		Value: map[string]any{"ready": false},
	}
	host.waitResult = api.PollResult{
		Done:  true,
		Value: map[string]any{"ready": true},
	}

	mod := &module.Module{
		Magic:     "IAVM",
		Version:   1,
		Target:    "ialang",
		Types:     []core.FuncType{{}},
		Constants: []any{"network", int64(37), "ready"},
		Capabilities: []module.CapabilityDecl{
			{
				Kind: module.CapabilityNetwork,
				Config: map[string]any{
					"retry_call_enabled": false,
					"retry_count":        int64(2),
					"retry_backoff_ms":   int64(1),
				},
			},
		},
		Functions: []module.Function{
			{
				Name:      "entry",
				TypeIndex: 0,
				Code: []core.Instruction{
					{Op: core.OpImportCap, A: 0},
					{Op: core.OpConst, A: 1},
					{Op: core.OpHostPoll},
					{Op: core.OpAwait},
					{Op: core.OpGetProp, A: 2},
					{Op: core.OpReturn},
				},
			},
		},
	}

	vm, err := New(mod, Options{
		Host:         host,
		HostTimeout:  5 * time.Millisecond,
		RetryCount:   2,
		RetryBackoff: time.Millisecond,
	})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if err := vm.RunUntilSettled(context.Background()); err != nil {
		t.Fatalf("RunUntilSettled failed: %v", err)
	}
	if len(host.pollLog) < 3 {
		t.Fatalf("expected poll retry attempts to remain enabled, got %#v", host.pollLog)
	}
}

func TestRetryCallLikeStopsWhenElapsedBudgetIsExhausted(t *testing.T) {
	vm := &VM{}
	attempts := 0

	_, err := vm.retryCallLike(context.Background(), true, 3, 20*time.Millisecond, 0, 5*time.Millisecond, 2, 0, func() (api.CallResult, error) {
		attempts++
		return api.CallResult{}, api.MarkRetryable(errors.New("retry later"))
	})
	if err == nil || !api.IsRetryableError(err) {
		t.Fatalf("expected retryable error once budget is exhausted, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
	}
}

func TestRetryPollLikeStopsWhenElapsedBudgetIsExhausted(t *testing.T) {
	vm := &VM{}
	attempts := 0

	_, err := vm.retryPollLike(context.Background(), true, 3, 20*time.Millisecond, 0, 5*time.Millisecond, 2, 0, func() (api.PollResult, error) {
		attempts++
		return api.PollResult{}, api.MarkRetryable(errors.New("retry later"))
	})
	if err == nil || !api.IsRetryableError(err) {
		t.Fatalf("expected retryable error once budget is exhausted, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1", attempts)
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
