package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"application-logging-audit-module/internal/common"
	"application-logging-audit-module/internal/exporter"
	"application-logging-audit-module/internal/logger"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type LogHandler struct {
	logger   logger.Logger
	logRepo  logger.LogRepository
}

func NewLogHandler(svc logger.Logger, repo logger.LogRepository) *LogHandler {
	return &LogHandler{logger: svc, logRepo: repo}
}

type demoLogRequest struct {
	Message  string         `json:"message"`
	Metadata map[string]any `json:"metadata"`
}

func (h *LogHandler) DemoLogInfo(w http.ResponseWriter, r *http.Request) {
	var req demoLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Message == "" {
		req.Message = "Demo info log message"
	}
	if err := h.logger.Info(r.Context(), req.Message, req.Metadata); err != nil {
		common.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.WriteJSON(w, http.StatusCreated, map[string]string{"status": "logged"})
}

func (h *LogHandler) DemoLogError(w http.ResponseWriter, r *http.Request) {
	var req demoLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if req.Message == "" {
		req.Message = "Demo error log message"
	}
	err := &demoError{msg: "simulated error for demo"}
	if logErr := h.logger.Error(r.Context(), req.Message, err, req.Metadata); logErr != nil {
		common.WriteError(w, http.StatusInternalServerError, logErr.Error())
		return
	}
	common.WriteJSON(w, http.StatusCreated, map[string]string{"status": "logged"})
}

type demoError struct{ msg string }

func (e *demoError) Error() string { return e.msg }

func (h *LogHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	filter := parseLogFilter(r)
	entries, err := h.logRepo.Find(r.Context(), filter)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	common.WriteJSON(w, http.StatusOK, map[string]any{
		"data":  entries,
		"page":  filter.Pagination.Page,
		"limit": filter.Pagination.Limit,
	})
}

func (h *LogHandler) GetLog(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		common.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	entry, err := h.logRepo.FindByID(r.Context(), id)
	if err != nil {
		common.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if entry == nil {
		common.WriteError(w, http.StatusNotFound, "log not found")
		return
	}
	common.WriteJSON(w, http.StatusOK, entry)
}

func (h *LogHandler) ExportLogs(w http.ResponseWriter, r *http.Request) {
	filter := parseLogFilter(r)
	entries, err := h.logRepo.Find(r.Context(), filter)
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
		data, err := exporter.ExportLogsCSV(entries)
		if err != nil {
			common.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=logs.csv")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	case "json":
		data, err := exporter.ExportLogsJSON(entries)
		if err != nil {
			common.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=logs.json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	default:
		common.WriteError(w, http.StatusBadRequest, "format must be json or csv")
	}
}

func parseLogFilter(r *http.Request) logger.LogFilter {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	filter := logger.LogFilter{
		Level:     q.Get("level"),
		RequestID: q.Get("request_id"),
		Source:    q.Get("source"),
		Pagination: common.NormalizePagination(page, limit),
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
