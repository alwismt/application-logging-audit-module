package database

import (
	"context"
	"fmt"

	databasemigrations "application-logging-audit-module/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	tableApplicationLogs = "application_logs"
	tableAuditEvents     = "audit_events"
	tableAdminUsers      = "admin_users"
)

// TablesExist reports whether both core tables are present in the public schema.
func TablesExist(ctx context.Context, pool *pgxpool.Pool) (bool, error) {
	const q = `
		SELECT COUNT(*) = 2
		FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_name IN ($1, $2)`

	var ok bool
	if err := pool.QueryRow(ctx, q, tableApplicationLogs, tableAuditEvents).Scan(&ok); err != nil {
		return false, fmt.Errorf("check tables exist: %w", err)
	}
	return ok, nil
}

// RunMigrations applies the embedded V1_1 schema SQL.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, databasemigrations.MigrationV11); err != nil {
		return fmt.Errorf("run migration V1_1: %w", err)
	}
	return nil
}

// EnsureSchema pings the database, runs migrations only when core tables are missing,
// and verifies the schema is ready. When autoMigrate is false, tables must already exist.
func EnsureSchema(ctx context.Context, pool *pgxpool.Pool, autoMigrate bool) error {
	if err := Ping(ctx, pool); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	exists, err := TablesExist(ctx, pool)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	if !autoMigrate {
		return fmt.Errorf("required tables %q and %q are missing and DB_AUTO_MIGRATE is disabled",
			tableApplicationLogs, tableAuditEvents)
	}

	if err := RunMigrations(ctx, pool); err != nil {
		return err
	}

	exists, err = TablesExist(ctx, pool)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("migration completed but tables %q and %q are still missing",
			tableApplicationLogs, tableAuditEvents)
	}
	return nil
}

func AdminTableExists(ctx context.Context, pool *pgxpool.Pool) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)`
	var ok bool
	if err := pool.QueryRow(ctx, q, tableAdminUsers).Scan(&ok); err != nil {
		return false, fmt.Errorf("check admin_users table: %w", err)
	}
	return ok, nil
}

func RunAdminMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if _, err := pool.Exec(ctx, databasemigrations.MigrationV12); err != nil {
		return fmt.Errorf("run migration V1_2 admin_users: %w", err)
	}
	return nil
}

func EnsureAdminSchema(ctx context.Context, pool *pgxpool.Pool, autoMigrate bool) error {
	exists, err := AdminTableExists(ctx, pool)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if !autoMigrate {
		return fmt.Errorf("required table %q is missing and DB_AUTO_MIGRATE is disabled", tableAdminUsers)
	}
	if err := RunAdminMigrations(ctx, pool); err != nil {
		return err
	}
	exists, err = AdminTableExists(ctx, pool)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("migration completed but table %q is still missing", tableAdminUsers)
	}
	return nil
}
