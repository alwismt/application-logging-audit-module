//go:build integration

package database

import (
	"context"
	"testing"

	"github.com/alwismt/application-logging-audit-module/internal/config"
)

func TestConnect(t *testing.T) {
	pool := SetupTestPool(t)
	defer pool.Close()
	if err := Ping(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
}

func TestMigrationCreatesTables(t *testing.T) {
	url, err := config.BuildTestDatabaseURL()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	pool, err := Connect(ctx, url)
	if err != nil {
		t.Skipf("database not available: %v", err)
	}
	defer pool.Close()

	if err := EnsureSchema(ctx, pool, true); err != nil {
		t.Fatal(err)
	}

	var logsExists, auditExists bool
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'application_logs'
		)`).Scan(&logsExists)
	if err != nil {
		t.Fatal(err)
	}
	err = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'audit_events'
		)`).Scan(&auditExists)
	if err != nil {
		t.Fatal(err)
	}
	if !logsExists || !auditExists {
		t.Fatalf("tables missing: logs=%v audit=%v", logsExists, auditExists)
	}
}
