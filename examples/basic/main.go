// Package main provides a basic example of using autopool to process tasks.
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/AshvinBambhaniya/autopool"
)

func main() {
	// Initialize a new worker pool with custom configuration.
	// - MinWorkers(2): Always keep 2 workers active.
	// - MaxWorkers(5): Scale up to 5 workers under load.
	// - QueueSize(10): Buffer up to 10 tasks before Submit blocks.
	pool := autopool.New(
		autopool.WithMinWorkers(2),
		autopool.WithMaxWorkers(5),
		autopool.WithQueueSize(10),
		autopool.WithIdleTimeout(5*time.Second),
	)

	fmt.Println("Starting autopool example...")
	fmt.Printf("Initial Stats: %+v\n", pool.Stats())

	// totalTasks to process.
	const totalTasks = 20
	start := time.Now()

	for i := 1; i <= totalTasks; i++ {
		taskID := i

		// Log the submission attempt.
		fmt.Printf("[%4s] Submitting task %d...\n", time.Since(start).Truncate(time.Millisecond), taskID)

		// Submit a task to the pool.
		// Submit will block once the queue (10) and active workers (5) are full.
		err := pool.Submit(func(ctx context.Context) error {
			// Simulate processing time.
			time.Sleep(1 * time.Second)
			fmt.Printf("[%4s] Processed task %d\n", time.Since(start).Truncate(time.Millisecond), taskID)
			return nil
		})

		if err != nil {
			log.Printf("Failed to submit task %d: %v", taskID, err)
		}
	}

	fmt.Println("All tasks submitted. Waiting for completion...")

	// Create a context with timeout for graceful shutdown.
	// This ensures the program doesn't hang indefinitely if tasks are stuck.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := pool.Shutdown(shutdownCtx); err != nil {
		log.Printf("Pool shutdown failed: %v", err)
	}

	fmt.Printf("Example finished in %s. Final Stats: %+v\n", time.Since(start).Truncate(time.Millisecond), pool.Stats())
}
