package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresLogRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresLogRepository {
	return &PostgresLogRepository{pool: pool}
}

func (r *PostgresLogRepository) Insert(ctx context.Context, entry LogEntry) error {
	meta, err := json.Marshal(entry.Metadata)
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}
	if entry.Metadata == nil {
		meta = []byte("{}")
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO application_logs (
			id, level, message, source, request_id, user_id,
			error_code, stack_trace, metadata, created_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		entry.ID, entry.Level, entry.Message, entry.Source, entry.RequestID,
		entry.UserID, entry.ErrorCode, entry.StackTrace, meta, entry.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert log: %w", err)
	}
	return nil
}

func (r *PostgresLogRepository) Find(ctx context.Context, filter LogFilter) ([]LogEntry, error) {
	var conditions []string
	var args []any
	n := 1

	if filter.Level != "" {
		conditions = append(conditions, fmt.Sprintf("level = $%d", n))
		args = append(args, strings.ToUpper(filter.Level))
		n++
	}
	if filter.RequestID != "" {
		conditions = append(conditions, fmt.Sprintf("request_id = $%d", n))
		args = append(args, filter.RequestID)
		n++
	}
	if filter.Source != "" {
		conditions = append(conditions, fmt.Sprintf("source = $%d", n))
		args = append(args, filter.Source)
		n++
	}
	if filter.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", n))
		args = append(args, *filter.UserID)
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
		SELECT id, level, message, source, request_id, user_id,
		       error_code, stack_trace, metadata, created_at
		FROM application_logs %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, n, n+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find logs: %w", err)
	}
	defer rows.Close()

	return scanLogRows(rows)
}

func (r *PostgresLogRepository) FindByID(ctx context.Context, id uuid.UUID) (*LogEntry, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, level, message, source, request_id, user_id,
		       error_code, stack_trace, metadata, created_at
		FROM application_logs WHERE id = $1`, id)
	entry, err := scanLogRow(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &entry, nil
}

func scanLogRows(rows pgx.Rows) ([]LogEntry, error) {
	var entries []LogEntry
	for rows.Next() {
		entry, err := scanLogFromRows(rows)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func scanLogRow(row pgx.Row) (LogEntry, error) {
	var entry LogEntry
	var meta []byte
	err := row.Scan(
		&entry.ID, &entry.Level, &entry.Message, &entry.Source, &entry.RequestID,
		&entry.UserID, &entry.ErrorCode, &entry.StackTrace, &meta, &entry.CreatedAt,
	)
	if err != nil {
		return entry, err
	}
	if len(meta) > 0 {
		_ = json.Unmarshal(meta, &entry.Metadata)
	}
	return entry, nil
}

func scanLogFromRows(rows pgx.Rows) (LogEntry, error) {
	var entry LogEntry
	var meta []byte
	err := rows.Scan(
		&entry.ID, &entry.Level, &entry.Message, &entry.Source, &entry.RequestID,
		&entry.UserID, &entry.ErrorCode, &entry.StackTrace, &meta, &entry.CreatedAt,
	)
	if err != nil {
		return entry, err
	}
	if len(meta) > 0 {
		_ = json.Unmarshal(meta, &entry.Metadata)
	}
	return entry, nil
}
