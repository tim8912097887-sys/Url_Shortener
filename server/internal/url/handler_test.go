package url_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	urlerror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/url_error"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response/envelope"
	urlschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/url_schema"
	userschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/user_schema"
	jwttoken "github.com/tim8912097887-sys/url-shortener/internal/shared/util/jwt_token"
	"github.com/tim8912097887-sys/url-shortener/internal/url"
)

func decodeResponse[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	var payload T
	err := json.NewDecoder(resp.Body).Decode(&payload)
	if err != nil {
		t.Fatal(err)
	}
	return payload
}

func wireupHandler(
	t *testing.T,
	repo *MockRepository,
	cache *MockCache,
) *url.Handler {
	t.Helper()

	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, handlerOpts))

	service := url.NewService(&url.ServiceConfig{
		UrlRepository: repo,
		UserRepository: InitMockUserRepository(),
		Cache:         cache,
		Logger:        logger,
	})
	tokens := jwttoken.NewTokenManager("access-secret", "refresh-secret")
	handler := url.NewHandler(url.HandlerConfig{
		Logger:  logger,
		Service: service,
		Tokens:  *tokens,
	})

	return &handler
}

func accessToken(t *testing.T, userID string) string {
	t.Helper()
	tokens := jwttoken.NewTokenManager("access-secret", "refresh-secret")
	token, err := tokens.GenerateAccessToken(userID, 1)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func setupRouter(t *testing.T, h *url.Handler) *fiber.App {
	t.Helper()
	app := fiber.New()
	urlGroup := app.Group("/api/v1/urls")
	h.RegisterRoutes(urlGroup)
	return app
}

func shortenUrlRequest(t *testing.T, app *fiber.App, payload url.CreateUrlSchema) *http.Response {
	t.Helper()
	// Serialize payload
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	// Construct request
	req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := app.Test(req, fiber.TestConfig{
		Timeout: -1,
	})
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func getUrlRequest(t *testing.T, app *fiber.App, params string) *http.Response {
	t.Helper()
	urlString := "/api/v1/urls/" + params
	// Construct request
	req := httptest.NewRequest(http.MethodGet, urlString, bytes.NewReader([]byte{}))

	// Make request
	resp, err := app.Test(req, fiber.TestConfig{
		Timeout: -1,
	})
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func getUrlsForUserRequest(t *testing.T, app *fiber.App, token string, query ...string) *http.Response {
	t.Helper()
	path := "/api/v1/urls"
	if len(query) > 0 && query[0] != "" {
		path += "?" + query[0]
	}
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
	}
	resp, err := app.Test(req, fiber.TestConfig{Timeout: -1})
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestShortenUrlValidation(t *testing.T) {

	mockRepository := InitMockRepository()
	mockCache := InitMockCache()
	handler := wireupHandler(t, mockRepository, mockCache)

	t.Run("when provide invalid url,should response with Invalid Input Error", func(t *testing.T) {
		// Arrange
		payload := url.CreateUrlSchema{Url: "invalid url"}
		app := setupRouter(t, handler)
		// Act
		resp := shortenUrlRequest(t, app, payload)
		// Assert
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d but got %d", http.StatusBadRequest, resp.StatusCode)
		}

		errorResponse := decodeResponse[envelope.ErrorResponse](t, resp)
		if errorResponse.Error.Code != "INVALID_INPUT" {
			t.Errorf("expected error code %s but got %s", "invalid_input", errorResponse.Error.Code)
		}
		if !strings.Contains(errorResponse.Error.Message, "url") {
			t.Errorf("expected error message contains %s but got %s", "url", errorResponse.Error.Message)
		}
	})
}

func TestShortenUrlSuccess(t *testing.T) {

	t.Run("when provide valid url,should response with Success", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockCache := InitMockCache()
		handler := wireupHandler(t, mockRepository, mockCache)
		payload := url.CreateUrlSchema{Url: "https://www.google.com/"}
		app := setupRouter(t, handler)
		// Act
		resp := shortenUrlRequest(t, app, payload)
		// Assert
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d but got %d", http.StatusOK, resp.StatusCode)
		}

		successResponse := decodeResponse[envelope.SuccessResponse](t, resp)
		if successResponse.Data.(map[string]any)["message"] != "Successfully shorten url" {
			t.Errorf("expected message %s but got %s", "Successfully shorten url", successResponse.Data.(map[string]string)["message"])
		}
		if len(successResponse.Data.(map[string]any)["shortUrl"].(string)) != 8 {
			t.Errorf("expected short url length %d but got %d", 8, len(successResponse.Data.(map[string]any)["shortUrl"].(string)))
		}
	})

	t.Run("when set cache failed,should still response with Success", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockCache := InitMockCache()
		mockCache.SetFunc = func(ctx context.Context, key string, value any, expiration time.Duration) error {
			return errors.New("error")
		}
		handler := wireupHandler(t, mockRepository, mockCache)
		payload := url.CreateUrlSchema{Url: "https://www.google.com/"}
		app := setupRouter(t, handler)
		// Act
		resp := shortenUrlRequest(t, app, payload)
		// Assert
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d but got %d", http.StatusOK, resp.StatusCode)
		}

		successResponse := decodeResponse[envelope.SuccessResponse](t, resp)
		if successResponse.Data.(map[string]any)["message"] != "Successfully shorten url" {
			t.Errorf("expected message %s but got %s", "Successfully shorten url", successResponse.Data.(map[string]string)["message"])
		}
		if len(successResponse.Data.(map[string]any)["shortUrl"].(string)) != 8 {
			t.Errorf("expected short url length %d but got %d", 8, len(successResponse.Data.(map[string]any)["shortUrl"].(string)))
		}
	})
}

