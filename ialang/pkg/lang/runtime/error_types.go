package runtime

import "fmt"

// ErrorType defines the category of an error.
type ErrorType string

const (
	ErrorTypeGeneric      ErrorType = "Error"
	ErrorTypeRuntime      ErrorType = "RuntimeError"
	ErrorTypeTimeout      ErrorType = "TimeoutError"
	ErrorTypeSandbox      ErrorType = "SandboxError"
	ErrorTypeImport       ErrorType = "ImportError"
	ErrorTypeSyntax       ErrorType = "SyntaxError"
	ErrorTypeReference    ErrorType = "ReferenceError"
	ErrorTypeType         ErrorType = "TypeError"
	ErrorTypeRange        ErrorType = "RangeError"
)

// IaError is the base error type for all ialang runtime errors.
type IaError struct {
	Type    ErrorType
	Name    string
	Code    string
	Message string
	Cause   error
	Retryable bool
	// Optional context fields
	ModulePath string
	IP         int
	OpCode     int
	StackDepth int
}

func (e *IaError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Name, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Name, e.Message)
}

func (e *IaError) Unwrap() error {
	return e.Cause
}

// ToObject converts an IaError to an Object value for catch blocks.
func (e *IaError) ToObject() Object {
	obj := Object{
		"name":      string(e.Type),
		"code":      e.Code,
		"message":   e.Message,
		"retryable": e.Retryable,
	}
	if e.ModulePath != "" {
		obj["module"] = e.ModulePath
	}
	if e.IP >= 0 {
		obj["ip"] = float64(e.IP)
	}
	if e.OpCode >= 0 {
		obj["op"] = float64(e.OpCode)
	}
	if e.StackDepth >= 0 {
		obj["stack_depth"] = float64(e.StackDepth)
	}
	if e.Cause != nil {
		obj["cause"] = e.Cause.Error()
	}
	return obj
}

// NewRuntimeError creates a new runtime error.
func NewRuntimeError(message string, opts ...func(*IaError)) *IaError {
	err := &IaError{
		Type:    ErrorTypeRuntime,
		Name:    "RuntimeError",
		Code:    "RUNTIME_ERROR",
		Message: message,
	}
	for _, opt := range opts {
		opt(err)
	}
	return err
}

// NewTimeoutError creates a new timeout error.
func NewTimeoutError(message string, retryable bool) *IaError {
	return &IaError{
		Type:      ErrorTypeTimeout,
		Name:      "TimeoutError",
		Code:      "TIMEOUT",
		Message:   message,
		Retryable: retryable,
	}
}

// NewSandboxError creates a new sandbox policy violation error.
func NewSandboxError(violation, limit, current string) *IaError {
	return &IaError{
		Type:    ErrorTypeSandbox,
		Name:    "SandboxError",
		Code:    "SANDBOX_VIOLATION",
		Message: fmt.Sprintf("%s (limit: %s, current: %s)", violation, limit, current),
	}
}

// NewImportError creates a new module import error.
func NewImportError(message string, modulePath string) *IaError {
	return &IaError{
		Type:       ErrorTypeImport,
		Name:       "ImportError",
		Code:       "IMPORT_ERROR",
		Message:    message,
		ModulePath: modulePath,
	}
}

// NewTypeError creates a new type error.
func NewTypeError(message string) *IaError {
	return &IaError{
		Type:    ErrorTypeType,
		Name:    "TypeError",
		Code:    "TYPE_ERROR",
		Message: message,
	}
}

// NewReferenceError creates a new reference error (undefined variable).
func NewReferenceError(name string) *IaError {
	return &IaError{
		Type:    ErrorTypeReference,
		Name:    "ReferenceError",
		Code:    "REFERENCE_ERROR",
		Message: fmt.Sprintf("%s is not defined", name),
	}
}

// WithContext adds execution context to an error.
func WithContext(modulePath string, ip, op, stackDepth int) func(*IaError) {
	return func(e *IaError) {
		e.ModulePath = modulePath
		e.IP = ip
		e.OpCode = op
		e.StackDepth = stackDepth
	}
}

// WithCause sets the underlying cause of an error.
func WithCause(cause error) func(*IaError) {
	return func(e *IaError) {
		e.Cause = cause
	}
}

// IsTimeoutError checks if an error is a timeout error.
func IsTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	iaErr, ok := err.(*IaError)
	if ok {
		return iaErr.Type == ErrorTypeTimeout
	}
	return false
}

// IsSandboxError checks if an error is a sandbox policy violation.
func IsSandboxError(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(*SandboxError); ok {
		return true
	}
	iaErr, ok := err.(*IaError)
	if ok {
		return iaErr.Type == ErrorTypeSandbox
	}
	return false
}

// IsImportError checks if an error is an import error.
func IsImportError(err error) bool {
	if err == nil {
		return false
	}
	iaErr, ok := err.(*IaError)
	if ok {
		return iaErr.Type == ErrorTypeImport
	}
	return false
}
