package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type cleanupService struct {
	pool *pgxpool.Pool
	logger *slog.Logger
}

func NewCleanupService(pool *pgxpool.Pool, logger *slog.Logger) *cleanupService {
	return &cleanupService{
		pool: pool,
		logger: logger,
	}
}

func (c *cleanupService) Start(ctx context.Context) error {
	ticker := time.NewTicker(time.Hour * 24)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Expected exit
			return nil
		case <-ticker.C:
			// Cleanup timeout prevent infinite hang
			queryContext, cancel := context.WithTimeout(ctx,30*time.Second)
			err := cleanupExpiredUrls(queryContext, c.pool)
		    cancel()
			if err != nil {
				c.logger.Error("failed to cleanup expired urls",slog.Any("error", err))
			}
		}
	}
}

func cleanupExpiredUrls(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}

	// Rollback is safe even if Commit succeeds.
	defer tx.Rollback(ctx)

	sql := `
	WITH expired AS (
		SELECT id
		FROM urls_map
		WHERE expired_at < NOW()
		LIMIT 100
	)
	DELETE FROM urls_map
	WHERE id IN (SELECT id FROM expired);
	`
	if _, err := tx.Exec(ctx, sql); err != nil {
		return err
	}

	sql = `DELETE FROM unique_urls WHERE id NOT IN (SELECT short_url_id FROM urls_map);`
	if _, err := tx.Exec(ctx, sql); err != nil {
		return err
	}

	return tx.Commit(ctx)
}