package pool

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/AshvinBambhaniya/autopool/v2/internal/backoff"
	"github.com/AshvinBambhaniya/autopool/v2/pkg/types"
)

func (p *Pool) SpawnWorker() {
	atomic.AddInt64(&p.CurrentWorkers, 1)
	p.Wg.Add(1)
	go p.Worker()
}

func (p *Pool) Worker() {
	defer func() {
		atomic.AddInt64(&p.CurrentWorkers, -1)
		p.Wg.Done()
	}()

	for {
		atomic.AddInt64(&p.IdleWorkers, 1)

		// Use PopWithTimeout to allow for scaling down
		task, timedOut := p.TaskQueue.PopWithTimeout(p.IdleTimeout)
		atomic.AddInt64(&p.IdleWorkers, -1)

		if task == nil {
			if timedOut {
				if p.ShouldScaleDown() {
					return
				}
				continue
			}
			return // Queue closed
		}

		p.Execute(*task)
	}
}

func (p *Pool) Execute(t types.TaskWrapper) {
	defer func() {
		if r := recover(); r != nil && p.PanicHandler != nil {
			p.PanicHandler(r)
		}
	}()

	// Use a simple exponential backoff for retries
	strategy := backoff.NewExponential(
		t.Opts.RetryDelay,
		10*time.Second,
	)
	if t.Opts.RetryDelay == 0 {
		strategy.Base = 100 * time.Millisecond
	}

	for i := 0; i <= t.Opts.RetryCount; i++ {
		var ctx context.Context
		var cancel context.CancelFunc

		if t.Opts.Timeout > 0 {
			ctx, cancel = context.WithTimeout(p.Ctx, t.Opts.Timeout)
		} else {
			ctx = p.Ctx
		}

		err := t.Fn(ctx)
		if cancel != nil {
			cancel()
		}

		if err == nil {
			return
		}

		if i < t.Opts.RetryCount {
			delay := strategy.Next(i)

			select {
			case <-time.After(delay):
			case <-p.Ctx.Done():
				return
			}
		}
	}
}
