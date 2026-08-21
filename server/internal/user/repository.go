package user

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	usererror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/user_error"
	userschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/user_schema"
)

type User struct {
	ID           string
	Username     string
	Email        string
	PasswordHash string
	TokenVersion int
	CreatedAt    time.Time
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{
		pool: pool,
	}
}

// CreateUser inserts a new user. If the email is already taken it returns
// (nil, nil) instead of an error — ON CONFLICT DO NOTHING means the insert
// is a no-op and RETURNING yields no row, which we treat as "nothing to do"
// rather than a failure. This keeps signup idempotent and race-free without
// a separate existence check.
func (r *repository) CreateUser(ctx context.Context, userInsert userschema.UserInsert) (*userschema.CreateUserRepositoryResponse, error) {
	var u userschema.CreateUserRepositoryResponse

	sql := `
		INSERT INTO users (id, username, email, password_hash)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO NOTHING
		RETURNING id, username, email;
	`
	err := r.pool.QueryRow(ctx, sql, userInsert.ID, userInsert.Username, userInsert.Email, userInsert.PasswordHash).
		Scan(&u.ID, &u.Username, &u.Email)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*userschema.GetUserbyEmailRepositoryResponse, error) {
	var u userschema.GetUserbyEmailRepositoryResponse

	sql := `SELECT id, email, password_hash, token_version FROM users WHERE email = $1;`
	err := r.pool.QueryRow(ctx, sql, email).
		Scan(&u.ID, &u.Email, &u.PasswordHash, &u.TokenVersion)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, usererror.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (r *repository) GetUserByID(ctx context.Context, id string) (*userschema.GetUserByIDRepositoryResponse, error) {
	var u userschema.GetUserByIDRepositoryResponse

	sql := `SELECT id, token_version FROM users WHERE id = $1;`
	err := r.pool.QueryRow(ctx, sql, id).
		Scan(&u.ID, &u.TokenVersion)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, usererror.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// IncrementTokenVersion bumps token_version by one, which invalidates every
// access/refresh token issued before the call (they carry the old version).
func (r *repository) IncrementTokenVersion(ctx context.Context, id string) (int, error) {
	var newVersion int

	sql := `UPDATE users SET token_version = token_version + 1 WHERE id = $1 RETURNING token_version;`
	err := r.pool.QueryRow(ctx, sql, id).Scan(&newVersion)

	if errors.Is(err, pgx.ErrNoRows) {
		return 0, usererror.ErrUserNotFound
	}
	if err != nil {
		return 0, err
	}

	return newVersion, nil
}