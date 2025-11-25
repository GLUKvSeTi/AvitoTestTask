package postgres

import (
	"AvitoTestTask/internal/domain"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepo struct {
	pool *pgxpool.Pool
}

func NewTeamRepo(pool *pgxpool.Pool) *TeamRepo {
	return &TeamRepo{pool: pool}
}

func (r *TeamRepo) CreateTeam(ctx context.Context, teamName string) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Printf("warning: failed to rollback transaction: %v", err)
		}
	}()
	if _, err := tx.Exec(ctx, "INSERT INTO teams(team_name) VALUES($1) ON CONFLICT (team_name) DO NOTHING", teamName); err != nil {
		return "", err
	}
	var id string
	if err := tx.QueryRow(ctx, "SELECT id::text FROM teams WHERE team_name=$1", teamName).Scan(&id); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return id, nil
}

func (r *TeamRepo) GetTeamByName(ctx context.Context, teamName string) (*domain.Team, error) {
	var teamID string
	if err := r.pool.QueryRow(ctx, "SELECT id::text FROM teams WHERE team_name=$1", teamName).Scan(&teamID); err != nil {
		return nil, errors.New("team not found")
	}
	rows, err := r.pool.Query(ctx, "SELECT id::text, username, is_active FROM users WHERE team_id=$1", teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var members []domain.TeamMember
	for rows.Next() {
		var uid, uname string
		var active bool
		if err := rows.Scan(&uid, &uname, &active); err != nil {
			return nil, err
		}
		members = append(members, domain.TeamMember{UserID: uid, Username: uname, IsActive: active})
	}
	return &domain.Team{TeamName: teamName, Members: members}, nil
}

func (r *TeamRepo) UpdateTeam(ctx context.Context, oldName, newName string) error {
	ct, err := r.pool.Exec(ctx, "UPDATE teams SET team_name=$1 WHERE team_name=$2", newName, oldName)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return fmt.Errorf("team not found")
	}
	return nil
}

func (r *TeamRepo) DeleteTeam(ctx context.Context, teamName string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Printf("warning: failed to rollback transaction: %v", err)
		}
	}()
	var teamID string
	if err := tx.QueryRow(ctx, "SELECT id::text FROM teams WHERE team_name=$1", teamName).Scan(&teamID); err != nil {
		return fmt.Errorf("team not found")
	}
	if _, err := tx.Exec(ctx, "UPDATE users SET team_id=NULL WHERE team_id=$1::uuid", teamID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM teams WHERE id=$1::uuid", teamID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
