package infra

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}
	config.MaxConns = 10
	return pgxpool.NewWithConfig(ctx, config)
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	stmts := []string{
		`CREATE EXTENSION IF NOT EXISTS "pgcrypto";`,
		`CREATE TABLE IF NOT EXISTS teams (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    team_name text NOT NULL UNIQUE,
    created_at timestamptz DEFAULT now()
);`,g
		`CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY,
    username text NOT NULL,
    team_id uuid REFERENCES teams(id) ON DELETE SET NULL,
    is_active boolean DEFAULT true,
    created_at timestamptz DEFAULT now()
);`,
		`CREATE TABLE IF NOT EXISTS pull_requests (
    id text PRIMARY KEY,
    name text,
    author_id uuid REFERENCES users(id) ON DELETE SET NULL,
    status text NOT NULL DEFAULT 'OPEN',
    created_at timestamptz DEFAULT now(),
    merged_at timestamptz
);`,
		`CREATE TABLE IF NOT EXISTS pull_request_reviewers (
    pull_request_id text REFERENCES pull_requests(id) ON DELETE CASCADE,
    user_id uuid REFERENCES users(id) ON DELETE CASCADE,
    assigned_at timestamptz DEFAULT now(),
    PRIMARY KEY (pull_request_id, user_id)
);`,
		`CREATE INDEX IF NOT EXISTS idx_users_team ON users(team_id);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_user ON pull_request_reviewers(user_id);`,
	}
	for _, stmt := range stmts {
		if _, err := pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}
