package urltest

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/tim8912097887-sys/url-shortener/internal/shared/response/envelope"
	"github.com/tim8912097887-sys/url-shortener/internal/url"
	"github.com/tim8912097887-sys/url-shortener/tests/helper"
)

func createURL(t *testing.T, app *helper.App, longURL, accessToken string) *http.Response {
	t.Helper()
	return helper.Request(t, app, http.MethodPost, "/api/v1/urls", url.CreateUrlSchema{Url: longURL}, bearer(accessToken), "")
}

func bearer(accessToken string) string {
	if accessToken == "" {
		return ""
	}
	return "Bearer " + accessToken
}

func shortURLFromResponse(t *testing.T, response *http.Response) string {
	t.Helper()
	payload := helper.DecodeResponse[envelope.SuccessResponse](t, response)
	data, ok := payload.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected response data object, got %T", payload.Data)
	}
	shortURL, ok := data["shortUrl"].(string)
	if !ok || len(shortURL) != 8 {
		t.Fatalf("expected eight-character short URL, got %#v", data["shortUrl"])
	}
	return shortURL
}

func TestShortenURL(t *testing.T) {
	app := helper.NewApp(t)
	tests := []struct {
		name       string
		payload    any
		statusCode int
		errorCode  string
	}{
		{"empty body", nil, http.StatusBadRequest, "INVALID_INPUT"},
		{"malformed URL", map[string]string{"url": "not a URL"}, http.StatusBadRequest, "INVALID_INPUT"},
		{"missing URL", map[string]string{}, http.StatusBadRequest, "INVALID_INPUT"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := helper.Request(t, app, http.MethodPost, "/api/v1/urls", test.payload, "", "")
			if response.StatusCode != test.statusCode {
				t.Fatalf("expected status %d, got %d", test.statusCode, response.StatusCode)
			}
			if helper.ErrorCode(t, response) != test.errorCode {
				t.Fatalf("expected error %s", test.errorCode)
			}
		})
	}

	t.Run("creates anonymous URL", func(t *testing.T) {

		response := createURL(t, app, "https://example.com/anonymous", "")
	
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
		}
		shortURL := shortURLFromResponse(t, response)
		var userID *string
		var expiredAt time.Time
		if err := app.Pool.QueryRow(context.Background(), "SELECT user_id, expired_at FROM urls_map WHERE short_url = $1", shortURL).Scan(&userID, &expiredAt); err != nil {
			t.Fatal(err)
		}
		if userID != nil {
			t.Fatalf("expected anonymous URL to have no user")
		}
		if time.Until(expiredAt) < 6*24*time.Hour {
			t.Fatalf("expected anonymous URL to expire in about 7 days, got expiry %s", expiredAt)
		}
		helper.Cleanup(t, app.Pool, app.Cache)
	})

	t.Run("creates authenticated URL with long expiry", func(t *testing.T) {
		
		if response := helper.SignupUser(t, app, "alice", "alice@example.com", "password1"); response.StatusCode != http.StatusOK {
			t.Fatalf("signup: got %d", response.StatusCode)
		}
	
		accessToken, _, response := helper.LoginUser(t, app, "alice@example.com", "password1")
		if response.StatusCode != http.StatusOK {
			t.Fatalf("login: got %d", response.StatusCode)
		}
		shortURL := shortURLFromResponse(t, createURL(t, app, "https://example.com/authenticated", accessToken))
		var userID string
		var expiredAt time.Time
		if err := app.Pool.QueryRow(context.Background(), "SELECT user_id, expired_at FROM urls_map WHERE short_url = $1", shortURL).Scan(&userID, &expiredAt); err != nil {
			t.Fatal(err)
		}
		if userID == "" || time.Until(expiredAt) < 29*24*time.Hour {
			t.Fatalf("expected authenticated URL to belong to user and expire in about 30 days, got user %q and expiry %s", userID, expiredAt)
		}
		helper.Cleanup(t, app.Pool, app.Cache)
	})

	t.Run("returns existing authenticated URL for duplicate long URL", func(t *testing.T) {
		
		if response := helper.SignupUser(t, app, "alice", "alice@example.com", "password1"); response.StatusCode != http.StatusOK {
			t.Fatalf("signup: got %d", response.StatusCode)
		}
		accessToken, _, response := helper.LoginUser(t, app, "alice@example.com", "password1")
		if response.StatusCode != http.StatusOK {
			t.Fatalf("login: got %d", response.StatusCode)
		}
		first := shortURLFromResponse(t, createURL(t, app, "https://example.com/duplicate", accessToken))
		second := shortURLFromResponse(t, createURL(t, app, "https://example.com/duplicate", accessToken))
		if first != second {
			t.Fatalf("expected duplicate URL to return %q, got %q", first, second)
		}
		var count int
		if err := app.Pool.QueryRow(context.Background(), "SELECT count(*) FROM urls_map WHERE long_url = $1", "https://example.com/duplicate").Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("expected one stored URL, got %d", count)
		}
		helper.Cleanup(t, app.Pool, app.Cache)
	})
}

