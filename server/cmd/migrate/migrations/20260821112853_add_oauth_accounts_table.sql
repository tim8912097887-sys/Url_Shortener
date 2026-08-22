-- +goose Up
CREATE TABLE oauth_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,

    provider VARCHAR(50) NOT NULL,

    provider_account_id TEXT NOT NULL,

    provider_email TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT oauth_accounts_provider_account_unique
        UNIQUE (provider, provider_account_id)
);

CREATE INDEX idx_oauth_accounts_user_id
ON oauth_accounts(user_id);

-- +goose Down
DROP TABLE oauth_accounts;
