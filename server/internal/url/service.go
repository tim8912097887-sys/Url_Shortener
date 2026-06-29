package url

import "context"

type UrlRepository interface {
	ShortCodeExists(ctx context.Context, shortUrl string) (bool, error)
	GetLongUrl(ctx context.Context, shortUrl string) (string, error)
	CreateShortenUrl(ctx context.Context, longUrl string, shortUrl string) (string, error)
}

type service struct {
	repository UrlRepository
}

func NewService(repository UrlRepository) *service {
	return &service{
		repository: repository,
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

	return shortUrl, nil
}

func (s *service) GetUrl(ctx context.Context, shortUrl string) (string, error) {
	var longUrl string
	var err error
	if longUrl, err = s.repository.GetLongUrl(ctx, shortUrl); err != nil {
		if err == ErrUrlNotFound {
			return "", ErrUrlNotFound
		}
		return "", err
	}

	return longUrl, nil
}