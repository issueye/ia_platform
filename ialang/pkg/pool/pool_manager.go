package pool

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	rttypes "ialang/pkg/lang/runtime/types"
)

// PoolType 协程池的类型
type PoolType string

const (
	// DefaultPool 默认通用池
	DefaultPool PoolType = "default"
	// CPUPool CPU密集型任务池
	CPUPool PoolType = "cpu"
	// IOPool IO密集型任务池
	IOPool PoolType = "io"
	// HighPriorityPool 高优先级任务池
	HighPriorityPool PoolType = "high_priority"
)

// PoolConfig 协程池配置
type PoolConfig struct {
	Type     PoolType      `json:"type"`     // 池类型
	Options  PoolOptions   `json:"options"`  // 池选项
	Enabled  bool          `json:"enabled"`  // 是否启用
}

// PoolManager 统一的协程池管理器
type PoolManager struct {
	mu            sync.RWMutex
	pools         map[PoolType]*GoroutinePool
	configs       map[PoolType]PoolConfig
	isInitialized bool
	isShutdown    bool

	// 全局统计
	totalSubmitted  int64
	totalCompleted  int64
	totalFailed     int64
	totalRejected   int64

	// 监控回调
	OnTaskSubmit   func(poolType PoolType, task *Task)
	OnTaskComplete func(poolType PoolType, task *Task)
	OnTaskFailed   func(poolType PoolType, task *Task)
	OnPoolCreated  func(poolType PoolType, pool *GoroutinePool)
	OnPoolShutdown func(poolType PoolType)

	// 关闭超时
	shutdownTimeout time.Duration
}

// globalPoolManager 全局协程池管理器实例
var globalPoolManager *PoolManager
var globalPoolManagerOnce sync.Once

// PoolManagerOptions 协程池管理器选项
type PoolManagerOptions struct {
	ShutdownTimeout time.Duration
	EnableDefault   bool
	EnableCPUPool   bool
	EnableIOPool    bool
	EnableHighPriorityPool bool
}

// DefaultPoolManagerOptions 默认管理器选项
func DefaultPoolManagerOptions() PoolManagerOptions {
	return PoolManagerOptions{
		ShutdownTimeout:    30 * time.Second,
		EnableDefault:      true,
		EnableCPUPool:      false,
		EnableIOPool:       false,
		EnableHighPriorityPool: false,
	}
}

// GetPoolManager 获取全局协程池管理器（单例）
func GetPoolManager() *PoolManager {
	globalPoolManagerOnce.Do(func() {
		globalPoolManager = &PoolManager{
			pools:           make(map[PoolType]*GoroutinePool),
			configs:         make(map[PoolType]PoolConfig),
			shutdownTimeout: 30 * time.Second,
		}
	})
	return globalPoolManager
}

// NewPoolManager 创建新的协程池管理器
func NewPoolManager(opts PoolManagerOptions) *PoolManager {
	pm := &PoolManager{
		pools:           make(map[PoolType]*GoroutinePool),
		configs:         make(map[PoolType]PoolConfig),
		shutdownTimeout: opts.ShutdownTimeout,
	}
	pm.applyOptionsLocked(opts)

	return pm
}

// RegisterPool 注册一个协程池配置
func (pm *PoolManager) RegisterPool(config PoolConfig) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.isShutdown {
		return fmt.Errorf("pool manager is shutdown")
	}

	pm.configs[config.Type] = config
	return nil
}

// Initialize 初始化所有已注册的协程池
func (pm *PoolManager) Initialize() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.isInitialized {
		return fmt.Errorf("pool manager already initialized")
	}
	if pm.isShutdown {
		return fmt.Errorf("pool manager is shutdown")
	}
	return pm.initializeEnabledPoolsLocked()
}

// GetPool 获取指定类型的协程池
func (pm *PoolManager) GetPool(poolType PoolType) (*GoroutinePool, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if !pm.isInitialized {
		return nil, fmt.Errorf("pool manager not initialized")
	}

	pool, exists := pm.pools[poolType]
	if !exists {
		return nil, fmt.Errorf("pool %s not found", poolType)
	}

	return pool, nil
}

// Submit 提交任务到指定类型的池
func (pm *PoolManager) Submit(poolType PoolType, taskFunc func() (rttypes.Value, error)) (rttypes.Awaitable, error) {
	pool, err := pm.GetPool(poolType)
	if err != nil {
		return nil, err
	}

	// 包装任务以统计
	wrappedTask := func() (rttypes.Value, error) {
		atomic.AddInt64(&pm.totalSubmitted, 1)
		return taskFunc()
	}

	task := pool.Submit(wrappedTask)

	if pm.OnTaskSubmit != nil {
		if t, ok := task.(*Task); ok {
			pm.OnTaskSubmit(poolType, t)
		}
	}

	return task, nil
}

