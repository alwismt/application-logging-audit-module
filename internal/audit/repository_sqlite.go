package audit

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type SQLiteAuditRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteAuditRepository {
	return &SQLiteAuditRepository{db: db}
}

func (r *SQLiteAuditRepository) Insert(ctx context.Context, event AuditEvent) error {
	oldVal, err := marshalJSON(event.OldValue)
	if err != nil {
		return err
	}
	newVal, err := marshalJSON(event.NewValue)
	if err != nil {
		return err
	}
	meta, err := marshalJSON(event.Metadata)
	if err != nil {
		return err
	}

	var userID sql.NullString
	if event.UserID != nil {
		userID = sql.NullString{String: event.UserID.String(), Valid: true}
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO audit_events (
			id, user_id, username, action, resource_type, resource_id,
			old_value, new_value, ip_address, user_agent, request_id,
			status, metadata, created_at
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		event.ID.String(), userID, event.Username, event.Action, event.ResourceType,
		event.ResourceID, string(oldVal), string(newVal), event.IPAddress, event.UserAgent,
		event.RequestID, event.Status, string(meta), event.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

func (r *SQLiteAuditRepository) Find(ctx context.Context, filter AuditFilter) ([]AuditEvent, error) {
	var conditions []string
	var args []any

	if filter.UserID != nil {
		conditions = append(conditions, "user_id = ?")
		args = append(args, filter.UserID.String())
	}
	if filter.Username != "" {
		conditions = append(conditions, "username = ?")
		args = append(args, filter.Username)
	}
	if filter.Action != "" {
		conditions = append(conditions, "action = ?")
		args = append(args, strings.ToUpper(filter.Action))
	}
	if filter.ResourceType != "" {
		conditions = append(conditions, "resource_type = ?")
		args = append(args, filter.ResourceType)
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, strings.ToUpper(filter.Status))
	}
	if filter.RequestID != "" {
		conditions = append(conditions, "request_id = ?")
		args = append(args, filter.RequestID)
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
		SELECT id, user_id, username, action, resource_type, resource_id,
		       old_value, new_value, ip_address, user_agent, request_id,
		       status, metadata, created_at
		FROM audit_events %s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`, where)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find audit events: %w", err)
	}
	defer rows.Close()

	return scanSQLiteAuditRows(rows)
}

func (r *SQLiteAuditRepository) FindByID(ctx context.Context, id uuid.UUID) (*AuditEvent, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, username, action, resource_type, resource_id,
		       old_value, new_value, ip_address, user_agent, request_id,
		       status, metadata, created_at
		FROM audit_events WHERE id = ?`, id.String())
	event, err := scanSQLiteAuditRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &event, nil
}

func scanSQLiteAuditRows(rows *sql.Rows) ([]AuditEvent, error) {
	var events []AuditEvent
	for rows.Next() {
		event, err := scanSQLiteAuditFromRows(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func scanSQLiteAuditRow(row *sql.Row) (AuditEvent, error) {
	var event AuditEvent
	var idStr, oldVal, newVal, meta, createdStr string
	var userID sql.NullString
	err := row.Scan(
		&idStr, &userID, &event.Username, &event.Action,
		&event.ResourceType, &event.ResourceID, &oldVal, &newVal,
		&event.IPAddress, &event.UserAgent, &event.RequestID,
		&event.Status, &meta, &createdStr,
	)
	if err != nil {
		return event, err
	}
	return finishSQLiteAuditScan(&event, idStr, userID, oldVal, newVal, meta, createdStr)
}

func scanSQLiteAuditFromRows(rows *sql.Rows) (AuditEvent, error) {
	var event AuditEvent
	var idStr, oldVal, newVal, meta, createdStr string
	var userID sql.NullString
	err := rows.Scan(
		&idStr, &userID, &event.Username, &event.Action,
		&event.ResourceType, &event.ResourceID, &oldVal, &newVal,
		&event.IPAddress, &event.UserAgent, &event.RequestID,
		&event.Status, &meta, &createdStr,
	)
	if err != nil {
		return event, err
	}
	return finishSQLiteAuditScan(&event, idStr, userID, oldVal, newVal, meta, createdStr)
}

func finishSQLiteAuditScan(event *AuditEvent, idStr string, userID sql.NullString, oldVal, newVal, meta, createdStr string) (AuditEvent, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return *event, fmt.Errorf("parse audit id: %w", err)
	}
	event.ID = id
	if userID.Valid {
		uid, err := uuid.Parse(userID.String)
		if err != nil {
			return *event, fmt.Errorf("parse user_id: %w", err)
		}
		event.UserID = &uid
	}
	unmarshalJSON([]byte(oldVal), &event.OldValue)
	unmarshalJSON([]byte(newVal), &event.NewValue)
	unmarshalJSON([]byte(meta), &event.Metadata)
	event.CreatedAt, err = parseSQLiteAuditTime(createdStr)
	if err != nil {
		return *event, err
	}
	return *event, nil
}

func parseSQLiteAuditTime(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05Z07:00"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("parse time %q", s)
}
