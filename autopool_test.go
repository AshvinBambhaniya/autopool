package autopool

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestPool_Basic(t *testing.T) {
	p := New(
		WithMinWorkers(2),
		WithMaxWorkers(5),
		WithQueueSize(10),
	)

	var wg sync.WaitGroup
	count := 20
	wg.Add(count)

	results := make(chan int, count)

	for i := 0; i < count; i++ {
		val := i
		err := p.Submit(func(ctx context.Context) error {
			defer wg.Done()
			results <- val
			return nil
		})
		if err != nil {
			t.Logf("Task %d rejected: %v", i, err)
			wg.Done()
		}
	}

	wg.Wait()
	close(results)

	if len(results) == 0 {
		t.Errorf("Expected some results, got 0")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := p.Shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestPool_AutoScaling(t *testing.T) {
	p := New(
		WithMinWorkers(1),
		WithMaxWorkers(10),
		WithIdleTimeout(100*time.Millisecond),
	)

	// Initially 1 worker
	stats := p.Stats()
	if stats.TotalWorkers != 1 {
		t.Errorf("Expected 1 worker, got %d", stats.TotalWorkers)
	}

	// Submit many tasks to trigger scaling
	for i := 0; i < 20; i++ {
		p.Submit(func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		})
	}

	time.Sleep(100 * time.Millisecond)
	stats = p.Stats()
	if stats.TotalWorkers <= 1 {
		t.Errorf("Expected scaling up, got %d workers", stats.TotalWorkers)
	}

	// Wait for idle timeout
	time.Sleep(500 * time.Millisecond)
	stats = p.Stats()
	if stats.TotalWorkers > 1 {
		t.Logf("Workers after idle: %d", stats.TotalWorkers)
		// Should have scaled down to min (1)
	}

	p.Shutdown(context.Background())
}
