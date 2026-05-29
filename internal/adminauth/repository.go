package adminauth

import (
	"context"
)

type Repository interface {
	FindByUsername(ctx context.Context, username string) (*AdminUser, error)
	Create(ctx context.Context, user *AdminUser) error
}
