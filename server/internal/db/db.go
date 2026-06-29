package db

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Init(logger *slog.Logger,ctx context.Context,dbUrl string) (*pgxpool.Pool,error) {
	 // Initialize the connection pool
    pool, err := pgxpool.New(ctx, dbUrl)
    if err != nil {
       return nil,err
    }

    // Verify the connection
    if err := pool.Ping(ctx); err != nil {
        return nil,err
    }

    logger.Info("Connected to database")
	return pool, nil
}