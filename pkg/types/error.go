package types

import "errors"

var (
	// ErrQueueFull is returned when a task is submitted to a pool that has
	// reached its maximum queue capacity.
	ErrQueueFull = errors.New("autopool: task queue is full")

	// ErrAlreadyClosed is returned when an operation (like submitting a task
	// or shutting down) is attempted on a pool that has already been closed.
	ErrAlreadyClosed = errors.New("autopool: pool is already closed")
)
