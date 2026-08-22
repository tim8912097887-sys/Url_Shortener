package oauthschema

type Provider string

type OAuthUser struct {
	Provider   Provider
	ProviderID string
	Email      string
	Name       string
	AvatarURL  string
}

type GoogleUserResponse struct {
	ID            string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
