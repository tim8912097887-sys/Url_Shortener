package cache

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

func CacheInit(ctx context.Context,logger *slog.Logger,rdb *redis.Client) (*redis.Client,error) {
	
	// Test the connection using Ping
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	logger.Info("Connected to Redis",slog.String("pong", pong))

	return rdb, nil
}

func NewRedisClient(logger *slog.Logger,redisURL string) *redis.Client {
    // Configure the client with production-ready settings
    opts, err := redis.ParseURL(redisURL)
    if err != nil {
        logger.Error("failed to parse redis url",slog.Any("error", err))
        return nil
    }

    // Apply custom connection pool & timeout overrides
    opts.PoolSize = 10
    opts.MinIdleConns = 5
    opts.PoolTimeout = 30 * time.Second
    opts.DialTimeout = 5 * time.Second
    opts.ReadTimeout = 3 * time.Second
    opts.WriteTimeout = 3 * time.Second
    opts.MaxRetries = 3
    opts.MinRetryBackoff = 8 * time.Millisecond
    opts.MaxRetryBackoff = 512 * time.Millisecond

    rdb := redis.NewClient(opts)

    return rdb
}