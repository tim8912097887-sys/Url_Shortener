package user

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/tim8912097887-sys/url-shortener/internal/configs"
	usererror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/user_error"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/middleware"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response/envelope"
	writeresponse "github.com/tim8912097887-sys/url-shortener/internal/shared/response/write_response"
	userschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/user_schema"
	jwttoken "github.com/tim8912097887-sys/url-shortener/internal/shared/util/jwt_token"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/validation"
)

const (
	refreshCookieName = "refresh_token"
	// Scope this to your auth router's mount path (e.g. "/api/auth") if you
	// don't want the browser sending the refresh cookie on every request.
	refreshCookiePath = "/"
)

type UserService interface {
	Signup(ctx context.Context, createUserInput userschema.CreateUserInput) error
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context, userID string, tokenVersion int) error
	LogoutAll(ctx context.Context, userID string, tokenVersion int) error
	Refresh(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error)
}

type HandlerConfig struct {
	Logger  *slog.Logger
	Service UserService
	Tokens  jwttoken.TokenManager
    Cfg     *configs.Configs
}

type Handler struct {
	logger  *slog.Logger
	service UserService
	tokens  jwttoken.TokenManager // needed to wire AuthMiddleware in RegisterRoutes
    cfg     *configs.Configs
}

func NewHandler(handlerConfig HandlerConfig) Handler {
	return Handler{
		logger:  handlerConfig.Logger,
		service: handlerConfig.Service,
		tokens:  handlerConfig.Tokens,
		cfg:     handlerConfig.Cfg,
	}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
	router.Post("/signup", h.Signup)
	router.Post("/login", h.Login)
	router.Post("/refresh", h.Refresh)

	router.Post("/logout", middleware.AuthMiddleware(h.tokens, h.logger), h.Logout)
	router.Post("/logout-all", middleware.AuthMiddleware(h.tokens, h.logger), h.LogoutAll)
}

func (h *Handler) Signup(c fiber.Ctx) {
	// Validate input
	validatedInput, err := validation.BindAndValidate[userschema.SignupRequest](c)

	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(envelope.NewErrorResponse(envelope.Error{
			Code:    "INVALID_INPUT",
			Message: err.Error(),
		}))
		return
	}

	if err := h.service.Signup(c.RequestCtx(), userschema.CreateUserInput{
		Email:    validatedInput.Email,
		Username: validatedInput.Username,
		Password: validatedInput.Password,
	}); err != nil {
		writeresponse.ErrorHandler(c, err, h.logger, "failed to signup")
		return
	}

	// Same success response whether or not the email was already registered —
	// signup can never be used to enumerate existing accounts.
	writeresponse.SuccessJson(c, fiber.StatusOK, map[string]string{
		"message": "Successfully signed up",
	})
}

func (h *Handler) Login(c fiber.Ctx) {
	// Validate input
	validatedInput, err := validation.BindAndValidate[userschema.LoginRequest](c)

	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(envelope.NewErrorResponse(envelope.Error{
			Code:    "INVALID_INPUT",
			Message: err.Error(),
		}))
		return
	}

	accessToken, refreshToken, err := h.service.Login(c.RequestCtx(), validatedInput.Email, validatedInput.Password)

	if err != nil {
		writeresponse.ErrorHandler(c, err, h.logger, "failed to login")
		return
	}

	h.setRefreshCookie(c, refreshToken)

	writeresponse.SuccessJson(c, fiber.StatusOK, map[string]string{
		"accessToken": accessToken,
		"message":     "Successfully logged in",
	})
}

func (h *Handler) Refresh(c fiber.Ctx) {
	refreshToken := c.Cookies(refreshCookieName)

	if refreshToken == "" {
		writeresponse.ErrorHandler(c, usererror.ErrInvalidToken, h.logger, "failed to refresh user")
		return
	}

	newAccessToken, newRefreshToken, err := h.service.Refresh(c.RequestCtx(), refreshToken)

	if err != nil {
		writeresponse.ErrorHandler(c, err, h.logger, "failed to refresh user")
		return
	}

	h.setRefreshCookie(c, newRefreshToken)

	writeresponse.SuccessJson(c, fiber.StatusOK, map[string]string{
		"accessToken": newAccessToken,
		"message":     "Successfully refreshed token",
	})
}

// Logout requires AuthMiddleware to have already validated the bearer access
// token and populated Locals with the user id and token version.
func (h *Handler) Logout(c fiber.Ctx) {
	userID, tokenVersion, ok := h.authFromLocals(c)
	if !ok {
		writeresponse.ErrorHandler(c, usererror.ErrInvalidToken, h.logger, "failed to logout user")
		return
	}

	err := h.service.Logout(c.RequestCtx(), userID, tokenVersion)

	if err != nil {
		writeresponse.ErrorHandler(c, err, h.logger, "failed to logout user")
		return
	}

	h.clearRefreshCookie(c)

	writeresponse.SuccessJson(c, fiber.StatusOK, map[string]string{
		"message": "Successfully logged out",
	})
}

// LogoutAll requires AuthMiddleware, same as Logout.
func (h *Handler) LogoutAll(c fiber.Ctx) {
	userID, tokenVersion, ok := h.authFromLocals(c)
	if !ok {
		writeresponse.ErrorHandler(c, usererror.ErrInvalidToken, h.logger, "failed to logout all user")
		return
	}

	err := h.service.LogoutAll(c.RequestCtx(), userID, tokenVersion)

	if err != nil {
		writeresponse.ErrorHandler(c, err, h.logger, "failed to logout all user")
		return
	}

	h.clearRefreshCookie(c)

	writeresponse.SuccessJson(c, fiber.StatusOK, map[string]string{
		"message": "Successfully logged out all",
	})
}

func (h *Handler) authFromLocals(c fiber.Ctx) (userID string, tokenVersion int, ok bool) {
	userID, ok = c.Locals(middleware.LocalsUserID).(string)
	if !ok {
		return "", 0, false
	}

	tokenVersion, ok = c.Locals(middleware.LocalsTokenVersion).(int)
	if !ok {
		return "", 0, false
	}

	return userID, tokenVersion, true
}

func (h *Handler) setRefreshCookie(c fiber.Ctx, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    refreshToken,
		Path:     refreshCookiePath,
		Domain:   h.cfg.CookieDomain,
		Expires:  time.Now().Add(jwttoken.RefreshTokenTTL),
		MaxAge:   int(jwttoken.RefreshTokenTTL.Seconds()),
		HTTPOnly: true,
		Secure:   h.cfg.CookieSecure,
		SameSite: h.cfg.CookieSameSite,
	})
}

func (h *Handler) clearRefreshCookie(c fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     refreshCookiePath,
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
	})
}
