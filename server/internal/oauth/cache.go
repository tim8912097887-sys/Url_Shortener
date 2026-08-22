package oauth

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	oautherror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/oauth_error"
)

type Cache struct {
	cache *redis.Client
}

func NewCache(cache *redis.Client) *Cache {
	return &Cache{
		cache: cache,
	}
}

func (c *Cache) Save(
	ctx context.Context,
	state string,
) error {
	return c.cache.Set(
		ctx,
		oAuthStateKey(state),
		"1",
		1*time.Minute,
	).Err()
}

func (c *Cache) Consume(
	ctx context.Context,
	state string,
) error {
	key := "oauth:state:" + state

	value, err := c.cache.GetDel(ctx, key).Result()
	
	if errors.Is(err, redis.Nil) {
        return oautherror.ErrInvalidState
    }
	
	if err != nil {
		return err
	}

	if value == "" {
		return oautherror.ErrInvalidState
	}

	return nil
}