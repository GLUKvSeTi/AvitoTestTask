package pullrequest

import (
	"AvitoTestTask/internal/domain"
	"context"
)

type Repository interface {
	CreatePR(ctx context.Context, pr *domain.PullRequest) error
	SavePRReviewers(ctx context.Context, prID string, reviewerIDs []string) error
	GetPRByID(ctx context.Context, prID string) (*domain.PullRequest, error)
	UpdatePRStatus(ctx context.Context, prID, status string) error
	GetPRsForReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error)
	UpdatePRName(ctx context.Context, prID, name string) error
	DeletePR(ctx context.Context, prID string) error
}

type TeamRepository interface {
	GetTeamByName(ctx context.Context, userID string) (*domain.Team, error)
}

type UserRepository interface {
	GetUserByID(ctx context.Context, userID string) (*domain.User, error)
}

type Service interface {
	CreatePRWithAssignments(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (string, *domain.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*domain.PullRequest, error)
	GetPRsForReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error)
	GetPR(ctx context.Context, prID string) (*domain.PullRequest, error)
	UpdatePR(ctx context.Context, pr *domain.PullRequest) error
	DeletePR(ctx context.Context, prID string) error
}
