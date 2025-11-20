package models

import "time"

// Rule は通知ルールの基本情報。
type Rule struct {
	ID             int64     `db:"id" json:"id"`
	UserID         int64     `db:"user_id" json:"userId"`
	GuildID        string    `db:"guild_id" json:"guildId"`
	ChannelID      string    `db:"channel_id" json:"channelId"`
	ChannelName    string    `db:"channel_name" json:"channelName"`
	Name           string    `db:"name" json:"name"`
	Description    string    `db:"description" json:"description"`
	NotifyTypes    []string  `json:"notifyTypes"`
	Keywords       []string  `json:"keywords"`
	Tags           []string  `json:"tags"`
	Location       string    `db:"location" json:"location"`
	CapacityThresh int       `db:"capacity_threshold" json:"capacityThreshold"`
	IsActive       bool      `db:"is_active" json:"isActive"`
	CreatedAt      time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time `db:"updated_at" json:"updatedAt"`
}

// RuleKeyword はルールとキーワードのマッピング。
type RuleKeyword struct {
	RuleID  int64  `db:"rule_id"`
	Keyword string `db:"keyword"`
}

// RuleNotifyType はルールの通知条件マッピング。
type RuleNotifyType struct {
	RuleID    int64  `db:"rule_id"`
	NotifyKey string `db:"notify_key"`
}
