package usecase

import (
	"context"

	"layout/domain"

	"go.uber.org/zap"
)

//go:generate mockery --name=storage --structname=storageMock --filename=storage_mock.go --inpackage
type storage interface {
	CreateUser(ctx context.Context, user domain.User) (int64, error)
	GetUser(ctx context.Context, userID int64) (domain.User, error)

	CreateOrder(ctx context.Context, order domain.Order) (int64, error)
	GetOrder(ctx context.Context, orderID int64) (domain.Order, error)
}

type service struct {
	logger  *zap.Logger
	storage storage
}

func New(logger *zap.Logger, storage storage) *service {
	return &service{
		logger:  logger,
		storage: storage,
	}
}
