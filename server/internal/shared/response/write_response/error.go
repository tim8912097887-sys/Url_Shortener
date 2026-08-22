package writeresponse

import (
	"errors"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	oautherror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/oauth_error"
	urlerror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/url_error"
	usererror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/user_error"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response/envelope"
)

func ErrorHandler(c fiber.Ctx, err error, logger *slog.Logger, message string) {
	switch {
	case errors.Is(err, oautherror.ErrInvalidState):
		ErrorJson(c, fiber.StatusBadRequest, envelope.Error{Code: "INVALID_STATE", Message: err.Error()})
		return
	case errors.Is(err, oautherror.ErrInvalidProvider):
		ErrorJson(c, fiber.StatusBadRequest, envelope.Error{Code: "INVALID_PROVIDER", Message: err.Error()})
		return
	case errors.Is(err, usererror.ErrInvalidCredential):
		ErrorJson(c, fiber.StatusBadRequest, envelope.Error{Code: "INVALID_CREDENTIAL", Message: err.Error()})
		return
	case errors.Is(err, usererror.ErrInvalidToken):
		ErrorJson(c, fiber.StatusUnauthorized, envelope.Error{Code: "INVALID_TOKEN", Message: err.Error()})
		return
	case errors.Is(err, usererror.ErrUserNotFound):
		ErrorJson(c, fiber.StatusNotFound, envelope.Error{Code: "USER_NOT_FOUND", Message: err.Error()})
		return
	case errors.Is(err, urlerror.ErrUrlNotFound):
		ErrorJson(c, fiber.StatusNotFound, envelope.Error{Code: "URL_NOT_FOUND", Message: err.Error()})
		return
	default:
		logger.Error(message, slog.Any("error", err))
		ErrorJson(c, fiber.StatusInternalServerError, envelope.Error{Code: "INTERNAL_SERVER_ERROR", Message: "Internal server error"})
		return
	}
}

func ErrorJson(c fiber.Ctx, status int, errorBody envelope.Error) error {
	payload := envelope.NewErrorResponse(errorBody)
	return c.Status(status).JSON(payload)
}