func TestGetUrlSuccess(t *testing.T) {
	t.Run("when provide valid and exist short url,should response with Temporary Redirect", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockCache := InitMockCache()
		handler := wireupHandler(t, mockRepository, mockCache)
		payload := url.CreateUrlSchema{Url: "https://www.google.com/"}
		app := setupRouter(t, handler)
		resp := shortenUrlRequest(t, app, payload)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d but got %d", http.StatusOK, resp.StatusCode)
		}

		// Act
		params := decodeResponse[envelope.SuccessResponse](t, resp).Data.(map[string]any)["shortUrl"].(string)
		resp = getUrlRequest(t, app, params)
		// Assert
		if resp.StatusCode != http.StatusTemporaryRedirect {
			t.Fatalf("expected status code %d but got %d", http.StatusTemporaryRedirect, resp.StatusCode)
		}
	})
}

func TestGetUrlCache(t *testing.T) {

	t.Run("when cache hit,should response with Temporary Redirect", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockCache := InitMockCache()
		mockCache.GetFunc = func(ctx context.Context, key string) (string, error) { return "https://www.google.com", nil }
		handler := wireupHandler(t, mockRepository, mockCache)
		params := "sdfj32fo"
		app := setupRouter(t, handler)
		// Act
		resp := getUrlRequest(t, app, params)
		// Assert
		if resp.StatusCode != http.StatusTemporaryRedirect {
			t.Fatalf("expected status code %d but got %d", http.StatusTemporaryRedirect, resp.StatusCode)
		}
	})

	t.Run("when get cache failed,should response with Temporary Redirect", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockCache := InitMockCache()
		mockCache.GetFunc = func(ctx context.Context, key string) (string, error) { return "", errors.New("error") }
		handler := wireupHandler(t, mockRepository, mockCache)
		params := "sdfj32fo"
		app := setupRouter(t, handler)
		// Act
		resp := getUrlRequest(t, app, params)
		// Assert
		if resp.StatusCode != http.StatusTemporaryRedirect {
			t.Fatalf("expected status code %d but got %d", http.StatusTemporaryRedirect, resp.StatusCode)
		}
	})

	t.Run("when set cache fail,should response with Temporary Redirect", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockCache := InitMockCache()
		mockCache.SetFunc = func(ctx context.Context, key string, value any, expiration time.Duration) error {
			return errors.New("error")
		}
		handler := wireupHandler(t, mockRepository, mockCache)
		params := "sdfj32fo"
		app := setupRouter(t, handler)
		// Act
		resp := getUrlRequest(t, app, params)
		// Assert
		if resp.StatusCode != http.StatusTemporaryRedirect {
			t.Fatalf("expected status code %d but got %d", http.StatusTemporaryRedirect, resp.StatusCode)
		}
	})
}

