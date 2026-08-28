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

func (c *Cache) Increment(
	ctx context.Context,
	key string,
) (int64, error) {
	pipe := c.client.TxPipeline()

	incr := pipe.Incr(ctx, key)
    pipe.ExpireNX(ctx, key, ClickCacheTTL)   

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, err
	}

	return incr.Val(), nil
}

func (c *Cache) GetAndReset(
    ctx context.Context,
    key string,
) (int64, error) {
    value, err := c.client.GetDel(ctx, key).Int64()
    if err == redis.Nil {
        return 0, nil
    }
    if err != nil {
        return 0, err
    }

    return value, nil
}

func (c *Cache) AddPendingClick(
	ctx context.Context,
	shortURL string,
) error {
	return c.client.SAdd(
		ctx,
		pendingClicksKey,
		shortURL,
	).Err()
}

func (c *Cache) GetPendingClicks(
	ctx context.Context,
) ([]string, error) {
	return c.client.SMembers(
		ctx,
		pendingClicksKey,
	).Result()
}

func (c *Cache) RemovePendingClick(
	ctx context.Context,
	shortURL string,
) error {
	return c.client.SRem(
		ctx,
		pendingClicksKey,
		shortURL,
	).Err()
}
