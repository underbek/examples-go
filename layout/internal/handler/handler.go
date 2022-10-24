package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"layout/domain"

	"go.uber.org/zap"
)

type Handler struct {
	logger  *zap.Logger
	useCase useCase
}

//go:generate mockery --name=useCase --structname=useCaseMock --filename=usecase_mock.go --inpackage
type useCase interface {
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUser(ctx context.Context, userID int64) (domain.User, error)

	CreateOrder(ctx context.Context, order domain.Order) (domain.Order, error)
	GetOrder(ctx context.Context, orderID int64) (domain.Order, error)
}

func New(logger *zap.Logger, useCase useCase) *Handler {
	return &Handler{
		logger:  logger,
		useCase: useCase,
	}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
