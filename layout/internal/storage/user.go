package storage

import (
	"context"
	"fmt"

	"layout/domain"
)

func (s *storage) CreateUser(_ context.Context, user domain.User) (int64, error) {
	s.userMtx.Lock()
	defer s.userMtx.Unlock()

	id := s.currentUserID
	s.currentUserID++
	user.ID = id
	if _, ok := s.users[id]; ok {
		return 0, fmt.Errorf("user with id %d already exists", id)
	}
	s.users[id] = user
	return id, nil
}

func (s *storage) GetUser(_ context.Context, userID int64) (domain.User, error) {
	s.userMtx.RLock()
	defer s.userMtx.RUnlock()

	user, ok := s.users[userID]
	if !ok {
		return domain.User{}, fmt.Errorf("iser with id %d is not exists", userID)
	}
	return user, nil
}
