package logger

import (
	"time"

	"application-logging-audit-module/internal/common"

	"github.com/google/uuid"
)

var ValidLevels = []string{"INFO", "WARNING", "ERROR", "DEBUG"}

type LogEntry struct {
	ID         uuid.UUID      `json:"id"`
	Level      string         `json:"level"`
	Message    string         `json:"message"`
	Source     string         `json:"source,omitempty"`
	RequestID  string         `json:"request_id,omitempty"`
	UserID     *uuid.UUID     `json:"user_id,omitempty"`
	ErrorCode  string         `json:"error_code,omitempty"`
	StackTrace string         `json:"stack_trace,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

type LogFilter struct {
	Level     string
	RequestID string
	UserID    *uuid.UUID
	Source    string
	From      *time.Time
	To        *time.Time
	Pagination common.Pagination
}
