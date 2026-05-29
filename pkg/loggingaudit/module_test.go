package loggingaudit_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alwismt/application-logging-audit-module/pkg/loggingaudit"
	"github.com/google/uuid"
)

func TestNewFromEnvSQLite(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	t.Setenv("DB_DRIVER", loggingaudit.DriverSQLite)
	t.Setenv("SQLITE_PATH", dbPath)
	t.Setenv("DB_AUTO_MIGRATE", "true")

	mod, err := loggingaudit.NewFromEnv()
	if err != nil {
		t.Fatalf("NewFromEnv: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := mod.Logger().Info(ctx, "pkg smoke test", map[string]any{"test": true}); err != nil {
		t.Fatalf("Info: %v", err)
	}

	userID := uuid.New()
	if err := mod.Auditor().Record(ctx, loggingaudit.AuditEvent{
		UserID:       &userID,
		Username:     "tester",
		Action:       "CREATE_RECORD",
		ResourceType: "demo",
		ResourceID:   "1",
		Status:       "SUCCESS",
	}); err != nil {
		t.Fatalf("Record: %v", err)
	}

	if mod.Handler() == nil {
		t.Fatal("Handler() returned nil")
	}

	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("sqlite file: %v", err)
	}
}
