package postgres

import (
	"AvitoTestTask/internal/domain"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PRRepo struct {
	pool *pgxpool.Pool
}

func NewPRRepo(pool *pgxpool.Pool) *PRRepo {
	return &PRRepo{pool: pool}
}

func (r *PRRepo) CreatePR(ctx context.Context, pr *domain.PullRequest) error {
	_, err := r.pool.Exec(ctx, "INSERT INTO pull_requests(id, name, author_id, status, created_at) VALUES($1,$2,$3,$4,now())", pr.ID, pr.Name, pr.AuthorID, pr.Status)
	if err != nil {
		return err
	}
	if len(pr.AssignedReviewers) > 0 {
		parts := make([]string, 0, len(pr.AssignedReviewers))
		args := make([]interface{}, 0, len(pr.AssignedReviewers)*2)
		for i, uid := range pr.AssignedReviewers {
			parts = append(parts, fmt.Sprintf("($%d,$%d)", i*2+1, i*2+2))
			args = append(args, pr.ID, uid)
		}
		q := "INSERT INTO pull_request_reviewers(pull_request_id, user_id) VALUES " + strings.Join(parts, ",")
		if _, err := r.pool.Exec(ctx, q, args...); err != nil {
			return err
		}
	}
	return nil
}

func (r *PRRepo) SavePRReviewers(ctx context.Context, prID string, reviewerIDs []string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "DELETE FROM pull_request_reviewers WHERE pull_request_id=$1", prID); err != nil {
		return err
	}
	if len(reviewerIDs) > 0 {
		parts := make([]string, 0, len(reviewerIDs))
		args := make([]interface{}, 0, len(reviewerIDs)*2)
		for i, uid := range reviewerIDs {
			parts = append(parts, fmt.Sprintf("($%d,$%d)", i*2+1, i*2+2))
			args = append(args, prID, uid)
		}
		q := "INSERT INTO pull_request_reviewers(pull_request_id, user_id) VALUES " + strings.Join(parts, ",")
		if _, err := tx.Exec(ctx, q, args...); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *PRRepo) GetPRByID(ctx context.Context, prID string) (*domain.PullRequest, error) {
	var id, name, authorID, status string
	var createdAt, mergedAt *time.Time
	if err := r.pool.QueryRow(ctx, "SELECT id, name, author_id::text, status, created_at, merged_at FROM pull_requests WHERE id=$1", prID).Scan(&id, &name, &authorID, &status, &createdAt, &mergedAt); err != nil {
		return nil, err
	}
	pr := &domain.PullRequest{ID: id, Name: name, AuthorID: authorID, Status: domain.PRStatus(status), CreatedAt: createdAt, MergedAt: mergedAt}
	rows, err := r.pool.Query(ctx, "SELECT user_id::text FROM pull_request_reviewers WHERE pull_request_id=$1", prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, uid)
	}
	return pr, nil
}

func (r *PRRepo) UpdatePRStatus(ctx context.Context, prID, status string) error {
	if strings.ToUpper(status) == "MERGED" {
		_, err := r.pool.Exec(ctx, "UPDATE pull_requests SET status='MERGED', merged_at=now() WHERE id=$1", prID)
		return err
	}
	_, err := r.pool.Exec(ctx, "UPDATE pull_requests SET status=$2 WHERE id=$1", prID, status)
	return err
}

func (r *PRRepo) GetPRsForReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	rows, err := r.pool.Query(ctx, `
SELECT pr.id, pr.name, pr.author_id::text, pr.status
FROM pull_requests pr
JOIN pull_request_reviewers rr ON pr.id = rr.pull_request_id
WHERE rr.user_id = $1
ORDER BY pr.created_at DESC
`, reviewerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.PullRequest
	for rows.Next() {
		var p domain.PullRequest
		if err := rows.Scan(&p.ID, &p.Name, &p.AuthorID, &p.Status); err != nil {
			return nil, err
		}
		rvRows, err := r.pool.Query(ctx, "SELECT user_id::text FROM pull_request_reviewers WHERE pull_request_id=$1", p.ID)
		if err == nil {
			for rvRows.Next() {
				var rid string
				_ = rvRows.Scan(&rid)
				p.AssignedReviewers = append(p.AssignedReviewers, rid)
			}
			rvRows.Close()
		}
		out = append(out, p)
	}
	return out, nil
}

func (r *PRRepo) UpdatePRName(ctx context.Context, prID, name string) error {
	_, err := r.pool.Exec(ctx, "UPDATE pull_requests SET name=$1 WHERE id=$2", name, prID)
	return err
}

func (r *PRRepo) DeletePR(ctx context.Context, prID string) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM pull_requests WHERE id=$1", prID)
	return err
}
