package usertest

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/tim8912097887-sys/url-shortener/internal/shared/response/envelope"
	userschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/user_schema"
	"github.com/tim8912097887-sys/url-shortener/tests/helper"
)

func decode[T any](t *testing.T, response *http.Response) T {
	t.Helper()
	defer response.Body.Close()
	var payload T
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func errorCode(t *testing.T, response *http.Response) string {
	return decode[envelope.ErrorResponse](t, response).Error.Code
}

func signup(t *testing.T, app *helper.App, username, email, password string) *http.Response {
	t.Helper()
	return helper.Request(t, app, http.MethodPost, "/api/v1/users/signup", userschema.SignupRequest{
		Username: username,
		Email:    email,
		Password: password,
	}, "", "")
}

func login(t *testing.T, app *helper.App, email, password string) (string, string, *http.Response) {
	t.Helper()
	response := helper.Request(t, app, http.MethodPost, "/api/v1/users/login", userschema.LoginRequest{
		Email:    email,
		Password: password,
	}, "", "")
	if response.StatusCode != http.StatusOK {
		return "", "", response
	}
	payload := decode[envelope.SuccessResponse](t, response)
	data := payload.Data.(map[string]any)
	accessToken := data["accessToken"].(string)
	refreshToken := response.Header.Get("Set-Cookie")
	refreshToken = strings.Split(strings.TrimPrefix(refreshToken, "refresh_token="), ";")[0]
	return accessToken, refreshToken, response
}

func TestSignupEndpoint(t *testing.T) {
	app := helper.NewApp(t)
	cases := []struct {
		name     string
		payload  any
		status   int
		code     string
		verifyDB bool
	}{
		{"missing username", map[string]string{"email": "a@example.com", "password": "password1"}, http.StatusBadRequest, "INVALID_INPUT", false},
		{"invalid email", userschema.SignupRequest{Username: "alice", Email: "bad", Password: "password1"}, http.StatusBadRequest, "INVALID_INPUT", false},
		{"short password", userschema.SignupRequest{Username: "alice", Email: "a@example.com", Password: "short"}, http.StatusBadRequest, "INVALID_INPUT", false},
		{"valid signup", userschema.SignupRequest{Username: "alice", Email: "a@example.com", Password: "password1"}, http.StatusOK, "", true},
		{"duplicate email is idempotent", userschema.SignupRequest{Username: "other", Email: "a@example.com", Password: "password1"}, http.StatusOK, "", true},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			
			if test.name == "duplicate email is idempotent" {
				if response := signup(t, app, "alice", "a@example.com", "password1"); response.StatusCode != http.StatusOK {
					t.Fatalf("seed signup: got %d", response.StatusCode)
				}
			}
			response := helper.Request(t, app, http.MethodPost, "/api/v1/users/signup", test.payload, "", "")
			if response.StatusCode != test.status {
				t.Fatalf("expected status %d, got %d", test.status, response.StatusCode)
			}
			if test.code != "" && errorCode(t, response) != test.code {
				t.Fatalf("expected error %s", test.code)
			}
			if test.verifyDB {
				var count int
				if err := app.Pool.QueryRow(context.Background(), "SELECT count(*) FROM users WHERE email = 'a@example.com'").Scan(&count); err != nil {
					t.Fatal(err)
				}
				if count != 1 {
					t.Fatalf("expected one user, got %d", count)
				}
			}
			response.Body.Close()
			helper.Cleanup(t, app.Pool, app.Cache)
		})
	}
}

func TestLoginEndpoint(t *testing.T) {
	app := helper.NewApp(t)
	cases := []struct {
		name     string
		email    string
		password string
		status   int
		code     string
	}{
		{"missing email", "", "password1", http.StatusBadRequest, "INVALID_INPUT"},
		{"invalid email", "bad", "password1", http.StatusBadRequest, "INVALID_INPUT"},
		{"short password", "alice@example.com", "short", http.StatusBadRequest, "INVALID_INPUT"},
		{"unknown user", "unknown@example.com", "password1", http.StatusBadRequest, "INVALID_CREDENTIAL"},
		{"wrong password", "alice@example.com", "wrongpass", http.StatusBadRequest, "INVALID_CREDENTIAL"},
		{"valid login", "alice@example.com", "password1", http.StatusOK, ""},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			
			if response := signup(t, app, "alice", "alice@example.com", "password1"); response.StatusCode != http.StatusOK {
				t.Fatalf("seed signup: got %d", response.StatusCode)
			}
			_, _, response := login(t, app, test.email, test.password)
			if response.StatusCode != test.status {
				t.Fatalf("expected status %d, got %d", test.status, response.StatusCode)
			}
			if test.code != "" && errorCode(t, response) != test.code {
				t.Fatalf("expected error %s", test.code)
			}
			helper.Cleanup(t, app.Pool, app.Cache)
		})
	}
}

