package runtime

import (
	"context"
	"errors"
	"fmt"
	"ialang/pkg/pool"
	rttypes "ialang/pkg/lang/runtime/types"
	"time"
)

type Awaitable = rttypes.Awaitable
type ContextAwaitable = rttypes.ContextAwaitable
type AsyncTask = rttypes.AsyncTask
type AsyncRuntime = rttypes.AsyncRuntime

var (
	ErrAsyncTaskTimeout  = errors.New("async task timeout")
	ErrAsyncAwaitTimeout = errors.New("async await timeout")
)

const (
	AsyncErrorCodeTaskTimeout  = "ASYNC_TASK_TIMEOUT"
	AsyncErrorCodeAwaitTimeout = "ASYNC_AWAIT_TIMEOUT"
	AsyncErrorKindTimeout      = "timeout"
	RuntimeErrorCodeGeneric    = "RUNTIME_ERROR"
	RuntimeErrorKindGeneric    = "runtime"
)

type GoroutineRuntimeOptions struct {
	// TaskTimeout applies to each async task spawned by Spawn.
	// Zero means no timeout.
	TaskTimeout time.Duration
	// AwaitTimeout applies to each AwaitValue call.
	// Zero means no timeout.
	AwaitTimeout time.Duration
}

type GoroutineRuntime struct {
	options GoroutineRuntimeOptions
}

func NewGoroutineRuntime() AsyncRuntime {
	return NewGoroutineRuntimeWithOptions(GoroutineRuntimeOptions{})
}

func NewGoroutineRuntimeWithOptions(options GoroutineRuntimeOptions) AsyncRuntime {
	return &GoroutineRuntime{options: options}
}

func (r *GoroutineRuntime) Spawn(task AsyncTask) Awaitable {
	// 尝试使用协程池执行
	pm := pool.GetPoolManager()
	usePool := pm.IsInitialized() && !pm.IsShutdown()

	if r.options.TaskTimeout > 0 {
		// 带超时的任务执行
		return NewPromise(func() (Value, error) {
			type taskResult struct {
				value Value
				err   error
			}
			
			// 如果可以使用协程池，通过池提交
			if usePool {
				poolTask, err := pm.Submit(pool.IOPool, func() (Value, error) {
					return task()
				})
				if err != nil {
					return nil, err
				}
				
				done := make(chan taskResult, 1)
				go func() {
					v, err := poolTask.Await()
					done <- taskResult{value: v, err: err}
				}()
				
				select {
				case result := <-done:
					return result.value, result.err
				case <-time.After(r.options.TaskTimeout):
					return nil, fmt.Errorf("%w after %s", ErrAsyncTaskTimeout, r.options.TaskTimeout)
				}
			}
			
			// Fallback: 使用原始 goroutine
			done := make(chan taskResult, 1)
			go func() {
				v, err := task()
				done <- taskResult{value: v, err: err}
			}()
			select {
			case result := <-done:
				return result.value, result.err
			case <-time.After(r.options.TaskTimeout):
				return nil, fmt.Errorf("%w after %s", ErrAsyncTaskTimeout, r.options.TaskTimeout)
			}
		})
	}
	
	// 不带超时的任务执行
	if usePool {
		// 使用协程池
		poolTask, err := pm.Submit(pool.IOPool, task)
		if err != nil {
			// 如果提交失败，回退到原始实现
			return NewPromise(task)
		}
		return poolTask
	}
	
	// Fallback: 使用原始 Promise（创建 goroutine）
	return NewPromise(task)
}

func (r *GoroutineRuntime) AwaitValue(v Value) (Value, error) {
	awaitable, ok := v.(Awaitable)
	if !ok {
		return v, nil
	}
	if r.options.AwaitTimeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), r.options.AwaitTimeout)
		defer cancel()
		if contextAwaitable, ok := awaitable.(ContextAwaitable); ok {
			resolved, err := contextAwaitable.AwaitContext(ctx)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					return nil, fmt.Errorf("%w after %s", ErrAsyncAwaitTimeout, r.options.AwaitTimeout)
				}
				return nil, err
			}
			return resolved, nil
		}

		type awaitResult struct {
			value Value
			err   error
		}
		done := make(chan awaitResult, 1)
		go func() {
			resolved, err := awaitable.Await()
			done <- awaitResult{value: resolved, err: err}
		}()
		select {
		case result := <-done:
			return result.value, result.err
		case <-ctx.Done():
			return nil, fmt.Errorf("%w after %s", ErrAsyncAwaitTimeout, r.options.AwaitTimeout)
		}
	}
	return awaitable.Await()
}

func (r *GoroutineRuntime) Name() string {
	return "goroutine"
}

func NewAsyncRuntimeErrorValue(name, code, kind, message, runtimeName string, retryable bool) Object {
	return Object{
		"name":      name,
		"code":      code,
		"kind":      kind,
		"message":   message,
		"runtime":   runtimeName,
		"retryable": retryable,
	}
}
