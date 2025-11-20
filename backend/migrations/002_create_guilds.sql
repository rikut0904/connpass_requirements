CREATE TABLE IF NOT EXISTS guild_permissions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    guild_id TEXT NOT NULL,
    guild_name TEXT NOT NULL,
    permissions BIGINT NOT NULL,
    icon_url TEXT,
    can_manage BOOLEAN NOT NULL DEFAULT FALSE,
    can_manage_role BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_guild_permissions_unique ON guild_permissions(user_id, guild_id);