func TestGetURL(t *testing.T) {
	app := helper.NewApp(t)
	t.Run("redirects to the original URL", func(t *testing.T) {
		
		shortURL := shortURLFromResponse(t, createURL(t, app, "https://example.com/target", ""))
		response := helper.Request(t, app, http.MethodGet, "/api/v1/urls/"+shortURL, nil, "", "")
		if response.StatusCode != http.StatusTemporaryRedirect {
			t.Fatalf("expected status %d, got %d", http.StatusTemporaryRedirect, response.StatusCode)
		}
		if response.Header.Get("Location") != "https://example.com/target" {
			t.Fatalf("expected redirect location, got %q", response.Header.Get("Location"))
		}
		response.Body.Close()
		helper.Cleanup(t, app.Pool, app.Cache)
	})

	t.Run("serves a cache hit without the database row", func(t *testing.T) {
		
		if err := app.Cache.Set(context.Background(), "url:cached1x", "https://example.com/cached", time.Minute).Err(); err != nil {
			t.Fatal(err)
		}
		response := helper.Request(t, app, http.MethodGet, "/api/v1/urls/cached1x", nil, "", "")
		if response.StatusCode != http.StatusTemporaryRedirect || response.Header.Get("Location") != "https://example.com/cached" {
			t.Fatalf("expected cached redirect, got status %d and location %q", response.StatusCode, response.Header.Get("Location"))
		}
		response.Body.Close()
		helper.Cleanup(t, app.Pool, app.Cache)
	})

	tests := []struct {
		name       string
		shortURL   string
		statusCode int
		errorCode  string
	}{
		{"too short", "abc1234", http.StatusBadRequest, "INVALID_INPUT"},
		{"too long", "abc123456", http.StatusBadRequest, "INVALID_INPUT"},
		{"non alphanumeric", "abc1234!", http.StatusBadRequest, "INVALID_INPUT"},
		{"not found", "missng1x", http.StatusNotFound, "URL_NOT_FOUND"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			
			response := helper.Request(t, app, http.MethodGet, "/api/v1/urls/"+test.shortURL, nil, "", "")
			if response.StatusCode != test.statusCode {
				t.Fatalf("expected status %d, got %d", test.statusCode, response.StatusCode)
			}
			if helper.ErrorCode(t, response) != test.errorCode {
				t.Fatalf("expected error %s", test.errorCode)
			}
		})
	}

	t.Run("expired URL is not redirected", func(t *testing.T) {
		
		_, err := app.Pool.Exec(context.Background(), "INSERT INTO urls_map (short_url, long_url, expired_at) VALUES ($1, $2, NOW() - INTERVAL '1 hour')", "exprd01x", "https://example.com/expired")
		if err != nil {
			t.Fatal(err)
		}
		response := helper.Request(t, app, http.MethodGet, "/api/v1/urls/exprd01x", nil, "", "")
		if response.StatusCode != http.StatusNotFound || helper.ErrorCode(t, response) != "URL_NOT_FOUND" {
			t.Fatalf("expected expired URL to be not found, got %d", response.StatusCode)
		}
		response.Body.Close()
		helper.Cleanup(t, app.Pool, app.Cache)
	})
}

