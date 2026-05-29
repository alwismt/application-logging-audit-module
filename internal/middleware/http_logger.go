package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alwismt/application-logging-audit-module/internal/logger"

	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

type HTTPLogger struct {
	logger *logger.LoggerService
}

func NewHTTPLogger(svc *logger.LoggerService) *HTTPLogger {
	return &HTTPLogger{logger: svc}
}

func (h *HTTPLogger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		w.Header().Set(RequestIDHeader, requestID)

		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)

		latency := time.Since(start)
		ip := clientIP(r)
		ctx := logger.WithRequestID(r.Context(), requestID)

		meta := map[string]any{
			"method":      r.Method,
			"path":        r.URL.Path,
			"status_code": rw.Status(),
			"latency_ms":  latency.Milliseconds(),
			"ip_address":  ip,
			"user_agent":  r.UserAgent(),
			"request_id":  requestID,
		}
		msg := fmt.Sprintf("%s %s %d %dms", r.Method, r.URL.Path, rw.Status(), latency.Milliseconds())
		_ = h.logger.LogHTTP(ctx, msg, meta)
	})
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}