func TestGetUrlBusinessLogic(t *testing.T) {

	t.Run("when provide valid but not exist short url,should response with Not Found Error", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockRepository.GetLongUrlFunc = func(ctx context.Context, shortUrl string) (string, time.Time, error) {
			return "", time.Time{}, urlerror.ErrUrlNotFound
		}
		mockCache := InitMockCache()
		handler := wireupHandler(t, mockRepository, mockCache)
		app := setupRouter(t, handler)
		params := "sdfj32fo"

		// Act
		resp := getUrlRequest(t, app, params)
		// Assert
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status code %d but got %d", http.StatusNotFound, resp.StatusCode)
		}

		errorResponse := decodeResponse[envelope.ErrorResponse](t, resp)
		if errorResponse.Error.Code != "URL_NOT_FOUND" {
			t.Errorf("expected error code %s but got %s", "url_not_found", errorResponse.Error.Code)
		}
		if !strings.Contains(errorResponse.Error.Message, "url") {
			t.Errorf("expected error message contains %s but got %s", "url", errorResponse.Error.Message)
		}
	})

	t.Run("when provide expired url,should response with Not Found Error", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockRepository.GetLongUrlFunc = func(ctx context.Context, shortUrl string) (string, time.Time, error) {
			return "https://www.google.com", time.Now().Add(-24 * time.Hour), nil
		}
		mockCache := InitMockCache()
		handler := wireupHandler(t, mockRepository, mockCache)
		app := setupRouter(t, handler)
		params := "sdfj32fo"

		// Act
		resp := getUrlRequest(t, app, params)
		// Assert
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status code %d but got %d", http.StatusNotFound, resp.StatusCode)
		}

		errorResponse := decodeResponse[envelope.ErrorResponse](t, resp)
		if errorResponse.Error.Code != "URL_NOT_FOUND" {
			t.Errorf("expected error code %s but got %s", "url_not_found", errorResponse.Error.Code)
		}
		if !strings.Contains(errorResponse.Error.Message, "url") {
			t.Errorf("expected error message contains %s but got %s", "url", errorResponse.Error.Message)
		}
	})
}

func TestGetUrlParams(t *testing.T) {

	mockRepository := InitMockRepository()
	mockCache := InitMockCache()
	handler := wireupHandler(t, mockRepository, mockCache)

	t.Run("when provide less than 8 chars short url in params,should response with Invalid Input Error", func(t *testing.T) {
		// Arrange
		app := setupRouter(t, handler)
		params := "sdfj3"

		// Act
		resp := getUrlRequest(t, app, params)
		// Assert
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d but got %d", http.StatusBadRequest, resp.StatusCode)
		}

		errorResponse := decodeResponse[envelope.ErrorResponse](t, resp)
		if errorResponse.Error.Code != "INVALID_INPUT" {
			t.Errorf("expected error code %s but got %s", "invalid_input", errorResponse.Error.Code)
		}
	})

	t.Run("when provide more than 8 chars short url in params,should response with Invalid Input Error", func(t *testing.T) {
		// Arrange
		app := setupRouter(t, handler)
		params := "sdfj32fof"

		// Act
		resp := getUrlRequest(t, app, params)
		// Assert
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d but got %d", http.StatusBadRequest, resp.StatusCode)
		}

		errorResponse := decodeResponse[envelope.ErrorResponse](t, resp)
		if errorResponse.Error.Code != "INVALID_INPUT" {
			t.Errorf("expected error code %s but got %s", "invalid_input", errorResponse.Error.Code)
		}
	})

	t.Run("when provide non alphanumeric chars short url in params,should response with Invalid Input Error", func(t *testing.T) {
		// Arrange
		app := setupRouter(t, handler)
		params := "sdfj32f!"

		// Act
		resp := getUrlRequest(t, app, params)
		// Assert
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d but got %d", http.StatusBadRequest, resp.StatusCode)
		}

		errorResponse := decodeResponse[envelope.ErrorResponse](t, resp)
		if errorResponse.Error.Code != "INVALID_INPUT" {
			t.Errorf("expected error code %s but got %s", "invalid_input", errorResponse.Error.Code)
		}
	})

}

