package url

import (
	"context"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response/envelope"
	writeresponse "github.com/tim8912097887-sys/url-shortener/internal/shared/response/write_response"
)

type UrlService interface {
	ShortenUrl(ctx context.Context, url string) (string, error)
	GetUrl(ctx context.Context, shortUrl string) (string, error)
}

type HandlerConfig struct {
	Logger  *slog.Logger
	Service UrlService
}

type Handler struct {
	logger  *slog.Logger
	service UrlService
}

func NewHandler(handlerConfig HandlerConfig) Handler {
	return Handler{
		logger:  handlerConfig.Logger,
		service: handlerConfig.Service,
	}
}

func (h *Handler) RegisterRoutes(router fiber.Router) {
     router.Post("/",h.ShortenUrl)
	 router.Get("/:short_url",h.GetUrl)
}

func (h *Handler) ShortenUrl(c fiber.Ctx) {
	// Validate input
	validatedInput, err := BindAndValidate[CreateUrlSchema](c)

	if err != nil {
		c.Status(fiber.StatusBadRequest).JSON(envelope.NewErrorResponse(envelope.Error{Code: "INVALID_INPUT", Message: err.Error()}))
		return
	}

	shortUrl, err := h.service.ShortenUrl(c.RequestCtx(),validatedInput.Url)

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

	longUrl, err := h.service.GetUrl(c.RequestCtx(),validatedParams.ShortURL)

	if err != nil {
		writeresponse.ErrorHandler(c, err, h.logger, "failed to get url")
		return
	}

	c.Redirect().Status(fiber.StatusTemporaryRedirect).To(longUrl)
}