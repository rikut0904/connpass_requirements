package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"connpass-requirement/internal/models"
)

// UserRepository はユーザー関連のDB操作を担当する。
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Upsert はDiscordユーザー情報を保存または更新する。
func (r *UserRepository) Upsert(ctx context.Context, user *models.User) error {
	query := `
	INSERT INTO users (
		discord_user_id, discord_username, avatar_url,
		access_token, refresh_token, token_expires_at
	) VALUES ($1, $2, $3, $4, $5, $6)
	ON CONFLICT (discord_user_id)
	DO UPDATE SET
		discord_username = EXCLUDED.discord_username,
		avatar_url = EXCLUDED.avatar_url,
		access_token = EXCLUDED.access_token,
		refresh_token = EXCLUDED.refresh_token,
		token_expires_at = EXCLUDED.token_expires_at,
		updated_at = NOW()
	RETURNING id, created_at, updated_at
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		user.DiscordUserID,
		user.DiscordUsername,
		user.AvatarURL,
		user.AccessToken,
		user.RefreshToken,
		user.TokenExpiresAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// SaveGuildPermissions はユーザーのギルド権限情報を保存する。
func (r *UserRepository) SaveGuildPermissions(ctx context.Context, userID int64, guilds []models.GuildPermission) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback() //nolint:errcheck
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM guild_permissions WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete old guild permissions: %w", err)
	}

	if len(guilds) == 0 {
		return tx.Commit()
	}

	stmt, err := tx.PrepareContext(ctx, `
	INSERT INTO guild_permissions (
		user_id, guild_id, guild_name, permissions, icon_url,
		can_manage, can_manage_role
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert guild permissions: %w", err)
	}
	defer stmt.Close()

	for _, g := range guilds {
		if _, err = stmt.ExecContext(
			ctx,
			userID,
			g.GuildID,
			g.GuildName,
			g.Permissions,
			g.IconURL,
			g.CanManage,
			g.CanManageRole,
		); err != nil {
			return fmt.Errorf("insert guild permission: %w", err)
		}
	}

	return tx.Commit()
}

// ListGuildPermissions はユーザーのギルド権限を取得する。
func (r *UserRepository) ListGuildPermissions(ctx context.Context, userID int64) ([]models.GuildPermission, error) {
	rows, err := r.db.QueryContext(ctx, `
	SELECT id, user_id, guild_id, guild_name, permissions, icon_url,
		can_manage, can_manage_role
	FROM guild_permissions
	WHERE user_id = $1
	ORDER BY guild_name ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("query guild permissions: %w", err)
	}
	defer rows.Close()

	var guilds []models.GuildPermission

	for rows.Next() {
		var g models.GuildPermission
		if err := rows.Scan(
			&g.ID,
			&g.UserID,
			&g.GuildID,
			&g.GuildName,
			&g.Permissions,
			&g.IconURL,
			&g.CanManage,
			&g.CanManageRole,
		); err != nil {
			return nil, fmt.Errorf("scan guild permission: %w", err)
		}
		guilds = append(guilds, g)
	}

	return guilds, rows.Err()
}

// FindByDiscordID はDiscordユーザーIDでユーザーを検索する。
func (r *UserRepository) FindByDiscordID(ctx context.Context, discordID string) (*models.User, error) {
	var user models.User
	if err := r.db.QueryRowContext(ctx, `
	SELECT id, discord_user_id, discord_username, avatar_url,
		access_token, refresh_token, token_expires_at,
		created_at, updated_at
	FROM users
	WHERE discord_user_id = $1
	`, discordID).Scan(
		&user.ID,
		&user.DiscordUserID,
		&user.DiscordUsername,
		&user.AvatarURL,
		&user.AccessToken,
		&user.RefreshToken,
		&user.TokenExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select user: %w", err)
	}
	return &user, nil
}

// FindByID はユーザーIDでユーザーを検索する。
func (r *UserRepository) FindByID(ctx context.Context, userID int64) (*models.User, error) {
	var user models.User
	if err := r.db.QueryRowContext(ctx, `
	SELECT id, discord_user_id, discord_username, avatar_url,
		access_token, refresh_token, token_expires_at,
		created_at, updated_at
	FROM users
	WHERE id = $1
	`, userID).Scan(
		&user.ID,
		&user.DiscordUserID,
		&user.DiscordUsername,
		&user.AvatarURL,
		&user.AccessToken,
		&user.RefreshToken,
		&user.TokenExpiresAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("select user: %w", err)
	}
	return &user, nil
}

// GetGuildPermissions はユーザーのギルド権限を取得する（ListGuildPermissionsのエイリアス）。
func (r *UserRepository) GetGuildPermissions(ctx context.Context, userID int64) ([]models.GuildPermission, error) {
	return r.ListGuildPermissions(ctx, userID)
}
