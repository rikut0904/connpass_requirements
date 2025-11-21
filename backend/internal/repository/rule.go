package repository

import (
	"context"
	"database/sql"
	"fmt"

	"connpass-requirement/internal/models"
)

// RuleRepository は通知ルールの操作を司る。
type RuleRepository struct {
	db *sql.DB
}

func NewRuleRepository(db *sql.DB) *RuleRepository {
	return &RuleRepository{db: db}
}

// ListActive はアクティブなルールを全件取得する。スケジューラ専用。
func (r *RuleRepository) ListActive(ctx context.Context) ([]models.Rule, error) {
	rows, err := r.db.QueryContext(ctx, `
	SELECT id, user_id, guild_id, channel_id, channel_name, name,
		description, location, capacity_threshold, is_active,
		created_at, updated_at
	FROM rules
	WHERE is_active = TRUE
	ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("select active rules: %w", err)
	}
	defer rows.Close()

	var rules []models.Rule
	for rows.Next() {
		var rule models.Rule
		if err := rows.Scan(
			&rule.ID,
			&rule.UserID,
			&rule.GuildID,
			&rule.ChannelID,
			&rule.ChannelName,
			&rule.Name,
			&rule.Description,
			&rule.Location,
			&rule.CapacityThresh,
			&rule.IsActive,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan active rule: %w", err)
		}
		if err := r.attachKeywordsAndTypes(ctx, &rule); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

// ListByUserAndGuild は指定ユーザー・ギルドのルールを一覧取得する。
func (r *RuleRepository) ListByUserAndGuild(ctx context.Context, userID int64, guildID string) ([]models.Rule, error) {
	rows, err := r.db.QueryContext(ctx, `
	SELECT id, user_id, guild_id, channel_id, channel_name, name,
		description, location, capacity_threshold, is_active,
		created_at, updated_at
	FROM rules
	WHERE user_id = $1 AND guild_id = $2
	ORDER BY created_at DESC
	`, userID, guildID)
	if err != nil {
		return nil, fmt.Errorf("select rules: %w", err)
	}
	defer rows.Close()

	rules := make([]models.Rule, 0)

	for rows.Next() {
		var rule models.Rule
		if err := rows.Scan(
			&rule.ID,
			&rule.UserID,
			&rule.GuildID,
			&rule.ChannelID,
			&rule.ChannelName,
			&rule.Name,
			&rule.Description,
			&rule.Location,
			&rule.CapacityThresh,
			&rule.IsActive,
			&rule.CreatedAt,
			&rule.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan rule: %w", err)
		}

		if err := r.attachKeywordsAndTypes(ctx, &rule); err != nil {
			return nil, err
		}

		rules = append(rules, rule)
	}

	return rules, rows.Err()
}

// Get は単一ルールを取得する。
func (r *RuleRepository) Get(ctx context.Context, ruleID int64) (*models.Rule, error) {
	var rule models.Rule
	if err := r.db.QueryRowContext(ctx, `
	SELECT id, user_id, guild_id, channel_id, channel_name, name,
		description, location, capacity_threshold, is_active,
		created_at, updated_at
	FROM rules
	WHERE id = $1
	`, ruleID).Scan(
		&rule.ID,
		&rule.UserID,
		&rule.GuildID,
		&rule.ChannelID,
		&rule.ChannelName,
		&rule.Name,
		&rule.Description,
		&rule.Location,
		&rule.CapacityThresh,
		&rule.IsActive,
		&rule.CreatedAt,
		&rule.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select rule: %w", err)
	}

	if err := r.attachKeywordsAndTypes(ctx, &rule); err != nil {
		return nil, err
	}

	return &rule, nil
}

// Create は新しいルールを作成する。
func (r *RuleRepository) Create(ctx context.Context, rule *models.Rule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback() //nolint:errcheck
		}
	}()

	err = tx.QueryRowContext(ctx, `
	INSERT INTO rules (
		user_id, guild_id, channel_id, channel_name, name,
		description, location, capacity_threshold, is_active
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id, created_at, updated_at
	`,
		rule.UserID,
		rule.GuildID,
		rule.ChannelID,
		rule.ChannelName,
		rule.Name,
		rule.Description,
		rule.Location,
		rule.CapacityThresh,
		rule.IsActive,
	).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert rule: %w", err)
	}

	if err = insertKeywords(ctx, tx, rule.ID, rule.Keywords); err != nil {
		return err
	}
	if err = insertNotifyTypes(ctx, tx, rule.ID, rule.NotifyTypes); err != nil {
		return err
	}

	return tx.Commit()
}

// Update は既存ルールを更新する。
func (r *RuleRepository) Update(ctx context.Context, rule *models.Rule) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback() //nolint:errcheck
		}
	}()

	_, err = tx.ExecContext(ctx, `
	UPDATE rules
	SET channel_id = $1,
		channel_name = $2,
		name = $3,
		description = $4,
		location = $5,
		capacity_threshold = $6,
		is_active = $7,
		updated_at = NOW()
	WHERE id = $8
	`,
		rule.ChannelID,
		rule.ChannelName,
		rule.Name,
		rule.Description,
		rule.Location,
		rule.CapacityThresh,
		rule.IsActive,
		rule.ID,
	)
	if err != nil {
		return fmt.Errorf("update rule: %w", err)
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM rule_keywords WHERE rule_id = $1`, rule.ID); err != nil {
		return fmt.Errorf("delete rule keywords: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM rule_notify_types WHERE rule_id = $1`, rule.ID); err != nil {
		return fmt.Errorf("delete notify types: %w", err)
	}

	if err = insertKeywords(ctx, tx, rule.ID, rule.Keywords); err != nil {
		return err
	}
	if err = insertNotifyTypes(ctx, tx, rule.ID, rule.NotifyTypes); err != nil {
		return err
	}

	return tx.Commit()
}

// Delete はルールを削除する。
func (r *RuleRepository) Delete(ctx context.Context, ruleID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM rules WHERE id = $1`, ruleID)
	if err != nil {
		return fmt.Errorf("delete rule: %w", err)
	}
	return nil
}

func (r *RuleRepository) attachKeywordsAndTypes(ctx context.Context, rule *models.Rule) error {
	keywordRows, err := r.db.QueryContext(ctx, `SELECT keyword FROM rule_keywords WHERE rule_id = $1 ORDER BY keyword`, rule.ID)
	if err != nil {
		return fmt.Errorf("select keywords: %w", err)
	}
	defer keywordRows.Close()

	var keywords []string
	for keywordRows.Next() {
		var keyword string
		if err := keywordRows.Scan(&keyword); err != nil {
			return fmt.Errorf("scan keyword: %w", err)
		}
		keywords = append(keywords, keyword)
	}
	rule.Keywords = keywords

	typeRows, err := r.db.QueryContext(ctx, `SELECT notify_key FROM rule_notify_types WHERE rule_id = $1 ORDER BY notify_key`, rule.ID)
	if err != nil {
		return fmt.Errorf("select notify types: %w", err)
	}
	defer typeRows.Close()

	var types []string
	for typeRows.Next() {
		var notify string
		if err := typeRows.Scan(&notify); err != nil {
			return fmt.Errorf("scan notify type: %w", err)
		}
		types = append(types, notify)
	}
	rule.NotifyTypes = types

	return nil
}

func insertKeywords(ctx context.Context, tx *sql.Tx, ruleID int64, keywords []string) error {
	if len(keywords) == 0 {
		return nil
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO rule_keywords (rule_id, keyword) VALUES ($1, $2)`)
	if err != nil {
		return fmt.Errorf("prepare insert keyword: %w", err)
	}
	defer stmt.Close()

	for _, keyword := range keywords {
		if keyword == "" {
			continue
		}
		if _, err := stmt.ExecContext(ctx, ruleID, keyword); err != nil {
			return fmt.Errorf("insert keyword: %w", err)
		}
	}
	return nil
}

func insertNotifyTypes(ctx context.Context, tx *sql.Tx, ruleID int64, notifyTypes []string) error {
	if len(notifyTypes) == 0 {
		return nil
	}
	stmt, err := tx.PrepareContext(ctx, `INSERT INTO rule_notify_types (rule_id, notify_key) VALUES ($1, $2)`)
	if err != nil {
		return fmt.Errorf("prepare insert notify type: %w", err)
	}
	defer stmt.Close()

	for _, notify := range notifyTypes {
		if notify == "" {
			continue
		}
		if _, err := stmt.ExecContext(ctx, ruleID, notify); err != nil {
			return fmt.Errorf("insert notify type: %w", err)
		}
	}
	return nil
}
