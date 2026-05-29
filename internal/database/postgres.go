package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultConnectAttempts = 5
	defaultConnectBackoff  = 2 * time.Second
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	return ConnectWithRetry(ctx, databaseURL, defaultConnectAttempts, defaultConnectBackoff)
}

// ConnectWithRetry connects and pings PostgreSQL, retrying on transient failures (e.g. Docker first boot).
func ConnectWithRetry(ctx context.Context, databaseURL string, attempts int, backoff time.Duration) (*pgxpool.Pool, error) {
	if attempts < 1 {
		attempts = 1
	}

	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnLifetime = time.Hour

	var lastErr error
	for i := 0; i < attempts; i++ {
		if i > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		pool, err := pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			lastErr = fmt.Errorf("create connection pool: %w", err)
			continue
		}
		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			lastErr = fmt.Errorf("ping database: %w", err)
			continue
		}
		return pool, nil
	}
	return nil, fmt.Errorf("connect after %d attempts: %w", attempts, lastErr)
}

func Ping(ctx context.Context, pool *pgxpool.Pool) error {
	return pool.Ping(ctx)
}
