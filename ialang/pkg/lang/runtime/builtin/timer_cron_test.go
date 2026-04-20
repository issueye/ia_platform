package builtin

import (
	"testing"
	"time"

	rt "ialang/pkg/lang/runtime"
)

// TestTimerCron 测试 Cron 表达式
func TestTimerCron(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	timerMod := mustModuleObject(t, modules, "timer")

	execCount := 0
	callback := NativeFunction(func(args []Value) (Value, error) {
		execCount++
		return true, nil
	})

	// 使用每分钟执行的表达式 (* * * * *)
	// 为了测试，我们无法等待一分钟，所以主要测试注册成功和表达式解析错误
	// 但我们可以测试错误的表达式
	
	// 测试正确表达式注册
	result := callNative(t, timerMod, "cron", "* * * * *", callback)
	id, ok := result.(float64)
	if !ok {
		t.Fatalf("cron result type = %T, want float64", result)
	}
	if id <= 0 {
		t.Fatal("cron should return positive job id")
	}

	// 移除任务
	callNative(t, timerMod, "removeJob", id)
}

// TestTimerCronInvalidExpression 测试无效的 Cron 表达式
func TestTimerCronInvalidExpression(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	timerMod := mustModuleObject(t, modules, "timer")

	callback := NativeFunction(func(args []Value) (Value, error) {
		return true, nil
	})

	// 测试无效表达式
	_, err := callNativeWithError(timerMod, "cron", "invalid cron string", callback)
	if err == nil {
		t.Fatal("cron should fail on invalid expression")
	}
}

// TestTimerCronEverySecond 测试每秒执行 (@every 1s)
func TestTimerCronEverySecond(t *testing.T) {
	modules := DefaultModules(rt.NewGoroutineRuntime())
	timerMod := mustModuleObject(t, modules, "timer")

	execCount := 0
	callback := NativeFunction(func(args []Value) (Value, error) {
		execCount++
		return true, nil
	})

	// 使用 @every 1s 表达式
	result := callNative(t, timerMod, "cron", "@every 1s", callback)
	id, ok := result.(float64)
	if !ok {
		t.Fatalf("cron result type = %T, want float64", result)
	}

	// 等待执行几次
	time.Sleep(2500 * time.Millisecond)

	// 移除任务
	callNative(t, timerMod, "removeJob", id)

	// 验证执行次数 (大约 2-3 次)
	if execCount < 2 {
		t.Fatalf("@every 1s should execute at least 2 times in 2.5s, got %d", execCount)
	}
}
