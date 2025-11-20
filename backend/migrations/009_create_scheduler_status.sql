CREATE TABLE IF NOT EXISTS scheduler_status (
    id SMALLINT PRIMARY KEY,
    last_run_at TIMESTAMPTZ,
    last_error TEXT,
    updated_at TIMESTAMPTZ
);
