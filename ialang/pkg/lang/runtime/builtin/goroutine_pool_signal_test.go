package builtin

import (
	"os"
	"strings"
	"testing"

	rt "ialang/pkg/lang/runtime"
)

type fakeSignal string

func (s fakeSignal) Signal() {}

func (s fakeSignal) String() string {
	return string(s)
}

func TestGoroutinePoolModuleSubmitRetryAndCreatePool(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "pool")
	defer func() {
		_, _ = callNativeWithError(mod, "shutdown", float64(100))
	}()

	result := awaitValue(t, callNative(t, mod, "submit", NativeFunction(func(args []Value) (Value, error) {
		return "done", nil
	})))
	if result != "done" {
		t.Fatalf("pool.submit result = %#v, want done", result)
	}

	attempts := 0
	retryResult := awaitValue(t, callNative(t, mod, "submitWithRetry", NativeFunction(func(args []Value) (Value, error) {
		attempts++
		if attempts < 2 {
			return nil, &testError{msg: "retry me"}
		}
		return float64(42), nil
	}), float64(2)))
	if got := retryResult.(float64); got != 42 {
		t.Fatalf("pool.submitWithRetry result = %#v, want 42", retryResult)
	}
	if attempts != 2 {
		t.Fatalf("pool.submitWithRetry attempts = %d, want 2", attempts)
	}

	afterStats := mustRuntimeObject(t, callNative(t, mod, "getStats"), "pool.getStats after")
	if afterStats["totalPools"].(float64) < 1 {
		t.Fatalf("pool totalPools after = %#v, want >= 1", afterStats["totalPools"])
	}
	if afterStats["totalWorkers"].(float64) < 1 {
		t.Fatalf("pool totalWorkers after = %#v, want >= 1", afterStats["totalWorkers"])
	}

	poolObj := mustRuntimeObject(t, callNative(t, mod, "createPool", Object{
		"minWorkers": float64(1),
		"maxWorkers": float64(2),
		"queueSize":  float64(4),
	}), "pool.createPool return")

	poolTask := callNative(t, poolObj, "submit", NativeFunction(func(args []Value) (Value, error) {
		return "local", nil
	}))
	if got := awaitValue(t, poolTask); got != "local" {
		t.Fatalf("created pool submit result = %#v, want local", got)
	}

	poolStats := mustRuntimeObject(t, callNative(t, poolObj, "getStats"), "pool object getStats")
	if poolStats["totalWorkers"].(float64) < 1 {
		t.Fatalf("created pool totalWorkers = %#v, want >= 1", poolStats["totalWorkers"])
	}
	if poolStats["completedTasks"].(float64) < 1 {
		t.Fatalf("created pool completedTasks = %#v, want >= 1", poolStats["completedTasks"])
	}

	shutdown := callNative(t, poolObj, "shutdown", float64(100))
	if shutdown != true {
		t.Fatalf("created pool shutdown = %#v, want true", shutdown)
	}

	if _, err := callNativeWithError(mod, "submit"); err == nil || !strings.Contains(err.Error(), "expects at least 1 arg") {
		t.Fatalf("pool.submit() error = %v, want arity error", err)
	}
	if _, err := callNativeWithError(mod, "submit", "not-a-function"); err == nil || !strings.Contains(err.Error(), "expects function") {
		t.Fatalf("pool.submit(invalid) error = %v, want type error", err)
	}
	if _, err := callNativeWithError(mod, "submitWithRetry", "not-a-function"); err == nil || !strings.Contains(err.Error(), "expects function") {
		t.Fatalf("pool.submitWithRetry(invalid) error = %v, want type error", err)
	}
}

