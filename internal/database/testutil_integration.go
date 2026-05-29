//go:build integration

package database

import (
	"context"
	"testing"
	"time"

	"github.com/alwismt/application-logging-audit-module/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url, err := config.BuildTestDatabaseURL()
	if err != nil {
		t.Fatalf("build test database url: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := Connect(ctx, url)
	if err != nil {
		t.Skipf("skipping integration test: database not available: %v", err)
	}
	if err := EnsureSchema(ctx, pool, true); err != nil {
		pool.Close()
		t.Fatalf("ensure schema: %v", err)
	}
	return pool
}

func TruncateTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	_, _ = pool.Exec(ctx, "TRUNCATE application_logs, audit_events")
}
