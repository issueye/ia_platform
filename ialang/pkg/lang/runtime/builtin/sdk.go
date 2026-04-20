package builtin

import (
	"fmt"
	"time"
)

func newAgentSDKModule(asyncRuntime AsyncRuntime) Object {
	return Object{
		"llm": Object{
			"chat": NativeFunction(func(args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("llm.chat expects 1 arg, got %d", len(args))
				}
				return "[mock-llm] " + toString(args[0]), nil
			}),
			"chatAsync": NativeFunction(func(args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("llm.chatAsync expects 1 arg, got %d", len(args))
				}
				p := asyncRuntime.Spawn(func() (Value, error) {
					time.Sleep(30 * time.Millisecond)
					return "[mock-llm-async] " + toString(args[0]), nil
				})
				return p, nil
			}),
		},
		"tool": Object{
			"call": NativeFunction(func(args []Value) (Value, error) {
				if len(args) < 1 {
					return nil, fmt.Errorf("tool.call expects at least 1 arg")
				}
				return "[mock-tool] called " + toString(args[0]), nil
			}),
		},
		"memory": Object{
			"get": NativeFunction(func(args []Value) (Value, error) {
				if len(args) != 1 {
					return nil, fmt.Errorf("memory.get expects 1 arg, got %d", len(args))
				}
				return "[mock-memory] " + toString(args[0]), nil
			}),
		},
	}
}
