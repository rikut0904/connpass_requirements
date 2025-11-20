package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"connpass-requirement/internal/models"
)

// EventRepository はイベントキャッシュ操作を担う。
type EventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{db: db}
}

func (r *EventRepository) Upsert(ctx context.Context, event *models.Event) error {
	query := `
	INSERT INTO events_cache (
		event_id, title, event_url, started_at, ended_at, "limit",
		accepted, waiting, updated_at, retrieved_at, owner_nickname,
		series_title, hash_digest
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	ON CONFLICT (event_id)
	DO UPDATE SET
		title = EXCLUDED.title,
		event_url = EXCLUDED.event_url,
		started_at = EXCLUDED.started_at,
		ended_at = EXCLUDED.ended_at,
		"limit" = EXCLUDED."limit",
		accepted = EXCLUDED.accepted,
		waiting = EXCLUDED.waiting,
		updated_at = EXCLUDED.updated_at,
		retrieved_at = EXCLUDED.retrieved_at,
		owner_nickname = EXCLUDED.owner_nickname,
		series_title = EXCLUDED.series_title,
		hash_digest = EXCLUDED.hash_digest
	RETURNING id
	`

	return r.db.QueryRowContext(
		ctx,
		query,
		event.EventID,
		event.Title,
		event.EventURL,
		event.StartedAt,
		event.EndedAt,
		event.Limit,
		event.Accepted,
		event.Waiting,
		event.UpdatedAt,
		event.RetrievedAt,
		event.OwnerNickname,
		event.SeriesTitle,
		event.HashDigest,
	).Scan(&event.ID)
}

func (r *EventRepository) FindByEventID(ctx context.Context, eventID int64) (*models.Event, error) {
	var event models.Event
	if err := r.db.QueryRowContext(ctx, `
	SELECT id, event_id, title, event_url, started_at, ended_at,
		"limit", accepted, waiting, updated_at, retrieved_at,
		owner_nickname, series_title, hash_digest
	FROM events_cache
	WHERE event_id = $1
	`, eventID).Scan(
		&event.ID,
		&event.EventID,
		&event.Title,
		&event.EventURL,
		&event.StartedAt,
		&event.EndedAt,
		&event.Limit,
		&event.Accepted,
		&event.Waiting,
		&event.UpdatedAt,
		&event.RetrievedAt,
		&event.OwnerNickname,
		&event.SeriesTitle,
		&event.HashDigest,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("select event: %w", err)
	}
	return &event, nil
}

func (r *EventRepository) Cleanup(ctx context.Context, before time.Time) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM events_cache WHERE retrieved_at < $1`, before); err != nil {
		return fmt.Errorf("cleanup events cache: %w", err)
	}
	return nil
}
