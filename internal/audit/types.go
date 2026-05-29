package audit

import (
	"time"

	"application-logging-audit-module/internal/common"

	"github.com/google/uuid"
)

var ValidActions = []string{
	"LOGIN", "LOGOUT", "CREATE_RECORD", "UPDATE_RECORD", "DELETE_RECORD",
	"DOWNLOAD_FILE", "FAILED_LOGIN", "PERMISSION_DENIED",
}

var ValidStatuses = []string{"SUCCESS", "FAILURE"}

type AuditEvent struct {
	ID           uuid.UUID      `json:"id"`
	UserID       *uuid.UUID     `json:"user_id,omitempty"`
	Username     string         `json:"username,omitempty"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type,omitempty"`
	ResourceID   string         `json:"resource_id,omitempty"`
	OldValue     map[string]any `json:"old_value,omitempty"`
	NewValue     map[string]any `json:"new_value,omitempty"`
	IPAddress    string         `json:"ip_address,omitempty"`
	UserAgent    string         `json:"user_agent,omitempty"`
	RequestID    string         `json:"request_id,omitempty"`
	Status       string         `json:"status"`
	Metadata     map[string]any `json:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

type AuditFilter struct {
	UserID       *uuid.UUID
	Username     string
	Action       string
	ResourceType string
	Status       string
	RequestID    string
	From         *time.Time
	To           *time.Time
	Pagination   common.Pagination
}
