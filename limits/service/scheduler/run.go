package scheduler

import (
	"context"
	"time"
)

func (m *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(m.cfg.Cleanup.RunInterval)
	defer ticker.Stop()

	for {
		m.cleanup(ctx)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (m *Scheduler) cleanup(ctx context.Context) {
	if err := m.storage.CleanupLimits(ctx); err != nil {
		m.logger.WithCtx(ctx).
			WithError(err).
			Error("limits cleanup failed")
	}

	if err := m.storage.CleanupCounters(ctx, m.cfg.Cleanup.OutdateInterval); err != nil {
		m.logger.WithCtx(ctx).
			WithError(err).
			Error("counters cleanup failed")
	}

	if err := m.storage.CleanupContext(ctx, m.cfg.Cleanup.OutdateInterval); err != nil {
		m.logger.WithCtx(ctx).
			WithError(err).
			Error("context cleanup failed")
	}
}
