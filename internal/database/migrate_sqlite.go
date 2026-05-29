package database

import (
	"context"
	"database/sql"
	"fmt"

	databasemigrations "github.com/alwismt/application-logging-audit-module/database"
)

func TablesExistSQLite(ctx context.Context, db *sql.DB) (bool, error) {
	const q = `
		SELECT COUNT(*) = 2
		FROM sqlite_master
		WHERE type = 'table'
		  AND name IN (?, ?)`

	var ok bool
	if err := db.QueryRowContext(ctx, q, tableApplicationLogs, tableAuditEvents).Scan(&ok); err != nil {
		return false, fmt.Errorf("check sqlite tables exist: %w", err)
	}
	return ok, nil
}

func RunMigrationsSQLite(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, databasemigrations.MigrationV11SQLite); err != nil {
		return fmt.Errorf("run sqlite migration V1_1: %w", err)
	}
	return nil
}

func EnsureSchemaSQLite(ctx context.Context, db *sql.DB, autoMigrate bool) error {
	if err := PingSQLite(ctx, db); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	exists, err := TablesExistSQLite(ctx, db)
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

	if err := RunMigrationsSQLite(ctx, db); err != nil {
		return err
	}

	exists, err = TablesExistSQLite(ctx, db)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("migration completed but tables %q and %q are still missing",
			tableApplicationLogs, tableAuditEvents)
	}
	return nil
}

func AdminTableExistsSQLite(ctx context.Context, db *sql.DB) (bool, error) {
	const q = `
		SELECT COUNT(*) > 0
		FROM sqlite_master
		WHERE type = 'table' AND name = ?`
	var ok bool
	if err := db.QueryRowContext(ctx, q, tableAdminUsers).Scan(&ok); err != nil {
		return false, fmt.Errorf("check admin_users table: %w", err)
	}
	return ok, nil
}

func RunAdminMigrationsSQLite(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, databasemigrations.MigrationV12SQLite); err != nil {
		return fmt.Errorf("run sqlite migration V1_2 admin_users: %w", err)
	}
	return nil
}

func EnsureAdminSchemaSQLite(ctx context.Context, db *sql.DB, autoMigrate bool) error {
	exists, err := AdminTableExistsSQLite(ctx, db)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if !autoMigrate {
		return fmt.Errorf("required table %q is missing and DB_AUTO_MIGRATE is disabled", tableAdminUsers)
	}
	if err := RunAdminMigrationsSQLite(ctx, db); err != nil {
		return err
	}
	exists, err = AdminTableExistsSQLite(ctx, db)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("migration completed but table %q is still missing", tableAdminUsers)
	}
	return nil
}
