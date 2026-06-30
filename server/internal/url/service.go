package url

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	CacheTTL = 24 * time.Hour * 7
)

type UrlRepository interface {
	ShortCodeExists(ctx context.Context, shortUrl string) (bool, error)
	GetLongUrl(ctx context.Context, shortUrl string) (string, time.Time, error)
	CreateShortenUrl(ctx context.Context, longUrl string, shortUrl string) (string, error)
}

type UrlRedis interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}

type service struct {
	repository UrlRepository
	cache      UrlRedis
	logger     *slog.Logger
}

func NewService(repository UrlRepository,cache UrlRedis, logger *slog.Logger) *service {
	return &service{
		repository: repository,
		cache:      cache,
		logger:     logger,
	}
}

func (s *service) ShortenUrl(ctx context.Context, url string) (string, error) {
	var shortUrl string
	var err error
	for {

		shortUrl, err = GenerateCode(8)

		if err != nil {
			return "", err
		}

		if existence, err := s.repository.ShortCodeExists(ctx, shortUrl); (err != nil || existence) {
			if err != nil {
				return "", err
			}
			if existence {
				continue
			}
		}

		break
	}

	if _, err = s.repository.CreateShortenUrl(ctx, url, shortUrl); err != nil {
		return "", err
	}

	// Write through cache
	if err = s.cache.Set(ctx, urlCacheKey(shortUrl), url, CacheTTL).Err(); err != nil {
	   s.logger.Error("failed to set cache",slog.Any("error", err))
	}

	return shortUrl, nil
}

func (s *service) GetUrl(ctx context.Context, shortUrl string) (string, error) {
	var longUrl string
	var err error
	var remainingTime time.Time

	// Read from cache
	if longUrl, err = s.cache.Get(ctx,urlCacheKey(shortUrl)).Result(); (err != nil && err != redis.Nil) {
		s.logger.Error("failed to get cache",slog.Any("error", err))
	}

	// Cache hit
	if err == nil && longUrl != "" {
		s.logger.Info("cache hit",slog.String("shortUrl", shortUrl))
		return longUrl, nil
	}

	if longUrl, remainingTime, err = s.repository.GetLongUrl(ctx, shortUrl); err != nil {
		if err == ErrUrlNotFound {
			return "", ErrUrlNotFound
		}
		return "", err
	}

	timeUntilExpiry := time.Until(remainingTime)
	
	if timeUntilExpiry < 0 {
		return "", ErrUrlNotFound
	}

	cacheTTL := min(CacheTTL, timeUntilExpiry)

	// Cache aside
	if err = s.cache.Set(ctx, urlCacheKey(shortUrl), longUrl, cacheTTL).Err(); err != nil {
		s.logger.Error("failed to set cache",slog.Any("error", err))
	}

	return longUrl, nil
}

func urlCacheKey(short string) string {
    return "url:" + short
}