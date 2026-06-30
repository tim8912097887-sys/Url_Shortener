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

func NewRedisClient(addr string, password string, db int) *redis.Client {
    // Configure the client with production-ready settings
    rdb := redis.NewClient(&redis.Options{
        // Server address and credentials
        Addr:     addr,
        Password: password,
        DB:       db,

        // Connection pool settings
        PoolSize:     10,              // Maximum number of connections
        MinIdleConns: 5,               // Minimum idle connections to maintain
        PoolTimeout:  30 * time.Second, // Time to wait for a connection from pool

        // Timeouts for various operations
        DialTimeout:  5 * time.Second,  // Timeout for establishing new connections
        ReadTimeout:  3 * time.Second,  // Timeout for read operations
        WriteTimeout: 3 * time.Second,  // Timeout for write operations

        // Retry configuration
        MaxRetries:      3,                      // Maximum number of retries
        MinRetryBackoff: 8 * time.Millisecond,   // Minimum backoff between retries
        MaxRetryBackoff: 512 * time.Millisecond, // Maximum backoff between retries

    })

    return rdb
}