package queue

import (
	"container/heap"
	"sync"
	"time"

	"github.com/AshvinBambhaniya/autopool/v2/pkg/types"
)

// AgingWeight defines the promotion speed for low-priority tasks.
// Higher values slow down the aging process.
const AgingWeight = 10 * time.Second

// TaskQueue is a thread-safe, bounded priority queue implementation.
//
// It coordinates Producers (Submitters) and Consumers (Workers) using two
// distinct condition variables to maximize throughput and prevent deadlocks.
type TaskQueue struct {
	mu sync.Mutex
	// hasTasks is signaled when a new task is pushed; workers wait here.
	hasTasks *sync.Cond
	// hasSpace is signaled when a task is popped; submitters wait here.
	hasSpace *sync.Cond
	// heap is the internal, non-thread-safe priority heap.
	heap taskHeap
	// capacity is the maximum number of tasks allowed in the queue.
	capacity int
	// closed indicates if the queue is no longer accepting submissions.
	closed bool
}

// New creates and initializes a new thread-safe priority task queue.
func New(capacity int) *TaskQueue {
	tq := &TaskQueue{
		capacity: capacity,
		heap:     make(taskHeap, 0, capacity),
	}
	tq.hasTasks = sync.NewCond(&tq.mu)
	tq.hasSpace = sync.NewCond(&tq.mu)
	return tq
}

// Push adds a task to the queue based on its priority and submission time.
//
// It uses a virtual runtime model to calculate a stable priority score,
// ensuring that high-priority tasks jump the line while preventing starvation
// of older tasks. This method blocks if the queue is at capacity.
func (tq *TaskQueue) Push(task types.TaskWrapper) bool {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	// Wait while the queue is full and still operational.
	for len(tq.heap) >= tq.capacity && !tq.closed {
		tq.hasSpace.Wait()
	}

	if tq.closed {
		return false
	}

	// Calculate a frozen urgency score at birth to maintain heap invariants.
	// score = now - (priority * AgingWeight)
	now := time.Now().UnixNano()
	boost := int64(task.Opts.Priority) * int64(AgingWeight)

	item := &TaskItem{
		Task:          task,
		priorityScore: now - boost,
	}
	heap.Push(&tq.heap, item)

	// Notify one waiting worker that work is available.
	tq.hasTasks.Signal()
	return true
}

// Pop retrieves and removes the highest priority task from the queue.
//
// If the queue is empty, this method will block until work arrives or the
// queue is closed.
func (tq *TaskQueue) Pop() *types.TaskWrapper {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	// Wait while the queue is empty and still operational.
	for len(tq.heap) == 0 && !tq.closed {
		tq.hasTasks.Wait()
	}

	if len(tq.heap) == 0 {
		return nil
	}

	item := heap.Pop(&tq.heap).(*TaskItem)

	// Notify one blocked submitter that space is now available.
	tq.hasSpace.Signal()
	return &item.Task
}

// PopWithTimeout attempts to retrieve a task within the specified duration.
//
// This is used by workers to implement idle timeout scaling. If no task
// is retrieved before the timeout, it returns (nil, true).
func (tq *TaskQueue) PopWithTimeout(timeout time.Duration) (*types.TaskWrapper, bool) {
	tq.mu.Lock()
	defer tq.mu.Unlock()

	if len(tq.heap) == 0 && !tq.closed {
		// Set a timer to wake the goroutine if it's still waiting
		// when the timeout expires.
		timer := time.AfterFunc(timeout, func() {
			tq.mu.Lock()
			tq.hasTasks.Broadcast()
			tq.mu.Unlock()
		})
		defer timer.Stop()

		deadline := time.Now().Add(timeout)
		for len(tq.heap) == 0 && !tq.closed && time.Now().Before(deadline) {
			tq.hasTasks.Wait()
		}
	}

	if len(tq.heap) == 0 {
		return nil, !tq.closed // Returns (nil, true) if it actually timed out.
	}

	item := heap.Pop(&tq.heap).(*TaskItem)
	tq.hasSpace.Signal()
	return &item.Task, false
}

// Len returns the current number of tasks waiting in the queue.
func (tq *TaskQueue) Len() int {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	return len(tq.heap)
}

// Close transitions the queue to a stopping state and wakes all blocked goroutines.
func (tq *TaskQueue) Close() {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	tq.closed = true

	// Wake all submitters and workers to handle the closure.
	tq.hasTasks.Broadcast()
	tq.hasSpace.Broadcast()
}
