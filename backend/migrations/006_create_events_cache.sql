CREATE TABLE IF NOT EXISTS events_cache (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    event_url TEXT NOT NULL,
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    "limit" INTEGER,
    accepted INTEGER,
    waiting INTEGER,
    updated_at TIMESTAMPTZ,
    retrieved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    owner_nickname TEXT,
    series_title TEXT,
    hash_digest TEXT
);

CREATE INDEX IF NOT EXISTS idx_events_cache_event_id ON events_cache(event_id);
