package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alwismt/application-logging-audit-module/internal/audit"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type mockAuditor struct {
	events []audit.AuditEvent
}

func (m *mockAuditor) Record(ctx context.Context, event audit.AuditEvent) error {
	m.events = append(m.events, event)
	return nil
}
func (m *mockAuditor) Find(ctx context.Context, filter audit.AuditFilter) ([]audit.AuditEvent, error) {
	return m.events, nil
}

type mockAuditRepoHandler struct {
	events []audit.AuditEvent
}

func (m *mockAuditRepoHandler) Insert(ctx context.Context, event audit.AuditEvent) error {
	m.events = append(m.events, event)
	return nil
}
func (m *mockAuditRepoHandler) Find(ctx context.Context, filter audit.AuditFilter) ([]audit.AuditEvent, error) {
	return m.events, nil
}
func (m *mockAuditRepoHandler) FindByID(ctx context.Context, id uuid.UUID) (*audit.AuditEvent, error) {
	for _, e := range m.events {
		if e.ID == id {
			return &e, nil
		}
	}
	return nil, nil
}

func TestAuditHandler_DemoAuditLogin(t *testing.T) {
	auditor := &mockAuditor{}
	repo := &mockAuditRepoHandler{}
	h := NewAuditHandler(auditor, repo)

	body, _ := json.Marshal(map[string]string{"username": "alice"})
	req := httptest.NewRequest(http.MethodPost, "/logs/audit-login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.DemoAuditLogin(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAuditHandler_GetAuditEvent(t *testing.T) {
	id := uuid.New()
	repo := &mockAuditRepoHandler{events: []audit.AuditEvent{{ID: id, Action: "LOGIN", Status: "SUCCESS"}}}
	h := NewAuditHandler(&mockAuditor{}, repo)

	r := chi.NewRouter()
	r.Get("/admin/audit-events/{id}", h.GetAuditEvent)
	req := httptest.NewRequest(http.MethodGet, "/admin/audit-events/"+id.String(), nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
