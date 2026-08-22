package url

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	urlerror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/url_error"
)

const (
	CacheTTL = 24 * time.Hour * 7
)

type UrlRepository interface {
	ShortCodeExists(ctx context.Context, shortUrl string) (bool, error)
	GetLongUrl(ctx context.Context, shortUrl string) (string, time.Time, error)
	CreateShortenUrl(ctx context.Context, longUrl string, shortUrl string) (string, error)
}

type UrlCache interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type ServiceConfig struct {
	Repository UrlRepository
	Cache      UrlCache
	Logger     *slog.Logger
}

type service struct {
	repository UrlRepository
	cache      UrlCache
	logger     *slog.Logger
}

func NewService(serviceConfig *ServiceConfig) *service {
	return &service{
		repository: serviceConfig.Repository,
		cache:      serviceConfig.Cache,
		logger:     serviceConfig.Logger,
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
	if err = s.cache.Set(ctx, urlCacheKey(shortUrl), url, CacheTTL); err != nil {
	   s.logger.Error("failed to set cache",slog.Any("error", err))
	}

	return shortUrl, nil
}

func (s *service) GetUrl(ctx context.Context, shortUrl string) (string, error) {
	var longUrl string
	var err error
	var remainingTime time.Time

	// Read from cache
	if longUrl, err = s.cache.Get(ctx,urlCacheKey(shortUrl)); (err != nil && err != redis.Nil) {
		s.logger.Error("failed to get cache",slog.Any("error", err))
	}

	// Cache hit
	if err == nil && longUrl != "" {
		s.logger.Info("cache hit",slog.String("shortUrl", shortUrl))
		return longUrl, nil
	}

	if longUrl, remainingTime, err = s.repository.GetLongUrl(ctx, shortUrl); err != nil {
		if err == urlerror.ErrUrlNotFound {
			return "", urlerror.ErrUrlNotFound
		}
		return "", err
	}

	timeUntilExpiry := time.Until(remainingTime)
	
	if timeUntilExpiry < 0 {
		return "", urlerror.ErrUrlNotFound
	}

	cacheTTL := min(CacheTTL, timeUntilExpiry)

	// Cache aside
	if err = s.cache.Set(ctx, urlCacheKey(shortUrl), longUrl, cacheTTL); err != nil {
		s.logger.Error("failed to set cache",slog.Any("error", err))
	}

	return longUrl, nil
}

func urlCacheKey(short string) string {
    return "url:" + short
}