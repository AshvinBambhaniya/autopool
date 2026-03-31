// Package queue provides a high-performance, priority-aware task queue.
package queue

import (
	"github.com/AshvinBambhaniya/autopool/pkg/types"
)

// TaskItem wraps a task with metadata required for priority queue operations.
type TaskItem struct {
	// Task contains the unit of work and its associated execution options.
	Task types.TaskWrapper
	// priorityScore is a static urgency value calculated at submission time.
	//
	// Formula: submission_time_unix - (priority * AgingWeight)
	//
	// By using a static score, we ensure the heap invariant is never violated
	// while the task is stored, while still achieving natural priority aging.
	// A smaller score indicates a more urgent task.
	priorityScore int64
	// Index is the current position of the item within the heap array.
	// This is maintained automatically by the container/heap implementation.
	Index int
}

// taskHeap is a min-priority heap implemented as a slice of TaskItem pointers.
// It is ordered by the static priorityScore, ensuring the most urgent
// task is always at index 0.
type taskHeap []*TaskItem

// Len returns the number of items currently in the heap.
func (h taskHeap) Len() int { return len(h) }

// Less determines the priority order between two tasks.
// It uses the constant priorityScore to ensure heap stability.
func (h taskHeap) Less(i, j int) bool { return h[i].priorityScore < h[j].priorityScore }

// Swap exchanges two items in the heap and updates their internal indices.
func (h taskHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

// Push adds a new item to the underlying slice.
// This is a low-level method used by container/heap.
func (h *taskHeap) Push(x any) {
	item := x.(*TaskItem)
	item.Index = len(*h)
	*h = append(*h, item)
}

// Pop removes and returns the last item from the underlying slice.
// This is a low-level method used by container/heap.
func (h *taskHeap) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // Explicitly nil the pointer to assist Garbage Collection.
	item.Index = -1
	*h = old[0 : n-1]
	return item
}
