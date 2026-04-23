package api

import (
	"errors"
	"time"
)

type RetryableError interface {
	error
	HostRetryable() bool
}

type RetryBackoffError interface {
	error
	HostRetryBackoff() (time.Duration, bool)
}

type hostRetryableError struct {
	err            error
	backoffHint    time.Duration
	hasBackoffHint bool
}

func (e *hostRetryableError) Error() string {
	return e.err.Error()
}

func (e *hostRetryableError) Unwrap() error {
	return e.err
}

func (e *hostRetryableError) HostRetryable() bool {
	return true
}

func (e *hostRetryableError) HostRetryBackoff() (time.Duration, bool) {
	return e.backoffHint, e.hasBackoffHint
}

func MarkRetryable(err error) error {
	return MarkRetryableWithBackoff(err, 0, false)
}

func MarkRetryableAfter(err error, backoff time.Duration) error {
	return MarkRetryableWithBackoff(err, backoff, true)
}

func MarkRetryableWithBackoff(err error, backoff time.Duration, hasHint bool) error {
	if err == nil {
		return nil
	}
	if IsRetryableError(err) {
		return err
	}
	if backoff < 0 {
		backoff = 0
	}
	return &hostRetryableError{
		err:            err,
		backoffHint:    backoff,
		hasBackoffHint: hasHint,
	}
}

func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	var retryable RetryableError
	return errors.As(err, &retryable) && retryable.HostRetryable()
}

func RetryBackoffHint(err error) (time.Duration, bool) {
	if err == nil {
		return 0, false
	}
	var hinted RetryBackoffError
	if errors.As(err, &hinted) {
		return hinted.HostRetryBackoff()
	}
	return 0, false
}
