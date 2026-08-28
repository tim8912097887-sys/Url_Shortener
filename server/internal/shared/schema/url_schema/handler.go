package urlschema

type AuthContext struct {
	UserID          string
	TokenVersion    int
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
