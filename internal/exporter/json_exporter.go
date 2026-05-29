package exporter

import (
	"bytes"
	"encoding/json"

	"github.com/alwismt/application-logging-audit-module/internal/audit"
	"github.com/alwismt/application-logging-audit-module/internal/logger"
)

func ExportLogsJSON(entries []logger.LogEntry) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(entries); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func ExportAuditJSON(events []audit.AuditEvent) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(events); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
