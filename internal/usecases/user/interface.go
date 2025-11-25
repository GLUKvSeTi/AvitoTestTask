package user

import (
	"AvitoTestTask/internal/domain"
	"context"
)

type Repository interface {
	CreateUser(ctx context.Context, u domain.User) error
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
	UpdateUser(ctx context.Context, u domain.User) error
	DeleteUser(ctx context.Context, userID string) error
	SetUserTeamByName(ctx context.Context, userID string, teamName *string) error
}

type Service interface {
	CreateUser(ctx context.Context, u domain.User) error
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	UpdateUser(ctx context.Context, u domain.User) error
	DeleteUser(ctx context.Context, userID string) error
}
