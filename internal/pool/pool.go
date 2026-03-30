// package pool contains the core implementation of the worker pool.
package pool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/AshvinBambhaniya/autopool/pkg/types"
)

// Pool is the concrete implementation of the worker pool.
// It manages workers, scaling, and the task queue.
type Pool struct {
	// Queue is the buffered channel where tasks are submitted.
	Queue chan types.TaskWrapper
	// MinWorkers is the minimum number of workers to keep alive.
	MinWorkers int
	// MaxWorkers is the maximum allowed number of concurrent workers.
	MaxWorkers int
	// CurrentWorkers is the atomic count of active goroutines.
	CurrentWorkers int64
	// IdleWorkers is the atomic count of workers currently waiting for tasks.
	IdleWorkers int64

	// Wg tracks the lifecycle of all spawned workers.
	Wg sync.WaitGroup
	// Ctx is used for signaling shutdown to workers.
	Ctx context.Context
	// Cancel function for Ctx.
	Cancel context.CancelFunc
	// State tracks the lifecycle state of the pool (Running, Stopping, Stopped).
	State int32

	// PanicHandler is called if a worker recovers from a panic.
	PanicHandler func(interface{})
	// ErrorHandler is called if a task fails after all retries.
	ErrorHandler func(error)
	// IdleTimeout is the duration a worker waits for a task before scaling down.
	IdleTimeout time.Duration
}

// Lifecycle states of the Pool.
const (
	Running int32 = iota
	Stopping
	Stopped
)

// New creates and initializes a new Pool implementation.
func New(opts ...Option) *Pool {
	p := &Pool{
		Queue:       make(chan types.TaskWrapper, 100),
		MinWorkers:  0,
		MaxWorkers:  10,
		IdleTimeout: 60 * time.Second,
		State:       Running,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.Ctx, p.Cancel = context.WithCancel(context.Background())

	// Spawn the minimum number of initial workers.
	for i := 0; i < p.MinWorkers; i++ {
		p.SpawnWorker()
	}

	return p
}

// Submit puts a task into the queue.
func (p *Pool) Submit(task types.Task) error {
	return p.SubmitWithOptions(task, types.TaskOptions{})
}

// SubmitWithOptions puts a task with custom options into the queue.
func (p *Pool) SubmitWithOptions(task types.Task, opts types.TaskOptions) error {
	if atomic.LoadInt32(&p.State) != Running {
		return types.ErrAlreadyClosed
	}

	// Dynamic scaling up if needed.
	p.ScaleUp()

	select {
	case p.Queue <- types.TaskWrapper{Fn: task, Opts: opts}:
		return nil
	case <-p.Ctx.Done():
		return types.ErrAlreadyClosed
	}
}

// Stats returns a snapshot of current pool performance metrics.
func (p *Pool) Stats() types.Stats {
	return types.Stats{
		TotalWorkers: int(atomic.LoadInt64(&p.CurrentWorkers)),
		IdleWorkers:  int(atomic.LoadInt64(&p.IdleWorkers)),
		QueueSize:    len(p.Queue),
	}
}

// Shutdown initiates a graceful stop of the pool.
func (p *Pool) Shutdown(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&p.State, Running, Stopping) {
		return types.ErrAlreadyClosed
	}

	// Close the queue to stop accepting new tasks and signal workers.
	close(p.Queue)

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
