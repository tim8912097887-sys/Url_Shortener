package url

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	urlerror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/url_error"
	urlschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/url_schema"
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

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) &&
        pgErr.Code == "23505" &&
		pgErr.ConstraintName == "unique_short_url" {
			return "", urlerror.ErrShortURLCollision
	}

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

func (r *repository) GetUrlsForUser(ctx context.Context, userId string) ([]urlschema.GetUrlsRepositoryResponse, error) {
	var urls []urlschema.GetUrlsRepositoryResponse
	sql := `SELECT short_url, long_url, clicks, expired_at FROM urls_map WHERE user_id = $1 AND expired_at > NOW();`
	row, err := r.pool.Query(ctx, sql, userId)
	
	if err != nil {
		return nil, err
	}
	
	for row.Next() {
		var url urlschema.GetUrlsRepositoryResponse
		if err := row.Scan(&url.ShortUrl, &url.LongUrl, &url.Clicks, &url.ExpiredAt); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	return urls, nil
}

func (r *repository) UpdateUrlClicks(ctx context.Context, shortUrl string, clicks int) error {
	sql := `UPDATE urls_map SET clicks = clicks + $1 WHERE short_url = $2;`
	_, err := r.pool.Exec(ctx, sql, clicks, shortUrl)
	return err
}