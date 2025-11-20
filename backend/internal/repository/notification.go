package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// NotificationRepository は通知履歴を扱う。
type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Exists(ctx context.Context, ruleID, eventID int64, notifyKey string) (bool, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `
	SELECT EXISTS (
		SELECT 1 FROM notifications
		WHERE rule_id = $1 AND event_id = $2 AND notify_key = $3
	)
	`, ruleID, eventID, notifyKey).Scan(&exists); err != nil {
		return false, fmt.Errorf("check notification existence: %w", err)
	}
	return exists, nil
}

func (r *NotificationRepository) Record(ctx context.Context, ruleID, eventID int64, notifyKey string) error {
	_, err := r.db.ExecContext(ctx, `
	INSERT INTO notifications (rule_id, event_id, notify_key, sent_at)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT (rule_id, event_id, notify_key)
	DO NOTHING
	`, ruleID, eventID, notifyKey, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("insert notification: %w", err)
	}
	return nil
}

func (r *NotificationRepository) Cleanup(ctx context.Context, before time.Time) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM notifications WHERE sent_at < $1`, before); err != nil {
		return fmt.Errorf("cleanup notifications: %w", err)
	}
	return nil
}
