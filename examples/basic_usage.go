// Basic usage example: embed logger and audit via the public pkg/loggingaudit API.
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

	"github.com/alwismt/application-logging-audit-module/pkg/loggingaudit"
	"github.com/google/uuid"
)

func main() {
	mod, err := loggingaudit.NewFromEnv()
	if err != nil {
		log.Fatalf("init module: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := mod.Logger().Info(ctx, "User dashboard loaded", map[string]any{
		"component": "dashboard",
	}); err != nil {
		log.Printf("info log failed: %v", err)
	}

	if err := mod.Logger().Error(ctx, "Failed to connect to payment service", fmt.Errorf("connection timeout"), map[string]any{
		"service": "payment",
	}); err != nil {
		log.Printf("error log failed: %v", err)
	}

	userID := uuid.New()
	if err := mod.Auditor().Record(ctx, loggingaudit.AuditEvent{
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
