package url

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache implements the UrlCache interface
type Cache struct {
    client *redis.Client
}

// NewCache creates a new Cache instance
func NewCache(client *redis.Client) *Cache {
    return &Cache{
        client: client,
    }
}

// Set stores a key-value pair with expiration
func (c *Cache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
    return c.client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value by key
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
    return c.client.Get(ctx, key).Result()
}
