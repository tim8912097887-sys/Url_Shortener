package url

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *repository {
	return &repository{
		pool: pool,
	}
}

func (r *repository) ShortCodeExists(ctx context.Context,shortUrl string) (bool,error) {

	var longUrl string
	sql := `SELECT url FROM unique_urls WHERE url = $1;`
	err := r.pool.QueryRow(ctx, sql, shortUrl).Scan(&longUrl)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false,err
	}

	return true, nil
}

func (r *repository) GetLongUrl(ctx context.Context,shortUrl string) (string, error) {

	var longUrl string
	sql := `SELECT urls_map.long_url FROM urls_map
	        JOIN unique_urls ON unique_urls.id = urls_map.short_url_id
	        WHERE unique_urls.url = $1;
			`
	err := r.pool.QueryRow(ctx, sql, shortUrl).Scan(&longUrl)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrUrlNotFound
	}
	if err != nil {
		return "", err
	}

	return longUrl, nil
}

func (r *repository) CreateShortenUrl(
	ctx context.Context,
	longUrl string,
	shortUrl string,
) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}

	// Rollback is safe even if Commit succeeds.
	defer tx.Rollback(ctx)

	var id string

	sql := `INSERT INTO unique_urls (url) VALUES ($1) RETURNING id;`
	if err := tx.QueryRow(ctx, sql, shortUrl).Scan(&id); err != nil {
		return "", err
	}

	sql = `INSERT INTO urls_map (short_url_id, long_url) VALUES ($1, $2);`
	if _, err := tx.Exec(ctx, sql, id, longUrl); err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return shortUrl, nil
}