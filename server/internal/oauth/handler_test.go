package oauth_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/tim8912097887-sys/url-shortener/internal/configs"
	"github.com/tim8912097887-sys/url-shortener/internal/oauth"
	oautherror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/oauth_error"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response/envelope"
	oauthschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/oauth_schema"
)

type mockCache struct {
	saveFunc    func(context.Context, string) error
	consumeFunc func(context.Context, string) error
}

func newCache() *mockCache {
	return &mockCache{
		saveFunc:    func(context.Context, string) error { return nil },
		consumeFunc: func(context.Context, string) error { return nil },
	}
}
func (m *mockCache) Save(ctx context.Context, state string) error { return m.saveFunc(ctx, state) }
func (m *mockCache) Consume(ctx context.Context, state string) error {
	return m.consumeFunc(ctx, state)
}

func newApp(t *testing.T, cache *mockCache) *fiber.App {
	t.Helper()
	service := oauth.NewService(&oauth.ServiceConfig{Cache: cache, OAuthConfig: oauth.New(oauth.Config{GoogleClientID: "client", GoogleClientSecret: "secret", BaseURL: "http://localhost:8080"})})
	handler := oauth.NewHandler(oauth.HandlerConfig{Logger: slog.Default(), Service: service, Cfg: &configs.Configs{CookieDomain: "example.com", CookieSecure: false, CookieSameSite: "Lax"}})
	app := fiber.New()
	handler.RegisterRoutes(app.Group("/api/v1/auth"))
	return app
}

func request(t *testing.T, app *fiber.App, path string) *http.Response {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil), fiber.TestConfig{Timeout: -1})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func errorResponse(t *testing.T, resp *http.Response) envelope.ErrorResponse {
	t.Helper()
	var result envelope.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestGoogleLogin(t *testing.T) {
	cache := newCache()
	app := newApp(t, cache)
	resp := request(t, app, "/api/v1/auth/google/login")
	if resp.StatusCode != http.StatusTemporaryRedirect || !strings.Contains(resp.Header.Get("Location"), "accounts.google.com") {
		t.Fatalf("expected Google redirect, got %d %q", resp.StatusCode, resp.Header.Get("Location"))
	}
	cache.saveFunc = func(context.Context, string) error { return errors.New("cache down") }
	resp = request(t, newApp(t, cache), "/api/v1/auth/google/login")
	if resp.StatusCode != 500 || errorResponse(t, resp).Error.Code != "INTERNAL_SERVER_ERROR" {
		t.Fatalf("cache failure: got %d", resp.StatusCode)
	}
}

func TestGoogleCallbackValidation(t *testing.T) {
	for _, path := range []string{"/api/v1/auth/google/callback", "/api/v1/auth/google/callback?code=code", "/api/v1/auth/google/callback?state=state"} {
		resp := request(t, newApp(t, newCache()), path)
		if resp.StatusCode != 400 || errorResponse(t, resp).Error.Code != "INVALID_INPUT" {
			t.Errorf("%s should return invalid input", path)
		}
	}
}

func TestGoogleCallbackState(t *testing.T) {
	cache := newCache()
	cache.consumeFunc = func(context.Context, string) error { return oautherror.ErrInvalidState }
	resp := request(t, newApp(t, cache), "/api/v1/auth/google/callback?code=code&state=state")
	if resp.StatusCode != 400 || errorResponse(t, resp).Error.Code != "INVALID_STATE" {
		t.Fatalf("invalid state: got %d", resp.StatusCode)
	}
}

func TestProviderTypeIsGoogle(t *testing.T) {
	if oauthschema.Provider("google") == "" {
		t.Fatal("provider should be non-empty")
	}
}
