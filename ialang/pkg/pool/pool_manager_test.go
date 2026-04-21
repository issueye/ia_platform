package pool

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	commonrt "iacommon/pkg/ialang/runtime"
)

// TestPoolManagerInitialization 测试池管理器初始化
func TestPoolManagerInitialization(t *testing.T) {
	opts := DefaultPoolManagerOptions()
	opts.EnableDefault = true
	opts.EnableCPUPool = true
	opts.EnableIOPool = true

	pm := NewPoolManager(opts)

	if pm.IsInitialized() {
		t.Error("Pool manager should not be initialized yet")
	}

	err := pm.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize pool manager: %v", err)
	}

	if !pm.IsInitialized() {
		t.Error("Pool manager should be initialized")
	}

	if pm.PoolCount() != 3 {
		t.Errorf("Expected 3 pools, got %d", pm.PoolCount())
	}

	// 测试重复初始化
	err = pm.Initialize()
	if err == nil {
		t.Error("Should error on duplicate initialization")
	}

	// 清理
	pm.Shutdown()
}

// TestPoolManagerSubmit 测试任务提交
func TestPoolManagerSubmit(t *testing.T) {
	opts := DefaultPoolManagerOptions()
	opts.EnableDefault = true
	pm := NewPoolManager(opts)
	pm.Initialize()
	defer pm.Shutdown()

	var counter int64

	// 提交多个任务
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		_, err := pm.Submit(DefaultPool, func() (commonrt.Value, error) {
			atomic.AddInt64(&counter, 1)
			wg.Done()
			return nil, nil
		})
		if err != nil {
			t.Fatalf("Failed to submit task: %v", err)
		}
	}

	// 等待所有任务完成
	done := make(chan bool)
	go func() {
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		if counter != 10 {
			t.Errorf("Expected counter=10, got %d", counter)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for tasks to complete")
	}
}

// TestPoolManagerStats 测试统计信息
func TestPoolManagerStats(t *testing.T) {
	opts := DefaultPoolManagerOptions()
	opts.EnableDefault = true
	pm := NewPoolManager(opts)
	pm.Initialize()
	defer pm.Shutdown()

	// 提交一些任务
	for i := 0; i < 5; i++ {
		pm.Submit(DefaultPool, func() (commonrt.Value, error) {
			time.Sleep(10 * time.Millisecond)
			return nil, nil
		})
	}

	// 等待任务完成
	time.Sleep(500 * time.Millisecond)

	stats := pm.GetGlobalStats()

	if stats.TotalPools != 1 {
		t.Errorf("Expected 1 pool, got %d", stats.TotalPools)
	}

	if stats.TotalSubmitted != 5 {
		t.Errorf("Expected 5 submitted tasks, got %d", stats.TotalSubmitted)
	}

	if stats.TotalCompleted != 5 {
		t.Errorf("Expected 5 completed tasks, got %d", stats.TotalCompleted)
	}
}

// TestPoolManagerMultiplePoolTypes 测试多种池类型
func TestPoolManagerMultiplePoolTypes(t *testing.T) {
	opts := DefaultPoolManagerOptions()
	opts.EnableDefault = true
	opts.EnableCPUPool = true
	opts.EnableIOPool = true
	opts.EnableHighPriorityPool = true

	pm := NewPoolManager(opts)
	pm.Initialize()
	defer pm.Shutdown()

	// 测试默认池
	_, err := pm.Submit(DefaultPool, func() (commonrt.Value, error) {
		return nil, nil
	})
	if err != nil {
		t.Errorf("Failed to submit to default pool: %v", err)
	}

	// 测试CPU池
	_, err = pm.Submit(CPUPool, func() (commonrt.Value, error) {
		return nil, nil
	})
	if err != nil {
		t.Errorf("Failed to submit to CPU pool: %v", err)
	}

	// 测试IO池
	_, err = pm.Submit(IOPool, func() (commonrt.Value, error) {
		return nil, nil
	})
	if err != nil {
		t.Errorf("Failed to submit to IO pool: %v", err)
	}

	// 测试高优先级池
	_, err = pm.Submit(HighPriorityPool, func() (commonrt.Value, error) {
		return nil, nil
	})
	if err != nil {
		t.Errorf("Failed to submit to high priority pool: %v", err)
	}

	stats := pm.GetGlobalStats()
	if stats.TotalPools != 4 {
		t.Errorf("Expected 4 pools, got %d", stats.TotalPools)
	}
}

// TestPoolManagerShutdown 测试优雅关闭
func TestPoolManagerShutdown(t *testing.T) {
	opts := DefaultPoolManagerOptions()
	opts.EnableDefault = true
	pm := NewPoolManager(opts)
	pm.Initialize()

	var completed int64

	// 提交长时间任务
	for i := 0; i < 5; i++ {
		pm.Submit(DefaultPool, func() (commonrt.Value, error) {
			time.Sleep(100 * time.Millisecond)
			atomic.AddInt64(&completed, 1)
			return nil, nil
		})
	}

	// 立即关闭（应该等待任务完成）
	err := pm.ShutdownWithTimeout(5 * time.Second)
	if err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	if completed != 5 {
		t.Errorf("Expected 5 completed tasks, got %d", completed)
	}

	// 测试重复关闭
	err = pm.Shutdown()
	if err == nil {
		t.Error("Should error on duplicate shutdown")
	}
}

