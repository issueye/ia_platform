package api

import (
	"errors"
	"testing"
	"time"
)

func TestMarkRetryablePreservesUnderlyingError(t *testing.T) {
	baseErr := errors.New("temporary failure")

	err := MarkRetryable(baseErr)
	if !IsRetryableError(err) {
		t.Fatal("expected retryable marker")
	}
	if !errors.Is(err, baseErr) {
		t.Fatal("expected wrapped error to preserve underlying error")
	}
}

func TestMarkRetryableIsIdempotent(t *testing.T) {
	baseErr := errors.New("temporary failure")

	first := MarkRetryable(baseErr)
	second := MarkRetryable(first)
	if first != second {
		t.Fatal("expected retryable marker to be idempotent")
	}
}

func TestMarkRetryableAfterExposesBackoffHint(t *testing.T) {
	baseErr := errors.New("temporary failure")

	err := MarkRetryableAfter(baseErr, 3*time.Second)
	if !IsRetryableError(err) {
		t.Fatal("expected retryable marker")
	}
	backoff, ok := RetryBackoffHint(err)
	if !ok {
		t.Fatal("expected retry backoff hint")
	}
	if backoff != 3*time.Second {
		t.Fatalf("backoff hint = %v, want 3s", backoff)
	}
}
