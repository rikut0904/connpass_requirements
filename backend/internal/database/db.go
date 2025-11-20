package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"connpass-requirement/internal/config"
)

// Connect はPostgreSQLへの接続を確立する。
func Connect(ctx context.Context, cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	const (
		maxAttempts   = 10
		retryInterval = 2 * time.Second
	)

	var pingErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		pingErr = db.PingContext(pingCtx)
		cancel()

		if pingErr == nil {
			return db, nil
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("ping database: %w", ctx.Err())
		case <-time.After(retryInterval):
		}
	}

	return nil, fmt.Errorf("ping database: %w", pingErr)
}
