package builtin

import (
	"fmt"
	"time"

	"ialang/pkg/lang/runtime"
)

// goroutinePoolModule 创建协程池模块
func newGoroutinePoolModule() Value {
	// 使用全局统一的池管理器
	poolManager := runtime.GetPoolManager()

	// 获取或初始化默认池
	getDefaultPool := func() (*runtime.GoroutinePool, error) {
		// 确保已初始化
		if err := poolManager.EnsureInitialized(); err != nil {
			return nil, err
		}
		return poolManager.GetPool(runtime.DefaultPool)
	}

	// submit 提交任务
	submitFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("pool.submit expects at least 1 arg: callback")
		}

		callback, ok := args[0].(NativeFunction)
		if !ok {
			return nil, fmt.Errorf("pool.submit arg[0] expects function, got %T", args[0])
		}

		pool, err := getDefaultPool()
		if err != nil {
			return nil, err
		}

		task := pool.Submit(func() (Value, error) {
			return callback([]Value{})
		})

		return task, nil
	})

	// submitWithRetry 提交任务（带重试）
	submitWithRetryFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("pool.submitWithRetry expects at least 1 arg: callback")
		}

		callback, ok := args[0].(NativeFunction)
		if !ok {
			return nil, fmt.Errorf("pool.submitWithRetry arg[0] expects function, got %T", args[0])
		}

		maxRetries := 3
		if len(args) > 1 {
			if retries, ok := args[1].(float64); ok {
				maxRetries = int(retries)
			}
		}

		pool, err := getDefaultPool()
		if err != nil {
			return nil, err
		}

		task := pool.SubmitWithRetry(func() (Value, error) {
			return callback([]Value{})
		}, maxRetries)

		return task, nil
	})

	// getStats 获取池统计信息
	getStatsFn := NativeFunction(func(args []Value) (Value, error) {
		stats := poolManager.GetGlobalStats()

		return Object{
			"totalSubmitted": float64(stats.TotalSubmitted),
			"totalCompleted": float64(stats.TotalCompleted),
			"totalFailed":    float64(stats.TotalFailed),
			"totalRejected":  float64(stats.TotalRejected),
			"totalPools":     float64(stats.TotalPools),
			"activePools":    float64(stats.ActivePools),
			"totalWorkers":   float64(stats.TotalWorkers),
			"activeWorkers":  float64(stats.ActiveWorkers),
			"queuedTasks":    float64(stats.QueuedTasks),
		}, nil
	})

	// createPool 创建新池
	createPoolFn := NativeFunction(func(args []Value) (Value, error) {
		opts := runtime.DefaultPoolOptions()

		if len(args) > 0 {
			if optsObj, ok := args[0].(Object); ok {
				if v, ok := optsObj["minWorkers"]; ok {
					if n, ok := v.(float64); ok {
						opts.MinWorkers = int(n)
					}
				}
				if v, ok := optsObj["maxWorkers"]; ok {
					if n, ok := v.(float64); ok {
						opts.MaxWorkers = int(n)
					}
				}
				if v, ok := optsObj["queueSize"]; ok {
					if n, ok := v.(float64); ok {
						opts.QueueSize = int(n)
					}
				}
				if v, ok := optsObj["maxRetries"]; ok {
					if n, ok := v.(float64); ok {
						opts.MaxRetries = int(n)
					}
				}
			}
		}

		pool, err := runtime.NewGoroutinePool(opts)
		if err != nil {
			return nil, err
		}

		if err := pool.Start(); err != nil {
			return nil, err
		}

		return createPoolObject(pool), nil
	})

	// shutdown 关闭默认池
	shutdownFn := NativeFunction(func(args []Value) (Value, error) {
		timeout := 30 * time.Second
		if len(args) > 0 {
			if ms, ok := args[0].(float64); ok {
				timeout = time.Duration(ms) * time.Millisecond
			}
		}
		
		err := poolManager.ShutdownWithTimeout(timeout)
		if err != nil {
			return false, err
		}
		return true, nil
	})

	return Object{
		"submit":           submitFn,
		"submitWithRetry":  submitWithRetryFn,
		"getStats":         getStatsFn,
		"createPool":       createPoolFn,
		"shutdown":         shutdownFn,
	}
}

// createPoolObject 创建池对象
func createPoolObject(pool *runtime.GoroutinePool) Object {
	// submit 方法
	submitFn := NativeFunction(func(args []Value) (Value, error) {
		if len(args) == 0 {
			return nil, fmt.Errorf("pool.submit expects at least 1 arg: callback")
		}

		callback, ok := args[0].(NativeFunction)
		if !ok {
			return nil, fmt.Errorf("pool.submit arg[0] expects function, got %T", args[0])
		}

		task := pool.Submit(func() (Value, error) {
			return callback([]Value{})
		})

		return task, nil
	})

	// getStats 方法
	getStatsFn := NativeFunction(func(args []Value) (Value, error) {
		stats := pool.GetStats()

		return Object{
			"activeWorkers":  float64(stats.ActiveWorkers),
			"idleWorkers":    float64(stats.IdleWorkers),
			"totalWorkers":   float64(stats.TotalWorkers),
			"queuedTasks":    float64(stats.QueuedTasks),
			"completedTasks": float64(stats.CompletedTasks),
			"failedTasks":    float64(stats.FailedTasks),
			"rejectedTasks":  float64(stats.RejectedTasks),
			"maxConcurrency": float64(stats.MaxConcurrency),
			"currentLoad":    stats.CurrentLoad,
		}, nil
	})

	// shutdown 方法
	shutdownFn := NativeFunction(func(args []Value) (Value, error) {
		timeout := 30 * time.Second
		if len(args) > 0 {
			if ms, ok := args[0].(float64); ok {
				timeout = time.Duration(ms) * time.Millisecond
			}
		}
		return pool.ShutdownWithTimeout(timeout) == nil, nil
	})

	return Object{
		"submit":   submitFn,
		"getStats": getStatsFn,
		"shutdown": shutdownFn,
	}
}
