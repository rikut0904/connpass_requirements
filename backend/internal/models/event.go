package models

import "time"

// Event はconnpassイベントキャッシュを表す。
type Event struct {
	ID            int64     `db:"id" json:"id"`
	EventID       int64     `db:"event_id" json:"eventId"`
	Title         string    `db:"title" json:"title"`
	EventURL      string    `db:"event_url" json:"eventUrl"`
	StartedAt     time.Time `db:"started_at" json:"startedAt"`
	EndedAt       time.Time `db:"ended_at" json:"endedAt"`
	Limit         int       `db:"limit" json:"limit"`
	Accepted      int       `db:"accepted" json:"accepted"`
	Waiting       int       `db:"waiting" json:"waiting"`
	UpdatedAt     time.Time `db:"updated_at" json:"updatedAt"`
	RetrievedAt   time.Time `db:"retrieved_at" json:"retrievedAt"`
	OwnerNickname string    `db:"owner_nickname" json:"ownerNickname"`
	SeriesTitle   string    `db:"series_title" json:"seriesTitle"`
	HashDigest    string    `db:"hash_digest" json:"hashDigest"`
}

// Notification は通知済みイベントの履歴。
type Notification struct {
	ID        int64     `db:"id" json:"id"`
	RuleID    int64     `db:"rule_id" json:"ruleId"`
	EventID   int64     `db:"event_id" json:"eventId"`
	NotifyKey string    `db:"notify_key" json:"notifyKey"`
	SentAt    time.Time `db:"sent_at" json:"sentAt"`
}
