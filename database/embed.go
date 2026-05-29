package databasemigrations

import _ "embed"

// MigrationV11 is the initial schema for application_logs and audit_events.
// Source file: database/migrations/V1_1__create_logging_audit_tables.sql
//
//go:embed migrations/V1_1__create_logging_audit_tables.sql
var MigrationV11 string

// MigrationV11SQLite is the initial schema for SQLite.
// Source file: database/migrations/V1_1_sqlite__create_logging_audit_tables.sql
//
//go:embed migrations/V1_1_sqlite__create_logging_audit_tables.sql
var MigrationV11SQLite string

// MigrationV12 adds admin_users for admin authentication.
//
//go:embed migrations/V1_2__create_admin_users.sql
var MigrationV12 string

// MigrationV12SQLite adds admin_users for SQLite.
//
//go:embed migrations/V1_2_sqlite__create_admin_users.sql
var MigrationV12SQLite string
