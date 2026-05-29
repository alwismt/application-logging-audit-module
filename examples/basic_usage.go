// Basic usage example: wire the logging and audit component in another Go application.
//
// Run (SQLite is the default; no PostgreSQL required):
//
//	go run ./examples/basic_usage.go
//
// For PostgreSQL, set DB_DRIVER=postgres and DB_* in .env before running.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"application-logging-audit-module/internal/audit"
	"application-logging-audit-module/internal/config"
	"application-logging-audit-module/internal/database"
	"application-logging-audit-module/internal/logger"

	"github.com/google/uuid"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var logRepo logger.LogRepository
	var auditRepo audit.AuditRepository

	switch cfg.DBDriver {
	case config.DriverSQLite:
		db, err := database.ConnectSQLite(cfg.SQLitePath)
		if err != nil {
			log.Fatalf("database: %v", err)
		}
		defer db.Close()
		if err := database.EnsureSchemaSQLite(ctx, db, cfg.DBAutoMigrate); err != nil {
			log.Fatalf("schema: %v", err)
		}
		logRepo = logger.NewSQLiteRepository(db)
		auditRepo = audit.NewSQLiteRepository(db)

	case config.DriverPostgres:
		pool, err := database.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("database: %v", err)
		}
		defer pool.Close()
		if err := database.EnsureSchema(ctx, pool, cfg.DBAutoMigrate); err != nil {
			log.Fatalf("schema: %v", err)
		}
		logRepo = logger.NewPostgresRepository(pool)
		auditRepo = audit.NewPostgresRepository(pool)

	default:
		log.Fatalf("unsupported DB_DRIVER: %s", cfg.DBDriver)
	}

	loggerSvc := logger.NewService(logRepo, cfg.ServiceName, true, true)
	auditSvc := audit.NewService(auditRepo)

	if err := loggerSvc.Info(ctx, "User dashboard loaded", map[string]any{
		"component": "dashboard",
	}); err != nil {
		log.Printf("info log failed: %v", err)
	}

	if err := loggerSvc.Error(ctx, "Failed to connect to payment service", fmt.Errorf("connection timeout"), map[string]any{
		"service": "payment",
	}); err != nil {
		log.Printf("error log failed: %v", err)
	}

	userID := uuid.New()
	if err := auditSvc.Record(ctx, audit.AuditEvent{
		UserID:       &userID,
		Username:     "demo_user",
		Action:       "UPDATE_RECORD",
		ResourceType: "invoice",
		ResourceID:   "1001",
		Status:       "SUCCESS",
		NewValue:     map[string]any{"status": "paid"},
	}); err != nil {
		log.Printf("audit record failed: %v", err)
	}

	fmt.Fprintln(os.Stdout, "Example completed. Check application_logs and audit_events tables.")
}
