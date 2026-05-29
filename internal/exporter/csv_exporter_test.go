package exporter

import (
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"github.com/alwismt/application-logging-audit-module/internal/logger"

	"github.com/google/uuid"
)

func TestExportLogsCSV(t *testing.T) {
	entries := []logger.LogEntry{{
		ID:        uuid.New(),
		Level:     "ERROR",
		Message:   "failed",
		CreatedAt: time.Now().UTC(),
	}}
	data, err := ExportLogsCSV(entries)
	if err != nil {
		t.Fatal(err)
	}
	reader := csv.NewReader(strings.NewReader(string(data)))
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected header + 1 row, got %d", len(rows))
	}
	if rows[0][0] != "id" {
		t.Fatalf("unexpected header: %v", rows[0])
	}
}
