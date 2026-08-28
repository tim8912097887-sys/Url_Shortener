package url

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	urlerror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/url_error"
	urlschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/url_schema"
)

type ServiceConfig struct {
	UrlRepository UrlRepository
	UserRepository UserRepository
	Cache      UrlCache
	Logger     *slog.Logger
}

type service struct {
	urlRepository UrlRepository
	userRepository UserRepository
	cache      UrlCache
	logger     *slog.Logger
}

func NewService(serviceConfig *ServiceConfig) *service {
	return &service{
		urlRepository: serviceConfig.UrlRepository,
		userRepository: serviceConfig.UserRepository,
		cache:      serviceConfig.Cache,
		logger:     serviceConfig.Logger,
	}
}

func (s *service) ShortenUrl(ctx context.Context, url string, authContext urlschema.AuthContext) (string, error) {

	var shortUrl string
	var err error
	var userId *string
	isUserValid := s.checkUserValid(ctx, authContext.UserID, authContext.TokenVersion)
	if isUserValid {
		userId = &authContext.UserID
	} else {
		userId = nil
	}
	// Retry on short url collision,and update short url when exists
	for {

		shortUrl, err = GenerateCode(8)

		if err != nil {
			return "", err
		}

		expiredAt := time.Now().Add(urlExpiry(authContext))

		if shortUrl, err = s.urlRepository.CreateShortenUrl(ctx, url, shortUrl, userId, expiredAt); err != nil {
			// Only retry on short url collision
			if errors.Is(err, urlerror.ErrShortURLCollision) {
				continue
			}
			return "", err
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
	if longUrl, err = s.cache.Get(ctx,UrlCacheKey(shortUrl)); (err != nil && err != redis.Nil) {
		s.logger.Error("failed to get cache",slog.Any("error", err))
	}

	// Cache hit
	if err == nil && longUrl != "" {
		s.logger.Info("cache hit",slog.String("shortUrl", shortUrl))
		// Increment cache hit counter
		if _, err = s.cache.Increment(ctx, UrlClickKey(shortUrl)); err != nil {
			s.logger.Error("failed to increment cache hit counter",slog.Any("error", err))
		}	

		if err = s.cache.AddPendingClick(ctx, shortUrl); err != nil {
			s.logger.Error("failed to add pending click",slog.Any("error", err))
		}
		
		return longUrl, nil
	}

	if longUrl, remainingTime, err = s.urlRepository.GetLongUrl(ctx, shortUrl); err != nil {
		if err == urlerror.ErrUrlNotFound {
			return "", urlerror.ErrUrlNotFound
		}
		return "", err
	}

	timeUntilExpiry := time.Until(remainingTime)
	
	if timeUntilExpiry < 0 {
		return "", urlerror.ErrUrlNotFound
	}

	// Check if user is valid
	isUserValid := s.checkUserValid(ctx, authContext.UserID, authContext.TokenVersion)
	if !isUserValid {
		authContext.IsAuthenticated = false
	}

	actualCacheTTL := min(cacheTTL(authContext), timeUntilExpiry)

	// Cache aside
	if err = s.cache.Set(ctx, UrlCacheKey(shortUrl), longUrl, actualCacheTTL); err != nil {
		s.logger.Error("failed to set cache",slog.Any("error", err))
	}

	return longUrl, nil
}

func (s *service) GetUrlsForUser(ctx context.Context, userId string) ([]urlschema.GetUrlsServiceResponse, error) {
	urls, err := s.urlRepository.GetUrlsForUser(ctx, userId)
    
	if err != nil {
		return nil, err
	}
	
	var response []urlschema.GetUrlsServiceResponse
	for _, url := range urls {
		response = append(response, urlschema.GetUrlsServiceResponse{
			ShortUrl:  url.ShortUrl,
			LongUrl:   url.LongUrl,
			Clicks:    url.Clicks,
			ExpiredAt: url.ExpiredAt,
		})
	}
	
	return response, nil
}

func (s *service) checkUserValid(ctx context.Context, userId string, tokenVersion int) bool {
	existUser, err := s.userRepository.GetUserByID(ctx, userId)
    if err != nil {
		return false
	}

	return existUser.TokenVersion == tokenVersion
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