func TestRefreshEndpoint(t *testing.T) {
	app := helper.NewApp(t)
	cases := []struct {
		name   string
		cookie string
		status int
		code   string
	}{
		{"missing cookie", "", http.StatusUnauthorized, "INVALID_TOKEN"},
		{"malformed cookie", "not-a-token", http.StatusUnauthorized, "INVALID_TOKEN"},
		{"valid cookie", "valid", http.StatusOK, ""},
		{"unknown user cookie", "unknown-user", http.StatusUnauthorized, "INVALID_TOKEN"},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			
			if response := signup(t, app, "alice", "alice@example.com", "password1"); response.StatusCode != http.StatusOK {
				t.Fatalf("seed signup: got %d", response.StatusCode)
			}
			_, refresh, response := login(t, app, "alice@example.com", "password1")
			cookie := test.cookie
			if cookie == "valid" {
				cookie = refresh
			} else if cookie == "unknown-user" {
				generatedToken, err := app.Tokens.GenerateRefreshToken("00000000-0000-0000-0000-000000000001", 0)
				if err != nil {
					t.Fatal(err)
				}
				cookie = generatedToken
			}
			response = helper.Request(t, app, http.MethodPost, "/api/v1/users/refresh", nil, "", cookie)
			if response.StatusCode != test.status {
				t.Fatalf("expected status %d, got %d", test.status, response.StatusCode)
			}
			if test.code != "" && errorCode(t, response) != test.code {
				t.Fatalf("expected error %s", test.code)
			}
			
			helper.Cleanup(t, app.Pool, app.Cache)
		})
	}

	t.Run("revoked refresh token", func(t *testing.T) {
	
		if response := signup(t, app, "alice", "alice@example.com", "password1"); response.StatusCode != http.StatusOK {
			t.Fatalf("seed signup: got %d", response.StatusCode)
		}
		access, refresh, response := login(t, app, "alice@example.com", "password1")
		if response.StatusCode != http.StatusOK {
			t.Fatalf("seed login: got %d", response.StatusCode)
		}
		if response := helper.Request(t, app, http.MethodPost, "/api/v1/users/logout-all", nil, "Bearer "+access, ""); response.StatusCode != http.StatusOK {
			t.Fatalf("logout-all: got %d", response.StatusCode)
		}
		response = helper.Request(t, app, http.MethodPost, "/api/v1/users/refresh", nil, "", refresh)
		if response.StatusCode != http.StatusUnauthorized || errorCode(t, response) != "INVALID_TOKEN" {
			t.Fatalf("expected revoked refresh token, got %d", response.StatusCode)
		}
		response.Body.Close()
		helper.Cleanup(t, app.Pool, app.Cache)
	})
}

func TestLogoutEndpoints(t *testing.T) {
	app := helper.NewApp(t)
	for _, endpoint := range []string{"logout", "logout-all"} {
		t.Run(endpoint, func(t *testing.T) {
			
			if response := signup(t, app, "alice", "alice@example.com", "password1"); response.StatusCode != http.StatusOK {
				t.Fatalf("seed signup: got %d", response.StatusCode)
			}
			access, _, response := login(t, app, "alice@example.com", "password1")
			if response.StatusCode != http.StatusOK {
				t.Fatalf("seed login: got %d", response.StatusCode)
			}

			path := "/api/v1/users/" + endpoint
			cases := []struct {
				name   string
				auth   string
				status int
				code   string
			}{
				{"missing token", "", http.StatusUnauthorized, "INVALID_TOKEN"},
				{"wrong auth scheme", "Basic invalid", http.StatusUnauthorized, "INVALID_TOKEN"},
				{"malformed token", "Bearer invalid", http.StatusUnauthorized, "INVALID_TOKEN"},
				{"valid token", "Bearer " + access, http.StatusOK, ""},
			}
			for _, test := range cases {
				t.Run(test.name, func(t *testing.T) {
					response := helper.Request(t, app, http.MethodPost, path, nil, test.auth, "")
					if response.StatusCode != test.status {
						t.Fatalf("expected status %d, got %d", test.status, response.StatusCode)
					}
					if test.code != "" && errorCode(t, response) != test.code {
						t.Fatalf("expected error %s", test.code)
					}
					
				})
			}
			unknownToken, err := app.Tokens.GenerateAccessToken("00000000-0000-0000-0000-000000000001", 0)
			if err != nil {
				t.Fatal(err)
			}
			response = helper.Request(t, app, http.MethodPost, path, nil, "Bearer "+unknownToken, "")
			if response.StatusCode != http.StatusUnauthorized || errorCode(t, response) != "INVALID_TOKEN" {
				t.Fatalf("expected unknown user rejection, got %d", response.StatusCode)
			}

			if endpoint == "logout-all" {
				response := helper.Request(t, app, http.MethodPost, path, nil, "Bearer "+access, "")
				if response.StatusCode != http.StatusUnauthorized || errorCode(t, response) != "INVALID_TOKEN" {
					t.Fatalf("expected stale token rejection, got %d", response.StatusCode)
				}
			}
		
			helper.Cleanup(t, app.Pool, app.Cache)
		})
	}
}
