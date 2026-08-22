package writeresponse

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response/envelope"
)

func SuccessJson(c fiber.Ctx, status int, successBody any) error {
	payload := envelope.NewSuccessResponse(successBody)
	return c.Status(status).JSON(payload)
}