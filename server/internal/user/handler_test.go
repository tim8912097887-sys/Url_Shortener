package user_test

import (
	"bytes"
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
	usererror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/user_error"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response/envelope"
	userschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/user_schema"
	jwttoken "github.com/tim8912097887-sys/url-shortener/internal/shared/util/jwt_token"
	"github.com/tim8912097887-sys/url-shortener/internal/user"
	"golang.org/x/crypto/bcrypt"
)

type mockRepository struct {
	createFunc    func(context.Context, userschema.UserInsert) (*userschema.CreateUserRepositoryResponse, error)
	byEmailFunc   func(context.Context, string) (*userschema.GetUserbyEmailRepositoryResponse, error)
	byIDFunc      func(context.Context, string) (*userschema.GetUserByIDRepositoryResponse, error)
	incrementFunc func(context.Context, string) (int, error)
}

func testPasswordHash() string {
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	return string(hash)
}

func newRepository() *mockRepository {
	return &mockRepository{
		createFunc: func(context.Context, userschema.UserInsert) (*userschema.CreateUserRepositoryResponse, error) {
			return &userschema.CreateUserRepositoryResponse{}, nil
		},
		byEmailFunc: func(context.Context, string) (*userschema.GetUserbyEmailRepositoryResponse, error) {
			return &userschema.GetUserbyEmailRepositoryResponse{ID: "user-1", Email: "a@example.com", PasswordHash: testPasswordHash(), TokenVersion: 2}, nil
		},
		byIDFunc: func(context.Context, string) (*userschema.GetUserByIDRepositoryResponse, error) {
			return &userschema.GetUserByIDRepositoryResponse{ID: "user-1", TokenVersion: 2}, nil
		},
		incrementFunc: func(context.Context, string) (int, error) { return 3, nil },
	}
}

func (m *mockRepository) CreateUser(ctx context.Context, input userschema.UserInsert) (*userschema.CreateUserRepositoryResponse, error) {
	return m.createFunc(ctx, input)
}
func (m *mockRepository) GetUserByEmail(ctx context.Context, email string) (*userschema.GetUserbyEmailRepositoryResponse, error) {
	return m.byEmailFunc(ctx, email)
}
func (m *mockRepository) GetUserByID(ctx context.Context, id string) (*userschema.GetUserByIDRepositoryResponse, error) {
	return m.byIDFunc(ctx, id)
}
func (m *mockRepository) IncrementTokenVersion(ctx context.Context, id string) (int, error) {
	return m.incrementFunc(ctx, id)
}

func newApp(t *testing.T, repository *mockRepository) (*fiber.App, *jwttoken.TokenManager) {
	t.Helper()
	tokens := jwttoken.NewTokenManager("access-secret", "refresh-secret")
	service := user.NewService(&user.ServiceConfig{Repository: repository, Tokens: *tokens, Logger: slog.Default()})
	handler := user.NewHandler(user.HandlerConfig{Logger: slog.Default(), Service: service, Tokens: jwttoken.TokenManager(*tokens), Cfg: &configs.Configs{CookieDomain: "example.com", CookieSecure: false, CookieSameSite: "Lax"}})
	app := fiber.New()
	handler.RegisterRoutes(app.Group("/api/v1/users"))
	return app, tokens
}

