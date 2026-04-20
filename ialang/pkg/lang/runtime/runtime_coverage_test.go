package runtime

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestEnvironmentDefineAndGet(t *testing.T) {
	env := NewEnvironment(nil)
	env.Define("x", float64(42))
	v, ok := env.Get("x")
	if !ok {
		t.Fatal("expected to find x")
	}
	if v != float64(42) {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestEnvironmentGetFromParent(t *testing.T) {
	parent := NewEnvironment(nil)
	parent.Define("x", "hello")
	child := NewEnvironment(parent)
	v, ok := child.Get("x")
	if !ok {
		t.Fatal("expected to find x in parent")
	}
	if v != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}
}

func TestEnvironmentSet(t *testing.T) {
	env := NewEnvironment(nil)
	env.Define("x", float64(1))
	ok := env.Set("x", float64(2))
	if !ok {
		t.Fatal("expected Set to succeed")
	}
	v, _ := env.Get("x")
	if v != float64(2) {
		t.Fatalf("expected 2, got %v", v)
	}
}

func TestEnvironmentSetNotFound(t *testing.T) {
	env := NewEnvironment(nil)
	ok := env.Set("missing", float64(1))
	if ok {
		t.Fatal("expected Set to fail for missing variable")
	}
}

func TestEnvironmentGetNotFound(t *testing.T) {
	env := NewEnvironment(nil)
	_, ok := env.Get("missing")
	if ok {
		t.Fatal("expected Get to fail for missing variable")
	}
}

func TestEnvironmentSized(t *testing.T) {
	env := NewEnvironmentSized(nil, 10)
	if env == nil {
		t.Fatal("expected non-nil environment")
	}
	env2 := NewEnvironmentSized(nil, -1)
	if env2 == nil {
		t.Fatal("expected non-nil environment for negative size")
	}
}

func TestIsTruthy(t *testing.T) {
	cases := []struct {
		val    Value
		truthy bool
	}{
		{nil, false},
		{false, false},
		{float64(0), false},
		{float64(1), true},
		{"", false},
		{"hello", true},
		{true, true},
		{Array{}, true},
		{Object{}, true},
	}
	for _, tc := range cases {
		if isTruthy(tc.val) != tc.truthy {
			t.Fatalf("isTruthy(%v) = %v, want %v", tc.val, !tc.truthy, tc.truthy)
		}
	}
}

func TestValueEqual(t *testing.T) {
	cases := []struct {
		a, b Value
		want bool
	}{
		{nil, nil, true},
		{nil, float64(0), false},
		{true, true, true},
		{true, false, false},
		{float64(42), float64(42), true},
		{float64(42), float64(43), false},
		{"hello", "hello", true},
		{"hello", "world", false},
		{float64(1), true, false},
	}
	for _, tc := range cases {
		if valueEqual(tc.a, tc.b) != tc.want {
			t.Fatalf("valueEqual(%v, %v) = %v, want %v", tc.a, tc.b, !tc.want, tc.want)
		}
	}
}

func TestToString(t *testing.T) {
	cases := []struct {
		val  Value
		want string
	}{
		{nil, "nil"},
		{true, "true"},
		{false, "false"},
		{float64(42), "42"},
		{float64(3.14), "3.14"},
		{"hello", "hello"},
	}
	for _, tc := range cases {
		got := toString(tc.val)
		if got != tc.want {
			t.Fatalf("toString(%v) = %q, want %q", tc.val, got, tc.want)
		}
	}
}

func TestToStringPromise(t *testing.T) {
	p := ResolvedPromise(float64(42))
	time.Sleep(10 * time.Millisecond)
	s := toString(p)
	if s != "[Promise resolved]" {
		t.Fatalf("expected [Promise resolved], got %q", s)
	}
}

func TestToStringUserFunction(t *testing.T) {
	fn := &UserFunction{Name: "test"}
	s := toString(fn)
	if s != "[Function test]" {
		t.Fatalf("expected [Function test], got %q", s)
	}
	fn2 := &UserFunction{Name: "test", Async: true}
	s2 := toString(fn2)
	if s2 != "[AsyncFunction test]" {
		t.Fatalf("expected [AsyncFunction test], got %q", s2)
	}
	fn3 := &UserFunction{}
	s3 := toString(fn3)
	if s3 != "[Function]" {
		t.Fatalf("expected [Function], got %q", s3)
	}
}

func TestToStringClassValue(t *testing.T) {
	cls := &ClassValue{Name: "Foo"}
	s := toString(cls)
	if s != "[Class Foo]" {
		t.Fatalf("expected [Class Foo], got %q", s)
	}
}

func TestToStringInstanceValue(t *testing.T) {
	inst := &InstanceValue{Class: &ClassValue{Name: "Foo"}}
	s := toString(inst)
	if s != "[Instance Foo]" {
		t.Fatalf("expected [Instance Foo], got %q", s)
	}
	inst2 := &InstanceValue{}
	s2 := toString(inst2)
	if s2 != "[Instance]" {
		t.Fatalf("expected [Instance], got %q", s2)
	}
}

func TestToStringBoundMethod(t *testing.T) {
	bm := &BoundMethod{Method: &UserFunction{Name: "fn"}}
	s := toString(bm)
	if s != "[BoundMethod fn]" {
		t.Fatalf("expected [BoundMethod fn], got %q", s)
	}
	bm2 := &BoundMethod{}
	s2 := toString(bm2)
	if s2 != "[BoundMethod]" {
		t.Fatalf("expected [BoundMethod], got %q", s2)
	}
}

func TestNewRuntimeError(t *testing.T) {
	err := NewRuntimeError("something broke")
	if err.Type != ErrorTypeRuntime {
		t.Fatalf("expected RuntimeError, got %s", err.Type)
	}
	if err.Message != "something broke" {
		t.Fatalf("unexpected message: %s", err.Message)
	}
	if err.Error() != "RuntimeError: something broke" {
		t.Fatalf("unexpected Error(): %s", err.Error())
	}
}

func TestRuntimeErrorWithCause(t *testing.T) {
	cause := errors.New("root cause")
	err := NewRuntimeError("wrapped", WithCause(cause))
	if err.Cause != cause {
		t.Fatal("expected cause to be set")
	}
	if err.Unwrap() != cause {
		t.Fatal("expected Unwrap to return cause")
	}
}

func TestRuntimeErrorWithContext(t *testing.T) {
	err := NewRuntimeError("test", WithContext("mod.ia", 10, 5, 3))
	if err.ModulePath != "mod.ia" {
		t.Fatalf("expected module path 'mod.ia', got %q", err.ModulePath)
	}
	if err.IP != 10 {
		t.Fatalf("expected IP 10, got %d", err.IP)
	}
}

func TestIaErrorToObject(t *testing.T) {
	err := NewRuntimeError("test")
	obj := err.ToObject()
	if obj["name"] != "RuntimeError" {
		t.Fatalf("expected name RuntimeError, got %v", obj["name"])
	}
	if obj["message"] != "test" {
		t.Fatalf("expected message 'test', got %v", obj["message"])
	}
}

func TestIaErrorToObjectWithFields(t *testing.T) {
	err := NewRuntimeError("test", WithContext("mod.ia", 10, 5, 3))
	obj := err.ToObject()
	if obj["module"] != "mod.ia" {
		t.Fatalf("expected module, got %v", obj["module"])
	}
	if obj["ip"] != float64(10) {
		t.Fatalf("expected ip 10, got %v", obj["ip"])
	}
}

func TestIaErrorToObjectWithCause(t *testing.T) {
	cause := errors.New("inner")
	err := NewRuntimeError("outer", WithCause(cause))
	obj := err.ToObject()
	if obj["cause"] != "inner" {
		t.Fatalf("expected cause 'inner', got %v", obj["cause"])
	}
}

func TestNewTimeoutError(t *testing.T) {
	err := NewTimeoutError("timed out", true)
	if err.Type != ErrorTypeTimeout {
		t.Fatalf("expected TimeoutError, got %s", err.Type)
	}
	if !err.Retryable {
		t.Fatal("expected retryable")
	}
}

func TestNewSandboxError(t *testing.T) {
	err := NewSandboxError("steps", "100", "200")
	if err.Type != ErrorTypeSandbox {
		t.Fatalf("expected SandboxError, got %s", err.Type)
	}
}

func TestNewImportError(t *testing.T) {
	err := NewImportError("not found", "mod.ia")
	if err.Type != ErrorTypeImport {
		t.Fatalf("expected ImportError, got %s", err.Type)
	}
	if err.ModulePath != "mod.ia" {
		t.Fatalf("expected module path 'mod.ia', got %q", err.ModulePath)
	}
}

func TestNewTypeError(t *testing.T) {
	err := NewTypeError("bad type")
	if err.Type != ErrorTypeType {
		t.Fatalf("expected TypeError, got %s", err.Type)
	}
}

func TestNewReferenceError(t *testing.T) {
	err := NewReferenceError("x")
	if err.Type != ErrorTypeReference {
		t.Fatalf("expected ReferenceError, got %s", err.Type)
	}
	if err.Message != "x is not defined" {
		t.Fatalf("unexpected message: %s", err.Message)
	}
}

func TestIsTimeoutError(t *testing.T) {
	err := NewTimeoutError("timeout", false)
	if !IsTimeoutError(err) {
		t.Fatal("expected IsTimeoutError to return true")
	}
	if IsTimeoutError(NewRuntimeError("test")) {
		t.Fatal("expected IsTimeoutError to return false for runtime error")
	}
	if IsTimeoutError(nil) {
		t.Fatal("expected IsTimeoutError(nil) to return false")
	}
}

func TestIsSandboxError(t *testing.T) {
	err := NewSandboxError("violation", "100", "200")
	if !IsSandboxError(err) {
		t.Fatal("expected IsSandboxError to return true for IaError")
	}
	sandboxErr := &SandboxError{Violation: "steps", Limit: "100", Current: "200"}
	if !IsSandboxError(sandboxErr) {
		t.Fatal("expected IsSandboxError to return true for SandboxError")
	}
	if IsSandboxError(NewRuntimeError("test")) {
		t.Fatal("expected IsSandboxError to return false for runtime error")
	}
	if IsSandboxError(nil) {
		t.Fatal("expected IsSandboxError(nil) to return false")
	}
}

func TestIsImportError(t *testing.T) {
	err := NewImportError("not found", "mod")
	if !IsImportError(err) {
		t.Fatal("expected IsImportError to return true")
	}
	if IsImportError(nil) {
		t.Fatal("expected IsImportError(nil) to return false")
	}
}

func TestSandboxPolicyDefault(t *testing.T) {
	p := DefaultSandboxPolicy()
	if p.MaxSteps != 100000 {
		t.Fatalf("expected 100000 steps, got %d", p.MaxSteps)
	}
	if p.AllowImport != true {
		t.Fatal("expected AllowImport true")
	}
	if p.AllowFS != false {
		t.Fatal("expected AllowFS false")
	}
}

func TestSandboxPolicyPermissive(t *testing.T) {
	p := PermissiveSandboxPolicy()
	if p.MaxSteps != 0 {
		t.Fatalf("expected 0 steps, got %d", p.MaxSteps)
	}
	if !p.AllowFS || !p.AllowNetwork || !p.AllowProcess {
		t.Fatal("expected all permissions true")
	}
}

func TestSandboxPolicyIsModuleAllowed(t *testing.T) {
	p := DefaultSandboxPolicy()
	if !p.IsModuleAllowed("json") {
		t.Fatal("expected json to be allowed with empty whitelist")
	}
	p.AllowImport = false
	if p.IsModuleAllowed("json") {
		t.Fatal("expected json to be disallowed when import disabled")
	}
}

func TestSandboxPolicyAllowedModulesWhitelist(t *testing.T) {
	p := DefaultSandboxPolicy()
	p.AllowedModules = map[string]bool{"json": true, "fs": false}
	if !p.IsModuleAllowed("json") {
		t.Fatal("expected json to be allowed")
	}
	if p.IsModuleAllowed("http") {
		t.Fatal("expected http to be disallowed (not in whitelist)")
	}
	if p.IsModuleAllowed("fs") {
		t.Fatal("expected fs to be disallowed (explicitly false)")
	}
}

func TestSandboxPolicyAddAllowedModule(t *testing.T) {
	p := &SandboxPolicy{AllowImport: true}
	p.AddAllowedModule("json")
	if !p.IsModuleAllowed("json") {
		t.Fatal("expected json to be allowed after AddAllowedModule")
	}
}

func TestSandboxErrorType(t *testing.T) {
	err := &SandboxError{Violation: "steps", Limit: "100", Current: "200"}
	s := err.Error()
	if s != "sandbox violation: steps (limit: 100, current: 200)" {
		t.Fatalf("unexpected error string: %s", s)
	}
}

func TestStepCounter(t *testing.T) {
	sc := NewStepCounter(3)
	if sc.Count() != 0 {
		t.Fatalf("expected count 0, got %d", sc.Count())
	}
	if err := sc.Increment(); err != nil {
		t.Fatal("unexpected error on increment")
	}
	if err := sc.Increment(); err != nil {
		t.Fatal("unexpected error on increment")
	}
	if sc.Count() != 2 {
		t.Fatalf("expected count 2, got %d", sc.Count())
	}
}

func TestStepCounterLimit(t *testing.T) {
	sc := NewStepCounter(2)
	sc.Increment()
	sc.Increment()
	err := sc.Increment()
	if err == nil {
		t.Fatal("expected error on exceeding limit")
	}
}

func TestStepCounterZeroLimit(t *testing.T) {
	sc := NewStepCounter(0)
	for i := 0; i < 100; i++ {
		if err := sc.Increment(); err != nil {
			t.Fatal("unexpected error with zero limit")
		}
	}
}

func TestPromiseResolved(t *testing.T) {
	p := ResolvedPromise(float64(42))
	if !p.IsDone() {
		t.Fatal("expected resolved promise to be done")
	}
	v, err := p.Await()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != float64(42) {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestPromiseAsync(t *testing.T) {
	p := NewPromise(func() (Value, error) {
		return "hello", nil
	})
	time.Sleep(50 * time.Millisecond)
	if !p.IsDone() {
		t.Fatal("expected promise to be done")
	}
	v, err := p.Await()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}
}

func TestPromiseAsyncError(t *testing.T) {
	p := NewPromise(func() (Value, error) {
		return nil, errors.New("fail")
	})
	time.Sleep(50 * time.Millisecond)
	_, err := p.Await()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPromiseAwaitContext(t *testing.T) {
	p := ResolvedPromise(float64(99))
	ctx := context.Background()
	v, err := p.AwaitContext(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != float64(99) {
		t.Fatalf("expected 99, got %v", v)
	}
}

func TestPromiseAwaitContextTimeout(t *testing.T) {
	p := NewPromise(func() (Value, error) {
		time.Sleep(5 * time.Second)
		return nil, nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := p.AwaitContext(ctx)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestPromiseAll(t *testing.T) {
	p1 := ResolvedPromise(float64(1))
	p2 := ResolvedPromise(float64(2))
	all := PromiseAll([]Awaitable{p1, p2})
	v, err := all.Await()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	arr, ok := v.(Array)
	if !ok {
		t.Fatalf("expected Array, got %T", v)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 results, got %d", len(arr))
	}
}

func TestPromiseAllReject(t *testing.T) {
	p1 := ResolvedPromise(float64(1))
	p2 := NewPromise(func() (Value, error) {
		return nil, errors.New("fail")
	})
	all := PromiseAll([]Awaitable{p1, p2})
	_, err := all.Await()
	if err == nil {
		t.Fatal("expected error from rejected promise")
	}
}

func TestPromiseRace(t *testing.T) {
	p1 := NewPromise(func() (Value, error) {
		time.Sleep(100 * time.Millisecond)
		return "slow", nil
	})
	p2 := ResolvedPromise("fast")
	race := PromiseRace([]Awaitable{p1, p2})
	v, err := race.Await()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "fast" {
		t.Fatalf("expected 'fast', got %v", v)
	}
}

func TestPromiseAllSettled(t *testing.T) {
	p1 := ResolvedPromise(float64(1))
	p2 := NewPromise(func() (Value, error) {
		return nil, errors.New("fail")
	})
	settled := PromiseAllSettled([]Awaitable{p1, p2})
	v, err := settled.Await()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	arr, ok := v.(Array)
	if !ok {
		t.Fatalf("expected Array, got %T", v)
	}
	if len(arr) != 2 {
		t.Fatalf("expected 2 results, got %d", len(arr))
	}
}

func TestNewGoroutineRuntime(t *testing.T) {
	rt := NewGoroutineRuntime()
	if rt.Name() != "goroutine" {
		t.Fatalf("expected name 'goroutine', got %q", rt.Name())
	}
}

func TestGoroutineRuntimeSpawn(t *testing.T) {
	rt := NewGoroutineRuntime()
	p := rt.Spawn(func() (Value, error) {
		return float64(42), nil
	})
	v, err := p.Await()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != float64(42) {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestGoroutineRuntimeAwaitValue(t *testing.T) {
	rt := NewGoroutineRuntime()
	p := ResolvedPromise("hello")
	v, err := rt.AwaitValue(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}
}

func TestGoroutineRuntimeAwaitNonAwaitable(t *testing.T) {
	rt := NewGoroutineRuntime()
	v, err := rt.AwaitValue(float64(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != float64(42) {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestNewAsyncRuntimeErrorValue(t *testing.T) {
	obj := NewAsyncRuntimeErrorValue("TimeoutError", "TIMEOUT", "timeout", "test", "goroutine", true)
	if obj["name"] != "TimeoutError" {
		t.Fatalf("unexpected name: %v", obj["name"])
	}
	if obj["retryable"] != true {
		t.Fatal("expected retryable true")
	}
}