func TestGetUrlsForUser(t *testing.T) {
	tests := []struct {
		name          string
		authenticated bool
		token         string
		repository    func(*MockRepository)
		statusCode    int
		errorCode     string
	}{
		{name: "missing authentication", statusCode: http.StatusUnauthorized, errorCode: "INVALID_TOKEN"},
		{name: "invalid authentication", token: "invalid-token", statusCode: http.StatusUnauthorized, errorCode: "INVALID_TOKEN"},
		{
			name:          "repository failure",
			authenticated: true,
			repository: func(repository *MockRepository) {
				repository.GetUrlsForUserFunc = func(context.Context, string, time.Time, int) ([]urlschema.GetUrlsRepositoryResponse, bool, error) {
					return nil, false, errors.New("database unavailable")
				}
			},
			statusCode: http.StatusInternalServerError,
			errorCode:  "INTERNAL_SERVER_ERROR",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := InitMockRepository()
			if test.repository != nil {
				test.repository(repository)
			}
			app := setupRouter(t, wireupHandler(t, repository, InitMockCache()))
			token := test.token
			if test.authenticated {
				token = accessToken(t, "user-1")
			}

			resp := getUrlsForUserRequest(t, app, token)
			if resp.StatusCode != test.statusCode {
				t.Fatalf("expected status code %d but got %d", test.statusCode, resp.StatusCode)
			}
			if test.errorCode != "" && decodeResponse[envelope.ErrorResponse](t, resp).Error.Code != test.errorCode {
				t.Fatalf("expected error code %s", test.errorCode)
			}
		})
	}

	t.Run("returns empty urls", func(t *testing.T) {
		repository := InitMockRepository()
		repository.GetUrlsForUserFunc = func(context.Context, string, time.Time, int) ([]urlschema.GetUrlsRepositoryResponse, bool, error) {
			return []urlschema.GetUrlsRepositoryResponse{}, false, nil
		}
		app := setupRouter(t, wireupHandler(t, repository, InitMockCache()))
		resp := getUrlsForUserRequest(t, app, accessToken(t, "user-1"))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d but got %d", http.StatusOK, resp.StatusCode)
		}
		response := decodeResponse[envelope.SuccessResponse](t, resp)
		data := response.Data.(map[string]any)
		if urls := data["urls"]; urls != nil {
			if list, ok := urls.([]any); !ok || len(list) != 0 {
				t.Fatalf("expected no urls, got %#v", urls)
			}
		}
	})

	t.Run("returns user urls", func(t *testing.T) {
		expiredAt := time.Now().Add(time.Hour).UTC().Truncate(time.Millisecond)
		repository := InitMockRepository()
		repository.GetUrlsForUserFunc = func(ctx context.Context, userID string, receivedExpiredAt time.Time, limit int) ([]urlschema.GetUrlsRepositoryResponse, bool, error) {
			if userID != "user-1" {
				t.Fatalf("expected user-1, got %s", userID)
			}
			if !receivedExpiredAt.After(time.Now()) {
				t.Fatalf("expected a future expiredAt, got %s", receivedExpiredAt)
			}
			if limit != url.UrlsMaxLimit {
				t.Fatalf("expected default limit %d, got %d", url.UrlsMaxLimit, limit)
			}
			return []urlschema.GetUrlsRepositoryResponse{{ShortUrl: "abc12345", LongUrl: "https://example.com", ExpiredAt: expiredAt}}, false, nil
		}
		app := setupRouter(t, wireupHandler(t, repository, InitMockCache()))
		resp := getUrlsForUserRequest(t, app, accessToken(t, "user-1"))
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d but got %d", http.StatusOK, resp.StatusCode)
		}
		response := decodeResponse[envelope.SuccessResponse](t, resp)
		data := response.Data.(map[string]any)
		urls := data["urls"].([]any)
		if len(urls) != 1 || urls[0].(map[string]any)["short_url"] != "abc12345" {
			t.Fatalf("unexpected urls response: %#v", urls)
		}
	})

	t.Run("applies pagination defaults and clamps limit and expiredAt", func(t *testing.T) {
		repository := InitMockRepository()
		repository.GetUrlsForUserFunc = func(ctx context.Context, userID string, expiredAt time.Time, limit int) ([]urlschema.GetUrlsRepositoryResponse, bool, error) {
			if userID != "user-1" {
				t.Fatalf("expected user-1, got %s", userID)
			}
			if limit != url.UrlsMaxLimit {
				t.Fatalf("expected clamped limit %d, got %d", url.UrlsMaxLimit, limit)
			}
			delta := time.Until(expiredAt)
			if delta < url.AuthURLExpiry-time.Minute || delta > url.AuthURLExpiry+time.Minute {
				t.Fatalf("expected expiredAt near %s, got %s (delta=%s)", time.Now().Add(url.AuthURLExpiry), expiredAt, delta)
			}
			return []urlschema.GetUrlsRepositoryResponse{}, false, nil
		}
		app := setupRouter(t, wireupHandler(t, repository, InitMockCache()))
		resp := getUrlsForUserRequest(t, app, accessToken(t, "user-1"), "limit=999&expiredAt=not-a-date")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d but got %d", http.StatusOK, resp.StatusCode)
		}
	})

	t.Run("respects minimum and maximum limit and past expiredAt fallback", func(t *testing.T) {
		testCases := []struct {
			name      string
			query     string
			expectedL int
			expectedA time.Time
		}{
			{name: "limit below minimum", query: "limit=0", expectedL: url.UrlsMinLimit},
			{name: "limit above maximum", query: "limit=999", expectedL: url.UrlsMaxLimit},
			{name: "past expiredAt fallback", query: "expiredAt=" + time.Now().Add(-time.Hour).Format(time.RFC3339Nano), expectedL: url.UrlsMaxLimit},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				repository := InitMockRepository()
				repository.GetUrlsForUserFunc = func(ctx context.Context, userID string, expiredAt time.Time, limit int) ([]urlschema.GetUrlsRepositoryResponse, bool, error) {
					if limit != tc.expectedL {
						t.Fatalf("expected limit %d, got %d", tc.expectedL, limit)
					}
					if tc.query == "" || strings.Contains(tc.query, "expiredAt=") {
						delta := time.Until(expiredAt)
						if delta < url.AuthURLExpiry-time.Minute || delta > url.AuthURLExpiry+time.Minute {
							t.Fatalf("expected expiredAt near default fallback, got %s (delta=%s)", expiredAt, delta)
						}
					}
					return []urlschema.GetUrlsRepositoryResponse{}, false, nil
				}
				app := setupRouter(t, wireupHandler(t, repository, InitMockCache()))
				resp := getUrlsForUserRequest(t, app, accessToken(t, "user-1"), tc.query)
				if resp.StatusCode != http.StatusOK {
					t.Fatalf("expected status code %d but got %d", http.StatusOK, resp.StatusCode)
				}
			})
		}
	})

	t.Run("returns hasMore from repository", func(t *testing.T) {
		repository := InitMockRepository()
		repository.GetUrlsForUserFunc = func(context.Context, string, time.Time, int) ([]urlschema.GetUrlsRepositoryResponse, bool, error) {
			return []urlschema.GetUrlsRepositoryResponse{{ShortUrl: "abc12345", LongUrl: "https://example.com", ExpiredAt: time.Now().Add(time.Hour)}}, true, nil
		}
		app := setupRouter(t, wireupHandler(t, repository, InitMockCache()))
		resp := getUrlsForUserRequest(t, app, accessToken(t, "user-1"), "limit=1")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d but got %d", http.StatusOK, resp.StatusCode)
		}
		response := decodeResponse[envelope.SuccessResponse](t, resp)
		data := response.Data.(map[string]any)
		if data["hasMore"] != true {
			t.Fatalf("expected hasMore to be true, got %#v", data["hasMore"])
		}
	})
}

