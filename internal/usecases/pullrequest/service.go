package pullrequest

import (
	"AvitoTestTask/internal/domain"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type service struct {
	repo     Repository
	teamRepo TeamRepository
	userRepo UserRepository
	limit    int
}

func NewService(r Repository, t TeamRepository, u UserRepository) Service {
	return &service{repo: r, teamRepo: t, userRepo: u, limit: 2}
}

func (s *service) CreatePRWithAssignments(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	if _, err := uuid.Parse(prID); err != nil {
		return nil, fmt.Errorf("invalid pr id: %w", err)
	}
	if _, err := uuid.Parse(authorID); err != nil {
		return nil, fmt.Errorf("invalid author id: %w", err)
	}
	user, err := s.userRepo.GetUserByID(ctx, authorID)
	if err != nil {
		return nil, err
	}
	team, err := s.teamRepo.GetTeamByName(ctx, *user.TeamName)
	if err != nil {
		return nil, err
	}
	members := team.Members
	pr := &domain.PullRequest{
		ID:       prID,
		Name:     prName,
		AuthorID: authorID,
		Status:   domain.StatusOpen,
	}
	pr.AssignReviewersFromMembers(members, s.limit)
	if err := s.repo.CreatePR(ctx, pr); err != nil {
		return nil, err
	}
	if len(pr.AssignedReviewers) > 0 {
		if err := s.repo.SavePRReviewers(ctx, pr.ID, pr.AssignedReviewers); err != nil {
			return nil, err
		}
	}
	return pr, nil
}

func (s *service) ReassignReviewer(ctx context.Context, prID, oldUserID string) (string, *domain.PullRequest, error) {
	pr, err := s.repo.GetPRByID(ctx, prID)
	if err != nil {
		return "", nil, err
	}
	if pr.Status == domain.StatusMerged {
		return "", nil, domain.ErrPRMerged
	}
	oldUser, err := s.userRepo.GetUserByID(ctx, oldUserID)
	if err != nil {
		return "", nil, err
	}
	team, err := s.teamRepo.GetTeamByName(ctx, *oldUser.TeamName)
	if err != nil {
		return "", nil, err
	}
	members := team.Members
	var candidate string
	for _, member := range members {
		if !member.IsActive || member.UserID == oldUserID {
			continue
		}
		already := false
		for _, r := range pr.AssignedReviewers {
			if r == member.UserID {
				already = true
				break
			}
		}
		if !already {
			candidate = member.UserID
			break
		}
	}
	if candidate == "" {
		return "", nil, domain.ErrNoCandidate
	}
	if err := pr.Reassign(oldUserID, candidate); err != nil {
		return "", nil, err
	}
	if err := s.repo.SavePRReviewers(ctx, pr.ID, pr.AssignedReviewers); err != nil {
		return "", nil, err
	}
	return candidate, pr, nil
}

func (s *service) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.repo.GetPRByID(ctx, prID)
	if err != nil {
		return nil, err
	}
	if pr.Status == domain.StatusMerged {
		return pr, nil
	}
	pr.Merge()
	if err := s.repo.UpdatePRStatus(ctx, pr.ID, string(pr.Status)); err != nil {
		return nil, err
	}
	return pr, nil
}

func (s *service) GetPRsForReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	return s.repo.GetPRsForReviewer(ctx, reviewerID)
}

func (s *service) GetPR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	return s.repo.GetPRByID(ctx, prID)
}

func (s *service) UpdatePR(ctx context.Context, pr *domain.PullRequest) error {
	if err := s.repo.UpdatePRStatus(ctx, pr.ID, string(pr.Status)); err != nil {
		return err
	}
	return s.repo.SavePRReviewers(ctx, pr.ID, pr.AssignedReviewers)
}

func (s *service) DeletePR(ctx context.Context, prID string) error {
	return s.repo.DeletePR(ctx, prID)
}
