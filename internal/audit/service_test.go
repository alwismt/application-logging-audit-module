package audit

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type mockAuditRepo struct {
	insertErr error
	last      AuditEvent
}

func (m *mockAuditRepo) Insert(ctx context.Context, event AuditEvent) error {
	m.last = event
	return m.insertErr
}

func (m *mockAuditRepo) Find(ctx context.Context, filter AuditFilter) ([]AuditEvent, error) {
	return nil, nil
}

func (m *mockAuditRepo) FindByID(ctx context.Context, id uuid.UUID) (*AuditEvent, error) {
	return nil, nil
}

func TestAuditService_RecordValidEvent(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := NewService(repo)
	err := svc.Record(context.Background(), AuditEvent{
		Action: "LOGIN",
		Status: "SUCCESS",
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.last.Action != "LOGIN" {
		t.Fatalf("expected LOGIN, got %s", repo.last.Action)
	}
}

func TestAuditService_RejectsInvalidStatus(t *testing.T) {
	svc := NewService(&mockAuditRepo{})
	err := svc.Record(context.Background(), AuditEvent{Action: "LOGIN", Status: "UNKNOWN"})
	if err == nil {
		t.Fatal("expected invalid status error")
	}
}

func TestAuditService_RequiresActionAndStatus(t *testing.T) {
	svc := NewService(&mockAuditRepo{})
	if err := svc.Record(context.Background(), AuditEvent{Status: "SUCCESS"}); err == nil {
		t.Fatal("expected action required error")
	}
	if err := svc.Record(context.Background(), AuditEvent{Action: "LOGIN"}); err == nil {
		t.Fatal("expected status required error")
	}
}

func TestAuditService_SanitizesValues(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := NewService(repo)
	_ = svc.Record(context.Background(), AuditEvent{
		Action:   "UPDATE_RECORD",
		Status:   "SUCCESS",
		OldValue: map[string]any{"password": "old"},
		NewValue: map[string]any{"token": "new"},
	})
	if repo.last.OldValue["password"] != "***MASKED***" {
		t.Fatal("old_value not sanitized")
	}
	if repo.last.NewValue["token"] != "***MASKED***" {
		t.Fatal("new_value not sanitized")
	}
}

func TestAuditService_RepositoryError(t *testing.T) {
	repo := &mockAuditRepo{insertErr: errors.New("db error")}
	svc := NewService(repo)
	err := svc.Record(context.Background(), AuditEvent{Action: "LOGIN", Status: "SUCCESS"})
	if err == nil {
		t.Fatal("expected repository error")
	}
}
