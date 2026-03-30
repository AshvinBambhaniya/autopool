package pool

import (
	"sync/atomic"
)

func (p *Pool) ScaleUp() {
	current := atomic.LoadInt64(&p.CurrentWorkers)
	idle := atomic.LoadInt64(&p.IdleWorkers)

	// If we have idle workers, no need to scale up
	if idle > 0 {
		return
	}

	// If we haven't reached maxWorkers, spawn a new one
	if current < int64(p.MaxWorkers) {
		p.SpawnWorker()
	}
}

func (p *Pool) ShouldScaleDown() bool {
	current := atomic.LoadInt64(&p.CurrentWorkers)

	// Never scale below minWorkers
	if current <= int64(p.MinWorkers) {
		return false
	}

	return true
}
