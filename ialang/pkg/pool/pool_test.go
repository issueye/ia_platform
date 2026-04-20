package pool

import (
	"sync/atomic"
	"testing"
	"time"

	rttypes "ialang/pkg/lang/runtime/types"
)

// testError 测试用错误类型
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// TestGoroutinePoolBasic 测试基本任务提交
func TestGoroutinePoolBasic(t *testing.T) {
	opts := DefaultPoolOptions()
	opts.MinWorkers = 2
	opts.MaxWorkers = 5

	pool, err := NewGoroutinePool(opts)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	if err := pool.Start(); err != nil {
		t.Fatalf("Failed to start pool: %v", err)
	}

	defer pool.Shutdown()

	// 提交任务
	completed := int32(0)
	for i := 0; i < 10; i++ {
		pool.Submit(func() (rttypes.Value, error) {
			atomic.AddInt32(&completed, 1)
			return true, nil
		})
	}

	// 等待任务完成
	time.Sleep(500 * time.Millisecond)

	if atomic.LoadInt32(&completed) != 10 {
		t.Fatalf("Expected 10 completed tasks, got %d", completed)
	}
}

// TestGoroutinePoolStats 测试统计信息
func TestGoroutinePoolStats(t *testing.T) {
	opts := DefaultPoolOptions()
	opts.MinWorkers = 2
	opts.MaxWorkers = 10

	pool, _ := NewGoroutinePool(opts)
	pool.Start()
	defer pool.Shutdown()

	// 提交一些任务
	for i := 0; i < 20; i++ {
		pool.Submit(func() (rttypes.Value, error) {
			time.Sleep(10 * time.Millisecond)
			return true, nil
		})
	}

	// 获取统计信息
	stats := pool.GetStats()

	if stats.TotalWorkers < opts.MinWorkers {
		t.Fatalf("Expected at least %d workers, got %d", opts.MinWorkers, stats.TotalWorkers)
	}

	if stats.QueuedTasks > opts.QueueSize {
		t.Fatalf("Queued tasks %d exceeds queue size %d", stats.QueuedTasks, opts.QueueSize)
	}

	t.Logf("Pool stats: %+v", stats)
}

// TestGoroutinePoolRetry 测试任务重试
func TestGoroutinePoolRetry(t *testing.T) {
	opts := DefaultPoolOptions()
	pool, _ := NewGoroutinePool(opts)
	pool.Start()
	defer pool.Shutdown()

	attempt := int32(0)

	task := pool.SubmitWithRetry(func() (rttypes.Value, error) {
		count := atomic.AddInt32(&attempt, 1)
		if count < 3 {
			return nil, &testError{msg: "fail"}
		}
		return "success", nil
	}, 3)

	// 等待任务完成
	time.Sleep(500 * time.Millisecond)

	result, err := task.Await()
	if err != nil {
		t.Fatalf("Task should succeed after retries, got error: %v", err)
	}

	if result != "success" {
		t.Fatalf("Expected result 'success', got %v", result)
	}

	if atomic.LoadInt32(&attempt) != 3 {
		t.Fatalf("Expected 3 attempts, got %d", attempt)
	}
}

// TestGoroutinePoolShutdown 测试优雅关闭
func TestGoroutinePoolShutdown(t *testing.T) {
	opts := DefaultPoolOptions()
	opts.MinWorkers = 2

	pool, _ := NewGoroutinePool(opts)
	pool.Start()

	// 提交任务
	for i := 0; i < 5; i++ {
		pool.Submit(func() (rttypes.Value, error) {
			time.Sleep(50 * time.Millisecond)
			return true, nil
		})
	}

	// 立即关闭（应该等待任务完成）
	err := pool.ShutdownWithTimeout(2 * time.Second)
	if err != nil {
		t.Fatalf("Shutdown should complete gracefully: %v", err)
	}
}

// TestPoolAsyncRuntime 测试 PoolAsyncRuntime
func TestPoolAsyncRuntime(t *testing.T) {
	opts := DefaultPoolOptions()
	opts.MinWorkers = 2

	runtime, err := NewPoolAsyncRuntime(opts)
	if err != nil {
		t.Fatalf("Failed to create pool runtime: %v", err)
	}

	defer runtime.GetPool().Shutdown()

	// 提交任务
	awaitable := runtime.Spawn(func() (rttypes.Value, error) {
		return 42, nil
	})

	result, err := runtime.AwaitValue(awaitable)
	if err != nil {
		t.Fatalf("Task should succeed: %v", err)
	}

	if result != 42 {
		t.Fatalf("Expected 42, got %v", result)
	}

	if runtime.Name() != "PoolRuntime" {
		t.Fatalf("Expected name 'PoolRuntime', got %s", runtime.Name())
	}
}

// TestPoolOptionsValidation 测试配置验证
func TestPoolOptionsValidation(t *testing.T) {
	// 测试无效配置
	testCases := []struct {
		name    string
		opts    PoolOptions
		wantErr bool
	}{
		{
			name:    "negative minWorkers",
			opts:    PoolOptions{MinWorkers: -1, MaxWorkers: 10, QueueSize: 100},
			wantErr: true,
		},
		{
			name:    "maxWorkers < minWorkers",
			opts:    PoolOptions{MinWorkers: 10, MaxWorkers: 5, QueueSize: 100},
			wantErr: true,
		},
		{
			name:    "zero queueSize",
			opts:    PoolOptions{MinWorkers: 2, MaxWorkers: 10, QueueSize: 0},
			wantErr: true,
		},
		{
			name:    "valid options",
			opts:    PoolOptions{MinWorkers: 2, MaxWorkers: 10, QueueSize: 100},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewGoroutinePool(tc.opts)
			if tc.wantErr && err == nil {
				t.Fatal("Expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}
		})
	}
}
