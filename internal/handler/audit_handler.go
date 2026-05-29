package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/alwismt/application-logging-audit-module/internal/audit"
	"github.com/alwismt/application-logging-audit-module/internal/common"
	"github.com/alwismt/application-logging-audit-module/internal/exporter"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AuditHandler struct {
	auditor   audit.Auditor
	auditRepo audit.AuditRepository
}

func NewAuditHandler(svc audit.Auditor, repo audit.AuditRepository) *AuditHandler {
	return &AuditHandler{auditor: svc, auditRepo: repo}
}

type demoAuditLoginRequest struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IP       string `json:"ip_address"`
}

type demoAuditUpdateRequest struct {
	UserID       string         `json:"user_id"`
	Username     string         `json:"username"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	OldValue     map[string]any `json:"old_value"`
	NewValue     map[string]any `json:"new_value"`
}

func (h *AuditHandler) DemoAuditLogin(w http.ResponseWriter, r *http.Request) {
	var req demoAuditLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	event := audit.AuditEvent{
		Username:  req.Username,
		Action:    "LOGIN",
		Status:    "SUCCESS",
		IPAddress: req.IP,
		UserAgent: r.UserAgent(),
		RequestID: r.Header.Get("X-Request-ID"),
	}
	if req.UserID != "" {
		if id, err := uuid.Parse(req.UserID); err == nil {
			event.UserID = &id
		}
	}
	if event.Username == "" {
		event.Username = "demo_user"
	}
	if err := h.auditor.Record(r.Context(), event); err != nil {
		common.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.WriteJSON(w, http.StatusCreated, map[string]string{"status": "recorded", "id": event.ID.String()})
}

func (h *AuditHandler) DemoAuditUpdate(w http.ResponseWriter, r *http.Request) {
	var req demoAuditUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	event := audit.AuditEvent{
		Username:     req.Username,
		Action:       "UPDATE_RECORD",
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		OldValue:     req.OldValue,
		NewValue:     req.NewValue,
		Status:       "SUCCESS",
		IPAddress:    r.RemoteAddr,
		UserAgent:    r.UserAgent(),
		RequestID:    r.Header.Get("X-Request-ID"),
	}
	if req.UserID != "" {
		if id, err := uuid.Parse(req.UserID); err == nil {
			event.UserID = &id
		}
	}
	if event.ResourceType == "" {
		event.ResourceType = "invoice"
	}
	if event.ResourceID == "" {
		event.ResourceID = "1001"
	}
	if err := h.auditor.Record(r.Context(), event); err != nil {
		common.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.WriteJSON(w, http.StatusCreated, map[string]string{"status": "recorded", "id": event.ID.String()})
}

func (h *AuditHandler) ListAuditEvents(w http.ResponseWriter, r *http.Request) {
	filter := parseAuditFilter(r)
	events, err := h.auditRepo.Find(r.Context(), filter)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{
		"data":  events,
		"page":  filter.Pagination.Page,
		"limit": filter.Pagination.Limit,
	})
}

func (h *AuditHandler) GetAuditEvent(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	event, err := h.auditRepo.FindByID(r.Context(), id)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if event == nil {
		common.WriteError(w, http.StatusNotFound, "audit event not found")
		return
	}
	common.WriteJSON(w, http.StatusOK, event)
}

func (h *AuditHandler) ExportAuditEvents(w http.ResponseWriter, r *http.Request) {
	filter := parseAuditFilter(r)
	events, err := h.auditRepo.Find(r.Context(), filter)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}
	switch format {
	case "csv":
		data, err := exporter.ExportAuditCSV(events)
		if err != nil {
			common.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=audit_events.csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	case "json":
		data, err := exporter.ExportAuditJSON(events)
		if err != nil {
			common.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=audit_events.json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	default:
		common.WriteError(w, http.StatusBadRequest, "format must be json or csv")
	}
}

func parseAuditFilter(r *http.Request) audit.AuditFilter {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	filter := audit.AuditFilter{
		Username:     q.Get("username"),
		Action:       q.Get("action"),
		ResourceType: q.Get("resource_type"),
		Status:       q.Get("status"),
		RequestID:    q.Get("request_id"),
		Pagination:   common.NormalizePagination(page, limit),
	}
	if uid := q.Get("user_id"); uid != "" {
		if id, err := uuid.Parse(uid); err == nil {
			filter.UserID = &id
		}
	}
	if from, err := common.ParseQueryTime(q.Get("from"), false); err == nil {
		filter.From = from
	}
	if to, err := common.ParseQueryTime(q.Get("to"), true); err == nil {
		filter.To = to
	}
	return filter
}
