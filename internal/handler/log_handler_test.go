package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"application-logging-audit-module/internal/logger"

	"github.com/google/uuid"
)

type mockLogger struct{}

func (m *mockLogger) Info(ctx context.Context, message string, metadata map[string]any) error {
	return nil
}
func (m *mockLogger) Warn(ctx context.Context, message string, metadata map[string]any) error {
	return nil
}
func (m *mockLogger) Error(ctx context.Context, message string, err error, metadata map[string]any) error {
	return nil
}
func (m *mockLogger) Debug(ctx context.Context, message string, metadata map[string]any) error {
	return nil
}

type mockLogRepoHandler struct {
	entries []logger.LogEntry
}

func (m *mockLogRepoHandler) Insert(ctx context.Context, entry logger.LogEntry) error {
	m.entries = append(m.entries, entry)
	return nil
}
func (m *mockLogRepoHandler) Find(ctx context.Context, filter logger.LogFilter) ([]logger.LogEntry, error) {
	return m.entries, nil
}
func (m *mockLogRepoHandler) FindByID(ctx context.Context, id uuid.UUID) (*logger.LogEntry, error) {
	for _, e := range m.entries {
		if e.ID == id {
			return &e, nil
		}
	}
	return nil, nil
}

func TestLogHandler_ListLogs(t *testing.T) {
	id := uuid.New()
	repo := &mockLogRepoHandler{entries: []logger.LogEntry{{ID: id, Level: "INFO", Message: "x"}}}
	h := NewLogHandler(&mockLogger{}, repo)

	req := httptest.NewRequest(http.MethodGet, "/admin/logs?level=INFO&page=1&limit=10", nil)
	rec := httptest.NewRecorder()
	h.ListLogs(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	data, ok := resp["data"].([]any)
	if !ok || len(data) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
