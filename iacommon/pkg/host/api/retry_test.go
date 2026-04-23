package api

import (
	"errors"
	"testing"
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
