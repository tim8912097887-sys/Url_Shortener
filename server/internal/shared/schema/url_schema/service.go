package urlschema

import "time"

type GetUrlsServiceResponse struct {
	ShortUrl  string    `json:"short_url"`
	LongUrl   string    `json:"long_url"`
	Clicks    int       `json:"clicks"`
	ExpiredAt time.Time `json:"expired_at"`
}