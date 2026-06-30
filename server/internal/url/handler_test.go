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
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response"
	"github.com/tim8912097887-sys/url-shortener/internal/url"
)

func decodeResponse[T any](t *testing.T,resp *http.Response) T {
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

	service := url.NewService(repo, cache, logger)
	handler := url.NewHandler(logger, service)

	return &handler
}

func setupRouter(t *testing.T,h *url.Handler) *fiber.App {
	t.Helper()
	app := fiber.New()
    urlGroup := app.Group("/api/v1/urls")
	h.RegisterRoutes(urlGroup)
	return app
}

func shortenUrlRequest(t *testing.T,app *fiber.App,payload url.CreateUrlSchema) *http.Response {
	t.Helper()
	// Serialize payload
    body,err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	// Construct request
	req := httptest.NewRequest(http.MethodPost,"/api/v1/urls",bytes.NewReader(body))
	req.Header.Set("Content-Type","application/json")

	// Make request
	resp, err := app.Test(req,fiber.TestConfig{
	    Timeout: -1,
	})
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func getUrlRequest(t *testing.T,app *fiber.App,params string) *http.Response {
	t.Helper()
	urlString := "/api/v1/urls/" + params
	// Construct request
	req := httptest.NewRequest(http.MethodGet,urlString,bytes.NewReader([]byte{}))

	// Make request
	resp, err := app.Test(req,fiber.TestConfig{
	    Timeout: -1,
	})
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestShortenUrlValidation(t *testing.T) {

	mockRepository := InitMockRepository()
	mockCache := InitMockCache()
	handler := wireupHandler(t,mockRepository, mockCache)

	t.Run("when provide invalid url,should response with Invalid Input Error", func(t *testing.T) {
		// Arrange
        payload := url.CreateUrlSchema{Url: "invalid url"}
		app := setupRouter(t,handler)
		// Act
		resp := shortenUrlRequest(t,app,payload)
		// Assert
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d but got %d",http.StatusBadRequest,resp.StatusCode)
		}

		errorResponse := decodeResponse[response.ErrorResponse](t,resp)
		if errorResponse.Error.Code != "invalid_input" {
			t.Errorf("expected error code %s but got %s", "invalid_input",errorResponse.Error.Code)
		}
		if !strings.Contains(errorResponse.Error.Message, "url") {
			t.Errorf("expected error message contains %s but got %s", "url",errorResponse.Error.Message)
		}
	})
}

func TestShortenUrlSuccess(t *testing.T) {

	t.Run("when provide valid url,should response with Success", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockCache := InitMockCache()
		handler := wireupHandler(t,mockRepository, mockCache)
		payload := url.CreateUrlSchema{Url: "https://www.google.com/"}
		app := setupRouter(t,handler)
		// Act
		resp := shortenUrlRequest(t,app,payload)
		// Assert
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d but got %d",http.StatusOK,resp.StatusCode)
		}

		successResponse := decodeResponse[response.SuccessResponse](t,resp)
		if successResponse.Data.(map[string]any)["message"] != "Successfully shorten url" {
			t.Errorf("expected message %s but got %s", "Successfully shorten url",successResponse.Data.(map[string]string)["message"])
		}
	})

	t.Run("when set cache failed,should still response with Success", func(t *testing.T) {
		// Arrange
	    mockRepository := InitMockRepository()
	    mockCache := InitMockCache()
	    mockCache.SetFunc = func(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {return redis.NewStatusResult("error",errors.New("error"))}
	    handler := wireupHandler(t,mockRepository, mockCache)
		payload := url.CreateUrlSchema{Url: "https://www.google.com/"}
		app := setupRouter(t,handler)
		// Act
		resp := shortenUrlRequest(t,app,payload)
		// Assert
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d but got %d",http.StatusOK,resp.StatusCode)
		}

		successResponse := decodeResponse[response.SuccessResponse](t,resp)
		if successResponse.Data.(map[string]any)["message"] != "Successfully shorten url" {
			t.Errorf("expected message %s but got %s", "Successfully shorten url",successResponse.Data.(map[string]string)["message"])
		}
	})
}

