package audit

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/alwismt/application-logging-audit-module/internal/logger"

	"github.com/google/uuid"
)

type AuditService struct {
	repo      AuditRepository
	sanitizer *logger.SensitiveDataSanitizer
}

func NewService(repo AuditRepository) *AuditService {
	return &AuditService{
		repo:      repo,
		sanitizer: logger.NewSanitizer(),
	}
}

func (s *AuditService) Record(ctx context.Context, event AuditEvent) error {
	action := strings.ToUpper(strings.TrimSpace(event.Action))
	status := strings.ToUpper(strings.TrimSpace(event.Status))

	if action == "" {
		return fmt.Errorf("audit action is required")
	}
	if status == "" {
		return fmt.Errorf("audit status is required")
	}
	if !slices.Contains(ValidActions, action) {
		return fmt.Errorf("invalid audit action: %s", action)
	}
	if !slices.Contains(ValidStatuses, status) {
		return fmt.Errorf("invalid audit status: %s", status)
	}

	event.Action = action
	event.Status = status
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now().UTC()
	}
	event.OldValue = s.sanitizer.SanitizeMap(event.OldValue)
	event.NewValue = s.sanitizer.SanitizeMap(event.NewValue)
	event.Metadata = s.sanitizer.SanitizeMap(event.Metadata)

	return s.repo.Insert(ctx, event)
}

func (s *AuditService) Find(ctx context.Context, filter AuditFilter) ([]AuditEvent, error) {
	return s.repo.Find(ctx, filter)
}
