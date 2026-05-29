package logger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SQLiteLogRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteLogRepository {
	return &SQLiteLogRepository{db: db}
}

func (r *SQLiteLogRepository) Insert(ctx context.Context, entry LogEntry) error {
	meta, err := json.Marshal(entry.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	if entry.Metadata == nil {
		meta = []byte("{}")
	}

	var userID sql.NullString
	if entry.UserID != nil {
		userID = sql.NullString{String: entry.UserID.String(), Valid: true}
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO application_logs (
			id, level, message, source, request_id, user_id,
			error_code, stack_trace, metadata, created_at
		) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		entry.ID.String(), entry.Level, entry.Message, entry.Source, entry.RequestID,
		userID, entry.ErrorCode, entry.StackTrace, string(meta), entry.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("insert log: %w", err)
	}
	return nil
}

func (r *SQLiteLogRepository) Find(ctx context.Context, filter LogFilter) ([]LogEntry, error) {
	var conditions []string
	var args []any

	if filter.Level != "" {
		conditions = append(conditions, "level = ?")
		args = append(args, strings.ToUpper(filter.Level))
	}
	if filter.RequestID != "" {
		conditions = append(conditions, "request_id = ?")
		args = append(args, filter.RequestID)
	}
	if filter.Source != "" {
		conditions = append(conditions, "source = ?")
		args = append(args, filter.Source)
	}
	if filter.UserID != nil {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID.String())
	}
	if filter.From != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, filter.From.UTC().Format(time.RFC3339Nano))
	}
	if filter.To != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, filter.To.UTC().Format(time.RFC3339Nano))
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	limit := filter.Pagination.Limit
	if limit < 1 {
		limit = 20
	}
	offset := filter.Pagination.Offset()

	query := fmt.Sprintf(`
		SELECT id, level, message, source, request_id, user_id,
		       error_code, stack_trace, metadata, created_at
		FROM application_logs %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, where)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find logs: %w", err)
	}
	defer rows.Close()

	return scanSQLiteLogRows(rows)
}

func (r *SQLiteLogRepository) FindByID(ctx context.Context, id uuid.UUID) (*LogEntry, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, level, message, source, request_id, user_id,
		       error_code, stack_trace, metadata, created_at
		FROM application_logs WHERE id = ?`, id.String())
	entry, err := scanSQLiteLogRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}

func scanSQLiteLogRows(rows *sql.Rows) ([]LogEntry, error) {
	var entries []LogEntry
	for rows.Next() {
		entry, err := scanSQLiteLogFromRows(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func scanSQLiteLogRow(row *sql.Row) (LogEntry, error) {
	var entry LogEntry
	var idStr, metaStr, createdStr string
	var userID sql.NullString
	err := row.Scan(
		&idStr, &entry.Level, &entry.Message, &entry.Source, &entry.RequestID,
		&userID, &entry.ErrorCode, &entry.StackTrace, &metaStr, &createdStr,
	)
	if err != nil {
		return entry, err
	}
	return finishSQLiteLogScan(&entry, idStr, userID, metaStr, createdStr)
}

func scanSQLiteLogFromRows(rows *sql.Rows) (LogEntry, error) {
	var entry LogEntry
	var idStr, metaStr, createdStr string
	var userID sql.NullString
	err := rows.Scan(
		&idStr, &entry.Level, &entry.Message, &entry.Source, &entry.RequestID,
		&userID, &entry.ErrorCode, &entry.StackTrace, &metaStr, &createdStr,
	)
	if err != nil {
		return entry, err
	}
	return finishSQLiteLogScan(&entry, idStr, userID, metaStr, createdStr)
}

func finishSQLiteLogScan(entry *LogEntry, idStr string, userID sql.NullString, metaStr, createdStr string) (LogEntry, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return *entry, fmt.Errorf("parse log id: %w", err)
	}
	entry.ID = id
	if userID.Valid {
		uid, err := uuid.Parse(userID.String)
		if err != nil {
			return *entry, fmt.Errorf("parse user_id: %w", err)
		}
		entry.UserID = &uid
	}
	if len(metaStr) > 0 {
		_ = json.Unmarshal([]byte(metaStr), &entry.Metadata)
	}
	entry.CreatedAt, err = parseSQLiteTime(createdStr)
	if err != nil {
		return *entry, err
	}
	return *entry, nil
}

func parseSQLiteTime(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05Z07:00"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("parse time %q", s)
}
