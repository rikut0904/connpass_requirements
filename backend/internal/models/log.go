package models

import "time"

// ImportantLog は重要ログテーブル。
type ImportantLog struct {
	ID        int64     `db:"id" json:"id"`
	Level     string    `db:"level" json:"level"`
	EventType string    `db:"event_type" json:"eventType"`
	Message   string    `db:"message" json:"message"`
	Metadata  string    `db:"metadata" json:"metadata"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

// SchedulerStatus はスケジューラの状態を記録。
type SchedulerStatus struct {
	ID        int64     `db:"id" json:"id"`
	LastRunAt time.Time `db:"last_run_at" json:"lastRunAt"`
	LastError string    `db:"last_error" json:"lastError"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}
