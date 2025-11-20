package models

import "time"

// User はDiscordユーザー情報を表現する。
type User struct {
	ID              int64     `db:"id" json:"id"`
	DiscordUserID   string    `db:"discord_user_id" json:"discordUserId"`
	DiscordUsername string    `db:"discord_username" json:"discordUsername"`
	AvatarURL       string    `db:"avatar_url" json:"avatarUrl"`
	AccessToken     string    `db:"access_token" json:"-"`
	RefreshToken    string    `db:"refresh_token" json:"-"`
	TokenExpiresAt  time.Time `db:"token_expires_at" json:"tokenExpiresAt"`
	CreatedAt       time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt       time.Time `db:"updated_at" json:"updatedAt"`
}

// GuildPermission はユーザーが所属するギルドの権限情報。
type GuildPermission struct {
	ID            int64  `db:"id" json:"id"`
	UserID        int64  `db:"user_id" json:"userId"`
	GuildID       string `db:"guild_id" json:"guildId"`
	GuildName     string `db:"guild_name" json:"guildName"`
	Permissions   int64  `db:"permissions" json:"permissions"`
	IconURL       string `db:"icon_url" json:"iconUrl"`
	CanManage     bool   `db:"can_manage" json:"canManage"`
	CanManageRole bool   `db:"can_manage_role" json:"canManageRole"`
}