// TestPoolManagerAsyncRuntime 测试 AsyncRuntime 创建
func TestPoolManagerAsyncRuntime(t *testing.T) {
	opts := DefaultPoolManagerOptions()
	opts.EnableDefault = true
	pm := NewPoolManager(opts)
	pm.Initialize()
	defer pm.Shutdown()

	// 创建 AsyncRuntime
	asyncRuntime, err := pm.CreateDefaultAsyncRuntime()
	if err != nil {
		t.Fatalf("Failed to create async runtime: %v", err)
	}

	// 使用 AsyncRuntime 提交任务
	awaitable := asyncRuntime.Spawn(func() (commonrt.Value, error) {
		return nil, nil
	})

	_, err = asyncRuntime.AwaitValue(awaitable)
	if err != nil {
		t.Errorf("Async runtime task failed: %v", err)
	}
}

// TestPoolManagerEnsureInitialized 测试 EnsureInitialized
func TestPoolManagerEnsureInitialized(t *testing.T) {
	pm := GetPoolManager()

	// 第一次调用应该初始化
	err := pm.EnsureInitialized()
	if err != nil {
		t.Fatalf("Failed to ensure initialized: %v", err)
	}

	if !pm.IsInitialized() {
		t.Error("Pool manager should be initialized")
	}

	// 第二次调用应该无错误
	err = pm.EnsureInitialized()
	if err != nil {
		t.Fatalf("Second ensure initialized failed: %v", err)
	}

	// 清理全局管理器（用于下次测试）
	pm.Shutdown()
}

// TestPoolManagerEnsureInitializedWithOptions 测试带 options 的初始化入口
func TestPoolManagerEnsureInitializedWithOptions(t *testing.T) {
	pm := NewPoolManager(PoolManagerOptions{})

	opts := DefaultPoolManagerOptions()
	opts.EnableDefault = true
	opts.EnableIOPool = true
	opts.EnableCPUPool = false
	opts.EnableHighPriorityPool = false
	opts.ShutdownTimeout = 3 * time.Second

	err := pm.EnsureInitializedWithOptions(opts)
	if err != nil {
		t.Fatalf("EnsureInitializedWithOptions() failed: %v", err)
	}
	defer pm.Shutdown()

	if !pm.IsInitialized() {
		t.Fatal("pool manager should be initialized")
	}

	if pm.PoolCount() != 2 {
		t.Fatalf("expected 2 pools (default + io), got %d", pm.PoolCount())
	}

	if _, err := pm.GetPool(DefaultPool); err != nil {
		t.Fatalf("default pool should exist: %v", err)
	}
	if _, err := pm.GetPool(IOPool); err != nil {
		t.Fatalf("io pool should exist: %v", err)
	}
	if _, err := pm.GetPool(CPUPool); err == nil {
		t.Fatal("cpu pool should not exist")
	}
}

// TestPoolManagerCallbacks 测试回调函数
func TestPoolManagerCallbacks(t *testing.T) {
	opts := DefaultPoolManagerOptions()
	opts.EnableDefault = true
	pm := NewPoolManager(opts)

	var submitCount int64
	var completeCount int64
	var failCount int64

	pm.OnTaskSubmit = func(poolType PoolType, task *Task) {
		atomic.AddInt64(&submitCount, 1)
	}
	pm.OnTaskComplete = func(poolType PoolType, task *Task) {
		atomic.AddInt64(&completeCount, 1)
	}
	pm.OnTaskFailed = func(poolType PoolType, task *Task) {
		atomic.AddInt64(&failCount, 1)
	}

	pm.Initialize()
	defer pm.Shutdown()

	// 提交成功任务
	pm.Submit(DefaultPool, func() (commonrt.Value, error) {
		return nil, nil
	})

	time.Sleep(100 * time.Millisecond)

	if submitCount == 0 {
		t.Error("OnTaskSubmit callback not called")
	}
	if completeCount == 0 {
		t.Error("OnTaskComplete callback not called")
	}
}

// TestPoolManagerGetPoolStats 测试获取单个池统计
func TestPoolManagerGetPoolStats(t *testing.T) {
	opts := DefaultPoolManagerOptions()
	opts.EnableDefault = true
	pm := NewPoolManager(opts)
	pm.Initialize()
	defer pm.Shutdown()

	stats, err := pm.GetPoolStats(DefaultPool)
	if err != nil {
		t.Fatalf("Failed to get pool stats: %v", err)
	}

	if stats.TotalWorkers < 1 {
		t.Error("Expected at least 1 worker")
	}
}

// TestPoolManagerGlobalSingleton 测试全局单例
func TestPoolManagerGlobalSingleton(t *testing.T) {
	pm1 := GetPoolManager()
	pm2 := GetPoolManager()

	if pm1 != pm2 {
		t.Error("GetPoolManager should return same instance")
	}
}
