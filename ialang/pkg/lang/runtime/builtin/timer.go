package builtin

import (
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"ialang/pkg/lang/runtime"
	rtvm "ialang/pkg/lang/runtime/vm"
)

func callTimerCallback(fn Value) {
	switch cb := fn.(type) {
	case NativeFunction:
		pm := runtime.GetPoolManager()
		if pm.IsInitialized() && !pm.IsShutdown() {
			_, _ = pm.Submit(runtime.IOPool, func() (runtime.Value, error) {
				return cb([]Value{})
			})
		} else {
			go func() { _, _ = cb([]Value{}) }()
		}
	case *UserFunction:
		pm := runtime.GetPoolManager()
		if pm.IsInitialized() && !pm.IsShutdown() {
			_, _ = pm.Submit(runtime.IOPool, func() (runtime.Value, error) {
				return rtvm.CallUserFunctionSync(cb, []Value{})
			})
		} else {
			go func() { _, _ = rtvm.CallUserFunctionSync(cb, []Value{}) }()
		}
	}
}

func expectTimerCallback(fn Value) error {
	switch fn.(type) {
	case NativeFunction, *UserFunction:
		return nil
	default:
		return fmt.Errorf("expects function, got %T", fn)
	}
}

// timerModule 创建定时任务模块
func newTimerModule(asyncRuntime AsyncRuntime) Value {
	var mu sync.Mutex
	var timerIDCounter float64
	timers := make(map[float64]*time.Timer)
	intervals := make(map[float64]chan bool)
	
	// 初始化 Cron
	c := cron.New()
	c.Start()

	// setTimeout
	setTimeoutFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("timer.setTimeout expects at least 2 args: callback, delay")
		}
		if err := expectTimerCallback(args[0]); err != nil {
			return nil, fmt.Errorf("timer.setTimeout arg[0] %w", err)
		}
		callback := args[0]
		delay, err := asIntValue("timer.setTimeout arg[1]", args[1])
		if err != nil {
			return nil, err
		}

		mu.Lock()
		timerIDCounter++
		id := timerIDCounter
		mu.Unlock()

		t := time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
			callTimerCallback(callback)
		})

		mu.Lock()
		timers[id] = t
		mu.Unlock()

		return id, nil
	})

	// setInterval
	setIntervalFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("timer.setInterval expects at least 2 args: callback, interval")
		}
		if err := expectTimerCallback(args[0]); err != nil {
			return nil, fmt.Errorf("timer.setInterval arg[0] %w", err)
		}
		callback := args[0]
		interval, err := asIntValue("timer.setInterval arg[1]", args[1])
		if err != nil {
			return nil, err
		}

		mu.Lock()
		timerIDCounter++
		id := timerIDCounter
		stopCh := make(chan bool)
		intervals[id] = stopCh
		mu.Unlock()

		go func() {
			ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					callTimerCallback(callback)
				case <-stopCh:
					return
				}
			}
		}()

		return id, nil
	})

	// clearTimeout
	clearTimeoutFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("timer.clearTimeout expects 1 arg: id")
		}

		id, err := asIntValue("timer.clearTimeout arg[0]", args[0])
		if err != nil {
			return nil, err
		}

		mu.Lock()
		defer mu.Unlock()

		if t, ok := timers[float64(id)]; ok {
			t.Stop()
			delete(timers, float64(id))
		}

		return true, nil
	})

	// clearInterval
	clearIntervalFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("timer.clearInterval expects 1 arg: id")
		}

		id, err := asIntValue("timer.clearInterval arg[0]", args[0])
		if err != nil {
			return nil, err
		}

		mu.Lock()
		defer mu.Unlock()

		if ch, ok := intervals[float64(id)]; ok {
			close(ch)
			delete(intervals, float64(id))
		}

		return true, nil
	})

	// sleep (阻塞)
	sleepFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("timer.sleep expects 1 arg: ms")
		}

		ms, err := asIntValue("timer.sleep arg[0]", args[0])
		if err != nil {
			return nil, err
		}

		time.Sleep(time.Duration(ms) * time.Millisecond)
		return true, nil
	})

	// sleepAsync (非阻塞)
	sleepAsyncFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("timer.sleepAsync expects 1 arg: ms")
		}

		ms, err := asIntValue("timer.sleepAsync arg[0]", args[0])
		if err != nil {
			return nil, err
		}

		return asyncRuntime.Spawn(func() (Value, error) {
			time.Sleep(time.Duration(ms) * time.Millisecond)
			return true, nil
		}), nil
	})

	// defer (延迟执行)
	deferFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("timer.defer expects at least 2 args: callback, delay")
		}
		if err := expectTimerCallback(args[0]); err != nil {
			return nil, fmt.Errorf("timer.defer arg[0] %w", err)
		}
		callback := args[0]
		delay, err := asIntValue("timer.defer arg[1]", args[1])
		if err != nil {
			return nil, err
		}

		return asyncRuntime.Spawn(func() (Value, error) {
			time.Sleep(time.Duration(delay) * time.Millisecond)
			switch cb := callback.(type) {
			case NativeFunction:
				return cb([]Value{})
			case *UserFunction:
				return rtvm.CallUserFunctionSync(cb, []Value{})
			default:
				return nil, nil
			}
		}), nil
	})

	// every (间隔执行)
	everyFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("timer.every expects at least 2 args: callback, interval")
		}
		if err := expectTimerCallback(args[0]); err != nil {
			return nil, fmt.Errorf("timer.every arg[0] %w", err)
		}
		callback := args[0]
		interval, err := asIntValue("timer.every arg[1]", args[1])
		if err != nil {
			return nil, err
		}

		return asyncRuntime.Spawn(func() (Value, error) {
			ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
			defer ticker.Stop()
			for range ticker.C {
				callTimerCallback(callback)
			}
			return true, nil
		}), nil
	})

	// cron (Cron 表达式任务)
	cronFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("timer.cron expects at least 2 args: expression, callback")
		}

		expr, err := asStringValue("timer.cron arg[0]", args[0])
		if err != nil {
			return nil, err
		}
		if err := expectTimerCallback(args[1]); err != nil {
			return nil, fmt.Errorf("timer.cron arg[1] %w", err)
		}
		callback := args[1]

		id, err := c.AddFunc(expr, func() {
			callTimerCallback(callback)
		})

		if err != nil {
			return nil, fmt.Errorf("timer.cron invalid expression: %w", err)
		}

		return float64(id), nil
	})

	// removeJob (移除 Cron 任务)
	removeJobFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("timer.removeJob expects 1 arg: id")
		}

		id, err := asIntValue("timer.removeJob arg[0]", args[0])
		if err != nil {
			return nil, err
		}

		c.Remove(cron.EntryID(id))
		return true, nil
	})

	// Create module with self-reference for namespace pattern
	module := Object{
		"setTimeout":    setTimeoutFn,
		"setInterval":   setIntervalFn,
		"clearTimeout":  clearTimeoutFn,
		"clearInterval": clearIntervalFn,
		"sleep":         sleepFn,
		"sleepAsync":    sleepAsyncFn,
		"defer":         deferFn,
		"every":         everyFn,
		"cron":          cronFn,
		"removeJob":     removeJobFn,
	}
	// Self-reference for namespace usage: import { timer } from "timer"; timer.setTimeout()
	module["timer"] = module
	
	return module
}
