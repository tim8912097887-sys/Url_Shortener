package oauth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response"
	oauthschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/oauth_schema"
	jwttoken "github.com/tim8912097887-sys/url-shortener/internal/shared/util/jwt_token"
)

type OAuthService interface {
	GetAuthorizationURL(
		ctx context.Context,
		provider oauthschema.Provider,
	) (string, error)

	Authenticate(
		ctx context.Context,
		provider oauthschema.Provider,
		state string,
		code string,
	) (*oauthschema.TokenResponse, error)
}

type Handler struct {
	logger  *slog.Logger
	service OAuthService
}

func NewHandler(logger *slog.Logger, service OAuthService) Handler {
	return Handler{
		logger:  logger,
		service: service,
	}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Get("/google/login", h.GoogleLogin)
	router.Get("/google/callback", h.GoogleCallback)
}

func (h *Handler) GoogleLogin(c fiber.Ctx) {
	url, err := h.service.GetAuthorizationURL(
		c.RequestCtx(),
		ProviderGoogle,
	)
	if err != nil {
		c.Status(fiber.StatusInternalServerError).JSON(response.NewErrorResponse("internal_error", "failed to get authorization url"))
		return
	}

	c.Redirect().Status(http.StatusTemporaryRedirect).To(url)
}

func (h *Handler) GoogleCallback(c fiber.Ctx) {
	code := c.Query("code")
	state := c.Query("state")
	

	if code == "" || state == "" {
		c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("invalid_input", "code or state is empty"))
		return
	}

	tokenResponse, err := h.service.Authenticate(
		c.RequestCtx(),
		ProviderGoogle,
		state,
		code,
	)
	if err != nil {
		c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("oauth_failed", "failed to login with google"))
		return
	}


    h.setRefreshCookie(c, tokenResponse.RefreshToken)

	c.Status(fiber.StatusOK).JSON(response.NewSuccessResponse(map[string]string{
		"accessToken": tokenResponse.AccessToken,
		"message":     "Successfully logged in",
	}))
}

func (h *Handler) setRefreshCookie(c fiber.Ctx, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Expires:  time.Now().Add(jwttoken.RefreshTokenTTL),
		MaxAge:   int(jwttoken.RefreshTokenTTL.Seconds()),
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
	})
}


