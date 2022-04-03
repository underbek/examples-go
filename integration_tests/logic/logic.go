package logic

import (
	"math/rand"
	"time"

	"github.com/AndreyAndreevich/examples-go/integration_tests/domain"
)

//go:generate mockery --name=storage --structname=storageMock --filename=storage_mock.go --inpackage
type storage interface {
	AddUser(user domain.User) error
	GetUser(id int) (domain.User, error)
	UpdateUser(user domain.User) error
}

type logic struct {
	storage storage
}

func New(storage storage) *logic {
	rand.Seed(time.Now().UnixNano())
	return &logic{
		storage: storage,
	}
}

func (l *logic) CreateUser(name string, balance float64) (domain.User, error) {
	user := domain.User{
		Id:      rand.Int(),
		Name:    name,
		Balance: balance,
	}

	if err := l.storage.AddUser(user); err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (l *logic) GetUser(id int) (domain.User, error) {
	return l.storage.GetUser(id)
}

func (l *logic) AddBalance(id int, amount float64) (domain.User, error) {
	user, err := l.storage.GetUser(id)
	if err != nil {
		return domain.User{}, err
	}

	user.Balance += amount
	if err := l.storage.UpdateUser(user); err != nil {
		return domain.User{}, err
	}

	return user, nil
}
