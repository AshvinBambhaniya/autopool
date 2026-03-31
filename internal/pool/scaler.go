package pool

import (
	"sync/atomic"
)

// ScaleUp spawns a new worker if there is pending work and capacity allows.
func (p *Pool) ScaleUp() {
	current := atomic.LoadInt64(&p.CurrentWorkers)
	idle := atomic.LoadInt64(&p.IdleWorkers)
	queueLen := len(p.Queue)

	// Scale up if no idle workers OR if work is backed up in the queue.
	if (idle == 0 || queueLen > 0) && current < int64(p.MaxWorkers) {
		p.SpawnWorker()
	}
}

// ShouldScaleDown checks if a worker can exit without dropping below MinWorkers.
func (p *Pool) ShouldScaleDown() bool {
	current := atomic.LoadInt64(&p.CurrentWorkers)

	// Never scale below the minimum baseline.
	if current <= int64(p.MinWorkers) {
		return false
	}

	return true
}
