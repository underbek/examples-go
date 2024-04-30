package scheduler

import (
	"context"
	"time"

	"github.com/underbek/examples-go/limits/config"
	"github.com/underbek/examples-go/logger"
)

//go:generate mockery --name storage --structname StorageMock --filename storage_mock.go --inpackage
type storage interface {
	CleanupLimits(ctx context.Context) error
	CleanupCounters(context.Context, time.Duration) error
	CleanupContext(context.Context, time.Duration) error
}

type Scheduler struct {
	logger  *logger.Logger
	storage storage
	cfg     config.Scheduler
}

func New(l *logger.Logger, st storage, cfg config.Scheduler) *Scheduler {
	return &Scheduler{
		logger:  l,
		storage: st,
		cfg:     cfg,
	}
}
