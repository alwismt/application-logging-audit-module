package audit

import (
	"context"

	"github.com/google/uuid"
)

type Auditor interface {
	Record(ctx context.Context, event AuditEvent) error
	Find(ctx context.Context, filter AuditFilter) ([]AuditEvent, error)
}

type AuditRepository interface {
	Insert(ctx context.Context, event AuditEvent) error
	Find(ctx context.Context, filter AuditFilter) ([]AuditEvent, error)
	FindByID(ctx context.Context, id uuid.UUID) (*AuditEvent, error)
}
