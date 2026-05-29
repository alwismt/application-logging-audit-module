package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresAuditRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresAuditRepository {
	return &PostgresAuditRepository{pool: pool}
}

func (r *PostgresAuditRepository) Insert(ctx context.Context, event AuditEvent) error {
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

	_, err = r.pool.Exec(ctx, `
		INSERT INTO audit_events (
			id, user_id, username, action, resource_type, resource_id,
			old_value, new_value, ip_address, user_agent, request_id,
			status, metadata, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		event.ID, event.UserID, event.Username, event.Action, event.ResourceType,
		event.ResourceID, oldVal, newVal, event.IPAddress, event.UserAgent,
		event.RequestID, event.Status, meta, event.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}
	return nil
}

func marshalJSON(m map[string]any) ([]byte, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshal json: %w", err)
	}
	return b, nil
}

func (r *PostgresAuditRepository) Find(ctx context.Context, filter AuditFilter) ([]AuditEvent, error) {
	var conditions []string
	var args []any
	n := 1

	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", n))
		args = append(args, *filter.UserID)
		n++
	}
	if filter.Username != "" {
		conditions = append(conditions, fmt.Sprintf("username = $%d", n))
		args = append(args, filter.Username)
		n++
	}
	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", n))
		args = append(args, strings.ToUpper(filter.Action))
		n++
	}
	if filter.ResourceType != "" {
		conditions = append(conditions, fmt.Sprintf("resource_type = $%d", n))
		args = append(args, filter.ResourceType)
		n++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", n))
		args = append(args, strings.ToUpper(filter.Status))
		n++
	}
	if filter.RequestID != "" {
		conditions = append(conditions, fmt.Sprintf("request_id = $%d", n))
		args = append(args, filter.RequestID)
		n++
	}
	if filter.From != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", n))
		args = append(args, *filter.From)
		n++
	}
	if filter.To != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", n))
		args = append(args, *filter.To)
		n++
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
		LIMIT $%d OFFSET $%d`, where, n, n+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find audit events: %w", err)
	}
	defer rows.Close()

	return scanAuditRows(rows)
}

func (r *PostgresAuditRepository) FindByID(ctx context.Context, id uuid.UUID) (*AuditEvent, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, username, action, resource_type, resource_id,
		       old_value, new_value, ip_address, user_agent, request_id,
		       status, metadata, created_at
		FROM audit_events WHERE id = $1`, id)
	event, err := scanAuditRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &event, nil
}

func scanAuditRows(rows pgx.Rows) ([]AuditEvent, error) {
	var events []AuditEvent
	for rows.Next() {
		event, err := scanAuditFromRows(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func scanAuditRow(row pgx.Row) (AuditEvent, error) {
	var event AuditEvent
	var oldVal, newVal, meta []byte
	err := row.Scan(
		&event.ID, &event.UserID, &event.Username, &event.Action,
		&event.ResourceType, &event.ResourceID, &oldVal, &newVal,
		&event.IPAddress, &event.UserAgent, &event.RequestID,
		&event.Status, &meta, &event.CreatedAt,
	)
	if err != nil {
		return event, err
	}
	unmarshalJSON(oldVal, &event.OldValue)
	unmarshalJSON(newVal, &event.NewValue)
	unmarshalJSON(meta, &event.Metadata)
	return event, nil
}

func scanAuditFromRows(rows pgx.Rows) (AuditEvent, error) {
	var event AuditEvent
	var oldVal, newVal, meta []byte
	err := rows.Scan(
		&event.ID, &event.UserID, &event.Username, &event.Action,
		&event.ResourceType, &event.ResourceID, &oldVal, &newVal,
		&event.IPAddress, &event.UserAgent, &event.RequestID,
		&event.Status, &meta, &event.CreatedAt,
	)
	if err != nil {
		return event, err
	}
	unmarshalJSON(oldVal, &event.OldValue)
	unmarshalJSON(newVal, &event.NewValue)
	unmarshalJSON(meta, &event.Metadata)
	return event, nil
}

func unmarshalJSON(data []byte, target *map[string]any) {
	if len(data) > 0 {
		_ = json.Unmarshal(data, target)
	}
}
