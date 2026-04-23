package api

import "errors"

type RetryableError interface {
	error
	HostRetryable() bool
}

type hostRetryableError struct {
	err error
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

func MarkRetryable(err error) error {
	if err == nil {
		return nil
	}
	if IsRetryableError(err) {
		return err
	}
	return &hostRetryableError{err: err}
}

func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}
	var retryable RetryableError
	return errors.As(err, &retryable) && retryable.HostRetryable()
}
