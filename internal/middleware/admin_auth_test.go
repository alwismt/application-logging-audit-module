package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alwismt/application-logging-audit-module/internal/adminauth"
)

func TestAdminAuth_JWTAndAPIKey(t *testing.T) {
	tokens := adminauth.NewTokenService("test-secret", 24)
	auth := NewAdminAuth(tokens, "test-api-key")

	token, err := tokens.IssueToken("admin")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		headers    map[string]string
		wantStatus int
	}{
		{
			name:       "valid bearer",
			headers:    map[string]string{"Authorization": "Bearer " + token},
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid api key header",
			headers:    map[string]string{"X-API-Key": "test-api-key"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid api key authorization",
			headers:    map[string]string{"Authorization": "ApiKey test-api-key"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing auth",
			headers:    nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid bearer",
			headers:    map[string]string{"Authorization": "Bearer invalid"},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodGet, "/admin/logs", nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			auth.Middleware(next).ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status: got %d want %d body=%s", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if tc.wantStatus == http.StatusOK && !called {
				t.Fatal("expected handler to be called")
			}
		})
	}
}
