package builtin

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
)

func newSignalModule() Object {
	notifyFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("signal.notify expects 1 arg: signals")
		}

		signals, err := parseSignalList(args[0])
		if err != nil {
			return nil, err
		}

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, signals...)
		return newSignalSubscriptionObject(ch), nil
	})

	ignoreFn := NativeFunction(func(args []Value) (Value, error) {
		signals, err := parseOptionalSignalList("signal.ignore", args)
		if err != nil {
			return nil, err
		}
		signal.Ignore(signals...)
		return true, nil
	})

	resetFn := NativeFunction(func(args []Value) (Value, error) {
		signals, err := parseOptionalSignalList("signal.reset", args)
		if err != nil {
			return nil, err
		}
		signal.Reset(signals...)
		return true, nil
	})

	namespace := Object{
		"notify": notifyFn,
		"ignore": ignoreFn,
		"reset":  resetFn,
		"SIGINT": "SIGINT",
	}
	for name, value := range platformSignalConstants() {
		namespace[name] = value
	}
	module := cloneObject(namespace)
	module["signal"] = namespace
	return module
}

func newSignalSubscriptionObject(ch chan os.Signal) Object {
	recvFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("signal.subscription.recv expects 0 args, got %d", len(args))
		}
		sig := <-ch
		return signalName(sig), nil
	})

	stopFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("signal.subscription.stop expects 0 args, got %d", len(args))
		}
		signal.Stop(ch)
		return true, nil
	})

	return Object{
		"recv": recvFn,
		"stop": stopFn,
	}
}

func parseOptionalSignalList(label string, args []Value) ([]os.Signal, error) {
	if len(args) > 1 {
		return nil, fmt.Errorf("%s expects 0-1 args, got %d", label, len(args))
	}
	if len(args) == 0 {
		return nil, nil
	}
	return parseSignalList(args[0])
}

func parseSignalList(v Value) ([]os.Signal, error) {
	switch vv := v.(type) {
	case string:
		sig, err := parseSignalName(vv)
		if err != nil {
			return nil, err
		}
		return []os.Signal{sig}, nil
	case Array:
		out := make([]os.Signal, 0, len(vv))
		for i, item := range vv {
			name, err := asStringValue(fmt.Sprintf("signal list arg[%d]", i), item)
			if err != nil {
				return nil, err
			}
			sig, err := parseSignalName(name)
			if err != nil {
				return nil, err
			}
			out = append(out, sig)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("signal list expects string or array, got %T", v)
	}
}

func parseSignalName(name string) (os.Signal, error) {
	normalized := strings.ToUpper(strings.TrimSpace(name))
	switch normalized {
	case "INT", "SIGINT", "INTERRUPT":
		return os.Interrupt, nil
	default:
		if sig, ok := parsePlatformSignalName(normalized); ok {
			return sig, nil
		}
		return nil, fmt.Errorf("unsupported signal: %q", name)
	}
}

func signalName(sig os.Signal) string {
	switch sig {
	case os.Interrupt:
		return "SIGINT"
	default:
		if name, ok := platformSignalName(sig); ok {
			return name
		}
		return sig.String()
	}
}
