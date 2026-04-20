package vm

import "errors"

type thrownError struct {
	value Value
}

func (e *thrownError) Error() string {
	return "thrown: " + toString(e.value)
}

type runtimeExecError struct {
	cause      error
	modulePath string
	ip         int
	op         OpCode
	stackDepth int
}

func (e *runtimeExecError) Error() string {
	if e.cause == nil {
		return ""
	}
	return e.cause.Error()
}

func (e *runtimeExecError) Unwrap() error {
	return e.cause
}

func (v *VM) attachRuntimeContext(err error, op OpCode) error {
	if err == nil {
		return nil
	}
	var runtimeErr *runtimeExecError
	if errors.As(err, &runtimeErr) {
		return err
	}
	return &runtimeExecError{
		cause:      err,
		modulePath: v.modulePath,
		ip:         v.ip - 1,
		op:         op,
		stackDepth: len(v.stack),
	}
}

func (v *VM) handleRuntimeError(err error) bool {
	if len(v.tryStack) == 0 {
		return false
	}
	last := len(v.tryStack) - 1
	frame := v.tryStack[last]
	v.tryStack = v.tryStack[:last]

	var catchVal Value
	var thrown *thrownError
	if errors.As(err, &thrown) {
		catchVal = thrown.value
	} else {
		catchVal = v.runtimeErrorToCatchValue(err)
	}
	if frame.stackBase >= 0 && frame.stackBase <= len(v.stack) {
		v.stack = v.stack[:frame.stackBase]
	}
	v.defineName(frame.catchName, catchVal)
	v.ip = frame.catchIP
	return true
}

func (v *VM) runtimeErrorToCatchValue(err error) Value {
	if v.options.StructuredRuntimeErrors {
		return v.runtimeErrorToStructuredValue(err)
	}
	return v.runtimeErrorToLegacyValue(err)
}

func (v *VM) runtimeErrorToLegacyValue(err error) Value {
	runtimeName := "unknown"
	if v.asyncRuntime != nil {
		runtimeName = v.asyncRuntime.Name()
	}
	switch {
	case errors.Is(err, ErrAsyncTaskTimeout):
		return NewAsyncRuntimeErrorValue(
			"AsyncTaskTimeout",
			AsyncErrorCodeTaskTimeout,
			AsyncErrorKindTimeout,
			err.Error(),
			runtimeName,
			true,
		)
	case errors.Is(err, ErrAsyncAwaitTimeout):
		return NewAsyncRuntimeErrorValue(
			"AsyncAwaitTimeout",
			AsyncErrorCodeAwaitTimeout,
			AsyncErrorKindTimeout,
			err.Error(),
			runtimeName,
			true,
		)
	default:
		return err.Error()
	}
}

func (v *VM) runtimeErrorToStructuredValue(err error) Value {
	runtimeName := "unknown"
	if v.asyncRuntime != nil {
		runtimeName = v.asyncRuntime.Name()
	}
	switch {
	case errors.Is(err, ErrAsyncTaskTimeout):
		return v.withRuntimeContextFields(err, NewAsyncRuntimeErrorValue(
			"AsyncTaskTimeout",
			AsyncErrorCodeTaskTimeout,
			AsyncErrorKindTimeout,
			err.Error(),
			runtimeName,
			true,
		))
	case errors.Is(err, ErrAsyncAwaitTimeout):
		return v.withRuntimeContextFields(err, NewAsyncRuntimeErrorValue(
			"AsyncAwaitTimeout",
			AsyncErrorCodeAwaitTimeout,
			AsyncErrorKindTimeout,
			err.Error(),
			runtimeName,
			true,
		))
	default:
		return v.withRuntimeContextFields(err, NewAsyncRuntimeErrorValue(
			"RuntimeError",
			RuntimeErrorCodeGeneric,
			RuntimeErrorKindGeneric,
			err.Error(),
			runtimeName,
			false,
		))
	}
}

func (v *VM) withRuntimeContextFields(err error, base Object) Object {
	modulePath, ip, op, stackDepth := v.runtimeErrorContext(err)
	out := Object{}
	for k, val := range base {
		out[k] = val
	}
	out["module"] = modulePath
	out["ip"] = float64(ip)
	out["op"] = float64(op)
	out["stack_depth"] = float64(stackDepth)
	return out
}

func (v *VM) runtimeErrorContext(err error) (modulePath string, ip int, op int, stackDepth int) {
	modulePath = v.modulePath
	ip = v.ip - 1
	op = -1
	stackDepth = len(v.stack)

	var runtimeErr *runtimeExecError
	if errors.As(err, &runtimeErr) {
		modulePath = runtimeErr.modulePath
		ip = runtimeErr.ip
		op = int(runtimeErr.op)
		stackDepth = runtimeErr.stackDepth
	}
	return
}

