// Package types contains common types and interfaces used throughout autopool.
package types

import (
	"context"
	"time"
)

// Task is a function that represents a unit of work.
// It receives a context that is cancelled if the pool shuts down
// or the task-specific timeout is exceeded.
type Task func(ctx context.Context) error

// TaskOptions provides granular control over individual task execution.
type TaskOptions struct {
	// RetryCount is the number of times a task should be retried on failure.
	// Default is 0 (no retries).
	RetryCount int
	// RetryDelay is the initial delay between retries.
	// This is the base for exponential backoff calculations.
	RetryDelay time.Duration
	// Timeout is the maximum execution time for a single task run.
	// If set, a context with timeout will be passed to the task.
	Timeout time.Duration
}

// TaskWrapper combines a task function with its associated options.
// Internal-only use.
type TaskWrapper struct {
	Fn   Task
	Opts TaskOptions
}
