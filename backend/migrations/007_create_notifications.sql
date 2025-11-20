CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    rule_id BIGINT NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
    event_id BIGINT NOT NULL,
    notify_key notify_trigger NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(rule_id, event_id, notify_key)
);

CREATE INDEX IF NOT EXISTS idx_notifications_event_id ON notifications(event_id);
