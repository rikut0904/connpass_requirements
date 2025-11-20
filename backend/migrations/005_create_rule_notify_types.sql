CREATE TABLE IF NOT EXISTS rule_notify_types (
    rule_id BIGINT NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
    notify_key notify_trigger NOT NULL,
    PRIMARY KEY(rule_id, notify_key)
);
