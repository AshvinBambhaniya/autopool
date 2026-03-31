// Package autopool provides a zero-config, auto-scaling worker pool for Go.
//
// It automatically manages a pool of goroutines, scaling up when tasks are
// submitted and scaling down when workers are idle, ensuring efficient
// resource usage and high throughput.
package autopool

import (
	"context"
	"time"

	"github.com/AshvinBambhaniya/autopool/internal/pool"
	"github.com/AshvinBambhaniya/autopool/pkg/types"
)

// Pool defines the behavior for the auto-scaling worker pool.
// It is the primary interface for submitting tasks and managing the lifecycle
// of the worker pool.
type Pool interface {
	// Submit adds a task to the queue for asynchronous execution.
	// It blocks if the task queue is full (natural backpressure).
	// Returns ErrAlreadyClosed if the pool is no longer running.
	Submit(task Task) error

	// SubmitWithOptions adds a task with custom execution options such as
	// retries and timeouts.
	// Returns ErrAlreadyClosed if the pool is no longer running.
	SubmitWithOptions(task Task, opts TaskOptions) error

	// Stats returns real-time metrics about the pool, including worker counts
	// and queue depth.
	Stats() Stats

	// Shutdown gracefully stops the pool. It stops accepting new tasks and
	// waits for all currently queued tasks to finish processing or until
	// the provided context is cancelled.
	Shutdown(ctx context.Context) error
}

// Task is a function that represents a unit of work to be executed by the pool.
// The provided context is cancelled if the pool shuts down or the task times out.
type Task = types.Task

// TaskOptions provides granular control over individual task execution,
// allowing for per-task retry logic and timeouts.
type TaskOptions = types.TaskOptions

// Stats provides real-time information about the pool's performance and load.
type Stats = types.Stats

// Option is a functional option for configuring the pool during initialization.
type Option = pool.Option

// Re-exported errors for public use.
var (
	// ErrQueueFull is returned when the task queue is at capacity.
	ErrQueueFull = types.ErrQueueFull
	// ErrAlreadyClosed is returned when an operation is attempted on a pool
	// that has already been shut down.
	ErrAlreadyClosed = types.ErrAlreadyClosed
)

// New creates and initializes a new worker pool with the provided options.
// If no options are provided, it uses sensible defaults (e.g., max 10 workers,
// queue size of 100).
func New(opts ...Option) Pool {
	return pool.New(opts...)
}

// WithMaxWorkers sets the maximum number of concurrent workers allowed in the pool.
func WithMaxWorkers(n int) Option { return pool.WithMaxWorkers(n) }

// WithMinWorkers sets the minimum number of workers that should always stay
// alive, even when idle.
func WithMinWorkers(n int) Option { return pool.WithMinWorkers(n) }

// WithQueueSize sets the capacity of the task queue. When the queue is full,
// Submit will block until a slot becomes available.
func WithQueueSize(n int) Option { return pool.WithQueueSize(n) }

// WithIdleTimeout sets the duration after which an idle worker will exit,
// provided the total number of workers stays above MinWorkers.
func WithIdleTimeout(d time.Duration) Option { return pool.WithIdleTimeout(d) }

// WithPanicHandler sets a custom function to handle panics that occur during
// task execution.
func WithPanicHandler(f func(interface{})) Option { return pool.WithPanicHandler(f) }

// WithErrorHandler sets a custom function to handle tasks that fail even
// after all retry attempts.
func WithErrorHandler(f func(error)) Option { return pool.WithErrorHandler(f) }
