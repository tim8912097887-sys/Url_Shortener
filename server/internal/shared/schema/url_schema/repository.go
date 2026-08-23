package urlschema

import "time"

type GetUrlsRepositoryResponse struct {
	ShortUrl  string    `json:"short_url"`
	LongUrl   string    `json:"long_url"`
	ExpiredAt time.Time `json:"expired_at"`
}