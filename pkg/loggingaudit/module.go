// Package loggingaudit is the public entry point for embedding the logging and audit
// component in another Go application via go get.
package loggingaudit

import (
	"context"
	"net/http"

	"github.com/alwismt/application-logging-audit-module/internal/app"
	"github.com/alwismt/application-logging-audit-module/internal/audit"
	"github.com/alwismt/application-logging-audit-module/internal/config"
)

// Config is the application configuration (env-backed defaults via LoadConfig).
type Config = config.Config

const (
	DriverSQLite   = config.DriverSQLite
	DriverPostgres = config.DriverPostgres
)

// AuditEvent is an audit trail record.
type AuditEvent = audit.AuditEvent

// Logger writes application logs.
type Logger interface {
	Info(ctx context.Context, message string, metadata map[string]any) error
	Warn(ctx context.Context, message string, metadata map[string]any) error
	Error(ctx context.Context, message string, err error, metadata map[string]any) error
	Debug(ctx context.Context, message string, metadata map[string]any) error
}

// Auditor records audit trail events.
type Auditor interface {
	Record(ctx context.Context, event AuditEvent) error
}

// Module wires database, logger, audit, and HTTP routes (/health, /logs, /admin).
type Module struct {
	application *app.App
}

// LoadConfig reads configuration from environment variables (and optional .env file).
func LoadConfig() (*Config, error) {
	return config.Load()
}

// New initializes the module from configuration.
func New(cfg *Config) (*Module, error) {
	application, err := app.New(cfg)
	if err != nil {
		return nil, err
	}
	return &Module{application: application}, nil
}

// NewFromEnv loads configuration from the environment and initializes the module.
func NewFromEnv() (*Module, error) {
	cfg, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	return New(cfg)
}

// Handler returns the HTTP handler with /health, /logs, and /admin routes.
// Mount on your router, e.g. r.Mount("/", mod.Handler()).
func (m *Module) Handler() http.Handler {
	return m.application.Router()
}

// Run starts the standalone HTTP server (same as cmd/server).
func (m *Module) Run() error {
	return m.application.Run()
}

// Logger returns the application logging service.
func (m *Module) Logger() Logger {
	return m.application.LoggerService()
}

// Auditor returns the audit trail service.
func (m *Module) Auditor() Auditor {
	return m.application.AuditService()
}
