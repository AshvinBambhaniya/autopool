package pool

import (
	"time"

	"github.com/AshvinBambhaniya/autopool/internal/queue"
)

// Option is a functional option for configuring the pool.
type Option func(*Pool)

// WithMaxWorkers sets the maximum number of workers allowed in the pool.
func WithMaxWorkers(n int) Option {
	return func(p *Pool) {
		if n > 0 {
			p.MaxWorkers = n
		}
	}
}

// WithMinWorkers sets the minimum number of workers that should always be active.
func WithMinWorkers(n int) Option {
	return func(p *Pool) {
		if n >= 0 {
			p.MinWorkers = n
		}
	}
}

// WithQueueSize sets the size of the task queue.
func WithQueueSize(n int) Option {
	return func(p *Pool) {
		if n > 0 {
			p.TaskQueue = queue.New(n)
		}
	}
}

// WithIdleTimeout sets the duration after which an idle worker will exit if
// total workers > minWorkers.
func WithIdleTimeout(d time.Duration) Option {
	return func(p *Pool) {
		if d > 0 {
			p.IdleTimeout = d
		}
	}
}

// WithPanicHandler sets a function to be called when a task panics.
func WithPanicHandler(f func(interface{})) Option {
	return func(p *Pool) {
		p.PanicHandler = f
	}
}

// WithErrorHandler sets a function to be called when a task fails
// after all retry attempts.
func WithErrorHandler(f func(error)) Option {
	return func(p *Pool) {
		p.ErrorHandler = f
	}
}
