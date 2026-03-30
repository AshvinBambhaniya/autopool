package autopool

import (
	"context"
	"sync"
	"testing"
)

// BenchmarkStandardGoroutines measures the cost of spawning a new goroutine for every task.
func BenchmarkStandardGoroutines(b *testing.B) {
	for n := 0; n < b.N; n++ {
		var wg sync.WaitGroup
		wg.Add(1000)
		for i := 0; i < 1000; i++ {
			go func() {
				defer wg.Done()
				// perform tiny work
				_ = 1 + 1
			}()
		}
		wg.Wait()
	}
}

// BenchmarkAutopool measures the cost of using the worker pool for the same tasks.
func BenchmarkAutopool(b *testing.B) {
	p := New(
		WithMinWorkers(10),
		WithMaxWorkers(50),
		WithQueueSize(1000),
	)
	defer p.Shutdown(context.Background())

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var wg sync.WaitGroup
		wg.Add(1000)
		for i := 0; i < 1000; i++ {
			p.Submit(func(ctx context.Context) error {
				defer wg.Done()
				// perform tiny work
				_ = 1 + 1
				return nil
			})
		}
		wg.Wait()
	}
}

// BenchmarkAutopoolHighConcurrency simulates heavy concurrent submission to the pool.
func BenchmarkAutopoolHighConcurrency(b *testing.B) {
	p := New(
		WithMinWorkers(50),
		WithMaxWorkers(100),
		WithQueueSize(5000),
	)
	defer p.Shutdown(context.Background())

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var wg sync.WaitGroup
			wg.Add(100)
			for i := 0; i < 100; i++ {
				p.Submit(func(ctx context.Context) error {
					defer wg.Done()
					return nil
				})
			}
			wg.Wait()
		}
	})
}
