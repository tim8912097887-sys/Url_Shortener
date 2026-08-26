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
	refreshCookiePath = "/api/v1"
)

type UserService interface {
	Signup(ctx context.Context, createUserInput userschema.CreateUserInput) error
	Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error)
	Logout(ctx context.Context, userID string, tokenVersion int) error
	LogoutAll(ctx context.Context, userID string, tokenVersion int) error
	Refresh(ctx context.Context, userID string, tokenVersion int) (newAccessToken, newRefreshToken string, err error)
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
	router.Post("/refresh", middleware.RefreshTokenMiddleware(h.tokens, h.logger), h.Refresh)

	router.Post("/logout", middleware.AccessTokenMiddleware(h.tokens, h.logger), h.Logout)
	router.Post("/logout-all", middleware.AccessTokenMiddleware(h.tokens, h.logger), h.LogoutAll)
}

// Signup godoc
//
// @Summary Create user account
// @Description Registers a new user account.
// @Description
// @Description For security, this endpoint returns the same successful response
// @Description whether the email is newly registered or already exists.
// @Description This prevents account enumeration.
//
// @Tags Users
// @Accept json
// @Produce json
//
// @Param request body userschema.SignupRequest true "Signup request"
//
// @Success 201 {object} userschema.SignupResponse
//
// @Failure 400 {object} envelope.ErrorResponse "Invalid request"
// @Failure 500 {object} envelope.ErrorResponse "Internal server error"
//
// @Router /users/signup [post]
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
	writeresponse.SuccessJson(c, fiber.StatusCreated, userschema.SignupResponse{
		Message: "Successfully signed up",
	})
}

// Login godoc
//
// @Summary Login
// @Description Authenticates a user with email and password.
// @Description
// @Description Returns a short-lived access token in the response body.
// @Description A refresh token is stored in an HttpOnly Secure cookie.
// @Description
// @Description For security, invalid email and invalid password return the same error.
//
// @Tags Users
// @Accept json
// @Produce json
//
// @Param request body userschema.LoginRequest true "Login credentials"
//
// @Success 200 {object} userschema.LoginResponse
//
// @Failure 400 {object} envelope.ErrorResponse "Invalid request"
// @Failure 401 {object} envelope.ErrorResponse "Invalid credentials"
// @Failure 500 {object} envelope.ErrorResponse "Internal server error"
//
// @Router /users/login [post]
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

	writeresponse.SuccessJson(c, fiber.StatusOK, userschema.LoginResponse{
		AccessToken: accessToken,
		Message:     "Successfully logged in",
	})
}

// Refresh godoc
//
// @Summary Refresh access token
// @Description Issues a new access token using the refresh_token HttpOnly cookie.
// @Description
// @Description The refresh token is rotated after a successful refresh.
// @Description The new refresh token is returned as a replacement HttpOnly Secure cookie.
// @Description
// @Description Clients should call this endpoint when the access token expires.
//
// @Tags Users
// @Produce json
//
// @Success 200 {object} userschema.RefreshResponse
//
// @Failure 401 {object} envelope.ErrorResponse "Missing, expired, or invalid refresh token"
// @Failure 500 {object} envelope.ErrorResponse "Internal server error"
//
// @Router /users/refresh [post]
func (h *Handler) Refresh(c fiber.Ctx) {
	userID, tokenVersion, ok := h.authFromLocals(c)
	if !ok {
		writeresponse.ErrorHandler(c, usererror.ErrInvalidToken, h.logger, "failed to refresh user")
		return
	}

	newAccessToken, newRefreshToken, err := h.service.Refresh(c.RequestCtx(), userID, tokenVersion)

	if err != nil {
		writeresponse.ErrorHandler(c, err, h.logger, "failed to refresh user")
		return
	}

	h.setRefreshCookie(c, newRefreshToken)

	writeresponse.SuccessJson(c, fiber.StatusOK, userschema.RefreshResponse{
		AccessToken: newAccessToken,
		Message:     "Successfully refreshed access token",
	})
}

// Logout godoc
//
// @Summary Logout current device
// @Description Clears the refresh token cookie from the current browser.
// @Description
// @Description This endpoint does not invalidate all tokens server-side.
// @Description Use Logout All Sessions to invalidate tokens across devices.
//
// @Tags Users
// @Produce json
// @Security BearerAuth
//
// @Success 200 {object} userschema.LogoutResponse
//
// @Failure 401 {object} envelope.ErrorResponse "Missing or invalid access token"
// @Failure 500 {object} envelope.ErrorResponse "Internal server error"
//
// @Router /users/logout [post]
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

	writeresponse.SuccessJson(c, fiber.StatusOK, userschema.LogoutResponse{
		Message: "Successfully logged out",
	})
}

// LogoutAll godoc
//
// @Summary Logout all sessions
// @Description Invalidates all active access and refresh tokens for the user.
// @Description
// @Description The user's token version is incremented, causing previously issued
// @Description access and refresh tokens to become invalid.
//
// @Tags Users
// @Produce json
// @Security BearerAuth
//
// @Success 200 {object} userschema.LogoutAllResponse
//
// @Failure 401 {object} envelope.ErrorResponse "Missing or invalid access token"
// @Failure 500 {object} envelope.ErrorResponse "Internal server error"
//
// @Router /users/logout-all [post]
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

	writeresponse.SuccessJson(c, fiber.StatusOK, userschema.LogoutAllResponse{
		Message: "Successfully logged out all sessions",
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
		Expires:  time.Now().Add(jwttoken.RefreshTokenTTL),
		MaxAge:   int(jwttoken.RefreshTokenTTL.Seconds()),
		HTTPOnly: true,
		Secure:   true,
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
		SameSite: h.cfg.CookieSameSite,
	})
}
