# autopool

**Zero-config, auto-scaling worker pool for Go with priority-aware scheduling.**

`autopool` is a lightweight, high-performance library that manages an elastic pool of workers to process tasks asynchronously. It combines dynamic auto-scaling with a sophisticated priority-aware queue, ensuring that critical tasks jump the line while background work remains resilient and memory-safe.

[![Go CI](https://github.com/AshvinBambhaniya/autopool/actions/workflows/go.yml/badge.svg)](https://github.com/AshvinBambhaniya/autopool/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/AshvinBambhaniya/autopool)](https://goreportcard.com/report/github.com/AshvinBambhaniya/autopool)
[![Go Reference](https://pkg.go.dev/badge/github.com/AshvinBambhaniya/autopool.svg)](https://pkg.go.dev/github.com/AshvinBambhaniya/autopool)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- **High Throughput:** Process millions of tasks per second with minimal overhead.
- **Priority-Aware Scheduling:** Support for `Critical`, `High`, `Normal`, and `Low` priority levels.
- **Anti-Starvation (Aging):** Virtual Runtime model ensures low-priority tasks eventually execute even under heavy load.
- **Auto-scaling:** Spawns workers on demand and terminates idle ones automatically.
- **Resource Safety:** Bounded priority queue provides natural backpressure and prevents memory exhaustion.
- **Graceful Shutdown:** Ensures active tasks are drained completely before stopping.
- **Panic Recovery:** Built-in recovery per worker prevents cascading pool failures.
- **Smart Retries:** Per-task exponential backoff strategy for resilient processing.
- **Real-time Metrics:** Deep visibility into worker counts and queue depth.

## Installation

```bash
go get github.com/AshvinBambhaniya/autopool
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"github.com/AshvinBambhaniya/autopool"
	"time"
)

func main() {
	// Initialize pool with custom limits
	p := autopool.New(
		autopool.WithMaxWorkers(10),
		autopool.WithQueueSize(100),
	)

	// Submit a task (defaults to PriorityNormal)
	p.Submit(func(ctx context.Context) error {
		fmt.Println("Processing task...")
		return nil
	})

	// Submit a high-priority task that jumps the line
	p.SubmitWithOptions(func(ctx context.Context) error {
		fmt.Println("Processing critical work...")
		return nil
	}, autopool.TaskOptions{Priority: autopool.PriorityHigh})

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	p.Shutdown(ctx)
}
```

## Configuration

The pool is highly configurable using functional options:

```go
p := autopool.New(
	autopool.WithMinWorkers(2),         // Always keep 2 workers alive
	autopool.WithMaxWorkers(50),        // Allow scaling up to 50
	autopool.WithQueueSize(1000),       // Buffer size (Submit blocks when full)
	autopool.WithIdleTimeout(5 * time.Minute), // When to scale down
	autopool.WithPanicHandler(myPanicHandler),
	autopool.WithErrorHandler(myErrorHandler), // Handle final failures after retries
)
```

## Per-Task Options

You can control retry behavior, timeouts, and priority for individual tasks:

```go
err := p.SubmitWithOptions(func(ctx context.Context) error {
    return doWork()
}, autopool.TaskOptions{
    Priority:   autopool.PriorityHigh,
    RetryCount: 3,
    RetryDelay: 500 * time.Millisecond,
    Timeout:    2 * time.Second,
})
```

## Task Prioritization & Aging

`autopool` uses a **Virtual Runtime Aging** model. When a task is submitted, it is assigned a virtual score based on its priority. As time passes, the relative "urgency" of older tasks increases naturally, ensuring that high-priority tasks jump the line without starving low-priority tasks indefinitely.

Supported levels: `PriorityLow`, `PriorityNormal` (default), `PriorityHigh`, `PriorityCritical`.

## Benchmarks

Measured on 12th Gen Intel(R) Core(TM) i7-12700:

| Operation | Performance |
| :--- | :--- |
| **Queue Push** | **~52 ns/op** |
| **Queue Pop** | **~204 ns/op** |
| **Throughput (100k tasks)** | **~1,200,000 tasks/sec** |
| **Goroutine Leaks** | **None (Verified with -race)** |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## License

MIT License. See [LICENSE](LICENSE) for details.
