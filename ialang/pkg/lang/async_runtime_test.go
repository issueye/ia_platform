package lang

import (
	"errors"
	rt "ialang/pkg/lang/runtime"
	"strings"
	"testing"
	"time"
)

func TestGoroutineRuntimeAwaitValue(t *testing.T) {
	runtime := rt.NewGoroutineRuntime()

	passthrough, err := runtime.AwaitValue("plain")
	if err != nil {
		t.Fatalf("AwaitValue(non-awaitable) unexpected error: %v", err)
	}
	if passthrough != "plain" {
		t.Fatalf("AwaitValue(non-awaitable) = %v, want plain", passthrough)
	}

	awaitable := runtime.Spawn(func() (rt.Value, error) {
		return float64(7), nil
	})
	got, err := runtime.AwaitValue(awaitable)
	if err != nil {
		t.Fatalf("AwaitValue(awaitable) unexpected error: %v", err)
	}
	if got != float64(7) {
		t.Fatalf("AwaitValue(awaitable) = %v, want 7", got)
	}
}

func TestGoroutineRuntimeAwaitValueError(t *testing.T) {
	runtime := rt.NewGoroutineRuntime()
	awaitable := runtime.Spawn(func() (rt.Value, error) {
		return nil, errors.New("boom")
	})
	_, err := runtime.AwaitValue(awaitable)
	if err == nil {
		t.Fatal("AwaitValue(awaitable) expected error, got nil")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("AwaitValue(awaitable) error = %v, want contains boom", err)
	}
}

func TestGoroutineRuntimeTaskTimeout(t *testing.T) {
	runtime := rt.NewGoroutineRuntimeWithOptions(rt.GoroutineRuntimeOptions{
		TaskTimeout: 10 * time.Millisecond,
	})
	awaitable := runtime.Spawn(func() (rt.Value, error) {
		time.Sleep(40 * time.Millisecond)
		return "done", nil
	})
	_, err := runtime.AwaitValue(awaitable)
	if err == nil {
		t.Fatal("AwaitValue(timeout task) expected error, got nil")
	}
	if !errors.Is(err, rt.ErrAsyncTaskTimeout) {
		t.Fatalf("AwaitValue(timeout task) error = %v, want ErrAsyncTaskTimeout", err)
	}
}

func TestGoroutineRuntimeAwaitTimeout(t *testing.T) {
	runtime := rt.NewGoroutineRuntimeWithOptions(rt.GoroutineRuntimeOptions{
		AwaitTimeout: 10 * time.Millisecond,
	})
	awaitable := runtime.Spawn(func() (rt.Value, error) {
		time.Sleep(40 * time.Millisecond)
		return "done", nil
	})
	_, err := runtime.AwaitValue(awaitable)
	if err == nil {
		t.Fatal("AwaitValue(timeout await) expected error, got nil")
	}
	if !errors.Is(err, rt.ErrAsyncAwaitTimeout) {
		t.Fatalf("AwaitValue(timeout await) error = %v, want ErrAsyncAwaitTimeout", err)
	}
}
