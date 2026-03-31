package queue

import (
	"context"
	"testing"
	"time"

	"github.com/AshvinBambhaniya/autopool/pkg/types"
)

// TestQueueBasicOperations tests basic push and pop operations.
func TestQueueBasicOperations(t *testing.T) {
	q := New(10)
	defer q.Close()

	task := types.TaskWrapper{
		Opts: types.TaskOptions{Priority: 10},
		Fn:   func(ctx context.Context) error { return nil },
	}

	// Push should succeed
	if !q.Push(task) {
		t.Error("push failed on non-full queue")
	}

	if q.Len() != 1 {
		t.Errorf("expected length 1, got %d", q.Len())
	}

	// Pop should return the task
	popped := q.Pop()
	if popped == nil {
		t.Error("expected task, got nil")
	}

	if q.Len() != 0 {
		t.Errorf("expected length 0 after pop, got %d", q.Len())
	}
}

// TestQueuePriorityOrdering tests that tasks are ordered by priority score.
func TestQueuePriorityOrdering(t *testing.T) {
	q := New(10)
	defer q.Close()

	// Submit tasks with different priorities
	tasks := []int{5, 100, 10, 50, 1}
	for _, priority := range tasks {
		q.Push(types.TaskWrapper{
			Opts: types.TaskOptions{Priority: priority},
			Fn:   func(ctx context.Context) error { return nil },
		})
	}

	// Should pop in order of priority: 100, 50, 10, 5, 1
	expected := []int{100, 50, 10, 5, 1}
	for _, exp := range expected {
		popped := q.Pop()
		if popped == nil {
			t.Fatalf("expected task, got nil")
		}
		if popped.Opts.Priority != exp {
			t.Errorf("expected priority %d, got %d", exp, popped.Opts.Priority)
		}
	}
}

// TestQueueCapacityBlocking tests that Push blocks when queue is full.
func TestQueueCapacityBlocking(t *testing.T) {
	q := New(2)
	defer q.Close()

	task := types.TaskWrapper{
		Opts: types.TaskOptions{Priority: 10},
		Fn:   func(ctx context.Context) error { return nil },
	}

	// Fill queue to capacity
	q.Push(task)
	q.Push(task)

	// Next push should block
	done := make(chan bool, 1)
	go func() {
		q.Push(task)
		done <- true
	}()

	// Verify it's blocked
	select {
	case <-time.After(100 * time.Millisecond):
		// Good, it's blocked
	case <-done:
		t.Error("push should have blocked on full queue")
	}

	// Pop to unblock
	q.Pop()

	// Now it should complete
	select {
	case <-time.After(1 * time.Second):
		t.Error("push did not complete after space available")
	case <-done:
		// Good!
	}
}

// TestQueueClosedBehavior tests behavior after queue is closed.
func TestQueueClosedBehavior(t *testing.T) {
	q := New(10)

	task := types.TaskWrapper{
		Opts: types.TaskOptions{Priority: 10},
		Fn:   func(ctx context.Context) error { return nil },
	}

	// Push task
	q.Push(task)

	// Close queue
	q.Close()

	// Pop should still return the task
	popped := q.Pop()
	if popped == nil {
		t.Error("expected task from closed queue")
	}

	// Next pop should return nil (queue empty)
	popped = q.Pop()
	if popped != nil {
		t.Error("expected nil from empty closed queue")
	}

	// Push should fail
	if q.Push(task) {
		t.Error("push should fail on closed queue")
	}
}

// TestPopWithTimeout tests timeout behavior.
func TestPopWithTimeout(t *testing.T) {
	q := New(10)
	defer q.Close()

	// Pop from empty queue with short timeout
	task, timedOut := q.PopWithTimeout(50 * time.Millisecond)
	if task != nil {
		t.Error("expected nil task on timeout")
	}
	if !timedOut {
		t.Error("expected timedOut=true")
	}

	// Push a task
	q.Push(types.TaskWrapper{
		Opts: types.TaskOptions{Priority: 10},
		Fn:   func(ctx context.Context) error { return nil },
	})

	// Pop with timeout should succeed immediately
	task, timedOut = q.PopWithTimeout(100 * time.Millisecond)
	if task == nil {
		t.Error("expected task")
	}
	if timedOut {
		t.Error("expected timedOut=false when task available")
	}
}

// TestQueueConcurrentOperations is a stress test for concurrent pushes and pops.
func TestQueueConcurrentOperations(t *testing.T) {
	q := New(100)
	defer q.Close()

	numTasks := 1000
	done := make(chan bool, 2)

	// Queue populator
	go func() {
		for i := 0; i < numTasks; i++ {
			q.Push(types.TaskWrapper{
				Opts: types.TaskOptions{Priority: i % 100},
				Fn:   func(ctx context.Context) error { return nil },
			})
		}
		done <- true
	}()

	// Task consumer
	popped := 0
	go func() {
		for popped < numTasks {
			task, _ := q.PopWithTimeout(100 * time.Millisecond)
			if task != nil {
				popped++
			}
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done

	if popped != numTasks {
		t.Errorf("expected %d tasks popped, got %d", numTasks, popped)
	}
}

// BenchmarkQueuePush benchmarks push performance.
func BenchmarkQueuePush(b *testing.B) {
	q := New(b.N)
	defer q.Close()

	task := types.TaskWrapper{
		Opts: types.TaskOptions{Priority: 10},
		Fn:   func(ctx context.Context) error { return nil },
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Push(task)
	}
}

// BenchmarkQueuePop benchmarks pop performance.
func BenchmarkQueuePop(b *testing.B) {
	q := New(b.N)
	defer q.Close()

	task := types.TaskWrapper{
		Opts: types.TaskOptions{Priority: 10},
		Fn:   func(ctx context.Context) error { return nil },
	}

	// Pre-fill queue
	for i := 0; i < b.N; i++ {
		q.Push(task)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Pop()
	}
}
