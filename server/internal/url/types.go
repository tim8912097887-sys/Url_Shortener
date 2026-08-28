package url

import (
	"context"
	"time"

	urlschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/url_schema"
	userschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/user_schema"
)

type UrlService interface {
	ShortenUrl(ctx context.Context, url string, authContext urlschema.AuthContext) (string, error)
	GetUrl(ctx context.Context, shortUrl string, authContext urlschema.AuthContext) (string, error)
	GetUrlsForUser(ctx context.Context, userId string) ([]urlschema.GetUrlsServiceResponse, error)
}

type UrlRepository interface {
	GetLongUrl(ctx context.Context, shortUrl string) (string, time.Time, error)
	CreateShortenUrl(ctx context.Context, longUrl string, shortUrl string, userId *string, expiredAt time.Time) (string, error)
    GetUrlsForUser(ctx context.Context, userId string) ([]urlschema.GetUrlsRepositoryResponse, error)
    UpdateUrlClicks(ctx context.Context, shortUrl string, clicks int) error
}

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (*userschema.GetUserByIDRepositoryResponse, error)
}

type UrlCache interface {
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Increment(ctx context.Context, key string) (int64, error)
	GetAndReset(ctx context.Context, key string) (int64, error)
	AddPendingClick(ctx context.Context, shortURL string) error
	GetPendingClicks(ctx context.Context) ([]string, error)
	RemovePendingClick(ctx context.Context, shortURL string) error
}