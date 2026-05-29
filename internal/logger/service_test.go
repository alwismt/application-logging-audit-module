package logger

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type mockLogRepo struct {
	insertErr error
	last      LogEntry
}

func (m *mockLogRepo) Insert(ctx context.Context, entry LogEntry) error {
	m.last = entry
	return m.insertErr
}

func (m *mockLogRepo) Find(ctx context.Context, filter LogFilter) ([]LogEntry, error) {
	return nil, nil
}

func (m *mockLogRepo) FindByID(ctx context.Context, id uuid.UUID) (*LogEntry, error) {
	return nil, nil
}

func TestLoggerService_InfoCreatesValidEntry(t *testing.T) {
	repo := &mockLogRepo{}
	svc := NewService(repo, "test", false, true)
	err := svc.Info(context.Background(), "hello", map[string]any{"k": "v"})
	if err != nil {
		t.Fatal(err)
	}
	if repo.last.Level != "INFO" {
		t.Fatalf("expected INFO, got %s", repo.last.Level)
	}
	if repo.last.Message != "hello" {
		t.Fatalf("unexpected message: %s", repo.last.Message)
	}
}

func TestLoggerService_RejectsInvalidLevel(t *testing.T) {
	repo := &mockLogRepo{}
	svc := NewService(repo, "test", false, true)
	err := svc.log(context.Background(), "INVALID", "msg", "", nil, nil)
	if err == nil {
		t.Fatal("expected error for invalid level")
	}
}

func TestLoggerService_SanitizesMetadata(t *testing.T) {
	repo := &mockLogRepo{}
	svc := NewService(repo, "test", false, true)
	_ = svc.Info(context.Background(), "msg", map[string]any{"password": "x"})
	if repo.last.Metadata["password"] != "***MASKED***" {
		t.Fatalf("expected sanitized metadata, got %v", repo.last.Metadata["password"])
	}
}

func TestLoggerService_RepositoryError(t *testing.T) {
	repo := &mockLogRepo{insertErr: errors.New("db down")}
	svc := NewService(repo, "test", false, true)
	err := svc.Error(context.Background(), "fail", errors.New("err"), nil)
	if err == nil {
		t.Fatal("expected repository error")
	}
}
