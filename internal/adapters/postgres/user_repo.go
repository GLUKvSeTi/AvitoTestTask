package postgres

import (
	"AvitoTestTask/internal/domain"
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) CreateUser(ctx context.Context, u domain.User) error {
	if _, err := uuid.Parse(u.ID); err != nil {
		return err
	}
	var teamUUID *uuid.UUID
	if u.TeamName != nil {
		t, err := uuid.Parse(*u.TeamName)
		if err != nil {
			return err
		}
		teamUUID = &t
	}
	if teamUUID != nil {
		_, err := r.pool.Exec(ctx, "INSERT INTO users(id, username, team_id, is_active, created_at) VALUES($1,$2,$3,$4,now())", u.ID, u.Username, *teamUUID, u.IsActive)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}
	} else {
		_, err := r.pool.Exec(ctx, "INSERT INTO users(id, username, is_active, created_at) VALUES($1,$2,$3,now())", u.ID, u.Username, u.IsActive)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}
	}
	return nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	if _, err := uuid.Parse(userID); err != nil {
		return nil, err
	}
	var username string
	var teamID *string
	var isActive bool
	if err := r.pool.QueryRow(ctx, "SELECT username, team_id::text, is_active FROM users WHERE id=$1", userID).Scan(&username, &teamID, &isActive); err != nil {
		return nil, err
	}
	return &domain.User{ID: userID, Username: username, TeamName: teamID, IsActive: isActive}, nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, u domain.User) error {
	if _, err := uuid.Parse(u.ID); err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			log.Printf("warning: failed to rollback transaction: %v", err)
		}
	}()
	if _, err := tx.Exec(ctx, "UPDATE users SET username=$1, is_active=$2 WHERE id=$3", u.Username, u.IsActive, u.ID); err != nil {
		return err
	}
	if u.TeamName == nil {
		if _, err := tx.Exec(ctx, "UPDATE users SET team_id=NULL WHERE id=$1", u.ID); err != nil {
			return err
		}
	} else {
		if _, err := uuid.Parse(*u.TeamName); err == nil {
			if _, err := tx.Exec(ctx, "UPDATE users SET team_id=$1 WHERE id=$2", *u.TeamName, u.ID); err != nil {
				return err
			}
		} else {
			var teamID string
			if err := tx.QueryRow(ctx, "SELECT id::text FROM teams WHERE team_name=$1", *u.TeamName).Scan(&teamID); err != nil {
				return errors.New("team not found")
			}
			if _, err := tx.Exec(ctx, "UPDATE users SET team_id=$1::uuid WHERE id=$2", teamID, u.ID); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func (r *UserRepo) DeleteUser(ctx context.Context, userID string) error {
	if _, err := uuid.Parse(userID); err != nil {
		return err
	}
	ct, err := r.pool.Exec(ctx, "DELETE FROM users WHERE id=$1", userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *UserRepo) SetUserTeamByName(ctx context.Context, userID string, teamName *string) error {
	if _, err := uuid.Parse(userID); err != nil {
		return err
	}
	if teamName == nil {
		_, err := r.pool.Exec(ctx, "UPDATE users SET team_id=NULL WHERE id=$1", userID)
		return err
	}
	var teamID string
	if err := r.pool.QueryRow(ctx, "SELECT id::text FROM teams WHERE team_name=$1", *teamName).Scan(&teamID); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, "UPDATE users SET team_id=$1::uuid WHERE id=$2", teamID, userID)
	return err
}
