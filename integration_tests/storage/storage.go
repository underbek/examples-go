package storage

import (
	"errors"
	"github.com/AndreyAndreevich/examples-go/integration_tests/domain"
)

type storage struct {
	data map[int]domain.User
}

var (
	AlreadyExistsErr = errors.New("already exists")
	NotFoundErr      = errors.New("not found")
)

func New() *storage {
	return &storage{
		data: make(map[int]domain.User),
	}
}

func (s *storage) AddUser(user domain.User) error {
	if _, ok := s.data[user.Id]; ok {
		return AlreadyExistsErr
	}
	s.data[user.Id] = user
	return nil
}

func (s *storage) GetUser(id int) (domain.User, error) {
	user, ok := s.data[id]
	if !ok {
		return domain.User{}, NotFoundErr
	}
	return user, nil
}

func (s *storage) UpdateUser(user domain.User) error {
	if _, ok := s.data[user.Id]; !ok {
		return NotFoundErr
	}
	s.data[user.Id] = user
	return nil
}