type MockRepository struct {
	GetLongUrlFunc       func(ctx context.Context, shortUrl string) (string, time.Time, error)
	ShortCodeExistsFunc  func(ctx context.Context, shortUrl string) (bool, error)
	CreateShortenUrlFunc func(ctx context.Context, longUrl string, shortUrl string, userID *string, expiredAt time.Time) (string, error)
	GetUrlsForUserFunc   func(ctx context.Context, userID string, expiredAt time.Time, limit int) ([]urlschema.GetUrlsRepositoryResponse, bool, error)
	UpdateUrlClicksFunc  func(ctx context.Context, shortUrl string, clicks int) error
}

func InitMockRepository() *MockRepository {
	return &MockRepository{
		GetLongUrlFunc: func(ctx context.Context, shortUrl string) (string, time.Time, error) {
			return "https://google.com", time.Now().Add(24 * time.Hour), nil
		},
		ShortCodeExistsFunc: func(ctx context.Context, shortUrl string) (bool, error) {
			return false, nil
		},
		CreateShortenUrlFunc: func(ctx context.Context, longUrl string, shortUrl string, userID *string, expiredAt time.Time) (string, error) {
			return shortUrl, nil
		},
		GetUrlsForUserFunc: func(ctx context.Context, userID string, expiredAt time.Time, limit int) ([]urlschema.GetUrlsRepositoryResponse, bool, error) {
			return []urlschema.GetUrlsRepositoryResponse{}, false, nil
		},
		UpdateUrlClicksFunc: func(ctx context.Context, shortUrl string, clicks int) error {
			return nil
		},
	}
}

