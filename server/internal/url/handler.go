package url

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response"
)

type UrlService interface {
	ShortenUrl(url string) (string, error)
	GetUrl(shortUrl string) (string, error)
}

type handler struct {
	logger  *slog.Logger
	service UrlService
}

func NewHandler(logger *slog.Logger, service UrlService) handler {
	return handler{
		logger:  logger,
		service: service,
	}
}

func (h *handler) RegisterRoutes(router fiber.Router) {
     router.Post("/",h.ShortenUrl)
	 router.Get("/:short_url",h.GetUrl)
}

func (h *handler) ShortenUrl(c fiber.Ctx) {
	// Validate input
	validatedInput, err := BindAndValidate[CreateUrlSchema](c)

	if err != nil {
		h.logger.Error("failed to validate input",slog.Any("error", err),slog.String("context","shorten url"))
		c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("invalid_input", err.Error()))
		return
	}

	shortUrl, err := h.service.ShortenUrl(validatedInput.Url)

	if err != nil {
		h.logger.Error("failed to shorten url",slog.Any("error", err),slog.String("context","shorten url"))
		c.Status(fiber.StatusInternalServerError).JSON(response.NewErrorResponse("internal_error", err.Error()))
		return
	}

	c.Status(fiber.StatusOK).JSON(response.NewSuccessResponse(map[string]string{"shortUrl": shortUrl, "message": "Successfully shorten url"}))
}

func (h *handler) GetUrl(c fiber.Ctx) {
	// Validate params
	validatedParams, err := Validate(GetUrlParams{ShortURL: c.Params("short_url")})

	if err != nil {
		h.logger.Error("failed to validate params",slog.Any("error", err),slog.String("context","get long url"))
		c.Status(fiber.StatusBadRequest).JSON(response.NewErrorResponse("invalid_input", err.Error()))
		return
	}

	longUrl, err := h.service.GetUrl(validatedParams.ShortURL)

	// Handle business logic error
	if err == ErrUrlNotFound {
		h.logger.Error("failed to get long url",slog.Any("error", err),slog.String("context","get long url"))
		c.Status(fiber.StatusNotFound).JSON(response.NewErrorResponse("url_not_found", err.Error()))
		return
	}

	if err != nil {
		h.logger.Error("failed to get long url",slog.Any("error", err),slog.String("context","get long url"))
		c.Status(fiber.StatusInternalServerError).JSON(response.NewErrorResponse("internal_error", err.Error()))
		return
	}

	c.Redirect().Status(fiber.StatusTemporaryRedirect).To(longUrl)
}