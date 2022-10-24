package usecase

import (
	"context"

	"layout/domain"
)

func (s *service) CreateUser(ctx context.Context, user domain.User) (domain.User, error) {
	userID, err := s.storage.CreateUser(ctx, user)
	if err != nil {
		return domain.User{}, err
	}

	return s.storage.GetUser(ctx, userID)
}

func (s *service) GetUser(ctx context.Context, userID int64) (domain.User, error) {
	return s.storage.GetUser(ctx, userID)
}
