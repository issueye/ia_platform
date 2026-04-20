package core

import "errors"

var (
	ErrInvalidModule     = errors.New("invalid module")
	ErrInvalidOpcode     = errors.New("invalid opcode")
	ErrVerifyFailed      = errors.New("verify failed")
	ErrCapabilityDenied  = errors.New("capability denied")
	ErrUnknownOperation  = errors.New("unknown host operation")
	ErrHandleNotFound    = errors.New("handle not found")
	ErrResourceExhausted = errors.New("resource exhausted")
)
