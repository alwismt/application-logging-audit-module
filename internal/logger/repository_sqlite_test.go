package logger

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/alwismt/application-logging-audit-module/internal/database"

	"github.com/google/uuid"
)

func setupSQLiteLogRepo(t *testing.T) (*SQLiteLogRepository, *sql.DB) {
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

func TestSQLiteLogRepository_InsertAndFind(t *testing.T) {
	repo, db := setupSQLiteLogRepo(t)
	defer db.Close()

	ctx := context.Background()
	id := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	entry := LogEntry{
		ID:        id,
		Level:     "INFO",
		Message:   "sqlite test",
		Source:    "test",
		RequestID: "req-1",
		Metadata:  map[string]any{"k": "v"},
		CreatedAt: now,
	}
	if err := repo.Insert(ctx, entry); err != nil {
		t.Fatal(err)
	}

	got, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected entry")
	}
	if got.Message != entry.Message {
		t.Fatalf("message: got %q want %q", got.Message, entry.Message)
	}

	entries, err := repo.Find(ctx, LogFilter{Level: "INFO"})
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("find count: %d", len(entries))
	}
}