func TestSignalParsingSubscriptionAndModule(t *testing.T) {
	if signals, err := parseOptionalSignalList("signal.ignore", nil); err != nil || signals != nil {
		t.Fatalf("parseOptionalSignalList(nil) = %#v, %v; want nil, nil", signals, err)
	}
	if _, err := parseOptionalSignalList("signal.ignore", []Value{"SIGINT", "SIGTERM"}); err == nil || !strings.Contains(err.Error(), "expects 0-1 args") {
		t.Fatalf("parseOptionalSignalList(too many) error = %v, want arity error", err)
	}
	if _, err := parseSignalList(float64(1)); err == nil || !strings.Contains(err.Error(), "expects string or array") {
		t.Fatalf("parseSignalList(invalid) error = %v, want type error", err)
	}

	signals, err := parseSignalList(Array{" SIGINT ", "SIGTERM"})
	if err != nil {
		t.Fatalf("parseSignalList(array) error: %v", err)
	}
	if len(signals) != 2 {
		t.Fatalf("parseSignalList(array) len = %d, want 2", len(signals))
	}
	if signals[0] != os.Interrupt {
		t.Fatalf("parseSignalList first signal = %#v, want os.Interrupt", signals[0])
	}
	if name := signalName(signals[1]); name != "SIGTERM" {
		t.Fatalf("signalName(SIGTERM) = %q, want SIGTERM", name)
	}
	if _, err := parseSignalName("unsupported"); err == nil || !strings.Contains(err.Error(), "unsupported signal") {
		t.Fatalf("parseSignalName(unsupported) error = %v, want unsupported signal", err)
	}
	if name := signalName(fakeSignal("CUSTOM")); name != "CUSTOM" {
		t.Fatalf("signalName(fake) = %q, want CUSTOM", name)
	}

	ch := make(chan os.Signal, 1)
	sub := newSignalSubscriptionObject(ch)
	ch <- os.Interrupt
	if got := callNative(t, sub, "recv"); got != "SIGINT" {
		t.Fatalf("signal.subscription.recv = %#v, want SIGINT", got)
	}
	if stopped := callNative(t, sub, "stop"); stopped != true {
		t.Fatalf("signal.subscription.stop = %#v, want true", stopped)
	}
	if _, err := callNativeWithError(sub, "recv", "extra"); err == nil || !strings.Contains(err.Error(), "expects 0 args") {
		t.Fatalf("signal.subscription.recv(extra) error = %v, want arity error", err)
	}
	if _, err := callNativeWithError(sub, "stop", "extra"); err == nil || !strings.Contains(err.Error(), "expects 0 args") {
		t.Fatalf("signal.subscription.stop(extra) error = %v, want arity error", err)
	}

	modules := DefaultModules(rt.NewGoroutineRuntime())
	mod := mustModuleObject(t, modules, "signal")
	if mod["SIGINT"] != "SIGINT" {
		t.Fatalf("signal.SIGINT = %#v, want SIGINT", mod["SIGINT"])
	}
	if _, ok := mod["signal"].(Object); !ok {
		t.Fatalf("signal namespace export type = %T, want Object", mod["signal"])
	}

	notifySub := mustRuntimeObject(t, callNative(t, mod, "notify", Array{"SIGINT"}), "signal.notify return")
	if stopped := callNative(t, notifySub, "stop"); stopped != true {
		t.Fatalf("signal.notify.stop = %#v, want true", stopped)
	}
	if ignored := callNative(t, mod, "ignore", Array{"SIGINT"}); ignored != true {
		t.Fatalf("signal.ignore = %#v, want true", ignored)
	}
	if reset := callNative(t, mod, "reset"); reset != true {
		t.Fatalf("signal.reset = %#v, want true", reset)
	}
	if _, err := callNativeWithError(mod, "notify"); err == nil || !strings.Contains(err.Error(), "expects 1 arg") {
		t.Fatalf("signal.notify() error = %v, want arity error", err)
	}
	if _, err := callNativeWithError(mod, "ignore", float64(1)); err == nil || !strings.Contains(err.Error(), "expects string or array") {
		t.Fatalf("signal.ignore(invalid) error = %v, want type error", err)
	}
}
