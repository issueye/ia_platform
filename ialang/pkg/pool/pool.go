package pool

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	commonrt "iacommon/pkg/ialang/runtime"
)

// PoolStats 协程池统计信息
type PoolStats struct {
	ActiveWorkers  int     `json:"activeWorkers"`  // 活跃工作协程数
	IdleWorkers    int     `json:"idleWorkers"`    // 空闲工作协程数
	TotalWorkers   int     `json:"totalWorkers"`   // 总工作协程数
	QueuedTasks    int     `json:"queuedTasks"`    // 队列中等待的任务数
	CompletedTasks int64   `json:"completedTasks"` // 已完成任务总数
	FailedTasks    int64   `json:"failedTasks"`    // 失败任务总数
	RejectedTasks  int64   `json:"rejectedTasks"`  // 被拒绝的任务数
	MaxConcurrency int     `json:"maxConcurrency"` // 最大并发数
	CurrentLoad    float64 `json:"currentLoad"`    // 当前负载百分比 (0-100)
}

// PoolOptions 协程池配置选项
type PoolOptions struct {
	MinWorkers      int           `json:"minWorkers"`      // 最小工作协程数
	MaxWorkers      int           `json:"maxWorkers"`      // 最大工作协程数
	QueueSize       int           `json:"queueSize"`       // 任务队列大小
	IdleTimeout     time.Duration `json:"idleTimeout"`     // 空闲超时时间
	MaxRetries      int           `json:"maxRetries"`      // 任务最大重试次数
	RejectPolicy    string        `json:"rejectPolicy"`    // 拒绝策略: "block", "discard", "error"
	TrackTaskTiming bool          `json:"trackTaskTiming"` // 是否记录任务时间戳
}

// DefaultPoolOptions 默认协程池配置
func DefaultPoolOptions() PoolOptions {
	return PoolOptions{
		MinWorkers:      runtime.NumCPU(),
		MaxWorkers:      runtime.NumCPU() * 10,
		QueueSize:       1000,
		IdleTimeout:     30 * time.Second,
		MaxRetries:      3,
		RejectPolicy:    "block",
		TrackTaskTiming: true,
	}
}

// Task 任务定义
type Task struct {
	ID          string
	Func        func() (commonrt.Value, error)
	Result      commonrt.Value
	Err         error
	Retries     int
	MaxRetries  int
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	wg          sync.WaitGroup
	doneFlag    uint32
	trackTiming bool
}

// Worker 工作协程
type Worker struct {
	ID         int
	TaskChan   chan *Task
	ResultChan chan *Task
	QuitChan   chan bool
	IsActive   bool
	LastActive time.Time
	pool       *GoroutinePool
}

// GoroutinePool 统一的协程池管理器
type GoroutinePool struct {
	mu             sync.RWMutex
	workers        []*Worker
	taskQueue      chan *Task
	resultQueue    chan *Task
	options        PoolOptions
	activeCount    int32
	workerCount    int32
	completedCount int64
	failedCount    int64
	rejectedCount  int64
	isShutdown     bool
	isStarted      bool
	startTime      time.Time
	wg             sync.WaitGroup

	// 监控回调
	OnTaskComplete func(task *Task)
	OnTaskFailed   func(task *Task)
	OnPoolShutdown func()
}

// NewGoroutinePool 创建新的协程池
func NewGoroutinePool(options PoolOptions) (*GoroutinePool, error) {
	if options.MinWorkers < 0 {
		return nil, fmt.Errorf("minWorkers must be >= 0")
	}
	if options.MaxWorkers < options.MinWorkers {
		return nil, fmt.Errorf("maxWorkers must be >= minWorkers")
	}
	if options.QueueSize <= 0 {
		return nil, fmt.Errorf("queueSize must be > 0")
	}

	pool := &GoroutinePool{
		workers:     make([]*Worker, 0, options.MaxWorkers),
		taskQueue:   make(chan *Task, options.QueueSize),
		resultQueue: make(chan *Task, options.QueueSize),
		options:     options,
		startTime:   time.Now(),
	}

	return pool, nil
}

// Start 启动协程池
func (p *GoroutinePool) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isStarted {
		return fmt.Errorf("pool already started")
	}

	// 启动最小工作协程
	for i := 0; i < p.options.MinWorkers; i++ {
		p.addWorkerLocked()
	}

	p.isStarted = true
	p.isShutdown = false

	return nil
}

