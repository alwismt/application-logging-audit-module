package adminauth

import (
	"context"
	"fmt"
	"time"

	"application-logging-audit-module/internal/config"

	"github.com/google/uuid"
)

func SeedDefaultAdmin(ctx context.Context, repo Repository, cfg *config.Config) error {
	existing, err := repo.FindByUsername(ctx, cfg.AdminUsername)
	if err != nil {
		return fmt.Errorf("find admin user: %w", err)
	}
	if existing != nil {
		return nil
	}

	hash, err := HashPassword(cfg.AdminPassword)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	user := &AdminUser{
		ID:           uuid.New(),
		Username:     cfg.AdminUsername,
		PasswordHash: hash,
		CreatedAt:    time.Now().UTC(),
	}
	if err := repo.Create(ctx, user); err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}
	return nil
}
