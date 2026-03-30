package autopool

import (
	"context"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool_LeakDetection(t *testing.T) {
	// 1. Capture initial goroutine count
	// Give some time for background runtime goroutines to settle
	time.Sleep(100 * time.Millisecond)
	initialGoroutines := runtime.NumGoroutine()

	t.Logf("Initial goroutines: %d", initialGoroutines)

	// 2. Run the pool through a full cycle
	p := New(
		WithMinWorkers(5),
		WithMaxWorkers(20),
		WithQueueSize(50),
	)

	for i := 0; i < 100; i++ {
		p.Submit(func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
	}

	// 3. Shutdown the pool
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	// 4. Verify goroutine count returns to baseline
	// Allow a tiny bit of time for final worker cleanup
	time.Sleep(200 * time.Millisecond)
	finalGoroutines := runtime.NumGoroutine()

	t.Logf("Final goroutines: %d", finalGoroutines)

	// We allow a small margin (e.g., 2) for runtime/testing background noise
	if finalGoroutines > initialGoroutines+2 {
		t.Errorf("Potential leak detected: Initial=%d, Final=%d", initialGoroutines, finalGoroutines)
	}
}

func TestPool_HighThroughputLoad(t *testing.T) {
	p := New(
		WithMinWorkers(10),
		WithMaxWorkers(50),
		WithQueueSize(1000),
	)

	const taskCount = 100000
	var completedTasks int64

	t.Logf("Submitting %d fast tasks...", taskCount)
	start := time.Now()

	for i := 0; i < taskCount; i++ {
		err := p.Submit(func(ctx context.Context) error {
			atomic.AddInt64(&completedTasks, 1)
			return nil
		})
		if err != nil {
			t.Errorf("Submit failed at task %d: %v", i, err)
			break
		}
	}

	// Wait for completion via Shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := p.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown during load test failed: %v", err)
	}

	duration := time.Since(start)
	finalCount := atomic.LoadInt64(&completedTasks)

	t.Logf("Processed %d tasks in %v (%.0f tasks/sec)", 
		finalCount, duration, float64(finalCount)/duration.Seconds())

	if finalCount != taskCount {
		t.Errorf("Expected %d tasks, but only %d completed", taskCount, finalCount)
	}
}
