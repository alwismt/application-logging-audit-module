package exporter

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/alwismt/application-logging-audit-module/internal/logger"

	"github.com/google/uuid"
)

func TestExportLogsJSON(t *testing.T) {
	entries := []logger.LogEntry{{
		ID:        uuid.New(),
		Level:     "INFO",
		Message:   "test",
		CreatedAt: time.Now().UTC(),
	}}
	data, err := ExportLogsJSON(entries)
	if err != nil {
		t.Fatal(err)
	}
	var decoded []logger.LogEntry
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded) != 1 || decoded[0].Message != "test" {
		t.Fatalf("unexpected decoded data: %+v", decoded)
	}
}
