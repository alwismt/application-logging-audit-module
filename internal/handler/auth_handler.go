package handler

import (
	"encoding/json"
	"net/http"

	"application-logging-audit-module/internal/adminauth"
	"application-logging-audit-module/internal/common"
)

type AuthHandler struct {
	adminRepo adminauth.Repository
	tokens    *adminauth.TokenService
}

func NewAuthHandler(repo adminauth.Repository, tokens *adminauth.TokenService) *AuthHandler {
	return &AuthHandler{adminRepo: repo, tokens: tokens}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Username == "" || req.Password == "" {
		common.WriteError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	user, err := h.adminRepo.FindByUsername(r.Context(), req.Username)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil || !adminauth.CheckPassword(user.PasswordHash, req.Password) {
		common.WriteError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := h.tokens.IssueToken(user.Username)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	common.WriteJSON(w, http.StatusOK, loginResponse{Token: token})
}
