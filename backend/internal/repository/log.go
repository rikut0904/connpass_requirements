package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"connpass-requirement/internal/models"
)

// LogRepository は重要ログおよびスケジューラステータスを扱う。
type LogRepository struct {
	db *sql.DB
}

func NewLogRepository(db *sql.DB) *LogRepository {
	return &LogRepository{db: db}
}

func (r *LogRepository) Save(ctx context.Context, log models.ImportantLog) error {
	_, err := r.db.ExecContext(ctx, `
	INSERT INTO important_logs (level, event_type, message, metadata)
	VALUES ($1, $2, $3, $4)
	`, log.Level, log.EventType, log.Message, log.Metadata)
	if err != nil {
		return fmt.Errorf("insert important log: %w", err)
	}
	return nil
}

func (r *LogRepository) ListRecent(ctx context.Context, limit int) ([]models.ImportantLog, error) {
	rows, err := r.db.QueryContext(ctx, `
	SELECT id, level, event_type, message, metadata, created_at
	FROM important_logs
	ORDER BY created_at DESC
	LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("select important logs: %w", err)
	}
	defer rows.Close()

	logs := make([]models.ImportantLog, 0, limit)
	for rows.Next() {
		var log models.ImportantLog
		if err := rows.Scan(
			&log.ID,
			&log.Level,
			&log.EventType,
			&log.Message,
			&log.Metadata,
			&log.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan important log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

func (r *LogRepository) Cleanup(ctx context.Context, before time.Time) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM important_logs WHERE created_at < $1`, before); err != nil {
		return fmt.Errorf("cleanup important logs: %w", err)
	}
	return nil
}

func (r *LogRepository) UpdateSchedulerStatus(ctx context.Context, status models.SchedulerStatus) error {
	_, err := r.db.ExecContext(ctx, `
	INSERT INTO scheduler_status (id, last_run_at, last_error, updated_at)
	VALUES (1, $1, $2, NOW())
	ON CONFLICT (id)
	DO UPDATE SET
		last_run_at = EXCLUDED.last_run_at,
		last_error = EXCLUDED.last_error,
		updated_at = NOW()
	`, status.LastRunAt, status.LastError)
	if err != nil {
		return fmt.Errorf("upsert scheduler status: %w", err)
	}
	return nil
}

func (r *LogRepository) GetSchedulerStatus(ctx context.Context) (*models.SchedulerStatus, error) {
	var status models.SchedulerStatus
	if err := r.db.QueryRowContext(ctx, `
	SELECT id, last_run_at, last_error, updated_at
	FROM scheduler_status
	WHERE id = 1
	`).Scan(
		&status.ID,
		&status.LastRunAt,
		&status.LastError,
		&status.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select scheduler status: %w", err)
	}
	return &status, nil
}
