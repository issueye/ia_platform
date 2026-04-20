package runtime

import (
	"ialang/pkg/pool"
)

// 转发所有类型
type PoolStats = pool.PoolStats
type PoolOptions = pool.PoolOptions
type Task = pool.Task
type Worker = pool.Worker
type GoroutinePool = pool.GoroutinePool
type PoolAsyncRuntime = pool.PoolAsyncRuntime
type PoolManager = pool.PoolManager
type PoolType = pool.PoolType
type GlobalPoolStats = pool.GlobalPoolStats
type PoolConfig = pool.PoolConfig
type PoolManagerOptions = pool.PoolManagerOptions

// 转发所有常量
const (
	DefaultPool      = pool.DefaultPool
	CPUPool          = pool.CPUPool
	IOPool           = pool.IOPool
	HighPriorityPool = pool.HighPriorityPool
)

// 转发所有函数和变量
var (
	DefaultPoolOptions       = pool.DefaultPoolOptions
	DefaultPoolManagerOptions = pool.DefaultPoolManagerOptions
	NewGoroutinePool         = pool.NewGoroutinePool
	NewPoolAsyncRuntime      = pool.NewPoolAsyncRuntime
	GetPoolManager           = pool.GetPoolManager
	NewPoolManager           = pool.NewPoolManager
)
