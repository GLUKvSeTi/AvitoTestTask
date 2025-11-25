package team

import (
	"AvitoTestTask/internal/domain"
	"context"
)

type service struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &service{repository: r}
}

func (s *service) CreateTeam(ctx context.Context, teamName string) (string, error) {
	return s.repository.CreateTeam(ctx, teamName)
}

func (s *service) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	return s.repository.GetTeamByName(ctx, teamName)
}

func (s *service) UpdateTeam(ctx context.Context, oldName, NewName string) error {
	return s.repository.UpdateTeam(ctx, oldName, NewName)
}

func (s *service) DeleteTeam(ctx context.Context, teamName string) error {
	return s.repository.DeleteTeam(ctx, teamName)
}
