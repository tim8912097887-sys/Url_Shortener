package oauth

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	usererror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/user_error"
	oauthschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/oauth_schema"
	userschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/user_schema"
)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{
		pool: pool,
	}
}

func (r *repository) GetUserByOAuthAccount(
	ctx context.Context,
	provider oauthschema.Provider,
	providerAccountID string,
) (*oauthschema.GetUserByOAuthAccountRepositoryResponse, error) {
	var user oauthschema.GetUserByOAuthAccountRepositoryResponse

	sql := `
		SELECT
			u.id,
			u.token_version
		FROM oauth_accounts oa
		INNER JOIN users u
			ON u.id = oa.user_id
		WHERE
			oa.provider = $1
			AND oa.provider_account_id = $2;
	`

	err := r.pool.QueryRow(
		ctx,
		sql,
		provider,
		providerAccountID,
	).Scan(
		&user.UserID,
		&user.TokenVersion,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, usererror.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Const for PostgreSQL unique constraint violation error code
const pgUniqueViolation = "23505"

func (r *repository) CreateOAuthAccount(
	ctx context.Context,
	userInsert userschema.UserInsert,
	oauthInsert oauthschema.OAuthAccountInsert,
) (*oauthschema.CreateOAuthAccountRepositoryResponse, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var targetUserID string
	var targetTokenVersion int

	userSQL := `
		INSERT INTO users (
			id,
			username,
			email,
			password_hash
		)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO NOTHING
		RETURNING
			id,
			token_version;
	`

	err = tx.QueryRow(
		ctx,
		userSQL,
		userInsert.ID,
		userInsert.Username,
		userInsert.Email,
		nil,
	).Scan(&targetUserID, &targetTokenVersion)

	if err != nil {
	    if errors.Is(err, pgx.ErrNoRows) {
			getUserSQL := `
				SELECT id, token_version 
				FROM users 
				WHERE email = $1;
			`
			err = tx.QueryRow(ctx, getUserSQL, userInsert.Email).Scan(&targetUserID, &targetTokenVersion)
			if err != nil {
				return nil, fmt.Errorf("get existing user by email: %w", err)
			}
		} else {
            return nil, fmt.Errorf("create user: %w", err)
		}
	}

	var result oauthschema.CreateOAuthAccountRepositoryResponse
	oauthSQL := `
		INSERT INTO oauth_accounts (
			id,
			user_id,
			provider,
			provider_account_id,
			provider_email
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			user_id,
			provider,
			provider_account_id;
	`

	err = tx.QueryRow(
		ctx,
		oauthSQL,
		oauthInsert.ID,
		targetUserID,
		oauthInsert.Provider,
		oauthInsert.ProviderAccountID,
		oauthInsert.ProviderEmail,
	).Scan(
		&result.ID,
		&result.UserID,
		&result.Provider,
		&result.ProviderAccountID,
	)

	if err != nil {
		return nil, fmt.Errorf("create oauth account: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	// Attach TokenVersion to repository response if required for token generation downstream
	result.TokenVersion = targetTokenVersion

	return &result, nil
}

