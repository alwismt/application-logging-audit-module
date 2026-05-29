//go:build integration

package audit

import (
	"context"
	"testing"
	"time"

	"application-logging-audit-module/internal/common"
	"application-logging-audit-module/internal/database"

	"github.com/google/uuid"
)

func TestPostgresAuditRepository_InsertAndFind(t *testing.T) {
	pool := database.SetupTestPool(t)
	defer pool.Close()
	database.TruncateTables(t, pool)

	repo := NewPostgresRepository(pool)
	uid := uuid.New()
	event := AuditEvent{
		ID:           uuid.New(),
		UserID:       &uid,
		Username:     "tester",
		Action:       "UPDATE_RECORD",
		ResourceType: "invoice",
		ResourceID:   "1001",
		Status:       "SUCCESS",
		CreatedAt:    time.Now().UTC(),
	}
	if err := repo.Insert(context.Background(), event); err != nil {
		t.Fatal(err)
	}

	results, err := repo.Find(context.Background(), AuditFilter{
		UserID:     &uid,
		Action:     "UPDATE_RECORD",
		Status:     "SUCCESS",
		Pagination: common.Pagination{Page: 1, Limit: 20},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 event, got %d", len(results))
	}
}
