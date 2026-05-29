package exporter

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"application-logging-audit-module/internal/audit"
	"application-logging-audit-module/internal/logger"
)

func ExportLogsCSV(entries []logger.LogEntry) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "level", "message", "source", "request_id", "user_id", "error_code", "created_at", "metadata"})
	for _, e := range entries {
		userID := ""
		if e.UserID != nil {
			userID = e.UserID.String()
		}
		meta, _ := json.Marshal(e.Metadata)
		_ = w.Write([]string{
			e.ID.String(), e.Level, e.Message, e.Source, e.RequestID,
			userID, e.ErrorCode, e.CreatedAt.Format(time.RFC3339), string(meta),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ExportAuditCSV(events []audit.AuditEvent) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "user_id", "username", "action", "resource_type", "resource_id", "status", "ip_address", "request_id", "created_at", "metadata"})
	for _, e := range events {
		userID := ""
		if e.UserID != nil {
			userID = e.UserID.String()
		}
		meta, _ := json.Marshal(e.Metadata)
		_ = w.Write([]string{
			e.ID.String(), userID, e.Username, e.Action, e.ResourceType,
			e.ResourceID, e.Status, e.IPAddress, e.RequestID,
			e.CreatedAt.Format(time.RFC3339), string(meta),
		})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func FormatExportFilename(prefix string, format string) string {
	return fmt.Sprintf("%s_%s.%s", prefix, strconv.FormatInt(time.Now().Unix(), 10), format)
}
