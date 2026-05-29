package logger

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

type LoggerService struct {
	repo           LogRepository
	sanitizer      *SensitiveDataSanitizer
	source         string
	consoleEnabled bool
	dbEnabled      bool
}

func NewService(repo LogRepository, source string, consoleEnabled, dbEnabled bool) *LoggerService {
	return &LoggerService{
		repo:           repo,
		sanitizer:      NewSanitizer(),
		source:         source,
		consoleEnabled: consoleEnabled,
		dbEnabled:      dbEnabled,
	}
}

func (s *LoggerService) Info(ctx context.Context, message string, metadata map[string]any) error {
	return s.log(ctx, "INFO", message, "", metadata, nil)
}

func (s *LoggerService) Warn(ctx context.Context, message string, metadata map[string]any) error {
	return s.log(ctx, "WARNING", message, "", metadata, nil)
}

func (s *LoggerService) Debug(ctx context.Context, message string, metadata map[string]any) error {
	return s.log(ctx, "DEBUG", message, "", metadata, nil)
}

func (s *LoggerService) Error(ctx context.Context, message string, err error, metadata map[string]any) error {
	stack := ""
	if err != nil {
		stack = err.Error()
	}
	return s.log(ctx, "ERROR", message, stack, metadata, err)
}

func (s *LoggerService) log(ctx context.Context, level, message, stack string, metadata map[string]any, err error) error {
	level = strings.ToUpper(strings.TrimSpace(level))
	if !slices.Contains(ValidLevels, level) {
		return fmt.Errorf("invalid log level: %s", level)
	}

	sanitized := s.sanitizer.SanitizeMap(metadata)
	entry := LogEntry{
		ID:         uuid.New(),
		Level:      level,
		Message:    message,
		Source:     s.source,
		StackTrace: stack,
		Metadata:   sanitized,
		CreatedAt:  time.Now().UTC(),
	}

	if rid, ok := ctx.Value(requestIDKey{}).(string); ok {
		entry.RequestID = rid
	}
	if uid, ok := ctx.Value(userIDKey{}).(uuid.UUID); ok {
		entry.UserID = &uid
	}

	if s.consoleEnabled {
		fmt.Printf("[%s] %s: %s\n", entry.CreatedAt.Format(time.RFC3339), level, message)
		if err != nil {
			fmt.Printf("  error: %v\n", err)
		}
	}

	if !s.dbEnabled {
		return nil
	}
	return s.repo.Insert(ctx, entry)
}

type requestIDKey struct{}
type userIDKey struct{}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

func WithUserID(ctx context.Context, userID uuid.UUID) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// LogHTTP writes an INFO log for HTTP requests (used by middleware).
func (s *LoggerService) LogHTTP(ctx context.Context, message string, metadata map[string]any) error {
	return s.Info(ctx, message, metadata)
}
