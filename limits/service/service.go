package service

import (
	"context"
	"time"

	"github.com/underbek/examples-go/limits/config"
	"github.com/underbek/examples-go/limits/domain"
	"github.com/underbek/examples-go/logger"
	goKitPgx "github.com/underbek/examples-go/storage/pgx"
)

type Storage interface {
	CreateLimit(context.Context, domain.Limit) (domain.Limit, error)
	DeleteLimits(context.Context, []uint64, []uint64) ([]uint64, error)
	DeleteCounters(context.Context, []uint64) error
	GetLimitByID(context.Context, uint64) (domain.Limit, error)
	UpdateLimitValue(context.Context, domain.Limit) (domain.Limit, error)
	GetLimits(context.Context, domain.LimitsFilter) ([]domain.Limit, uint64, error)

	CreateContext(context.Context, domain.Attributes) (uint64, error)
	GetContextByID(context.Context, uint64) (domain.Context, error)
	UpdateContext(context.Context, domain.Context) error
	MatchLimits(context.Context, string, domain.Attributes) ([]domain.Limit, error)
	CreateOperation(context.Context, domain.Operation) (domain.Operation, error)
	GetOperationsByContextID(context.Context, uint64) ([]domain.Operation, error)
	CreateCountersIfNotExists(context.Context, []domain.Counter) ([]uint64, error)
	LinkCountersToOperation(ctx context.Context, counterIDs []uint64, operationID uint64) error
	IncrementCounters(context.Context, uint64, []uint64) ([]domain.ExceededCounters, error)
	IncrementCountersAndUpdateContext(context.Context, uint64, []uint64, domain.Context) ([]domain.ExceededCounters, error)

	CommitOperations(context.Context, []uint64) error
	RollbackOperations(context.Context, []uint64) ([]domain.ExceededCounters, error)
}

type timeProvider interface {
	Now() time.Time
}

type createStorage = func(ext goKitPgx.ExtContext) Storage

type service struct {
	logger        *logger.Logger
	storageTrxCfg config.StorageTransaction
	db            goKitPgx.Storage
	createStorage createStorage
	timeProvider  timeProvider
}

func New(
	logger *logger.Logger,
	storageTrxCfg config.StorageTransaction,
	db goKitPgx.Storage,
	createStorage createStorage,
	timeProvider timeProvider,
) *service {
	return &service{
		logger:        logger,
		storageTrxCfg: storageTrxCfg,
		db:            db,
		createStorage: createStorage,
		timeProvider:  timeProvider,
	}
}
