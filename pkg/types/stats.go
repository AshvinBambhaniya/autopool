package types

// Stats provides real-time information about the pool's performance and load.
type Stats struct {
	// TotalWorkers is the number of currently active goroutines managed by the pool.
	TotalWorkers int
	// IdleWorkers is the number of workers that are waiting for a new task.
	IdleWorkers int
	// QueueSize is the number of tasks currently waiting in the queue to be
	// picked up by a worker.
	QueueSize int
}
