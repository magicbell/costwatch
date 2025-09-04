package scheduler

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// LocalTicker is a minimal, in-process ticker that invokes a job function on an interval.
// Not suitable for distributed coordination; intended for local dev and single-instance workers.
type LocalTicker struct {
	log      *slog.Logger
	interval time.Duration
	job      func(context.Context) error

	stop chan struct{}
	wg   sync.WaitGroup
}

// NewLocalTicker constructs a new LocalTicker.
func NewLocalTicker(log *slog.Logger, interval time.Duration, job func(context.Context) error) *LocalTicker {
	return &LocalTicker{
		log:      log,
		interval: interval,
		job:      job,
		stop:     make(chan struct{}),
	}
}

// Start runs the ticker loop until the context is canceled or Stop is called.
func (lt *LocalTicker) Start(ctx context.Context) {
	lt.wg.Add(1)
	go func() {
		defer lt.wg.Done()
		ticker := time.NewTicker(lt.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-lt.stop:
				return
			case <-ticker.C:
				if lt.job != nil {
					_ = lt.job(ctx)
				}
			}
		}
	}()
}

// Stop signals the ticker to stop and waits for it to finish.
func (lt *LocalTicker) Stop() {
	select {
	case <-lt.stop:
		// already closed
	default:
		close(lt.stop)
	}
	lt.wg.Wait()
}
