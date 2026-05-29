package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"application-logging-audit-module/internal/adminauth"
	"application-logging-audit-module/internal/audit"
	"application-logging-audit-module/internal/config"
	"application-logging-audit-module/internal/database"
	"application-logging-audit-module/internal/logger"
	"application-logging-audit-module/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	cfg        *config.Config
	pool       *pgxpool.Pool
	sqliteDB   *sql.DB
	pingDB     func(context.Context) error
	loggerSvc  *logger.LoggerService
	auditSvc   *audit.AuditService
	logRepo    logger.LogRepository
	auditRepo  audit.AuditRepository
	adminRepo  adminauth.Repository
	tokens     *adminauth.TokenService
	adminAuth  *middleware.AdminAuth
	cors       *middleware.CORS
	httpLogger *middleware.HTTPLogger
	server     *http.Server
}

func New(cfg *config.Config) (*App, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var (
		logRepo   logger.LogRepository
		auditRepo audit.AuditRepository
		adminRepo adminauth.Repository
		pingDB    func(context.Context) error
	)

	app := &App{cfg: cfg}

	switch cfg.DBDriver {
	case config.DriverSQLite:
		db, err := database.ConnectSQLite(cfg.SQLitePath)
		if err != nil {
			return nil, fmt.Errorf("database connection: %w", err)
		}
		if err := database.EnsureSchemaSQLite(ctx, db, cfg.DBAutoMigrate); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("database schema: %w", err)
		}
		if err := database.EnsureAdminSchemaSQLite(ctx, db, cfg.DBAutoMigrate); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("admin schema: %w", err)
		}
		app.sqliteDB = db
		pingDB = func(c context.Context) error { return database.PingSQLite(c, db) }
		logRepo = logger.NewSQLiteRepository(db)
		auditRepo = audit.NewSQLiteRepository(db)
		adminRepo = adminauth.NewSQLiteRepository(db)

	case config.DriverPostgres:
		pool, err := database.Connect(ctx, cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("database connection: %w", err)
		}
		if err := database.EnsureSchema(ctx, pool, cfg.DBAutoMigrate); err != nil {
			pool.Close()
			return nil, fmt.Errorf("database schema: %w", err)
		}
		if err := database.EnsureAdminSchema(ctx, pool, cfg.DBAutoMigrate); err != nil {
			pool.Close()
			return nil, fmt.Errorf("admin schema: %w", err)
		}
		app.pool = pool
		pingDB = func(c context.Context) error { return database.Ping(c, pool) }
		logRepo = logger.NewPostgresRepository(pool)
		auditRepo = audit.NewPostgresRepository(pool)
		adminRepo = adminauth.NewPostgresRepository(pool)

	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER %q", cfg.DBDriver)
	}

	if err := adminauth.SeedDefaultAdmin(ctx, adminRepo, cfg); err != nil {
		app.closeDB()
		return nil, fmt.Errorf("seed admin user: %w", err)
	}

	app.pingDB = pingDB
	app.logRepo = logRepo
	app.auditRepo = auditRepo
	app.adminRepo = adminRepo
	app.tokens = adminauth.NewTokenService(cfg.JWTSecret, cfg.JWTExpiryHours)
	app.adminAuth = middleware.NewAdminAuth(app.tokens, cfg.AdminAPIKey)
	app.cors = middleware.NewCORS(cfg.CORSOrigins)

	loggerSvc := logger.NewService(
		logRepo,
		cfg.ServiceName,
		cfg.EnableConsoleLogging,
		cfg.EnableDatabaseLogging,
	)
	auditSvc := audit.NewService(auditRepo)
	httpLogger := middleware.NewHTTPLogger(loggerSvc)

	app.loggerSvc = loggerSvc
	app.auditSvc = auditSvc
	app.httpLogger = httpLogger

	handler := app.routes()
	app.server = &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	return app, nil
}

func (a *App) Run() error {
	errCh := make(chan error, 1)
	go func() {
		fmt.Printf("Server listening on %s\n", a.server.Addr)
		if a.cfg.EnableSwaggerUI {
			fmt.Printf("Swagger UI: http://localhost:%s/swagger/\n", a.cfg.AppPort)
		}
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-quit:
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.server.Shutdown(ctx); err != nil {
		return err
	}
	a.closeDB()
	return nil
}

func (a *App) closeDB() {
	if a.pool != nil {
		a.pool.Close()
	}
	if a.sqliteDB != nil {
		_ = a.sqliteDB.Close()
	}
}

func (a *App) Router() http.Handler {
	return a.routes()
}

func (a *App) Pool() *pgxpool.Pool {
	return a.pool
}

func (a *App) LoggerService() *logger.LoggerService {
	return a.loggerSvc
}

func (a *App) AuditService() *audit.AuditService {
	return a.auditSvc
}
