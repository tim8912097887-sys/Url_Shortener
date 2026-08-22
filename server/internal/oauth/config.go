package oauth

import (
	oauthschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/oauth_schema"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Config struct {
	GoogleClientID     string
	GoogleClientSecret string

	BaseURL string
}

type OAuth struct {
	providers map[oauthschema.Provider]*oauth2.Config
}

func New(config Config) *OAuth {
	return &OAuth{
		providers: map[oauthschema.Provider]*oauth2.Config{
			ProviderGoogle: {
				ClientID:     config.GoogleClientID,
				ClientSecret: config.GoogleClientSecret,
				RedirectURL:  config.BaseURL + "/api/v1/auth/google/callback",
				Scopes: []string{
					"openid",
					"email",
					"profile",
				},
				Endpoint: google.Endpoint,
			},
		},
	}
}

func (o *OAuth) GetProvider(provider oauthschema.Provider) (*oauth2.Config, bool) {
	config, ok := o.providers[provider]
	return config, ok
}