package runtime

import (
	"errors"
	"fmt"

	"iavm/pkg/core"
)

var ErrPromisePending = errors.New("promise is pending")

type promiseStatus string

const (
	promiseStatusPending  promiseStatus = "pending"
	promiseStatusResolved promiseStatus = "resolved"
	promiseStatusRejected promiseStatus = "rejected"
)

type promiseState struct {
	Status       promiseStatus
	Result       core.Value
	Error        string
	PollHandleID uint64
}

func pendingPromiseValue() core.Value {
	return core.Value{
		Kind: core.ValuePromise,
		Raw: &promiseState{
			Status: promiseStatusPending,
		},
	}
}

func resolvedPromiseValue(result core.Value) core.Value {
	return core.Value{
		Kind: core.ValuePromise,
		Raw: &promiseState{
			Status: promiseStatusResolved,
			Result: result,
		},
	}
}

func rejectedPromiseValue(message string) core.Value {
	return core.Value{
		Kind: core.ValuePromise,
		Raw: &promiseState{
			Status: promiseStatusRejected,
			Error:  message,
		},
	}
}

func promiseValueFromHostPoll(handleID uint64, result core.Value, done bool, errText string) core.Value {
	switch {
	case !done:
		return core.Value{
			Kind: core.ValuePromise,
			Raw: &promiseState{
				Status:       promiseStatusPending,
				PollHandleID: handleID,
			},
		}
	case errText != "":
		return rejectedPromiseValue(errText)
	default:
		return resolvedPromiseValue(result)
	}
}

func awaitValue(v core.Value) (core.Value, error) {
	if v.Kind != core.ValuePromise {
		return v, nil
	}

	state, ok := v.Raw.(*promiseState)
	if !ok || state == nil {
		return core.Value{}, fmt.Errorf("invalid promise value")
	}
	switch state.Status {
	case promiseStatusPending:
		return core.Value{}, ErrPromisePending
	case promiseStatusRejected:
		return core.Value{}, fmt.Errorf("await rejected promise: %s", state.Error)
	case promiseStatusResolved:
		return state.Result, nil
	default:
		return core.Value{}, fmt.Errorf("invalid promise status: %s", state.Status)
	}
}
