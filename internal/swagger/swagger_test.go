package swagger_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alwismt/application-logging-audit-module/internal/app"
	"github.com/alwismt/application-logging-audit-module/internal/config"
	"github.com/alwismt/application-logging-audit-module/internal/swagger"

	"github.com/go-chi/chi/v5"
)

func TestSpec_EmbeddedOpenAPI(t *testing.T) {
	body := swagger.Spec()
	if !strings.Contains(string(body), "openapi: 3.0.3") {
		t.Fatal("embedded spec missing openapi version")
	}
}

func TestMount_ServesUIAndSpec(t *testing.T) {
	r := chi.NewRouter()
	swagger.Mount(r)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("swagger UI: got %d", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("content-type: %s", rec.Header().Get("Content-Type"))
	}
	if !strings.Contains(rec.Body.String(), "swagger-ui") {
		t.Fatal("expected swagger-ui in HTML")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("openapi.yaml: got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "openapi: 3.0.3") {
		t.Fatal("expected openapi version in spec response")
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/swagger", nil)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		t.Fatalf("redirect: got %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/swagger/" {
		t.Fatalf("location: %s", rec.Header().Get("Location"))
	}
}

func TestApp_SwaggerDisabledByDefault(t *testing.T) {
	cfg := &config.Config{
		DBDriver:              config.DriverSQLite,
		SQLitePath:            ":memory:",
		DBAutoMigrate:         true,
		ServiceName:           "swagger-test",
		EnableConsoleLogging:  false,
		EnableDatabaseLogging: true,
		EnableSwaggerUI:       false,
		AdminUsername:         "admin",
		AdminPassword:         "12345678",
		JWTSecret:             "test-secret",
	}
	application, err := app.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger/", nil)
	application.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when disabled, got %d", rec.Code)
	}
}

func TestApp_SwaggerEnabled(t *testing.T) {
	cfg := &config.Config{
		DBDriver:              config.DriverSQLite,
		SQLitePath:            ":memory:",
		DBAutoMigrate:         true,
		ServiceName:           "swagger-test",
		EnableConsoleLogging:  false,
		EnableDatabaseLogging: true,
		EnableSwaggerUI:       true,
		AdminUsername:         "admin",
		AdminPassword:         "12345678",
		JWTSecret:             "test-secret",
	}
	application, err := app.New(cfg)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/swagger/openapi.yaml", nil)
	application.Router().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "openapi: 3.0.3") {
		t.Fatal("expected openapi spec body")
	}
}
