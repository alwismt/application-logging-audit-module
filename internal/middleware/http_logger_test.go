package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"application-logging-audit-module/internal/logger"

	"github.com/google/uuid"
)

type capturingRepo struct {
	last logger.LogEntry
}

func (c *capturingRepo) Insert(ctx context.Context, entry logger.LogEntry) error {
	c.last = entry
	return nil
}

func (c *capturingRepo) Find(ctx context.Context, filter logger.LogFilter) ([]logger.LogEntry, error) {
	return nil, nil
}

func (c *capturingRepo) FindByID(ctx context.Context, id uuid.UUID) (*logger.LogEntry, error) {
	return nil, nil
}

func TestHTTPLogger_Middleware(t *testing.T) {
	repo := &capturingRepo{}
	svc := logger.NewService(repo, "test", false, true)
	mw := NewHTTPLogger(svc)

	handler := mw.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	if rec.Header().Get(RequestIDHeader) == "" {
		t.Fatal("expected request id header")
	}
	if repo.last.Metadata["method"] != http.MethodGet {
		t.Fatalf("expected GET method in metadata, got %v", repo.last.Metadata["method"])
	}
	if repo.last.Metadata["path"] != "/api/test" {
		t.Fatalf("unexpected path: %v", repo.last.Metadata["path"])
	}
	if repo.last.Metadata["status_code"] != http.StatusCreated {
		t.Fatalf("unexpected status: %v", repo.last.Metadata["status_code"])
	}
	if repo.last.Metadata["user_agent"] != "test-agent" {
		t.Fatalf("unexpected user agent: %v", repo.last.Metadata["user_agent"])
	}
	if repo.last.Metadata["ip_address"] != "203.0.113.1" {
		t.Fatalf("unexpected ip: %v", repo.last.Metadata["ip_address"])
	}
	if repo.last.RequestID == "" && repo.last.Metadata["request_id"] == "" {
		t.Fatal("expected request id captured")
	}
}
