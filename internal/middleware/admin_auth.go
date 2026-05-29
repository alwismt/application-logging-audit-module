package middleware

import (
	"net/http"
	"strings"

	"application-logging-audit-module/internal/adminauth"
	"application-logging-audit-module/internal/common"
)

type AdminAuth struct {
	tokens *adminauth.TokenService
	apiKey string
}

func NewAdminAuth(tokens *adminauth.TokenService, apiKey string) *AdminAuth {
	return &AdminAuth{tokens: tokens, apiKey: apiKey}
}

func (a *AdminAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.authenticate(r) {
			next.ServeHTTP(w, r)
			return
		}
		common.WriteError(w, http.StatusUnauthorized, "unauthorized")
	})
}

func (a *AdminAuth) authenticate(r *http.Request) bool {
	if a.apiKey != "" {
		if key := r.Header.Get("X-API-Key"); key == a.apiKey {
			return true
		}
	}

	auth := r.Header.Get("Authorization")
	if auth == "" {
		return false
	}

	if strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		token := strings.TrimSpace(auth[7:])
		if token == "" {
			return false
		}
		_, err := a.tokens.ValidateToken(token)
		return err == nil
	}

	if a.apiKey != "" && strings.HasPrefix(strings.ToLower(auth), "apikey ") {
		key := strings.TrimSpace(auth[7:])
		return key == a.apiKey
	}

	return false
}
