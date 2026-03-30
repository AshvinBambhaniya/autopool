package pool

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/AshvinBambhaniya/autopool/internal/backoff"
	"github.com/AshvinBambhaniya/autopool/pkg/types"
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

	timer := time.NewTimer(p.IdleTimeout)
	defer timer.Stop()

	for {
		atomic.AddInt64(&p.IdleWorkers, 1)

		select {
		case <-p.Ctx.Done():
			atomic.AddInt64(&p.IdleWorkers, -1)
			return

		case task, ok := <-p.Queue:
			atomic.AddInt64(&p.IdleWorkers, -1)
			if !ok {
				return
			}
			p.Execute(task)
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(p.IdleTimeout)

		case <-timer.C:
			atomic.AddInt64(&p.IdleWorkers, -1)
			if p.ShouldScaleDown() {
				return
			}
			timer.Reset(p.IdleTimeout)
		}
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