// SubmitWithRetry 提交任务到指定类型的池（带重试）
func (pm *PoolManager) SubmitWithRetry(poolType PoolType, taskFunc func() (rttypes.Value, error), maxRetries int) (rttypes.Awaitable, error) {
	pool, err := pm.GetPool(poolType)
	if err != nil {
		return nil, err
	}

	// 包装任务以统计
	wrappedTask := func() (rttypes.Value, error) {
		atomic.AddInt64(&pm.totalSubmitted, 1)
		return taskFunc()
	}

	task := pool.SubmitWithRetry(wrappedTask, maxRetries)

	if pm.OnTaskSubmit != nil {
		if t, ok := task.(*Task); ok {
			pm.OnTaskSubmit(poolType, t)
		}
	}

	return task, nil
}

// GetGlobalStats 获取全局统计信息
func (pm *PoolManager) GetGlobalStats() GlobalPoolStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	stats := GlobalPoolStats{
		TotalSubmitted: atomic.LoadInt64(&pm.totalSubmitted),
		TotalCompleted: atomic.LoadInt64(&pm.totalCompleted),
		TotalFailed:    atomic.LoadInt64(&pm.totalFailed),
		TotalRejected:  atomic.LoadInt64(&pm.totalRejected),
		PoolStats:      make(map[PoolType]PoolStats),
		TotalPools:     len(pm.pools),
		ActivePools:    0,
	}

	for poolType, pool := range pm.pools {
		poolStats := pool.GetStats()
		stats.PoolStats[poolType] = poolStats
		if poolStats.ActiveWorkers > 0 {
			stats.ActivePools++
		}
		stats.TotalWorkers += poolStats.TotalWorkers
		stats.ActiveWorkers += poolStats.ActiveWorkers
		stats.QueuedTasks += poolStats.QueuedTasks
	}

	return stats
}

// GetPoolStats 获取指定池的统计信息
func (pm *PoolManager) GetPoolStats(poolType PoolType) (PoolStats, error) {
	pool, err := pm.GetPool(poolType)
	if err != nil {
		return PoolStats{}, err
	}
	return pool.GetStats(), nil
}

// Shutdown 优雅关闭所有协程池
func (pm *PoolManager) Shutdown() error {
	return pm.ShutdownWithTimeout(pm.shutdownTimeout)
}

// ShutdownWithTimeout 带超时的优雅关闭
func (pm *PoolManager) ShutdownWithTimeout(timeout time.Duration) error {
	pm.mu.Lock()
	if pm.isShutdown {
		pm.mu.Unlock()
		return fmt.Errorf("pool manager already shutdown")
	}
	if !pm.isInitialized {
		pm.isShutdown = true
		pm.mu.Unlock()
		return nil
	}
	pm.isShutdown = true

	// 复制池引用，避免在关闭过程中修改
	pools := make(map[PoolType]*GoroutinePool)
	for poolType, pool := range pm.pools {
		pools[poolType] = pool
	}
	pm.mu.Unlock()

	// 并发关闭所有池
	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex

	for poolType, pool := range pools {
		wg.Add(1)
		go func(pt PoolType, p *GoroutinePool) {
			defer wg.Done()

			err := p.ShutdownWithTimeout(timeout)
			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("pool %s shutdown error: %w", pt, err)
				}
				errMu.Unlock()
			}
		}(poolType, pool)
	}

	wg.Wait()

	if firstErr != nil {
		return firstErr
	}

	pm.mu.Lock()
	pm.pools = make(map[PoolType]*GoroutinePool)
	pm.mu.Unlock()

	return nil
}

// IsInitialized 检查是否已初始化
func (pm *PoolManager) IsInitialized() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.isInitialized
}

// IsShutdown 检查是否已关闭
func (pm *PoolManager) IsShutdown() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.isShutdown
}

// PoolCount 获取池数量
func (pm *PoolManager) PoolCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.pools)
}

// GlobalPoolStats 全局协程池统计信息
type GlobalPoolStats struct {
	TotalSubmitted int64              `json:"totalSubmitted"` // 总提交任务数
	TotalCompleted int64              `json:"totalCompleted"` // 总完成任务数
	TotalFailed    int64              `json:"totalFailed"`    // 总失败任务数
	TotalRejected  int64              `json:"totalRejected"`  // 总拒绝任务数
	TotalPools     int                `json:"totalPools"`     // 总池数
	ActivePools    int                `json:"activePools"`    // 活跃池数
	TotalWorkers   int                `json:"totalWorkers"`   // 总工作协程数
	ActiveWorkers  int                `json:"activeWorkers"`  // 总活跃工作协程数
	QueuedTasks    int                `json:"queuedTasks"`    // 总队列任务数
	PoolStats      map[PoolType]PoolStats `json:"poolStats"`  // 各池详细统计
}

// CreateAsyncRuntime 基于指定池类型创建 AsyncRuntime
func (pm *PoolManager) CreateAsyncRuntime(poolType PoolType) (*PoolAsyncRuntime, error) {
	pool, err := pm.GetPool(poolType)
	if err != nil {
		return nil, err
	}

	return &PoolAsyncRuntime{
		pool: pool,
		name: string(poolType),
	}, nil
}

