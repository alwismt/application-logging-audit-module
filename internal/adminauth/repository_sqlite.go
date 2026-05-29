package adminauth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) FindByUsername(ctx context.Context, username string) (*AdminUser, error) {
	const q = `
		SELECT id, username, password_hash, created_at
		FROM admin_users
		WHERE username = ?`

	var (
		idStr    string
		user     AdminUser
		created  string
	)
	err := r.db.QueryRowContext(ctx, q, username).Scan(
		&idStr, &user.Username, &user.PasswordHash, &created,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find admin user: %w", err)
	}
	user.ID, err = uuid.Parse(idStr)
	if err != nil {
		return nil, fmt.Errorf("parse admin id: %w", err)
	}
	t, err := time.Parse(time.RFC3339Nano, created)
	if err != nil {
		t, err = time.Parse("2006-01-02 15:04:05", created)
		if err != nil {
			return nil, fmt.Errorf("parse created_at: %w", err)
		}
	}
	user.CreatedAt = t
	return &user, nil
}

func (r *SQLiteRepository) Create(ctx context.Context, user *AdminUser) error {
	const q = `
		INSERT INTO admin_users (id, username, password_hash, created_at)
		VALUES (?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, q,
		user.ID.String(),
		user.Username,
		user.PasswordHash,
		user.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("insert admin user: %w", err)
	}
	return nil
}
