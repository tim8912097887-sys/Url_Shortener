package url

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/middleware"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response/envelope"
	writeresponse "github.com/tim8912097887-sys/url-shortener/internal/shared/response/write_response"
	urlschema "github.com/tim8912097887-sys/url-shortener/internal/shared/schema/url_schema"
	jwttoken "github.com/tim8912097887-sys/url-shortener/internal/shared/util/jwt_token"
)

type UrlService interface {
	ShortenUrl(ctx context.Context, url string, authContext urlschema.AuthContext) (string, error)
	GetUrl(ctx context.Context, shortUrl string, authContext urlschema.AuthContext) (string, error)
}

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
     router.Post("/", middleware.AuthMiddleware(h.tokens, h.logger),h.ShortenUrl)
	 router.Get("/:short_url", middleware.AuthMiddleware(h.tokens, h.logger),h.GetUrl)
}

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

	writeresponse.SuccessJson(c, fiber.StatusOK, map[string]string{"shortUrl": shortUrl, "message": "Successfully shorten url"})
}

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

func (h *Handler) authFromLocals(c fiber.Ctx) urlschema.AuthContext {
	userID, ok := c.Locals(middleware.LocalsUserID).(string)
	if !ok || userID == "" {
		return urlschema.AuthContext{
			UserID:          "",
			IsAuthenticated: false,
		}
	}


	return urlschema.AuthContext{
		UserID:          userID,
		IsAuthenticated: true,
	}
}