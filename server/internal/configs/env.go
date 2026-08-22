package configs

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Configs struct {
    Addr              string
    ClientOrigin      string
    DbUrl             string
    RedisUrl          string
    BaseURL           string
    GoogleClientID    string
    GoogleClientSecret string
    AccessTokenTTL    string
    RefreshTokenTTL   string
    AccessTokenSecret string
    RefreshTokenSecret string
    OAuthStateTTL     string

    CookieSecure   bool
    CookieDomain   string
    CookieSameSite string
}

func InitConfigs() (*Configs, error) {
    _ = godotenv.Load()

    return &Configs{
        Addr:              getEnv("ADDR", ":8080"),
        ClientOrigin:      getEnv("CLIENT_ORIGIN", "http://localhost:5173"),
        DbUrl:             getEnv("DB_URL", "postgres://postgres:password@db:5432/url_shortener?sslmode=disable"),
        RedisUrl:          getEnv("REDIS_URL", "redis://redis:6379"),
        BaseURL:           getEnv("BASE_URL", "http://localhost:8080"),
        GoogleClientID:    getEnv("GOOGLE_CLIENT_ID", ""),
        GoogleClientSecret:getEnv("GOOGLE_CLIENT_SECRET", ""),
        AccessTokenTTL:    getEnv("ACCESS_TOKEN_TTL", "15m"),
        RefreshTokenTTL:   getEnv("REFRESH_TOKEN_TTL", "24h"),
        AccessTokenSecret: getEnv("ACCESS_TOKEN_SECRET", ""),
        RefreshTokenSecret:getEnv("REFRESH_TOKEN_SECRET", ""),
        OAuthStateTTL:     getEnv("OAUTH_STATE_TTL", "5m"),
        
        CookieSecure:   getEnv("COOKIE_SECURE", "false") == "true",
        CookieDomain:   getEnv("COOKIE_DOMAIN", "example.com"),
        CookieSameSite: getEnv("COOKIE_SAME_SITE", "lax"),
    }, nil
}


func getEnv(key string, defaultValue string) string {

	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}

func getEnvFromInt(key string, defaultValue int) int {

	if value, ok := os.LookupEnv(key); ok {
		num, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}

		return num
	}

	return defaultValue
}