package builtin

import (
	"fmt"

	rttypes "ialang/pkg/lang/runtime/types"
	"ialang/pkg/lang/runtime"
)

func newPromiseModule() rttypes.Value {
	module := rttypes.Object{
		"all": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Promise.all expects 1 arg, got %d", len(args))
			}
			arr, ok := args[0].(rttypes.Array)
			if !ok {
				return nil, fmt.Errorf("Promise.all expects array of promises")
			}
			promises := make([]runtime.Awaitable, 0, len(arr))
			for _, v := range arr {
				if aw, ok := v.(runtime.Awaitable); ok {
					promises = append(promises, aw)
				} else {
					promises = append(promises, runtime.ResolvedPromise(v))
				}
			}
			return runtime.PromiseAll(promises), nil
		}),
		"race": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Promise.race expects 1 arg, got %d", len(args))
			}
			arr, ok := args[0].(rttypes.Array)
			if !ok {
				return nil, fmt.Errorf("Promise.race expects array of promises")
			}
			promises := make([]runtime.Awaitable, 0, len(arr))
			for _, v := range arr {
				if aw, ok := v.(runtime.Awaitable); ok {
					promises = append(promises, aw)
				} else {
					promises = append(promises, runtime.ResolvedPromise(v))
				}
			}
			return runtime.PromiseRace(promises), nil
		}),
		"allSettled": rttypes.NativeFunction(func(args []rttypes.Value) (rttypes.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("Promise.allSettled expects 1 arg, got %d", len(args))
			}
			arr, ok := args[0].(rttypes.Array)
			if !ok {
				return nil, fmt.Errorf("Promise.allSettled expects array of promises")
			}
			promises := make([]runtime.Awaitable, 0, len(arr))
			for _, v := range arr {
				if aw, ok := v.(runtime.Awaitable); ok {
					promises = append(promises, aw)
				} else {
					promises = append(promises, runtime.ResolvedPromise(v))
				}
			}
			return runtime.PromiseAllSettled(promises), nil
		}),
	}

	// Create module with self-reference for namespace pattern
	result := cloneObject(module)
	result["Promise"] = result
	
	return result
}
