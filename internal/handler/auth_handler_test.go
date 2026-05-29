package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"application-logging-audit-module/internal/adminauth"

	"github.com/google/uuid"
)

type mockAdminRepo struct {
	user *adminauth.AdminUser
}

func (m *mockAdminRepo) FindByUsername(ctx context.Context, username string) (*adminauth.AdminUser, error) {
	if m.user != nil && m.user.Username == username {
		return m.user, nil
	}
	return nil, nil
}

func (m *mockAdminRepo) Create(ctx context.Context, user *adminauth.AdminUser) error {
	m.user = user
	return nil
}

func TestAuthHandler_Login(t *testing.T) {
	hash, err := adminauth.HashPassword("12345678")
	if err != nil {
		t.Fatal(err)
	}
	repo := &mockAdminRepo{
		user: &adminauth.AdminUser{
			ID:           uuid.New(),
			Username:     "admin",
			PasswordHash: hash,
			CreatedAt:    time.Now(),
		},
	}
	tokens := adminauth.NewTokenService("secret", 24)
	h := NewAuthHandler(repo, tokens)

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "12345678"})
	req := httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	h.Login(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login: %d %s", rec.Code, rec.Body.String())
	}

	var resp loginResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Token == "" {
		t.Fatal("expected token")
	}

	body, _ = json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
	req = httptest.NewRequest(http.MethodPost, "/admin/login", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	h.Login(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad password: %d", rec.Code)
	}
}
