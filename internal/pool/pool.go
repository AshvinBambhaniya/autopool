// Package pool contains the core implementation of the worker pool orchestration.
package pool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AshvinBambhaniya/autopool/v2/internal/queue"
	"github.com/AshvinBambhaniya/autopool/v2/pkg/types"
)

// Pool is the concrete implementation of the worker pool.
// It manages the lifecycle of workers and coordinates task distribution
// using a priority-aware queue.
type Pool struct {
	// TaskQueue is the priority-aware queue where tasks are held before processing.
	TaskQueue *queue.TaskQueue
	// MinWorkers is the minimum number of workers to keep alive, even when idle.
	MinWorkers int
	// MaxWorkers is the maximum allowed number of concurrent worker goroutines.
	MaxWorkers int
	// CurrentWorkers is the atomic count of total workers currently running.
	CurrentWorkers int64
	// IdleWorkers is the atomic count of workers waiting for new tasks.
	IdleWorkers int64

	// Wg tracks the lifecycle of all spawned workers for graceful shutdown.
	Wg sync.WaitGroup
	// Ctx is used for signaling shutdown to all active workers.
	Ctx context.Context
	// Cancel is the function to signal the end of the pool's context.
	Cancel context.CancelFunc
	// State tracks the operational state of the pool (Running, Stopping, Stopped).
	State int32

	// PanicHandler is called if a worker recovers from an unexpected panic.
	PanicHandler func(interface{})
	// ErrorHandler is called if a task fails after all retry attempts.
	ErrorHandler func(error)
	// IdleTimeout is the duration a worker waits for a task before scaling down.
	IdleTimeout time.Duration
}

// Operational states of the Pool.
const (
	Running int32 = iota
	Stopping
	Stopped
)

// New creates and initializes a new Pool orchestration instance.
func New(opts ...Option) *Pool {
	p := &Pool{
		TaskQueue:   queue.New(100),
		MinWorkers:  0,
		MaxWorkers:  10,
		IdleTimeout: 60 * time.Second,
		State:       Running,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.Ctx, p.Cancel = context.WithCancel(context.Background())

	// Spawn the initial baseline of workers.
	for i := 0; i < p.MinWorkers; i++ {
		p.SpawnWorker()
	}

	return p
}

// Submit puts a task into the queue with normal priority.
func (p *Pool) Submit(task types.Task) error {
	return p.SubmitWithOptions(task, types.TaskOptions{Priority: types.PriorityNormal})
}

// SubmitWithOptions puts a task with custom options into the priority queue.
// It triggers a scaling check before adding the task.
func (p *Pool) SubmitWithOptions(task types.Task, opts types.TaskOptions) error {
	if atomic.LoadInt32(&p.State) != Running {
		return types.ErrAlreadyClosed
	}

	// Apply default priority if not explicitly specified.
	if opts.Priority == 0 {
		opts.Priority = types.PriorityNormal
	}

	// Trigger proactive scaling if capacity allows.
	p.ScaleUp()

	if !p.TaskQueue.Push(types.TaskWrapper{Fn: task, Opts: opts}) {
		return types.ErrAlreadyClosed
	}
	return nil
}

// Stats returns a snapshot of current pool performance and load metrics.
func (p *Pool) Stats() types.Stats {
	return types.Stats{
		TotalWorkers: int(atomic.LoadInt64(&p.CurrentWorkers)),
		IdleWorkers:  int(atomic.LoadInt64(&p.IdleWorkers)),
		QueueSize:    p.TaskQueue.Len(),
	}
}

// Shutdown initiates a graceful stop of the pool.
// It stops accepting new tasks and waits for the queue to drain.
func (p *Pool) Shutdown(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&p.State, Running, Stopping) {
		return types.ErrAlreadyClosed
	}

	// Signal the queue to stop accepting new submissions.
	p.TaskQueue.Close()

	done := make(chan struct{})
	go func() {
		p.Wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		atomic.StoreInt32(&p.State, Stopped)
		p.Cancel()
		return nil
	case <-ctx.Done():
		p.Cancel() // Force cancellation of worker contexts on timeout.
		return ctx.Err()
	}
}
