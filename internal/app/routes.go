package app

import (
	"net/http"

	"application-logging-audit-module/internal/handler"
	"application-logging-audit-module/internal/middleware"
	"application-logging-audit-module/internal/swagger"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

func (a *App) routes() http.Handler {
	r := chi.NewRouter()
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(a.cors.Middleware)
	r.Use(a.httpLogger.Middleware)

	health := handler.NewHealthHandler(a.pingDB)
	logH := handler.NewLogHandler(a.loggerSvc, a.logRepo)
	auditH := handler.NewAuditHandler(a.auditSvc, a.auditRepo)
	authH := handler.NewAuthHandler(a.adminRepo, a.tokens)

	r.Get("/health", health.ServeHTTP)

	if a.cfg.EnableSwaggerUI {
		swagger.Mount(r)
	}

	r.Route("/logs", func(r chi.Router) {
		r.Post("/log-info", logH.DemoLogInfo)
		r.Post("/log-error", logH.DemoLogError)
		r.Post("/audit-login", auditH.DemoAuditLogin)
		r.Post("/audit-update", auditH.DemoAuditUpdate)
	})

	r.Route("/admin", func(r chi.Router) {
		r.Post("/login", authH.Login)
		r.Group(func(r chi.Router) {
			r.Use(a.adminAuth.Middleware)
			r.Get("/logs", logH.ListLogs)
			r.Get("/logs/export", logH.ExportLogs)
			r.Get("/logs/{id}", logH.GetLog)
			r.Get("/audit-events", auditH.ListAuditEvents)
			r.Get("/audit-events/export", auditH.ExportAuditEvents)
			r.Get("/audit-events/{id}", auditH.GetAuditEvent)
		})
	})

	return r
}

// Expose middleware for tests
func (a *App) HTTPLogger() *middleware.HTTPLogger {
	return a.httpLogger
}
