package pool

import (
	"strings"
	"testing"
	"time"

	rttypes "ialang/pkg/lang/runtime/types"
)

func TestGoroutinePoolStartTwice(t *testing.T) {
	opts := DefaultPoolOptions()
	opts.MinWorkers = 1
	opts.MaxWorkers = 1

	p, err := NewGoroutinePool(opts)
	if err != nil {
		t.Fatalf("NewGoroutinePool() error = %v", err)
	}
	defer p.Shutdown()

	if err := p.Start(); err != nil {
		t.Fatalf("Start() first call error = %v", err)
	}
	if err := p.Start(); err == nil {
		t.Fatal("Start() second call expected error, got nil")
	}
}

func TestGoroutinePoolSubmitAfterShutdown(t *testing.T) {
	opts := DefaultPoolOptions()
	opts.MinWorkers = 1
	opts.MaxWorkers = 1

	p, err := NewGoroutinePool(opts)
	if err != nil {
		t.Fatalf("NewGoroutinePool() error = %v", err)
	}
	if err := p.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := p.Shutdown(); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}

	task := p.Submit(func() (rttypes.Value, error) {
		return "ok", nil
	})

	done := make(chan struct{})
	var gotErr error
	go func() {
		_, gotErr = task.Await()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Await() blocked unexpectedly for task submitted after shutdown")
	}

	if gotErr == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(gotErr.Error(), "shutdown") {
		t.Fatalf("error = %q, want contains %q", gotErr.Error(), "shutdown")
	}
}

func TestGoroutinePoolRejectPolicyErrorCompletesTask(t *testing.T) {
	opts := DefaultPoolOptions()
	opts.MinWorkers = 0
	opts.MaxWorkers = 1
	opts.QueueSize = 1
	opts.RejectPolicy = "error"

	p, err := NewGoroutinePool(opts)
	if err != nil {
		t.Fatalf("NewGoroutinePool() error = %v", err)
	}
	if err := p.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	defer p.Shutdown()

	first := p.Submit(func() (rttypes.Value, error) { return 1, nil })
	if first == nil {
		t.Fatal("first task should not be nil")
	}

	second := p.Submit(func() (rttypes.Value, error) { return 2, nil })
	done := make(chan struct{})
	var gotErr error
	go func() {
		_, gotErr = second.Await()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Await() blocked unexpectedly for rejected task")
	}

	if gotErr == nil {
		t.Fatal("expected rejection error, got nil")
	}
	if !strings.Contains(gotErr.Error(), "queue is full") {
		t.Fatalf("error = %q, want contains %q", gotErr.Error(), "queue is full")
	}

	stats := p.GetStats()
	if stats.RejectedTasks < 1 {
		t.Fatalf("RejectedTasks = %d, want >= 1", stats.RejectedTasks)
	}
}

func TestPoolManagerSubmitBeforeInitialize(t *testing.T) {
	opts := DefaultPoolManagerOptions()
	opts.EnableDefault = true

	pm := NewPoolManager(opts)

	_, err := pm.Submit(DefaultPool, func() (rttypes.Value, error) { return nil, nil })
	if err == nil {
		t.Fatal("Submit() before Initialize() should fail")
	}
}

