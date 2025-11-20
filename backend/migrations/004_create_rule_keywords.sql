CREATE TABLE IF NOT EXISTS rule_keywords (
    rule_id BIGINT NOT NULL REFERENCES rules(id) ON DELETE CASCADE,
    keyword TEXT NOT NULL,
    PRIMARY KEY(rule_id, keyword)
);
