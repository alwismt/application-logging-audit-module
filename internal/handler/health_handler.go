package handler

import (
	"context"
	"net/http"

	"github.com/alwismt/application-logging-audit-module/internal/common"
)

type HealthHandler struct {
	ping func(context.Context) error
}

func NewHealthHandler(ping func(context.Context) error) *HealthHandler {
	return &HealthHandler{ping: ping}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dbStatus := "up"
	if err := h.ping(r.Context()); err != nil {
		dbStatus = "down"
	}
	status := "ok"
	code := http.StatusOK
	if dbStatus == "down" {
		status = "degraded"
	}
	common.WriteJSON(w, code, map[string]string{
		"status":   status,
		"database": dbStatus,
	})
}

func (h *HealthHandler) Check(ctx context.Context) error {
	return h.ping(ctx)
}
