package url

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	urlerror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/url_error"
	urlschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/url_schema"
)

const (
	UnauthURLExpiry = 7 * 24 * time.Hour
	AuthURLExpiry   = 30 * 24 * time.Hour

	UnauthCacheTTL = 15 * time.Minute
	AuthCacheTTL   = 24 * time.Hour
)

type UrlRepository interface {
	GetLongUrl(ctx context.Context, shortUrl string) (string, time.Time, error)
	CreateShortenUrl(ctx context.Context, longUrl string, shortUrl string, userId *string, expiredAt time.Time) (string, error)
    GetUrlsForUser(ctx context.Context, userId string) ([]urlschema.GetUrlsRepositoryResponse, error)
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

func (s *service) ShortenUrl(ctx context.Context, url string, authContext urlschema.AuthContext) (string, error) {

	var shortUrl string
	var err error
	// Retry on short url collision,and update short url when exists
	for {

		shortUrl, err = GenerateCode(8)

		if err != nil {
			return "", err
		}

		expiredAt := time.Now().Add(urlExpiry(authContext))

		var userId *string
		if authContext.IsAuthenticated {
			userId = &authContext.UserID
		} else {
			userId = nil
		}

		if shortUrl, err = s.repository.CreateShortenUrl(ctx, url, shortUrl, userId, expiredAt); err != nil {
			log.Println("CreateShortenUrl error:", err)
			continue
		}

		break
	}
	return shortUrl, nil
}


func (s *service) GetUrl(ctx context.Context, shortUrl string, authContext urlschema.AuthContext) (string, error) {
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

	actualCacheTTL := min(cacheTTL(authContext), timeUntilExpiry)

	// Cache aside
	if err = s.cache.Set(ctx, urlCacheKey(shortUrl), longUrl, actualCacheTTL); err != nil {
		s.logger.Error("failed to set cache",slog.Any("error", err))
	}

	return longUrl, nil
}

func (s *service) GetUrlsForUser(ctx context.Context, userId string) ([]urlschema.GetUrlsServiceResponse, error) {
	urls, err := s.repository.GetUrlsForUser(ctx, userId)
    
	if err != nil {
		return nil, err
	}
	
	var response []urlschema.GetUrlsServiceResponse
	for _, url := range urls {
		response = append(response, urlschema.GetUrlsServiceResponse{
			ShortUrl:  url.ShortUrl,
			LongUrl:   url.LongUrl,
			ExpiredAt: url.ExpiredAt,
		})
	}
	
	return response, nil
}

func urlCacheKey(short string) string {
    return "url:" + short
}

func urlExpiry(auth urlschema.AuthContext) time.Duration {
	if auth.IsAuthenticated {
		return AuthURLExpiry
	}

	return UnauthURLExpiry
}

func cacheTTL(auth urlschema.AuthContext) time.Duration {
	if auth.IsAuthenticated {
		return AuthCacheTTL
	}

	return UnauthCacheTTL
}