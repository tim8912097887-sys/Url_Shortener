package url

import (
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	usererror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/user_error"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/middleware"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response/envelope"
	writeresponse "github.com/tim8912097887-sys/url-shortener/internal/shared/response/write_response"
	urlschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/url_schema"
	jwttoken "github.com/tim8912097887-sys/url-shortener/internal/shared/util/jwt_token"
)

type HandlerConfig struct {
	Logger  *slog.Logger
	Service UrlService
	Tokens  jwttoken.TokenManager
}

type Handler struct {
	logger  *slog.Logger
	service UrlService
	tokens  jwttoken.TokenManager
}

func NewHandler(handlerConfig HandlerConfig) Handler {
	return Handler{
		logger:  handlerConfig.Logger,
		service: handlerConfig.Service,
		tokens:  handlerConfig.Tokens,
	}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
     router.Post("/", middleware.AccessTokenMiddleware(h.tokens, h.logger),h.ShortenUrl)
	 router.Get("/:short_url", middleware.RefreshTokenMiddleware(h.tokens, h.logger), h.GetUrl)
	 router.Get("/", middleware.AccessTokenMiddleware(h.tokens, h.logger),h.GetUrlsForUser)
}

// ShortenUrl godoc
//
// @Summary Create a shortened URL
// @Description Creates a shortened URL.
// @Description
// @Description The URL expiration depends on authentication status:
// @Description - Authenticated users: longer expiration period.
// @Description - Anonymous users: shorter expiration period.
// @Description
// @Tags URLs
// @Accept json
// @Produce json
//
// @Param request body url.CreateUrlSchema true "URL to shorten"
//
// @Success 200 {object} urlschema.ShortenUrlResponse
//
// @Failure 400 {object} envelope.ErrorResponse "Invalid request body or URL"
// @Failure 429 {object} envelope.ErrorResponse "Rate limit exceeded"
// @Failure 500 {object} envelope.ErrorResponse "Internal server error"
//
// @Router /urls [post]
func (h *Handler) ShortenUrl(c fiber.Ctx) {
	// Validate input
	validatedInput, err := BindAndValidate[CreateUrlSchema](c)

	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(envelope.NewErrorResponse(envelope.Error{Code: "INVALID_INPUT", Message: err.Error()}))
		return
	}

	// Get user id from fiber context locals
	authContext := h.authFromLocals(c)

	shortUrl, err := h.service.ShortenUrl(c.RequestCtx(),validatedInput.Url, authContext)

	if err != nil {
		writeresponse.ErrorHandler(c, err, h.logger, "failed to shorten url")
		return
	}

	writeresponse.SuccessJson(c, fiber.StatusOK, urlschema.ShortenUrlResponse{ShortUrl: shortUrl, Message: "Successfully shorten url"})
}

// GetUrl godoc
//
// @Summary Resolve a shortened URL
// @Description Resolves a shortened URL and redirects the client to the original destination.
// @Description
// @Description The endpoint uses a cache-aside strategy:
// @Description 1. Redis cache is checked first.
// @Description 2. On cache miss, the URL is loaded from PostgreSQL.
// @Description 3. The result is cached with a TTL that never exceeds the URL expiration time.
// @Description
// @Description Expired URLs are treated as not found.
//
// @Tags URLs
// @Produce json
//
// @Param short_url path string true "Short URL code" example(abc12345)
//
// @Success 307 "Temporary Redirect"
//
// @Failure 400 {object} envelope.ErrorResponse "Invalid short URL"
// @Failure 404 {object} envelope.ErrorResponse "URL not found or expired"
// @Failure 500 {object} envelope.ErrorResponse "Internal server error"
//
// @Router /urls/{short_url} [get]
func (h *Handler) GetUrl(c fiber.Ctx) {
	// Validate params
	validatedParams, err := Validate(GetUrlParams{ShortURL: c.Params("short_url")})

	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(envelope.NewErrorResponse(envelope.Error{Code: "INVALID_INPUT", Message: err.Error()}))
		return
	}

	// Get user id from fiber context locals
	authContext := h.authFromLocals(c)

	longUrl, err := h.service.GetUrl(c.RequestCtx(),validatedParams.ShortURL, authContext)

	if err != nil {
		writeresponse.ErrorHandler(c, err, h.logger, "failed to get url")
		return
	}

	c.Redirect().Status(fiber.StatusTemporaryRedirect).To(longUrl)
}

// GetUrlsForUser godoc
//
// @Summary Get current user's shortened URLs
// @Description Returns all shortened URLs owned by the authenticated user.
// @Description
// @Description Only URLs created by the current authenticated user are returned.
//
// @Tags URLs
// @Produce json
// @Security BearerAuth
//
// @Success 200 {object} urlschema.GetUrlsResponse
//
// @Failure 401 {object} envelope.ErrorResponse "Missing or invalid access token"
// @Failure 500 {object} envelope.ErrorResponse "Internal server error"
//
// @Router /urls [get]
func (h *Handler) GetUrlsForUser(c fiber.Ctx) {
	// Get user id from fiber context locals
	authContext := h.authFromLocals(c)

	if !authContext.IsAuthenticated || authContext.UserID == "" {
		writeresponse.ErrorHandler(c, usererror.ErrInvalidToken, h.logger, "failed to get urls for user")
		return
	}
	// Get pagination params
    expiredAt, limit := getPaginationParams(c)

	urls, hasMore, err := h.service.GetUrlsForUser(c.RequestCtx(), authContext.UserID, expiredAt, limit)

	if err != nil {
		writeresponse.ErrorHandler(c, err, h.logger, "failed to get urls for user")
		return
	}

	writeresponse.SuccessJson(c, fiber.StatusOK, urlschema.GetUrlsResponse{Urls: urls, HasMore: hasMore, Message: "Successfully get urls for user"})
}

func (h *Handler) authFromLocals(c fiber.Ctx) urlschema.AuthContext {
	userID, ok := c.Locals(middleware.LocalsUserID).(string)
	if !ok || userID == "" {
		return urlschema.AuthContext{
			UserID:          "",
			TokenVersion:    0,
			IsAuthenticated: false,
		}
	}

	tokenVersion, ok := c.Locals(middleware.LocalsTokenVersion).(int)
	if !ok || tokenVersion == 0 {
		return urlschema.AuthContext{
			UserID:          userID,
			TokenVersion:    0,
			IsAuthenticated: false,
		}
	}

	return urlschema.AuthContext{
		UserID:          userID,
		TokenVersion:    tokenVersion,
		IsAuthenticated: true,
	}
}

func getPaginationParams(c fiber.Ctx) (expiredAt time.Time, limit int) {
	// Default values
	defaultExpiredAt := time.Now().Add(AuthURLExpiry)
	defaultLimit := UrlsMaxLimit

	// Parse 'expiredAt' parameter (expected ISO 8601 / RFC3339 format, e.g., 2026-08-29T15:04:05Z)
	expiredAtStr := c.Query("expiredAt")
	if expiredAtStr != "" {
		parsedTime, err := time.Parse(time.RFC3339Nano, expiredAtStr)
	
		if err == nil {
	        fmt.Println("parsedTime", parsedTime)
			expiredAt = parsedTime
		} else {
			fmt.Println("parsedTime err", err)
			expiredAt = defaultExpiredAt
		}
	} else {
		expiredAt = defaultExpiredAt
	}

	// Parse 'limit' parameter
	limitStr := c.Query("limit")
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil {
			limit = parsedLimit
		} else {
			limit = defaultLimit
		}
	} else {
		limit = defaultLimit
	}

	return expiredAt, limit
}