// addWorker 添加工作协程（外部调用，会获取锁）
func (p *GoroutinePool) addWorker() *Worker {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.addWorkerLocked()
}

// addWorkerLocked 添加工作协程（内部调用，假设已持有锁）
func (p *GoroutinePool) addWorkerLocked() *Worker {
	if len(p.workers) >= p.options.MaxWorkers {
		return nil
	}

	worker := &Worker{
		ID:         len(p.workers),
		TaskChan:   make(chan *Task, 1),
		ResultChan: make(chan *Task, 1),
		QuitChan:   make(chan bool),
		pool:       p,
	}

	p.workers = append(p.workers, worker)
	atomic.StoreInt32(&p.workerCount, int32(len(p.workers)))
	p.wg.Add(1)

	go worker.start()

	return worker
}

// removeIdleWorker 移除空闲工作协程
func (p *GoroutinePool) removeIdleWorker() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := len(p.workers) - 1; i >= p.options.MinWorkers; i-- {
		worker := p.workers[i]
		if time.Since(worker.LastActive) > p.options.IdleTimeout && !worker.IsActive {
			close(worker.QuitChan)
			p.workers = append(p.workers[:i], p.workers[i+1:]...)
			atomic.StoreInt32(&p.workerCount, int32(len(p.workers)))
			break
		}
	}
}

// Submit 提交任务到协程池
func (p *GoroutinePool) Submit(taskFunc func() (commonrt.Value, error)) commonrt.Awaitable {
	return p.SubmitWithRetry(taskFunc, p.options.MaxRetries)
}

// SubmitWithRetry 提交任务并支持重试
func (p *GoroutinePool) SubmitWithRetry(taskFunc func() (commonrt.Value, error), maxRetries int) commonrt.Awaitable {
	task := &Task{
		Func:        taskFunc,
		MaxRetries:  maxRetries,
		trackTiming: p.options.TrackTaskTiming,
	}
	if task.trackTiming {
		task.CreatedAt = time.Now()
	}
	task.wg.Add(1)

	// 检查是否已关闭
	p.mu.RLock()
	if p.isShutdown {
		p.mu.RUnlock()
		task.Err = fmt.Errorf("pool is shutdown")
		task.markDone()
		return task
	}
	p.mu.RUnlock()

	// 尝试提交任务
	select {
	case p.taskQueue <- task:
		// 如果活跃协程不足，动态添加
		p.ensureWorkers()
		return task
	default:
		// 队列已满，根据策略处理
		switch p.options.RejectPolicy {
		case "error":
			task.Err = fmt.Errorf("task queue is full")
			atomic.AddInt64(&p.rejectedCount, 1)
			task.markDone()
			return task
		case "discard":
			task.Err = fmt.Errorf("task discarded")
			atomic.AddInt64(&p.rejectedCount, 1)
			task.markDone()
			return task
		case "block":
			fallthrough
		default:
			// 阻塞等待
			p.taskQueue <- task
			return task
		}
	}
}

// ensureWorkers 确保有足够的工作协程
func (p *GoroutinePool) ensureWorkers() {
	total := atomic.LoadInt32(&p.workerCount)
	if int(total) < p.options.MinWorkers {
		p.addWorker()
	}
}

// GetStats 获取协程池统计信息
func (p *GoroutinePool) GetStats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	totalWorkers := len(p.workers)
	activeWorkers := atomic.LoadInt32(&p.activeCount)
	idleWorkers := totalWorkers - int(activeWorkers)

	var maxConcurrency int
	if p.options.MaxWorkers > 0 {
		maxConcurrency = p.options.MaxWorkers
	}

	var currentLoad float64
	if maxConcurrency > 0 {
		currentLoad = float64(activeWorkers) / float64(maxConcurrency) * 100
	}

	return PoolStats{
		ActiveWorkers:  int(activeWorkers),
		IdleWorkers:    idleWorkers,
		TotalWorkers:   totalWorkers,
		QueuedTasks:    len(p.taskQueue),
		CompletedTasks: atomic.LoadInt64(&p.completedCount),
		FailedTasks:    atomic.LoadInt64(&p.failedCount),
		RejectedTasks:  atomic.LoadInt64(&p.rejectedCount),
		MaxConcurrency: maxConcurrency,
		CurrentLoad:    currentLoad,
	}
}

