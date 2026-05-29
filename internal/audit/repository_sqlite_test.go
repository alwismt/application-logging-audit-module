package audit

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/alwismt/application-logging-audit-module/internal/database"

	"github.com/google/uuid"
)

func setupSQLiteAuditRepo(t *testing.T) (*SQLiteAuditRepository, *sql.DB) {
	t.Helper()
	db, err := database.ConnectSQLite(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := database.EnsureSchemaSQLite(ctx, db, true); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	return NewSQLiteRepository(db), db
}

func TestSQLiteAuditRepository_InsertAndFind(t *testing.T) {
	repo, db := setupSQLiteAuditRepo(t)
	defer db.Close()

	ctx := context.Background()
	id := uuid.New()
	userID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	event := AuditEvent{
		ID:       id,
		UserID:   &userID,
		Username: "alice",
		Action:   "LOGIN",
		Status:   "SUCCESS",
		Metadata: map[string]any{"ip": "127.0.0.1"},
		CreatedAt: now,
	}
	if err := repo.Insert(ctx, event); err != nil {
		t.Fatal(err)
	}

	got, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected event")
	}
	if got.Action != event.Action {
		t.Fatalf("action: got %q want %q", got.Action, event.Action)
	}

	events, err := repo.Find(ctx, AuditFilter{Action: "LOGIN"})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("find count: %d", len(events))
	}
}
