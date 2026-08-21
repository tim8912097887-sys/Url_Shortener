package user

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	usererror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/user_error"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/middleware"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response"
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

type Handler struct {
	logger  *slog.Logger
	service UserService
	tokens  jwttoken.TokenManager // needed to wire AuthMiddleware in RegisterRoutes
}

func NewHandler(logger *slog.Logger, service UserService, tokens jwttoken.TokenManager) Handler {
	return Handler{
		logger:  logger,
		service: service,
		tokens:  tokens,
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
		h.logger.Error("failed to validate input", slog.Any("error", err), slog.String("context", "signup"))
		c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("invalid_input", err.Error()))
		return
	}

	if err := h.service.Signup(c.RequestCtx(), userschema.CreateUserInput{
		Email:    validatedInput.Email,
		Username: validatedInput.Username,
		Password: validatedInput.Password,
	}); err != nil {
		h.logger.Error("failed to sign up", slog.Any("error", err), slog.String("context", "signup"))
		c.Status(fiber.StatusInternalServerError).JSON(response.NewErrorResponse("internal_error", err.Error()))
		return
	}

	// Same success response whether or not the email was already registered —
	// signup can never be used to enumerate existing accounts.
	c.Status(fiber.StatusOK).JSON(response.NewSuccessResponse(map[string]string{
		"message": "Successfully signed up",
	}))
}

func (h *Handler) Login(c fiber.Ctx) {
	// Validate input
	validatedInput, err := validation.BindAndValidate[userschema.LoginRequest](c)

	if err != nil {
		h.logger.Error("failed to validate input", slog.Any("error", err), slog.String("context", "login"))
		c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("invalid_input", err.Error()))
		return
	}

	accessToken, refreshToken, err := h.service.Login(c.RequestCtx(), validatedInput.Email, validatedInput.Password)

	if errors.Is(err, usererror.ErrInvalidCredential) {
		h.logger.Error("failed to login", slog.Any("error", err), slog.String("context", "login"))
		c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("invalid_credential", err.Error()))
		return
	}

	if err != nil {
		h.logger.Error("failed to login", slog.Any("error", err), slog.String("context", "login"))
		c.Status(fiber.StatusInternalServerError).JSON(response.NewErrorResponse("internal_error", err.Error()))
		return
	}

	h.setRefreshCookie(c, refreshToken)

	c.Status(fiber.StatusOK).JSON(response.NewSuccessResponse(map[string]string{
		"accessToken": accessToken,
		"message":     "Successfully logged in",
	}))
}

func (h *Handler) Refresh(c fiber.Ctx) {
	refreshToken := c.Cookies(refreshCookieName)

	if refreshToken == "" {
		h.logger.Error("missing refresh token cookie", slog.String("context", "refresh"))
		c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("invalid_token", usererror.ErrInvalidToken.Error()))
		return
	}

	newAccessToken, newRefreshToken, err := h.service.Refresh(c.RequestCtx(), refreshToken)

	if errors.Is(err, usererror.ErrInvalidToken) {
		h.logger.Error("failed to refresh token", slog.Any("error", err), slog.String("context", "refresh"))
		c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("invalid_token", err.Error()))
		return
	}

	if err != nil {
		h.logger.Error("failed to refresh token", slog.Any("error", err), slog.String("context", "refresh"))
		c.Status(fiber.StatusInternalServerError).JSON(response.NewErrorResponse("internal_error", err.Error()))
		return
	}

	h.setRefreshCookie(c, newRefreshToken)

	c.Status(fiber.StatusOK).JSON(response.NewSuccessResponse(map[string]string{
		"accessToken": newAccessToken,
		"message":     "Successfully refreshed token",
	}))
}

// Logout requires AuthMiddleware to have already validated the bearer access
// token and populated Locals with the user id and token version.
func (h *Handler) Logout(c fiber.Ctx) {
	userID, tokenVersion, ok := h.authFromLocals(c)
	if !ok {
		h.logger.Error("missing auth locals", slog.String("context", "logout"))
		c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("invalid_token", usererror.ErrInvalidToken.Error()))
		return
	}

	err := h.service.Logout(c.RequestCtx(), userID, tokenVersion)

	if errors.Is(err, usererror.ErrInvalidToken) {
		h.logger.Error("failed to logout", slog.Any("error", err), slog.String("context", "logout"))
		c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("invalid_token", err.Error()))
		return
	}

	if err != nil {
		h.logger.Error("failed to logout", slog.Any("error", err), slog.String("context", "logout"))
		c.Status(fiber.StatusInternalServerError).JSON(response.NewErrorResponse("internal_error", err.Error()))
		return
	}

	h.clearRefreshCookie(c)

	c.Status(fiber.StatusOK).JSON(response.NewSuccessResponse(map[string]string{
		"message": "Successfully logged out",
	}))
}

// LogoutAll requires AuthMiddleware, same as Logout.
func (h *Handler) LogoutAll(c fiber.Ctx) {
	userID, tokenVersion, ok := h.authFromLocals(c)
	if !ok {
		h.logger.Error("missing auth locals", slog.String("context", "logout all"))
		c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("invalid_token", usererror.ErrInvalidToken.Error()))
		return
	}

	err := h.service.LogoutAll(c.RequestCtx(), userID, tokenVersion)

	if errors.Is(err, usererror.ErrInvalidToken) {
		h.logger.Error("failed to logout all", slog.Any("error", err), slog.String("context", "logout all"))
		c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("invalid_token", err.Error()))
		return
	}

	if err != nil {
		h.logger.Error("failed to logout all", slog.Any("error", err), slog.String("context", "logout all"))
		c.Status(fiber.StatusInternalServerError).JSON(response.NewErrorResponse("internal_error", err.Error()))
		return
	}

	h.clearRefreshCookie(c)

	c.Status(fiber.StatusOK).JSON(response.NewSuccessResponse(map[string]string{
		"message": "Successfully logged out of all sessions",
	}))
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
		Expires:  time.Now().Add(jwttoken.RefreshTokenTTL),
		MaxAge:   int(jwttoken.RefreshTokenTTL.Seconds()),
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
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
