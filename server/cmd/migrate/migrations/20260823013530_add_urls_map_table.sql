-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE urls_map (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    long_url VARCHAR(500) NOT NULL,
    short_url VARCHAR(8) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    expired_at TIMESTAMP NOT NULL,
    -- Composite unique constraint for authenticated users
    CONSTRAINT unique_user_long_url UNIQUE (user_id, long_url)
);

-- +goose Down
DROP TABLE urls_map;
