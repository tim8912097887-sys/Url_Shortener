package oauth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tim8912097887-sys/url-shortener/internal/configs"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response/envelope"
	writeresponse "github.com/tim8912097887-sys/url-shortener/internal/shared/response/write_response"
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

type HandlerConfig struct {
	Logger  *slog.Logger
	Service OAuthService
	Cfg     *configs.Configs
}

type Handler struct {
	logger  *slog.Logger
	service OAuthService
	cfg     *configs.Configs
}

func NewHandler(handlerConfig HandlerConfig) Handler {
	return Handler{
		logger:  handlerConfig.Logger,
		service: handlerConfig.Service,
		cfg:     handlerConfig.Cfg,
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
		writeresponse.ErrorHandler(c, err, h.logger,"failed to get authorization url")
		return
	}

	c.Redirect().Status(http.StatusTemporaryRedirect).To(url)
}

func (h *Handler) GoogleCallback(c fiber.Ctx) {
	code := c.Query("code")
	state := c.Query("state")
	

	if code == "" || state == "" {
		writeresponse.ErrorJson(c, fiber.StatusBadRequest, envelope.Error{Code: "INVALID_INPUT", Message: "invalid input"})
		return
	}

	tokenResponse, err := h.service.Authenticate(
		c.RequestCtx(),
		ProviderGoogle,
		state,
		code,
	)
	if err != nil {
		writeresponse.ErrorHandler(c, err, h.logger,"failed to authenticate user")
		return
	}


    h.setRefreshCookie(c, tokenResponse.RefreshToken)

	redirectURL := fmt.Sprintf("%s/oauth-callback#access_token=%s", h.cfg.ClientOrigin, tokenResponse.AccessToken)
   
	// Prevent client caching on auth callback routes
    c.Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
    c.Set("Pragma", "no-cache")
	c.Redirect().Status(http.StatusSeeOther).To(redirectURL)
}

func (h *Handler) setRefreshCookie(c fiber.Ctx, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/api/v1",
		Expires:  time.Now().Add(jwttoken.RefreshTokenTTL),
		MaxAge:   int(jwttoken.RefreshTokenTTL.Seconds()),
		HTTPOnly: true,
		Secure:  true,
		SameSite: h.cfg.CookieSameSite,
	})
}


