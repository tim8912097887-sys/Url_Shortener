package urlschema

type AuthContext struct {
	UserID          string
	IsAuthenticated bool
}

type GetUrlsResponse struct {
	Urls    []GetUrlsServiceResponse `json:"urls"`
	Message string                   `json:"message"`
}

type ShortenUrlResponse struct {
	ShortUrl string `json:"shortUrl"`
	Message  string `json:"message"`
}
