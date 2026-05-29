//go:build integration

package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"application-logging-audit-module/internal/app"
	"application-logging-audit-module/internal/config"
	"application-logging-audit-module/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupIntegrationApp(t *testing.T) (*app.App, *pgxpool.Pool) {
	t.Helper()
	url, err := config.BuildTestDatabaseURL()
	if err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		AppPort:               "8080",
		DBDriver:              config.DriverPostgres,
		DatabaseURL:           url,
		DBAutoMigrate:         true,
		ServiceName:           "integration-test",
		EnableConsoleLogging:  false,
		EnableDatabaseLogging: true,
		AdminUsername:         "admin",
		AdminPassword:         "12345678",
		AdminAPIKey:           "integration-test-key",
		JWTSecret:             "integration-jwt-secret",
		JWTExpiryHours:        24,
		CORSOrigins:           "http://localhost:5173",
	}
	application, err := app.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	database.TruncateTables(t, application.Pool())
	return application, application.Pool()
}

func integrationAdminToken(t *testing.T, router http.Handler) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"username": "admin",
		"password": "12345678",
	})
	req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login: %d %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	return resp.Token
}

func TestAdminAPI_UnauthorizedWithoutToken(t *testing.T) {
	application, pool := setupIntegrationApp(t)
	defer pool.Close()
	defer application.Pool().Close()

	req := httptest.NewRequest(http.MethodGet, "/admin/logs", nil)
	rec := httptest.NewRecorder()
	application.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAdminAPI_LogsAndAuditEvents(t *testing.T) {
	application, pool := setupIntegrationApp(t)
	defer pool.Close()
	defer application.Pool().Close()

	router := application.Router()
	token := integrationAdminToken(t, router)

	body, _ := json.Marshal(map[string]string{"message": "admin api test"})
	req := httptest.NewRequest(http.MethodPost, "/logs/log-info", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("demo log: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/admin/logs?level=INFO&page=1&limit=20", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list logs: %d %s", rec.Code, rec.Body.String())
	}

	body, _ = json.Marshal(map[string]string{"username": "bob", "user_id": ""})
	req = httptest.NewRequest(http.MethodPost, "/logs/audit-login", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("demo audit: %d %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/admin/audit-events?action=LOGIN&status=SUCCESS&page=1&limit=20", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list audit: %d %s", rec.Code, rec.Body.String())
	}
}
