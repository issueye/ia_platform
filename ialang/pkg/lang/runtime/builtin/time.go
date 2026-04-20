package builtin

import (
	"fmt"
	"time"
)

func newTimeModule(asyncRuntime AsyncRuntime) Object {
	nowUnixFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("time.nowUnix expects 0 args, got %d", len(args))
		}
		return float64(time.Now().Unix()), nil
	})
	nowUnixMilliFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("time.nowUnixMilli expects 0 args, got %d", len(args))
		}
		return float64(time.Now().UnixMilli()), nil
	})
	nowUnixMicroFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("time.nowUnixMicro expects 0 args, got %d", len(args))
		}
		return float64(time.Now().UnixMicro()), nil
	})
	nowISOFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 0 {
			return nil, fmt.Errorf("time.nowISO expects 0 args, got %d", len(args))
		}
		return time.Now().Format(time.RFC3339Nano), nil
	})
	sleepFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("time.sleep expects 1 arg, got %d", len(args))
		}
		ms, err := asIntArg("time.sleep", args, 0)
		if err != nil {
			return nil, err
		}
		if ms < 0 {
			return nil, fmt.Errorf("time.sleep arg[0] expects non-negative integer, got %d", ms)
		}
		time.Sleep(time.Duration(ms) * time.Millisecond)
		return true, nil
	})
	sleepAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		return asyncRuntime.Spawn(func() (Value, error) {
			return sleepFn(args)
		}), nil
	})
	parseISOFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 1 {
			return nil, fmt.Errorf("time.parseISO expects 1 arg, got %d", len(args))
		}
		text, err := asStringArg("time.parseISO", args, 0)
		if err != nil {
			return nil, err
		}
		t, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			t, err = time.Parse(time.RFC3339, text)
			if err != nil {
				return nil, err
			}
		}
		return Object{
			"unix":      float64(t.Unix()),
			"unixMilli": float64(t.UnixMilli()),
			"iso":       t.Format(time.RFC3339Nano),
			"year":      float64(t.Year()),
			"month":     float64(t.Month()),
			"day":       float64(t.Day()),
			"hour":      float64(t.Hour()),
			"minute":    float64(t.Minute()),
			"second":    float64(t.Second()),
		}, nil
	})
	formatFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("time.format expects 1-2 args: unixMilli, [layout]")
		}
		ms, err := asIntArg("time.format", args, 0)
		if err != nil {
			return nil, err
		}
		layout := "2006-01-02 15:04:05"
		if len(args) == 2 {
			layout, err = asStringArg("time.format", args, 1)
			if err != nil {
				return nil, err
			}
		}
		return time.UnixMilli(int64(ms)).Format(layout), nil
	})
	parseFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 || len(args) > 2 {
			return nil, fmt.Errorf("time.parse expects 1-2 args: text, [layout]")
		}
		text, err := asStringArg("time.parse", args, 0)
		if err != nil {
			return nil, err
		}
		layout := "2006-01-02 15:04:05"
		if len(args) == 2 {
			layout, err = asStringArg("time.parse", args, 1)
			if err != nil {
				return nil, err
			}
		}
		t, err := time.Parse(layout, text)
		if err != nil {
			return nil, err
		}
		return float64(t.UnixMilli()), nil
	})
	addFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("time.add expects 2 args: unixMilli, durationMs")
		}
		ms, err := asIntArg("time.add", args, 0)
		if err != nil {
			return nil, err
		}
		dur, err := asIntArg("time.add", args, 1)
		if err != nil {
			return nil, err
		}
		return float64(time.UnixMilli(int64(ms)).Add(time.Duration(dur)*time.Millisecond).UnixMilli()), nil
	})
	diffFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) != 2 {
			return nil, fmt.Errorf("time.diff expects 2 args: unixMilliA, unixMilliB")
		}
		a, err := asIntArg("time.diff", args, 0)
		if err != nil {
			return nil, err
		}
		b, err := asIntArg("time.diff", args, 1)
		if err != nil {
			return nil, err
		}
		return float64(int64(b) - int64(a)), nil
	})

	namespace := Object{
		"nowUnix":      nowUnixFn,
		"nowUnixMilli": nowUnixMilliFn,
		"nowUnixMicro": nowUnixMicroFn,
		"nowISO":       nowISOFn,
		"sleep":        sleepFn,
		"sleepAsync":   sleepAsyncFn,
		"parseISO":     parseISOFn,
		"format":       formatFn,
		"parse":        parseFn,
		"add":          addFn,
		"diff":         diffFn,
	}
	module := cloneObject(namespace)
	module["time"] = namespace
	return module
}