func request(t *testing.T, app *fiber.App, method, path string, payload any, cookie, auth string) *http.Response {
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
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != "" {
		req.Header.Set("Cookie", "refresh_token="+cookie)
	}
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	resp, err := app.Test(req, fiber.TestConfig{Timeout: -1})
	if err != nil {
		t.Fatal(err)
	}
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

func TestSignup(t *testing.T) {
	for _, tc := range []struct {
		name    string
		repoErr error
		status  int
		code    string
	}{{"invalid input", nil, 400, "INVALID_INPUT"}, {"repository failure", errors.New("db down"), 500, "INTERNAL_SERVER_ERROR"}} {
		t.Run(tc.name, func(t *testing.T) {
			repository := newRepository()
			repository.createFunc = func(context.Context, userschema.UserInsert) (*userschema.CreateUserRepositoryResponse, error) {
				return nil, tc.repoErr
			}
			app, _ := newApp(t, repository)
			payload := any(userschema.SignupRequest{Username: "alice", Email: "a@example.com", Password: "password1"})
			if tc.name == "invalid input" {
				payload = map[string]string{"email": "bad"}
			}
			resp := request(t, app, http.MethodPost, "/api/v1/users/signup", payload, "", "")
			if resp.StatusCode != tc.status || errorResponse(t, resp).Error.Code != tc.code {
				t.Fatalf("expected %d/%s, got %d", tc.status, tc.code, resp.StatusCode)
			}
		})
	}
	app, _ := newApp(t, newRepository())
	if resp := request(t, app, http.MethodPost, "/api/v1/users/signup", userschema.SignupRequest{Username: "alice", Email: "a@example.com", Password: "password1"}, "", ""); resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestLogin(t *testing.T) {
	for _, tc := range []struct {
		name     string
		repoErr  error
		password string
		status   int
		code     string
	}{{"invalid input", nil, "", 400, "INVALID_INPUT"}, {"unknown user", usererror.ErrUserNotFound, "password1", 400, "INVALID_CREDENTIAL"}, {"repository failure", errors.New("db down"), "password1", 500, "INTERNAL_SERVER_ERROR"}, {"wrong password", nil, "wrongpass", 400, "INVALID_CREDENTIAL"}} {
		t.Run(tc.name, func(t *testing.T) {
			repository := newRepository()
			repository.byEmailFunc = func(context.Context, string) (*userschema.GetUserbyEmailRepositoryResponse, error) {
				if tc.repoErr != nil {
					return nil, tc.repoErr
				}
				return &userschema.GetUserbyEmailRepositoryResponse{ID: "user-1", Email: "a@example.com", PasswordHash: testPasswordHash(), TokenVersion: 2}, nil
			}
			app, _ := newApp(t, repository)
			payload := any(userschema.LoginRequest{Email: "a@example.com", Password: tc.password})
			if tc.name == "invalid input" {
				payload = map[string]string{"email": "bad"}
			}
			resp := request(t, app, http.MethodPost, "/api/v1/users/login", payload, "", "")
			if resp.StatusCode != tc.status || errorResponse(t, resp).Error.Code != tc.code {
				t.Fatalf("expected %d/%s, got %d", tc.status, tc.code, resp.StatusCode)
			}
		})
	}
	app, _ := newApp(t, newRepository())
	resp := request(t, app, http.MethodPost, "/api/v1/users/login", userschema.LoginRequest{Email: "a@example.com", Password: "password"}, "", "")
	if resp.StatusCode != 200 || !strings.Contains(resp.Header.Get("Set-Cookie"), "refresh_token=") {
		t.Fatalf("expected successful login, got %d", resp.StatusCode)
	}
}

func TestRefresh(t *testing.T) {
	repository := newRepository()
	app, tokens := newApp(t, repository)
	if resp := request(t, app, http.MethodPost, "/api/v1/users/refresh", nil, "", ""); resp.StatusCode != 401 {
		t.Fatalf("missing cookie: got %d", resp.StatusCode)
	}
	refresh, err := tokens.GenerateRefreshToken("user-1", 2)
	if err != nil {
		t.Fatal(err)
	}
	if resp := request(t, app, http.MethodPost, "/api/v1/users/refresh", nil, refresh, ""); resp.StatusCode != 200 {
		t.Fatalf("valid cookie: got %d", resp.StatusCode)
	}
	repository.byIDFunc = func(context.Context, string) (*userschema.GetUserByIDRepositoryResponse, error) {
		return nil, usererror.ErrUserNotFound
	}
	if resp := request(t, app, http.MethodPost, "/api/v1/users/refresh", nil, refresh, ""); resp.StatusCode != 401 {
		t.Fatalf("unknown user: got %d", resp.StatusCode)
	}
	repository.byIDFunc = func(context.Context, string) (*userschema.GetUserByIDRepositoryResponse, error) {
		return &userschema.GetUserByIDRepositoryResponse{ID: "user-1", TokenVersion: 12}, nil
	}
	if resp := request(t, app, http.MethodPost, "/api/v1/users/refresh", nil, refresh, ""); resp.StatusCode != 401 {
		t.Fatalf("stale token: got %d", resp.StatusCode)
	}
}

func TestLogoutRoutes(t *testing.T) {
	for _, path := range []string{"/api/v1/users/logout", "/api/v1/users/logout-all"} {
		t.Run(path, func(t *testing.T) {
			repository := newRepository()
			app, tokens := newApp(t, repository)
			access, err := tokens.GenerateAccessToken("user-1", 2)
			if err != nil {
				t.Fatal(err)
			}
			if resp := request(t, app, http.MethodPost, path, nil, "", ""); resp.StatusCode != 401 {
				t.Fatalf("missing auth: got %d", resp.StatusCode)
			}
			if resp := request(t, app, http.MethodPost, path, nil, "", "Bearer "+access); resp.StatusCode != 200 {
				t.Fatalf("valid auth: got %d", resp.StatusCode)
			}
			repository.byIDFunc = func(context.Context, string) (*userschema.GetUserByIDRepositoryResponse, error) {
				return nil, errors.New("db down")
			}
			if resp := request(t, app, http.MethodPost, path, nil, "", "Bearer "+access); resp.StatusCode != 500 {
				t.Fatalf("repository failure: got %d", resp.StatusCode)
			}
			repository.byIDFunc = func(context.Context, string) (*userschema.GetUserByIDRepositoryResponse, error) {
				return &userschema.GetUserByIDRepositoryResponse{ID: "user-1", TokenVersion: 12}, nil
			}
			if resp := request(t, app, http.MethodPost, path, nil, "", "Bearer "+access); resp.StatusCode != 401 {
				t.Fatalf("stale token: got %d", resp.StatusCode)
			}
			repository.byIDFunc = func(context.Context, string) (*userschema.GetUserByIDRepositoryResponse, error) {
				return nil, usererror.ErrUserNotFound
			}
			if resp := request(t, app, http.MethodPost, path, nil, "", "Bearer "+access); resp.StatusCode != 401 {
				t.Fatalf("unknown user: got %d", resp.StatusCode)
			}
		})
	}
}