// Shutdown 优雅关闭协程池
func (p *GoroutinePool) Shutdown() error {
	return p.ShutdownWithTimeout(30 * time.Second)
}

// ShutdownWithTimeout 带超时的优雅关闭
func (p *GoroutinePool) ShutdownWithTimeout(timeout time.Duration) error {
	p.mu.Lock()
	if p.isShutdown {
		p.mu.Unlock()
		return fmt.Errorf("pool already shutdown")
	}
	p.isShutdown = true
	// 关闭任务队列，通知所有 Worker 退出
	close(p.taskQueue)
	p.mu.Unlock()

	// 等待所有任务完成
	done := make(chan bool)
	go func() {
		p.wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		if p.OnPoolShutdown != nil {
			p.OnPoolShutdown()
		}
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("shutdown timeout after %v", timeout)
	}
}

// Start 启动工作协程
func (w *Worker) start() {
	defer w.pool.wg.Done()

	for {
		select {
		case task, ok := <-w.pool.taskQueue:
			if !ok {
				// channel 已关闭，退出
				return
			}
			w.IsActive = true
			w.LastActive = time.Now()
			atomic.AddInt32(&w.pool.activeCount, 1)

			// 执行任务
			w.executeTask(task)

			atomic.AddInt32(&w.pool.activeCount, -1)
			w.IsActive = false
			w.LastActive = time.Now()

		case <-w.QuitChan:
			return
		}
	}
}

// executeTask 执行任务
func (w *Worker) executeTask(task *Task) {
	if task.trackTiming {
		task.StartedAt = time.Now()
	}

	// 重试循环
	for task.Retries <= task.MaxRetries {
		task.Result, task.Err = task.Func()

		if task.Err == nil {
			// 成功
			atomic.AddInt64(&w.pool.completedCount, 1)
			task.markDone()

			if w.pool.OnTaskComplete != nil {
				w.pool.OnTaskComplete(task)
			}
			return
		}

		// 失败，准备重试
		task.Retries++
		if task.Retries <= task.MaxRetries {
			time.Sleep(time.Duration(task.Retries) * 100 * time.Millisecond)
		}
	}

	// 所有重试失败
	atomic.AddInt64(&w.pool.failedCount, 1)
	task.markDone()

	if w.pool.OnTaskFailed != nil {
		w.pool.OnTaskFailed(task)
	}
}

// PoolAsyncRuntime 基于协程池的 AsyncRuntime 实现
type PoolAsyncRuntime struct {
	pool *GoroutinePool
	name string
}

// NewPoolAsyncRuntime 创建基于协程池的 AsyncRuntime
func NewPoolAsyncRuntime(options PoolOptions) (*PoolAsyncRuntime, error) {
	pool, err := NewGoroutinePool(options)
	if err != nil {
		return nil, err
	}

	if err := pool.Start(); err != nil {
		return nil, err
	}

	return &PoolAsyncRuntime{
		pool: pool,
		name: "PoolRuntime",
	}, nil
}

// Spawn 提交任务到协程池
func (r *PoolAsyncRuntime) Spawn(task commonrt.AsyncTask) commonrt.Awaitable {
	return r.pool.Submit(task)
}

// AwaitValue 等待值完成
func (r *PoolAsyncRuntime) AwaitValue(v commonrt.Value) (commonrt.Value, error) {
	if awaitable, ok := v.(commonrt.Awaitable); ok {
		return awaitable.Await()
	}
	return v, nil
}

// Name 获取运行时名称
func (r *PoolAsyncRuntime) Name() string {
	return r.name
}

// GetPool 获取底层协程池
func (r *PoolAsyncRuntime) GetPool() *GoroutinePool {
	return r.pool
}

// Task 实现 Awaitable 接口
var _ commonrt.Awaitable = (*Task)(nil)

// Await 等待任务完成
func (t *Task) Await() (commonrt.Value, error) {
	t.wg.Wait()
	return t.Result, t.Err
}

// IsDone 检查任务是否完成
func (t *Task) IsDone() bool {
	return atomic.LoadUint32(&t.doneFlag) == 1
}

func (t *Task) markDone() {
	if t.trackTiming {
		t.CompletedAt = time.Now()
	}
	atomic.StoreUint32(&t.doneFlag, 1)
	t.wg.Done()
}
