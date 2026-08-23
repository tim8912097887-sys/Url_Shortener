package url

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	urlerror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/url_error"
)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{
		pool: pool,
	}
}

func (r *repository) GetLongUrl(ctx context.Context,shortUrl string) (string, time.Time, error) {

	var longUrl string
	var expiredAt time.Time

	sql := `SELECT long_url,expired_at FROM urls_map
	        WHERE short_url = $1
			AND expired_at > NOW();
			`
	err := r.pool.QueryRow(ctx, sql, shortUrl).Scan(&longUrl,&expiredAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", time.Time{}, urlerror.ErrUrlNotFound
	}
	if err != nil {
		return "", time.Time{}, err
	}

	return longUrl, expiredAt, nil
}

func (r *repository) CreateShortenUrl(
	ctx context.Context,
	longURL string,
	shortURL string,
	userID *string,
	expiredAt time.Time,
) (string, error) {
	// Update short url if exists
	sql := `
		INSERT INTO urls_map (
			user_id,
			short_url,
			long_url,
			expired_at
		)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, long_url)
		DO NOTHING
		RETURNING short_url
	`

	var resultShortURL string

	err := r.pool.QueryRow(
		ctx,
		sql,
		userID,
		shortURL,
		longURL,
		expiredAt,
	).Scan(&resultShortURL)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", err
	}
    // Query short url when already exists
	if errors.Is(err, pgx.ErrNoRows) {
		sql = `
		    SELECT short_url
		    FROM urls_map
		    WHERE long_url = $1 AND user_id = $2
		    AND expired_at > NOW()
		`
		err = r.pool.QueryRow(ctx, sql, longURL, userID).Scan(&resultShortURL)
		if err != nil {
			return "", err
		}
	}

	return resultShortURL, nil
}