func TestGetURLsForUser(t *testing.T) {
	app := helper.NewApp(t)
	t.Run("requires authentication", func(t *testing.T) {
		for _, auth := range []string{"", "Basic invalid", "Bearer invalid"} {
			t.Run(auth, func(t *testing.T) {
				
				response := helper.Request(t, app, http.MethodGet, "/api/v1/urls", nil, auth, "")
				if response.StatusCode != http.StatusUnauthorized || helper.ErrorCode(t, response) != "INVALID_TOKEN" {
					t.Fatalf("expected invalid token response, got %d", response.StatusCode)
				}
				response.Body.Close()
			})
		}
	})

	t.Run("returns only active URLs owned by the authenticated user", func(t *testing.T) {
		
		if response := helper.SignupUser(t, app, "alice", "alice@example.com", "password1"); response.StatusCode != http.StatusOK {
			t.Fatalf("alice signup: got %d", response.StatusCode)
		}
		aliceToken, _, response := helper.LoginUser(t, app, "alice@example.com", "password1")
		if response.StatusCode != http.StatusOK {
			t.Fatalf("alice login: got %d", response.StatusCode)
		}
		if response := helper.SignupUser(t, app, "bob", "bob@example.com", "password1"); response.StatusCode != http.StatusOK {
			t.Fatalf("bob signup: got %d", response.StatusCode)
		}
		bobToken, _, response := helper.LoginUser(t, app, "bob@example.com", "password1")
		if response.StatusCode != http.StatusOK {
			t.Fatalf("bob login: got %d", response.StatusCode)
		}
		aliceURL := shortURLFromResponse(t, createURL(t, app, "https://example.com/alice", aliceToken))
		if response := createURL(t, app, "https://example.com/bob", bobToken); response.StatusCode != http.StatusOK {
			t.Fatalf("bob URL: got %d", response.StatusCode)
		}
		if _, err := app.Pool.Exec(context.Background(), "INSERT INTO urls_map (user_id, short_url, long_url, expired_at) SELECT id, $1, $2, NOW() - INTERVAL '1 hour' FROM users WHERE email = $3", "exprd02x", "https://example.com/expired", "alice@example.com"); err != nil {
			t.Fatal(err)
		}

		response = helper.Request(t, app, http.MethodGet, "/api/v1/urls", nil, bearer(aliceToken), "")
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
		}
		payload := helper.DecodeResponse[envelope.SuccessResponse](t, response)
		data := payload.Data.(map[string]any)
		urls := data["urls"].([]any)
		if len(urls) != 1 || urls[0].(map[string]any)["short_url"] != aliceURL {
			t.Fatalf("expected only Alice's active URL, got %#v", urls)
		}
		
		helper.Cleanup(t, app.Pool, app.Cache)
	})

	t.Run("returns an empty list for a user without URLs", func(t *testing.T) {
		
		if response := helper.SignupUser(t, app, "alice", "alice@example.com", "password1"); response.StatusCode != http.StatusOK {
			t.Fatalf("signup: got %d", response.StatusCode)
		}
		accessToken, _, response := helper.LoginUser(t, app, "alice@example.com", "password1")
		if response.StatusCode != http.StatusOK {
			t.Fatalf("login: got %d", response.StatusCode)
		}
		response = helper.Request(t, app, http.MethodGet, "/api/v1/urls", nil, bearer(accessToken), "")
		if response.StatusCode != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
		}
		payload := helper.DecodeResponse[envelope.SuccessResponse](t, response)
		data := payload.Data.(map[string]any)
		if data["urls"] != nil {
			t.Fatalf("expected current empty-list response to contain null, got %#v", data["urls"])
		}
		
		helper.Cleanup(t, app.Pool, app.Cache)
	})
}
