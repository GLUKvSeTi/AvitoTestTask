package domain

import "time"

type PRStatus string

const (
	StatusOpen   PRStatus = "OPEN"
	StatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	ID                string     `json:"pull_request_id"`
	Name              string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            PRStatus   `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

func (pr *PullRequest) AssignReviewersFromMembers(members []TeamMember, limit int) {
	candidates := make([]string, 0, 2)
	for _, member := range members {
		if member.IsActive && member.UserID != pr.AuthorID {
			candidates = append(candidates, member.UserID)
		}
		if len(candidates) == limit {
			break
		}
	}
	pr.AssignedReviewers = candidates
}

func (pr *PullRequest) Reassign(oldReviewer string, newReviewer string) error {
	if pr.Status == StatusMerged {
		return ErrPRMerged
	}
	for i, candidate := range pr.AssignedReviewers {
		if candidate == oldReviewer {
			pr.AssignedReviewers[i] = newReviewer
			return nil
		}
	}
	return ErrReviewerNotAssigned
}

func (pr *PullRequest) Merge() {
	if pr.Status == StatusMerged {
		return
	}
	pr.Status = StatusMerged
	now := time.Now().UTC()
	pr.MergedAt = &now
}
