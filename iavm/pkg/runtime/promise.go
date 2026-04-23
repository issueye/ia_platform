package runtime

import (
	"fmt"

	"iavm/pkg/core"
)

type promiseState struct {
	Done   bool
	Result core.Value
	Error  string
}

func resolvedPromiseValue(result core.Value) core.Value {
	return core.Value{
		Kind: core.ValuePromise,
		Raw: &promiseState{
			Done:   true,
			Result: result,
		},
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
	if !state.Done {
		return core.Value{}, fmt.Errorf("await on pending promise is not supported yet")
	}
	if state.Error != "" {
		return core.Value{}, fmt.Errorf("await rejected promise: %s", state.Error)
	}
	return state.Result, nil
}
