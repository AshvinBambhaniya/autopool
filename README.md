# autopool

**Zero-config, auto-scaling worker pool for Go.**

autopool is a lightweight, high-performance library that manages an elastic pool of workers to process tasks asynchronously. It automatically scales workers up under load and down when idle, ensuring your application remains responsive and memory-safe.

[![Go CI](https://github.com/AshvinBambhaniya/autopool/actions/workflows/go.yml/badge.svg)](https://github.com/AshvinBambhaniya/autopool/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/AshvinBambhaniya/autopool)](https://goreportcard.com/report/github.com/AshvinBambhaniya/autopool)
[![Go Reference](https://pkg.go.dev/badge/github.com/AshvinBambhaniya/autopool.svg)](https://pkg.go.dev/github.com/AshvinBambhaniya/autopool)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- **High Throughput:** Process millions of tasks per second with minimal overhead.
- **Auto-scaling:** Spawns workers on demand and terminates idle ones automatically.
- **Resource Safety:** Bounded queue provides natural backpressure and prevents memory exhaustion.
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

	// Submit a task
	p.Submit(func(ctx context.Context) error {
		fmt.Println("Processing task...")
		return nil
	})

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

You can control retry behavior and timeouts for individual tasks:

```go
err := p.SubmitWithOptions(func(ctx context.Context) error {
    return doWork()
}, autopool.TaskOptions{
    RetryCount: 3,
    RetryDelay: 500 * time.Millisecond,
    Timeout:    2 * time.Second,
})
```

## Benchmarks

Measured on 12th Gen Intel(R) Core(TM) i7-12700:

| Operation | Performance |
| :--- | :--- |
| **High Concurrency** | **~28,000 ns/op** |
| **Throughput (1k tasks)** | **~835,000 ns** |
| **Goroutine Leaks** | **None (Verified)** |

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## License

MIT License. See [LICENSE](LICENSE) for details.