func TestGetUrlSuccess(t *testing.T) {
	t.Run("when provide valid and exist short url,should response with Temporary Redirect", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockCache := InitMockCache()
		handler := wireupHandler(t,mockRepository, mockCache)
		payload := url.CreateUrlSchema{Url: "https://www.google.com/"}
		app := setupRouter(t,handler)
		resp := shortenUrlRequest(t,app,payload)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status code %d but got %d",http.StatusOK,resp.StatusCode)
		}

		// Act
		params := decodeResponse[response.SuccessResponse](t,resp).Data.(map[string]any)["shortUrl"].(string)
		resp = getUrlRequest(t,app,params)
		// Assert
		if resp.StatusCode != http.StatusTemporaryRedirect {
			t.Fatalf("expected status code %d but got %d",http.StatusTemporaryRedirect,resp.StatusCode)
		}
	})
}

func TestGetUrlCache(t *testing.T) {

	t.Run("when cache hit,should response with Temporary Redirect", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockCache := InitMockCache()
		mockCache.GetFunc = func(ctx context.Context, key string) *redis.StringCmd {return redis.NewStringResult("https://www.google.com",nil)}
		handler := wireupHandler(t,mockRepository, mockCache)
	    params := "sdfj32fo"
		app := setupRouter(t,handler)
		// Act
		resp := getUrlRequest(t,app,params)
		// Assert
		if resp.StatusCode != http.StatusTemporaryRedirect {
			t.Fatalf("expected status code %d but got %d",http.StatusTemporaryRedirect,resp.StatusCode)
		}
	})

	t.Run("when get cache failed,should response with Temporary Redirect", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockCache := InitMockCache()
		mockCache.GetFunc = func(ctx context.Context, key string) *redis.StringCmd {return redis.NewStringResult("",errors.New("error"))}
		handler := wireupHandler(t,mockRepository, mockCache)
	    params := "sdfj32fo"
		app := setupRouter(t,handler)
		// Act
		resp := getUrlRequest(t,app,params)
		// Assert
		if resp.StatusCode != http.StatusTemporaryRedirect {
			t.Fatalf("expected status code %d but got %d",http.StatusTemporaryRedirect,resp.StatusCode)
		}
	})

	t.Run("when set cache fail,should response with Temporary Redirect", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockCache := InitMockCache()
		mockCache.SetFunc = func(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {return redis.NewStatusResult("error",errors.New("error"))}
		handler := wireupHandler(t,mockRepository, mockCache)
	    params := "sdfj32fo"
		app := setupRouter(t,handler)
		// Act
		resp := getUrlRequest(t,app,params)
		// Assert
		if resp.StatusCode != http.StatusTemporaryRedirect {
			t.Fatalf("expected status code %d but got %d",http.StatusTemporaryRedirect,resp.StatusCode)
		}
	})
}

func TestGetUrlBusinessLogic(t *testing.T) {

	t.Run("when provide valid but not exist short url,should response with Not Found Error", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockRepository.GetLongUrlFunc = func(ctx context.Context, shortUrl string) (string, time.Time, error) {return "",time.Time{},url.ErrUrlNotFound}
		mockCache := InitMockCache()
		handler := wireupHandler(t,mockRepository, mockCache)
		app := setupRouter(t,handler)
		params := "sdfj32fo"

		// Act
		resp := getUrlRequest(t,app,params)
		// Assert
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status code %d but got %d",http.StatusNotFound,resp.StatusCode)
		}

		errorResponse := decodeResponse[response.ErrorResponse](t,resp)
		if errorResponse.Error.Code != "url_not_found" {
			t.Errorf("expected error code %s but got %s", "url_not_found",errorResponse.Error.Code)
		}
		if !strings.Contains(errorResponse.Error.Message, "url") {
			t.Errorf("expected error message contains %s but got %s", "url",errorResponse.Error.Message)
		}
	})

	t.Run("when provide expired url,should response with Not Found Error", func(t *testing.T) {
		// Arrange
		mockRepository := InitMockRepository()
		mockRepository.GetLongUrlFunc = func(ctx context.Context, shortUrl string) (string, time.Time, error) {return "https://www.google.com",time.Now().Add(-24 * time.Hour),nil}
		mockCache := InitMockCache()
		handler := wireupHandler(t,mockRepository, mockCache)
		app := setupRouter(t,handler)
		params := "sdfj32fo"

		// Act
		resp := getUrlRequest(t,app,params)
		// Assert
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status code %d but got %d",http.StatusNotFound,resp.StatusCode)
		}

		errorResponse := decodeResponse[response.ErrorResponse](t,resp)
		if errorResponse.Error.Code != "url_not_found" {
			t.Errorf("expected error code %s but got %s", "url_not_found",errorResponse.Error.Code)
		}
		if !strings.Contains(errorResponse.Error.Message, "url") {
			t.Errorf("expected error message contains %s but got %s", "url",errorResponse.Error.Message)
		}
	})
}

