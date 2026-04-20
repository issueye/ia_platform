package lang

import (
	"context"
	"errors"
	rt "ialang/pkg/lang/runtime"
	"testing"
	"time"
)

func TestPromiseAwaitContextDone(t *testing.T) {
	p := rt.ResolvedPromise(float64(3))
	got, err := p.AwaitContext(context.Background())
	if err != nil {
		t.Fatalf("AwaitContext(resolved) unexpected error: %v", err)
	}
	if got != float64(3) {
		t.Fatalf("AwaitContext(resolved) = %v, want 3", got)
	}
}

func TestPromiseAwaitContextTimeout(t *testing.T) {
	p := rt.NewPromise(func() (rt.Value, error) {
		time.Sleep(30 * time.Millisecond)
		return "ok", nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer cancel()
	_, err := p.AwaitContext(ctx)
	if err == nil {
		t.Fatal("AwaitContext(timeout) expected error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("AwaitContext(timeout) error = %v, want DeadlineExceeded", err)
	}
}
