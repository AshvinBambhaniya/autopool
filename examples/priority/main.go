package main

import (
	"context"
	"fmt"
	"time"

	"github.com/AshvinBambhaniya/autopool/v2"
)

func main() {
	// Initialize pool with 1 worker to clearly see the priority jumping
	p := autopool.New(
		autopool.WithMaxWorkers(1),
		autopool.WithQueueSize(10),
	)
	defer p.Shutdown(context.Background())

	fmt.Println("Submitting a long-running 'Normal' task to fill the worker...")
	p.Submit(func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("[Worker] Initial task finished.")
		return nil
	})

	// Wait a bit to ensure the task is being processed
	time.Sleep(20 * time.Millisecond)

	fmt.Println("Submitting Low priority task...")
	p.SubmitWithOptions(func(ctx context.Context) error {
		fmt.Println("[Worker] Executed: LOW Priority Task")
		return nil
	}, autopool.TaskOptions{Priority: autopool.PriorityLow})

	fmt.Println("Submitting High priority task...")
	p.SubmitWithOptions(func(ctx context.Context) error {
		fmt.Println("[Worker] Executed: HIGH Priority Task")
		return nil
	}, autopool.TaskOptions{Priority: autopool.PriorityHigh})

	fmt.Println("Submitting Critical priority task...")
	p.SubmitWithOptions(func(ctx context.Context) error {
		fmt.Println("[Worker] Executed: CRITICAL Priority Task")
		return nil
	}, autopool.TaskOptions{Priority: autopool.PriorityCritical})

	// Wait for completion
	time.Sleep(1 * time.Second)
}
