package logic

import (
	"context"

	"github.com/AndreyAndreevich/examples-go/integration_tests/domain"
	"github.com/shopspring/decimal"
)

//go:generate mockery --name=storage --structname=storageMock --filename=storage_mock.go --inpackage
type storage interface {
	CreatUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUser(ctx context.Context, id int) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
}

type logic struct {
	storage storage
}

func New(storage storage) *logic {
	return &logic{
		storage: storage,
	}
}

func (l *logic) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {

	user, err := l.storage.CreatUser(ctx, user)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (l *logic) GetUser(ctx context.Context, id int) (domain.User, error) {
	return l.storage.GetUser(ctx, id)
}

func (l *logic) AddBalance(ctx context.Context, id int, amount decimal.Decimal) (domain.User, error) {
	user, err := l.storage.GetUser(ctx, id)
	if err != nil {
		return domain.User{}, err
	}

	user.Balance = user.Balance.Add(amount)
	return l.storage.UpdateUser(ctx, user)
}
