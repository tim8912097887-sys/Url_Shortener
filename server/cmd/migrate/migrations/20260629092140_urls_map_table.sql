-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE urls_map (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    long_url VARCHAR(500) NOT NULL,
    short_url_id UUID UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW() NOT NULL,
    expired_at TIMESTAMP NOT NULL
);

-- +goose Down
DROP TABLE urls_map;
