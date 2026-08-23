package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/tim8912097887-sys/url-shortener/cmd/api"
	"github.com/tim8912097887-sys/url-shortener/internal/cache"
	"github.com/tim8912097887-sys/url-shortener/internal/configs"
	jwttoken "github.com/tim8912097887-sys/url-shortener/internal/shared/util/jwt_token"
)

type App struct {
	Handler http.Handler
	Pool    *pgxpool.Pool
	Cache   *redis.Client
	Tokens  *jwttoken.TokenManager
}

func NewApp(t *testing.T) *App {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	databaseURL := os.Getenv("DB_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:password@localhost:5432/url_shortener?sslmode=disable"
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Skipf("integration database is unavailable: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("integration database is unavailable: %v", err)
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://:password@localhost:6379/0"
	}
	redisClient := cache.NewRedisClient(slog.Default(), redisURL)
	if redisClient == nil {
		pool.Close()
		t.Skip("integration Redis URL is invalid")
	}
	if _, err := cache.CacheInit(ctx, slog.Default(), redisClient); err != nil {
		redisClient.Close()
		pool.Close()
		t.Skipf("integration Redis is unavailable: %v", err)
	}

	cleanup(t, pool)
	cleanupCache(t, redisClient)
	t.Cleanup(func() {
		cleanup(t, pool)
		cleanupCache(t, redisClient)
		redisClient.Close()
		pool.Close()
	})

	tokens := jwttoken.NewTokenManager("integration-access-secret", "integration-refresh-secret")
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	cfg := &configs.Configs{
		ClientOrigin:       "http://localhost:5173",
		AccessTokenSecret:  "integration-access-secret",
		RefreshTokenSecret: "integration-refresh-secret",
		CookieDomain:       "localhost",
		CookieSecure:       false,
		CookieSameSite:     "lax",
	}
	handler := api.NewApi(api.ApiConfig{Logger: logger, Cfg: cfg}).Mount(pool, redisClient)

	return &App{Handler: handler, Pool: pool, Cache: redisClient, Tokens: tokens}
}

func Request(t *testing.T, app *App, method, path string, payload any, auth, refreshCookie string) *http.Response {
	t.Helper()

	var body *bytes.Reader
	if payload == nil {
		body = bytes.NewReader(nil)
	} else {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatal(err)
		}
		body = bytes.NewReader(encoded)
	}

	req := httptest.NewRequest(method, path, body)
	if payload != nil {
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	}
	if auth != "" {
		req.Header.Set(fiber.HeaderAuthorization, auth)
	}
	if refreshCookie != "" {
		req.Header.Set(fiber.HeaderCookie, "refresh_token="+refreshCookie)
	}

	recorder := httptest.NewRecorder()
	app.Handler.ServeHTTP(recorder, req)
	response := recorder.Result()

	return response
}

func Cleanup(t *testing.T, pool *pgxpool.Pool, redisClient *redis.Client) {
	t.Helper()
	cleanup(t, pool)
	cleanupCache(t, redisClient)
}

func cleanup(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, "TRUNCATE TABLE urls_map, oauth_accounts, users RESTART IDENTITY CASCADE"); err != nil {
		t.Fatalf("clean integration data: %v", err)
	}
}

func cleanupCache(t *testing.T, redisClient *redis.Client) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("clean integration cache: %v", err)
	}
}
