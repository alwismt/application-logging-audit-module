//go:build integration

package database

import (
	"context"
	"testing"
)

func TestEnsureSchema_CreatesTablesOnce(t *testing.T) {
	pool := SetupTestPool(t)
	defer pool.Close()

	ctx := context.Background()

	exists, err := TablesExist(ctx, pool)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("expected tables after SetupTestPool")
	}

	if err := EnsureSchema(ctx, pool, true); err != nil {
		t.Fatalf("second EnsureSchema should be no-op: %v", err)
	}
}
