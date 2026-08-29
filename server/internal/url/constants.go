package url

import "time"

const (
	UnauthURLExpiry = 7 * 24 * time.Hour
	AuthURLExpiry   = 30 * 24 * time.Hour

	UnauthCacheTTL = 15 * time.Minute
	AuthCacheTTL   = 24 * time.Hour

	ClickCacheTTL = 2 * time.Hour

	pendingClicksKey = "url:clicks:pending"

	WorkerInterval = 1 * time.Minute
	WorkerTimeout  = 10 * time.Second
	UrlsMaxLimit   = 10
	UrlsMinLimit   = 1
)