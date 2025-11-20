CREATE TABLE IF NOT EXISTS important_logs (
    id BIGSERIAL PRIMARY KEY,
    level TEXT NOT NULL,
    event_type TEXT NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_important_logs_level ON important_logs(level);
CREATE INDEX IF NOT EXISTS idx_important_logs_event_type ON important_logs(event_type);
