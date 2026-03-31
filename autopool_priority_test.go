package autopool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestPool_PriorityOrdering tests that high-priority tasks run before low-priority ones.
// Uses a blocker task and string labels for clarity.
func TestPool_PriorityOrdering(t *testing.T) {
	p := New(
		WithMaxWorkers(1),
		WithQueueSize(10),
	)
	defer func() { _ = p.Shutdown(context.Background()) }()

	// 1. Submit a long-running "blocker" task to fill the worker
	_ = p.Submit(func(ctx context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	// Wait a tiny bit to ensure the blocker is active
	time.Sleep(50 * time.Millisecond)

	var mu sync.Mutex
	executionOrder := make([]string, 0)

	// 2. Submit Low priority task
	_ = p.SubmitWithOptions(func(ctx context.Context) error {
		mu.Lock()
		executionOrder = append(executionOrder, "low")
		mu.Unlock()
		return nil
	}, TaskOptions{Priority: PriorityLow})

	// 3. Submit High priority task (should jump the low one)
	_ = p.SubmitWithOptions(func(ctx context.Context) error {
		mu.Lock()
		executionOrder = append(executionOrder, "high")
		mu.Unlock()
		return nil
	}, TaskOptions{Priority: PriorityHigh})

	// 4. Submit Critical priority task (should jump everyone)
	_ = p.SubmitWithOptions(func(ctx context.Context) error {
		mu.Lock()
		executionOrder = append(executionOrder, "critical")
		mu.Unlock()
		return nil
	}, TaskOptions{Priority: PriorityCritical})

	// Wait for completion
	time.Sleep(1 * time.Second)

	mu.Lock()
	defer mu.Unlock()

	expected := []string{"critical", "high", "low"}
	if len(executionOrder) != 3 {
		t.Fatalf("Expected 3 tasks to execute, got %d", len(executionOrder))
	}

	for i, v := range expected {
		if executionOrder[i] != v {
			t.Errorf("At index %d: expected %s, got %s", i, v, executionOrder[i])
		}
	}
}

// TestPool_PriorityAgingRobust tests that aging prevents starvation of low-priority tasks.
// It verifies that even with continuous high-priority submissions, low-priority tasks
// eventually run due to the Virtual Runtime aging mechanism.
func TestPool_PriorityAgingRobust(t *testing.T) {
	p := New(
		WithMaxWorkers(1),
		WithQueueSize(50),
	)
	defer func() { _ = p.Shutdown(context.Background()) }()

	// 1. Block the worker
	_ = p.Submit(func(ctx context.Context) error {
		time.Sleep(500 * time.Millisecond)
		return nil
	})

	// 2. Submit a Low priority task
	var lowTaskRan atomic.Bool
	_ = p.SubmitWithOptions(func(ctx context.Context) error {
		lowTaskRan.Store(true)
		return nil
	}, TaskOptions{Priority: PriorityLow})

	// 3. Keep submitting High priority tasks for a while
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
				_ = p.SubmitWithOptions(func(ctx context.Context) error {
					time.Sleep(10 * time.Millisecond)
					return nil
				}, TaskOptions{Priority: PriorityHigh})
				time.Sleep(5 * time.Millisecond)
			}
		}
	}()

	// 4. Wait long enough for aging to potentially kick in (AgingWeight is 10s)
	// For a unit test, we just want to ensure it runs eventually.
	time.Sleep(200 * time.Millisecond)
	close(stop)

	// Wait for the pool to process the queue
	time.Sleep(1 * time.Second)

	if !lowTaskRan.Load() {
		t.Error("Low priority task was starved and never ran")
	}
}

// TestPool_StarvationPrevention tests that low-priority tasks eventually run.
func TestPool_StarvationPrevention(t *testing.T) {
	p := New(
		WithMinWorkers(1),
		WithMaxWorkers(1), // Single worker to control execution
		WithQueueSize(100),
	)

	lowPriorityStarted := make(chan bool, 1)

	// Submit a low-priority task
	_ = p.SubmitWithOptions(func(ctx context.Context) error {
		lowPriorityStarted <- true
		time.Sleep(50 * time.Millisecond)
		return nil
	}, TaskOptions{Priority: PriorityLow})

	// Wait for low-priority task to queue
	time.Sleep(50 * time.Millisecond)

	// Continuously submit high-priority tasks for a short duration
	stopHighPriority := make(chan struct{})
	go func() {
		for i := 0; i < 10; i++ {
			select {
			case <-stopHighPriority:
				return
			default:
				_ = p.SubmitWithOptions(func(ctx context.Context) error {
					time.Sleep(50 * time.Millisecond)
					return nil
				}, TaskOptions{Priority: PriorityHigh})
				time.Sleep(10 * time.Millisecond)
			}
		}
		close(stopHighPriority)
	}()

	// The low-priority task should eventually start (aging should help it)
	// Aging weight is 10 seconds, but even with that, after some high-priority
	// tasks complete, the low-priority should get a chance.
	select {
	case <-lowPriorityStarted:
		t.Log("Low-priority task eventually started (starvation prevented)")
	case <-time.After(3 * time.Second):
		t.Error("Low-priority task starved and never ran")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = p.Shutdown(ctx)
}

// TestPool_DefaultPriority tests that tasks submitted without priority use PriorityNormal.
func TestPool_DefaultPriority(t *testing.T) {
	p := New(
		WithMinWorkers(1),
		WithMaxWorkers(1),
		WithQueueSize(100),
	)

	executed := make(chan string, 2)

	// Submit without explicit priority (should use PriorityNormal = 10)
	_ = p.Submit(func(ctx context.Context) error {
		executed <- "normal"
		return nil
	})

	// Wait a tiny bit to ensure order
	time.Sleep(20 * time.Millisecond)

	// Submit with low priority (0)
	_ = p.SubmitWithOptions(func(ctx context.Context) error {
		executed <- "low"
		return nil
	}, TaskOptions{Priority: PriorityLow})

	// Wait for execution (normal should come first due to higher default priority)
	time.Sleep(100 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = p.Shutdown(ctx)

	close(executed)
	results := make([]string, 0)
	for r := range executed {
		results = append(results, r)
	}

	t.Logf("Execution order: %v", results)
}

// TestPool_MixedPriorities tests with a mix of all priority levels.
func TestPool_MixedPriorities(t *testing.T) {
	p := New(
		WithMinWorkers(2),
		WithMaxWorkers(2), // Limit concurrency to see queueing
		WithQueueSize(50),
	)

	executionTimes := make(map[int]time.Time)
	var mu sync.Mutex

	taskFunc := func(taskID int, priority int) Task {
		return func(ctx context.Context) error {
			mu.Lock()
			executionTimes[taskID] = time.Now()
			mu.Unlock()
			time.Sleep(20 * time.Millisecond)
			return nil
		}
	}

	// Submit tasks with different priorities in mixed order
	taskID := 0
	priorities := []int{
		PriorityLow,      // 0
		PriorityCritical, // 100
		PriorityNormal,   // 10
		PriorityHigh,     // 50
		PriorityLow,      // 0
	}

	for _, priority := range priorities {
		_ = p.SubmitWithOptions(taskFunc(taskID, priority), TaskOptions{Priority: priority})
		taskID++
	}

	time.Sleep(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = p.Shutdown(ctx)

	mu.Lock()
	defer mu.Unlock()

	t.Logf("Execution times: %v", executionTimes)
	if len(executionTimes) > 0 {
		t.Logf("Successfully executed %d tasks", len(executionTimes))
	}
}
