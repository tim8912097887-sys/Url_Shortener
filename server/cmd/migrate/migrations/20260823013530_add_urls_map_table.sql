-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE urls_map (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    long_url VARCHAR(500) NOT NULL,
    short_url VARCHAR(8) NOT NULL,
    clicks INT DEFAULT 0 NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expired_at TIMESTAMPTZ NOT NULL,
    -- Composite unique constraint for authenticated users
    CONSTRAINT unique_user_long_url UNIQUE (user_id, long_url),
    CONSTRAINT unique_short_url UNIQUE (short_url)
);
CREATE INDEX idx_urls_map_user_expired
    ON urls_map (user_id, expired_at DESC);

-- +goose Down
DROP TABLE urls_map;
