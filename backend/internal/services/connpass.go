package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"connpass-requirement/internal/config"
	"connpass-requirement/internal/models"
)

// ConnpassService はconnpass APIとの連携を行う。
type ConnpassService struct {
	client  *http.Client
	cfg     config.Config
	limiter <-chan time.Time
}

// NewConnpassService はレートリミットを考慮したクライアントを生成する。
func NewConnpassService(cfg config.Config) *ConnpassService {
	return &ConnpassService{
		client:  &http.Client{Timeout: 10 * time.Second},
		cfg:     cfg,
		limiter: time.Tick(cfg.ConnpassRequestInterval),
	}
}

// FetchEvents はキーワードと開催地からイベントを取得する。
func (s *ConnpassService) FetchEvents(ctx context.Context, keyword, location string) ([]models.Event, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-s.limiter:
	}

	u, err := url.Parse(s.cfg.ConnpassBaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse connpass url: %w", err)
	}

	q := u.Query()
	q.Set("keyword", keyword)
	q.Set("count", strconv.Itoa(20))
	if location != "" {
		q.Set("address", location)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	// connpass API v2は X-API-Key ヘッダーが必須
	req.Header.Set("X-API-Key", s.cfg.ConnpassAPIKey)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request connpass: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return nil, fmt.Errorf("connpass status %d: %s", resp.StatusCode, string(body))
	}

	var payload struct {
		Count  int `json:"count"`
		Events []struct {
			ID            int64  `json:"id"`
			Title         string `json:"title"`
			URL           string `json:"url"`
			StartedAt     string `json:"started_at"`
			EndedAt       string `json:"ended_at"`
			Limit         int    `json:"limit"`
			Accepted      int    `json:"accepted"`
			Waiting       int    `json:"waiting"`
			UpdatedAt     string `json:"updated_at"`
			OwnerNickname string `json:"owner_nickname"`
			Group         struct {
				Title string `json:"title"`
			} `json:"group"`
		} `json:"events"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode connpass response: %w", err)
	}

	var events []models.Event
	for _, ev := range payload.Events {
		startedAt, _ := time.Parse(time.RFC3339, ev.StartedAt)
		endedAt, _ := time.Parse(time.RFC3339, ev.EndedAt)
		updatedAt, _ := time.Parse(time.RFC3339, ev.UpdatedAt)

		hash := sha1.Sum([]byte(fmt.Sprintf("%d:%s:%s:%d:%d", ev.ID, ev.Title, ev.UpdatedAt, ev.Accepted, ev.Limit)))

		events = append(events, models.Event{
			EventID:       ev.ID,
			Title:         ev.Title,
			EventURL:      ev.URL,
			StartedAt:     startedAt,
			EndedAt:       endedAt,
			Limit:         ev.Limit,
			Accepted:      ev.Accepted,
			Waiting:       ev.Waiting,
			UpdatedAt:     updatedAt,
			RetrievedAt:   time.Now().UTC(),
			OwnerNickname: ev.OwnerNickname,
			SeriesTitle:   ev.Group.Title,
			HashDigest:    hex.EncodeToString(hash[:]),
		})
	}

	return events, nil
}
