package helper

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/tim8912097887-sys/url-shortener/internal/shared/response/envelope"
	userschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/user_schema"
)

func DecodeResponse[T any](t *testing.T, response *http.Response) T {
	t.Helper()
	defer response.Body.Close()

	var payload T
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatal(err)
	}
	return payload
}

func ErrorCode(t *testing.T, response *http.Response) string {
	t.Helper()
	return DecodeResponse[envelope.ErrorResponse](t, response).Error.Code
}

func SignupUser(t *testing.T, app *App, username, email, password string) *http.Response {
	t.Helper()
	return Request(t, app, http.MethodPost, "/api/v1/users/signup", userschema.SignupRequest{
		Username: username,
		Email:    email,
		Password: password,
	}, "", "")
}

func LoginUser(t *testing.T, app *App, email, password string) (accessToken, refreshToken string, response *http.Response) {
	t.Helper()
	response = Request(t, app, http.MethodPost, "/api/v1/users/login", userschema.LoginRequest{
		Email:    email,
		Password: password,
	}, "", "")
	if response.StatusCode != http.StatusOK {
		return "", "", response
	}

	payload := DecodeResponse[envelope.SuccessResponse](t, response)
	data, ok := payload.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected response data object, got %T", payload.Data)
	}
	accessToken, ok = data["accessToken"].(string)
	if !ok || accessToken == "" {
		t.Fatal("expected access token in login response")
	}

	cookie := response.Header.Get("Set-Cookie")
	refreshToken = strings.Split(strings.TrimPrefix(cookie, "refresh_token="), ";")[0]
	if refreshToken == "" {
		t.Fatal("expected refresh token cookie in login response")
	}
	return accessToken, refreshToken, response
}