// CreateDefaultAsyncRuntime 创建基于默认池的 AsyncRuntime
func (pm *PoolManager) CreateDefaultAsyncRuntime() (*PoolAsyncRuntime, error) {
	return pm.CreateAsyncRuntime(DefaultPool)
}

// EnsureInitialized 确保池管理器已初始化，如果未初始化则使用默认配置初始化
func (pm *PoolManager) EnsureInitialized() error {
	if pm.IsInitialized() {
		return nil
	}
	if pm.IsShutdown() {
		return fmt.Errorf("pool manager is shutdown")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 双重检查
	if pm.isInitialized {
		return nil
	}

	// 如果没有配置，使用默认配置
	if len(pm.configs) == 0 {
		pm.applyOptionsLocked(DefaultPoolManagerOptions())
	}

	return pm.initializeEnabledPoolsLocked()
}

// EnsureInitializedWithOptions 确保池管理器使用指定 options 完成初始化。
// 如果管理器已初始化，则保持现状并返回 nil。
func (pm *PoolManager) EnsureInitializedWithOptions(opts PoolManagerOptions) error {
	if pm.IsInitialized() {
		return nil
	}
	if pm.IsShutdown() {
		return fmt.Errorf("pool manager is shutdown")
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.isInitialized {
		return nil
	}

	pm.applyOptionsLocked(opts)
	return pm.initializeEnabledPoolsLocked()
}

func (pm *PoolManager) initializeEnabledPoolsLocked() error {
	// 初始化所有启用的池
	for poolType, config := range pm.configs {
		if !config.Enabled {
			continue
		}

		pool, err := NewGoroutinePool(config.Options)
		if err != nil {
			return fmt.Errorf("failed to create pool %s: %w", poolType, err)
		}

		if err := pool.Start(); err != nil {
			return fmt.Errorf("failed to start pool %s: %w", poolType, err)
		}

		// 设置监控回调
		poolTypeCopy := poolType // 避免闭包捕获问题
		pool.OnTaskComplete = func(task *Task) {
			atomic.AddInt64(&pm.totalCompleted, 1)
			if pm.OnTaskComplete != nil {
				pm.OnTaskComplete(poolTypeCopy, task)
			}
		}
		pool.OnTaskFailed = func(task *Task) {
			atomic.AddInt64(&pm.totalFailed, 1)
			if pm.OnTaskFailed != nil {
				pm.OnTaskFailed(poolTypeCopy, task)
			}
		}
		pool.OnPoolShutdown = func() {
			if pm.OnPoolShutdown != nil {
				pm.OnPoolShutdown(poolTypeCopy)
			}
		}

		pm.pools[poolType] = pool

		if pm.OnPoolCreated != nil {
			pm.OnPoolCreated(poolType, pool)
		}
	}

	pm.isInitialized = true
	return nil
}

func (pm *PoolManager) applyOptionsLocked(opts PoolManagerOptions) {
	if opts.ShutdownTimeout > 0 {
		pm.shutdownTimeout = opts.ShutdownTimeout
	}

	pm.configs = make(map[PoolType]PoolConfig)

	// 注册默认池配置
	if opts.EnableDefault {
		pm.configs[DefaultPool] = PoolConfig{
			Type:    DefaultPool,
			Enabled: true,
			Options: DefaultPoolOptions(),
		}
	}

	// 注册CPU密集型池
	if opts.EnableCPUPool {
		pm.configs[CPUPool] = PoolConfig{
			Type:    CPUPool,
			Enabled: true,
			Options: PoolOptions{
				MinWorkers:   runtime.NumCPU(),
				MaxWorkers:   runtime.NumCPU() * 2,
				QueueSize:    500,
				IdleTimeout:  60 * time.Second,
				MaxRetries:   1,
				RejectPolicy: "error",
			},
		}
	}

	// 注册IO密集型池
	if opts.EnableIOPool {
		pm.configs[IOPool] = PoolConfig{
			Type:    IOPool,
			Enabled: true,
			Options: PoolOptions{
				MinWorkers:   runtime.NumCPU() * 2,
				MaxWorkers:   runtime.NumCPU() * 20,
				QueueSize:    5000,
				IdleTimeout:  120 * time.Second,
				MaxRetries:   3,
				RejectPolicy: "block",
			},
		}
	}

	// 注册高优先级池
	if opts.EnableHighPriorityPool {
		pm.configs[HighPriorityPool] = PoolConfig{
			Type:    HighPriorityPool,
			Enabled: true,
			Options: PoolOptions{
				MinWorkers:   2,
				MaxWorkers:   runtime.NumCPU(),
				QueueSize:    100,
				IdleTimeout:  30 * time.Second,
				MaxRetries:   0,
				RejectPolicy: "error",
			},
		}
	}
}
