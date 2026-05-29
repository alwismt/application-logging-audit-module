package adminauth

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) FindByUsername(ctx context.Context, username string) (*AdminUser, error) {
	const q = `
		SELECT id, username, password_hash, created_at
		FROM admin_users
		WHERE username = $1`

	var user AdminUser
	err := r.pool.QueryRow(ctx, q, username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find admin user: %w", err)
	}
	return &user, nil
}

func (r *PostgresRepository) Create(ctx context.Context, user *AdminUser) error {
	const q = `
		INSERT INTO admin_users (id, username, password_hash, created_at)
		VALUES ($1, $2, $3, $4)`
	_, err := r.pool.Exec(ctx, q, user.ID, user.Username, user.PasswordHash, user.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert admin user: %w", err)
	}
	return nil
}

// Ensure interface compliance at compile time.
var _ Repository = (*PostgresRepository)(nil)
