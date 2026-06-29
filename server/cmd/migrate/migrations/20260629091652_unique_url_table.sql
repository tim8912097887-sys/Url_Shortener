-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE unique_urls (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    url VARCHAR(8) UNIQUE NOT NULL,
    createdAt TIMESTAMP DEFAULT NOW() NOT NULL
);

-- +goose Down
DROP TABLE unique_urls;
