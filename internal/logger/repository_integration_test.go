//go:build integration

package logger

import (
	"context"
	"testing"
	"time"

	"application-logging-audit-module/internal/common"
	"application-logging-audit-module/internal/database"

	"github.com/google/uuid"
)

func TestPostgresLogRepository_InsertAndFind(t *testing.T) {
	pool := database.SetupTestPool(t)
	defer pool.Close()
	database.TruncateTables(t, pool)

	repo := NewPostgresRepository(pool)
	entry := LogEntry{
		ID:        uuid.New(),
		Level:     "ERROR",
		Message:   "integration test",
		Source:    "test",
		RequestID: "req-1",
		Metadata:  map[string]any{"k": "v"},
		CreatedAt: time.Now().UTC(),
	}
	if err := repo.Insert(context.Background(), entry); err != nil {
		t.Fatal(err)
	}

	found, err := repo.FindByID(context.Background(), entry.ID)
	if err != nil || found == nil {
		t.Fatalf("find by id: %v", err)
	}

	results, err := repo.Find(context.Background(), LogFilter{
		Level:      "ERROR",
		RequestID:  "req-1",
		Pagination: common.Pagination{Page: 1, Limit: 10},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