func (m *MockRepository) GetLongUrl(ctx context.Context, shortUrl string) (string, time.Time, error) {
	return m.GetLongUrlFunc(ctx, shortUrl)
}

func (m *MockRepository) ShortCodeExists(ctx context.Context, shortUrl string) (bool, error) {
	return m.ShortCodeExistsFunc(ctx, shortUrl)
}

func (m *MockRepository) CreateShortenUrl(ctx context.Context, longUrl string, shortUrl string, userID *string, expiredAt time.Time) (string, error) {
	return m.CreateShortenUrlFunc(ctx, longUrl, shortUrl, userID, expiredAt)
}

func (m *MockRepository) GetUrlsForUser(ctx context.Context, userID string, expiredAt time.Time, limit int) ([]urlschema.GetUrlsRepositoryResponse, bool, error) {
	return m.GetUrlsForUserFunc(ctx, userID, expiredAt, limit)
}

func (m *MockRepository) UpdateUrlClicks(ctx context.Context, shortUrl string, clicks int) error {
	return m.UpdateUrlClicksFunc(ctx, shortUrl, clicks)
}

type MockUserRepository struct {
	GetUserByIDFunc func(ctx context.Context, id string) (*userschema.GetUserByIDRepositoryResponse, error)
}

func InitMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		GetUserByIDFunc: func(ctx context.Context, id string) (*userschema.GetUserByIDRepositoryResponse, error) {
			return &userschema.GetUserByIDRepositoryResponse{ID: id, TokenVersion: 1}, nil
		},
	}
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*userschema.GetUserByIDRepositoryResponse, error) {
	return m.GetUserByIDFunc(ctx, id)
}

type MockCache struct {
	GetFunc               func(ctx context.Context, key string) (string, error)
	SetFunc               func(ctx context.Context, key string, value any, expiration time.Duration) error
	IncrementFunc         func(ctx context.Context, key string) (int64, error)
	GetAndResetFunc       func(ctx context.Context, key string) (int64, error)
	AddPendingClickFunc   func(ctx context.Context, shortURL string) error
	GetPendingClicksFunc  func(ctx context.Context) ([]string, error)
	RemovePendingClickFunc func(ctx context.Context, shortURL string) error
}

func InitMockCache() *MockCache {
	return &MockCache{
		GetFunc: func(ctx context.Context, key string) (string, error) {
			return "", redis.Nil
		},
		SetFunc: func(ctx context.Context, key string, value any, expiration time.Duration) error {
			return nil
		},
		IncrementFunc: func(ctx context.Context, key string) (int64, error) {
			return 1, nil
		},
		GetAndResetFunc: func(ctx context.Context, key string) (int64, error) {
			return 0, nil
		},
		AddPendingClickFunc: func(ctx context.Context, shortURL string) error {
			return nil
		},
		GetPendingClicksFunc: func(ctx context.Context) ([]string, error) {
			return nil, nil
		},
		RemovePendingClickFunc: func(ctx context.Context, shortURL string) error {
			return nil
		},
	}
}

func (m *MockCache) Get(ctx context.Context, key string) (string, error) {
	return m.GetFunc(ctx, key)
}

func (m *MockCache) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return m.SetFunc(ctx, key, value, expiration)
}

func (m *MockCache) Increment(ctx context.Context, key string) (int64, error) {
	return m.IncrementFunc(ctx, key)
}

func (m *MockCache) GetAndReset(ctx context.Context, key string) (int64, error) {
	return m.GetAndResetFunc(ctx, key)
}

func (m *MockCache) AddPendingClick(ctx context.Context, shortURL string) error {
	return m.AddPendingClickFunc(ctx, shortURL)
}

func (m *MockCache) GetPendingClicks(ctx context.Context) ([]string, error) {
	return m.GetPendingClicksFunc(ctx)
}

func (m *MockCache) RemovePendingClick(ctx context.Context, shortURL string) error {
	return m.RemovePendingClickFunc(ctx, shortURL)
}
