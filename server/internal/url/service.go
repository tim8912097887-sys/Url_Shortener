package url

import "errors"

var (
	ErrUrlNotFound = errors.New("url not found")
)

type service struct{}

func NewService() *service {
	return &service{}
}

var urlMap = map[string]string{}

func (s *service) ShortenUrl(url string) (string, error) {
	shortUrl, err := GenerateCode(8)

	if err != nil {
		return "", err
	}

	urlMap[shortUrl] = url

	return shortUrl, nil
}

func (s *service) GetUrl(shortUrl string) (string, error) {

	if url, ok := urlMap[shortUrl]; ok {
		return url, nil
	}

	return "", ErrUrlNotFound
}