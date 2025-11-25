package user

import (
	"AvitoTestTask/internal/domain"
	"context"
	"errors"
	"github.com/google/uuid"
)

type service struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &service{repository: r}
}

func (s *service) CreateUser(ctx context.Context, u domain.User) error {
	if _, err := uuid.Parse(u.ID); err != nil {
		return errors.New("invalid user_id")
	}
	return s.repository.CreateUser(ctx, u)
}

func (s *service) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, errors.New("invalid user_id")
	}
	return s.repository.GetUserByID(ctx, userID)
}

func (s *service) UpdateUser(ctx context.Context, u domain.User) error {
	if _, err := uuid.Parse(u.ID); err != nil {
		return errors.New("invalid user_id")
	}
	return s.repository.UpdateUser(ctx, u)
}

func (s *service) DeleteUser(ctx context.Context, userID string) error {
	if _, err := uuid.Parse(userID); err != nil {
		return errors.New("invalid user_id")
	}
	return s.repository.DeleteUser(ctx, userID)
}
