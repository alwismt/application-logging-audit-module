package logger

import (
	"context"

	"github.com/google/uuid"
)

type Logger interface {
	Info(ctx context.Context, message string, metadata map[string]any) error
	Warn(ctx context.Context, message string, metadata map[string]any) error
	Error(ctx context.Context, message string, err error, metadata map[string]any) error
	Debug(ctx context.Context, message string, metadata map[string]any) error
}

type LogRepository interface {
	Insert(ctx context.Context, entry LogEntry) error
	Find(ctx context.Context, filter LogFilter) ([]LogEntry, error)
	FindByID(ctx context.Context, id uuid.UUID) (*LogEntry, error)
}
