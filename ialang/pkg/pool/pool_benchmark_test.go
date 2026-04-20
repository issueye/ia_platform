package pool

import (
	"testing"

	rttypes "ialang/pkg/lang/runtime/types"
)

func BenchmarkGoroutinePoolSubmitAwait(b *testing.B) {
	benchPoolThroughput(b, 4, 64)
}

func BenchmarkGoroutinePoolSubmitAwaitParallel(b *testing.B) {
	opts := DefaultPoolOptions()
	opts.MinWorkers = 8
	opts.MaxWorkers = 8
	opts.QueueSize = 4096
	opts.TrackTaskTiming = false

	p, err := NewGoroutinePool(opts)
	if err != nil {
		b.Fatalf("NewGoroutinePool() error = %v", err)
	}
	if err := p.Start(); err != nil {
		b.Fatalf("Start() error = %v", err)
	}
	defer p.Shutdown()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			task := p.Submit(func() (rttypes.Value, error) {
				return 1, nil
			})
			if _, err := task.Await(); err != nil {
				b.Fatalf("task.Await() error = %v", err)
			}
		}
	})
}

func BenchmarkPoolManagerSubmitAwait(b *testing.B) {
	opts := DefaultPoolManagerOptions()
	opts.EnableDefault = true
	pm := NewPoolManager(opts)
	if err := pm.Initialize(); err != nil {
		b.Fatalf("Initialize() error = %v", err)
	}
	defer pm.Shutdown()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		task, err := pm.Submit(DefaultPool, func() (rttypes.Value, error) {
			return 1, nil
		})
		if err != nil {
			b.Fatalf("Submit() error = %v", err)
		}
		if _, err := task.Await(); err != nil {
			b.Fatalf("task.Await() error = %v", err)
		}
	}
}

func benchPoolThroughput(b *testing.B, workers int, batchSize int) {
	opts := DefaultPoolOptions()
	opts.MinWorkers = workers
	opts.MaxWorkers = workers
	opts.QueueSize = batchSize * 4
	opts.TrackTaskTiming = false

	p, err := NewGoroutinePool(opts)
	if err != nil {
		b.Fatalf("NewGoroutinePool() error = %v", err)
	}
	if err := p.Start(); err != nil {
		b.Fatalf("Start() error = %v", err)
	}
	defer p.Shutdown()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tasks := make([]rttypes.Awaitable, batchSize)
		for j := 0; j < batchSize; j++ {
			tasks[j] = p.Submit(func() (rttypes.Value, error) {
				return 1, nil
			})
		}
		for _, task := range tasks {
			if _, err := task.Await(); err != nil {
				b.Fatalf("task.Await() error = %v", err)
			}
		}
	}
}
