package builtin

import (
	"testing"
	"time"

	rt "ialang/pkg/lang/runtime"
)

// TestTimerSetTimeout 测试 setTimeout
func TestTimerSetTimeout(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	timerMod := mustModuleObject(t, modules, "timer")

	executed := false
	callback := NativeFunction(func(args []Value) (Value, error) {
		executed = true
		return true, nil
	})

	// 创建 100ms 的超时
	result := callNative(t, timerMod, "setTimeout", callback, float64(100))
	id, ok := result.(float64)
	if !ok {
		t.Fatalf("setTimeout result type = %T, want float64", result)
	}

	// 等待执行完成
	time.Sleep(200 * time.Millisecond)

	if !executed {
		t.Fatal("setTimeout callback should have been executed")
	}

	// 清除定时器（虽然已经执行）
	callNative(t, timerMod, "clearTimeout", id)
}

// TestTimerSetInterval 测试 setInterval
func TestTimerSetInterval(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	timerMod := mustModuleObject(t, modules, "timer")

	execCount := 0
	callback := NativeFunction(func(args []Value) (Value, error) {
		execCount++
		return true, nil
	})

	// 创建 50ms 的间隔
	result := callNative(t, timerMod, "setInterval", callback, float64(50))
	id, ok := result.(float64)
	if !ok {
		t.Fatalf("setInterval result type = %T, want float64", result)
	}

	// 等待执行几次
	time.Sleep(200 * time.Millisecond)

	// 清除定时器
	callNative(t, timerMod, "clearInterval", id)

	if execCount < 3 {
		t.Fatalf("setInterval should have executed at least 3 times, got %d", execCount)
	}
}

// TestTimerClearTimeout 测试 clearTimeout
func TestTimerClearTimeout(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	timerMod := mustModuleObject(t, modules, "timer")

	executed := false
	callback := NativeFunction(func(args []Value) (Value, error) {
		executed = true
		return true, nil
	})

	// 创建 100ms 的超时
	result := callNative(t, timerMod, "setTimeout", callback, float64(100))
	id, ok := result.(float64)
	if !ok {
		t.Fatalf("setTimeout result type = %T, want float64", result)
	}

	// 立即清除
	callNative(t, timerMod, "clearTimeout", id)

	// 等待确认不执行
	time.Sleep(200 * time.Millisecond)

	if executed {
		t.Fatal("clearTimeout should have prevented callback execution")
	}
}

// TestTimerClearInterval 测试 clearInterval
func TestTimerClearInterval(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	timerMod := mustModuleObject(t, modules, "timer")

	execCount := 0
	callback := NativeFunction(func(args []Value) (Value, error) {
		execCount++
		return true, nil
	})

	// 创建 50ms 的间隔
	result := callNative(t, timerMod, "setInterval", callback, float64(50))
	id, ok := result.(float64)
	if !ok {
		t.Fatalf("setInterval result type = %T, want float64", result)
	}

	// 等待执行几次
	time.Sleep(100 * time.Millisecond)

	// 清除
	callNative(t, timerMod, "clearInterval", id)

	// 记录当前执行次数
	countAfterClear := execCount

	// 等待确认不再执行
	time.Sleep(150 * time.Millisecond)

	// 允许有 1 次误差（因为 goroutine 调度）
	if execCount > countAfterClear+1 {
		t.Fatalf("clearInterval should have stopped execution, execCount = %d, expected <= %d", execCount, countAfterClear+1)
	}
}

// TestTimerSleep 测试 sleep（同步）
func TestTimerSleep(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	timerMod := mustModuleObject(t, modules, "timer")

	start := time.Now()
	callNative(t, timerMod, "sleep", float64(100))
	elapsed := time.Since(start)

	if elapsed < 100*time.Millisecond {
		t.Fatalf("sleep should block for at least 100ms, got %v", elapsed)
	}
}

// TestTimerSleepAsync 测试 sleepAsync（异步）
func TestTimerSleepAsync(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	timerMod := mustModuleObject(t, modules, "timer")

	// 测试 sleepAsync 返回 Promise
	result := callNative(t, timerMod, "sleepAsync", float64(50))

	// 应该返回 Awaitable
	if _, ok := result.(rt.Awaitable); !ok {
		t.Fatalf("sleepAsync should return Awaitable, got %T", result)
	}
}

// TestTimerDefer 测试 defer
func TestTimerDefer(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	timerMod := mustModuleObject(t, modules, "timer")

	executed := false
	callback := NativeFunction(func(args []Value) (Value, error) {
		executed = true
		return "done", nil
	})

	// 创建延迟
	result := callNative(t, timerMod, "defer", callback, float64(100))

	// 应该返回 Awaitable
	if _, ok := result.(rt.Awaitable); !ok {
		t.Fatalf("defer should return Awaitable, got %T", result)
	}

	// 等待执行
	time.Sleep(200 * time.Millisecond)

	if !executed {
		t.Fatal("defer callback should have been executed")
	}
}

// TestTimerEvery 测试 every
func TestTimerEvery(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	timerMod := mustModuleObject(t, modules, "timer")

	execCount := 0
	callback := NativeFunction(func(args []Value) (Value, error) {
		execCount++
		return true, nil
	})

	// 创建间隔执行
	result := callNative(t, timerMod, "every", callback, float64(50))

	// 应该返回 Awaitable
	if _, ok := result.(rt.Awaitable); !ok {
		t.Fatalf("every should return Awaitable, got %T", result)
	}

	// 等待执行几次
	time.Sleep(200 * time.Millisecond)

	if execCount < 3 {
		t.Fatalf("every should have executed multiple times, got %d", execCount)
	}
}