func TestGetUrlParams(t *testing.T) {

	mockRepository := InitMockRepository()
	mockCache := InitMockCache()
	handler := wireupHandler(t,mockRepository, mockCache)

	t.Run("when provide less than 8 chars short url in params,should response with Invalid Input Error", func(t *testing.T) {
		// Arrange
		app := setupRouter(t,handler)
		params := "sdfj3"

		// Act
		resp := getUrlRequest(t,app,params)
		// Assert
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d but got %d",http.StatusBadRequest,resp.StatusCode)
		}

		errorResponse := decodeResponse[response.ErrorResponse](t,resp)
		if errorResponse.Error.Code != "invalid_input" {
			t.Errorf("expected error code %s but got %s", "invalid_input",errorResponse.Error.Code)
		}
	})

	t.Run("when provide more than 8 chars short url in params,should response with Invalid Input Error", func(t *testing.T) {
		// Arrange
		app := setupRouter(t,handler)
		params := "sdfj32fof"

		// Act
		resp := getUrlRequest(t,app,params)
		// Assert
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d but got %d",http.StatusBadRequest,resp.StatusCode)
		}

		errorResponse := decodeResponse[response.ErrorResponse](t,resp)
		if errorResponse.Error.Code != "invalid_input" {
			t.Errorf("expected error code %s but got %s", "invalid_input",errorResponse.Error.Code)
		}
	})

	t.Run("when provide non alphanumeric chars short url in params,should response with Invalid Input Error", func(t *testing.T) {
		// Arrange
		app := setupRouter(t,handler)
		params := "sdfj32f!"

		// Act
		resp := getUrlRequest(t,app,params)
		// Assert
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d but got %d",http.StatusBadRequest,resp.StatusCode)
		}

		errorResponse := decodeResponse[response.ErrorResponse](t,resp)
		if errorResponse.Error.Code != "invalid_input" {
			t.Errorf("expected error code %s but got %s", "invalid_input",errorResponse.Error.Code)
		}
	})

}

type MockRepository struct {
	GetLongUrlFunc       func(ctx context.Context, shortUrl string) (string, time.Time, error)
	ShortCodeExistsFunc  func(ctx context.Context, shortUrl string) (bool, error)
	CreateShortenUrlFunc func(ctx context.Context, longUrl string, shortUrl string) (string, error)
}

func InitMockRepository() *MockRepository {
	return &MockRepository{
		GetLongUrlFunc: func(ctx context.Context, shortUrl string) (string, time.Time, error) {
			return "https://google.com", time.Now().Add(24 * time.Hour), nil
		},
		ShortCodeExistsFunc: func(ctx context.Context, shortUrl string) (bool, error) {
			return false, nil
		},
		CreateShortenUrlFunc: func(ctx context.Context, longUrl string, shortUrl string) (string, error) {
			return shortUrl, nil
		},
	}
}

func (m *MockRepository) GetLongUrl(ctx context.Context, shortUrl string) (string, time.Time, error) {
	return m.GetLongUrlFunc(ctx, shortUrl)
}

func (m *MockRepository) ShortCodeExists(ctx context.Context, shortUrl string) (bool, error) {
	return m.ShortCodeExistsFunc(ctx, shortUrl)
}

func (m *MockRepository) CreateShortenUrl(ctx context.Context, longUrl string, shortUrl string) (string, error) {
	return m.CreateShortenUrlFunc(ctx, longUrl, shortUrl)
}

type MockCache struct {
	GetFunc func(ctx context.Context, key string) *redis.StringCmd
	SetFunc func(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd
}

func InitMockCache() *MockCache {
	return &MockCache{
		GetFunc: func(ctx context.Context, key string) *redis.StringCmd {
			return redis.NewStringResult("", redis.Nil)
		},
		SetFunc: func(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
			return redis.NewStatusResult("OK", nil)
		},
	}
}

func (m *MockCache) Get(ctx context.Context, key string) *redis.StringCmd {
	return m.GetFunc(ctx, key)
}

func (m *MockCache) Set(ctx context.Context, key string, value any, expiration time.Duration) *redis.StatusCmd {
	return m.SetFunc(ctx, key, value, expiration)
}