package team

import (
	"AvitoTestTask/internal/domain"
	"context"
)

type Repository interface {
	CreateTeam(ctx context.Context, teamName string) (string, error)
	GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error)
	UpdateTeam(ctx context.Context, oldName, newName string) error
	DeleteTeam(ctx context.Context, teamName string) error
}

type Service interface {
	CreateTeam(ctx context.Context, teamName string) (string, error)
	GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error)
	UpdateTeam(ctx context.Context, oldName, newName string) error
	DeleteTeam(ctx context.Context, teamName string) error
}
