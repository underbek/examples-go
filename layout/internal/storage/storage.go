package storage

import (
	"sync"

	"layout/domain"

	"go.uber.org/zap"
)

type storage struct {
	logger *zap.Logger
	users  map[int64]domain.User
	orders map[int64]domain.Order

	currentUserID  int64
	currentOrderID int64

	userMtx  sync.RWMutex
	orderMtx sync.RWMutex
}

func New(logger *zap.Logger) *storage {
	return &storage{
		logger: logger,
		users:  make(map[int64]domain.User),
		orders: make(map[int64]domain.Order),
	}
}
