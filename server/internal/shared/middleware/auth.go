package middleware

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	usererror "github.com/tim8912097887-sys/url-shortener/internal/shared/error/user_error"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response"
	jwttoken "github.com/tim8912097887-sys/url-shortener/internal/shared/util/jwt_token"
)

const (
	LocalsUserID       = "user_id"
	LocalsTokenVersion = "token_version"
)

// AuthMiddleware extracts and validates the bearer access token from the
// Authorization header, then attaches the user id and token version to the
// fiber context locals for downstream handlers. Every failure path — missing
// header, wrong scheme, malformed/expired token — reports the same generic
// invalid_token error so nothing about *why* it failed leaks to the client.
func AuthMiddleware(tokens jwttoken.TokenManager, logger *slog.Logger) fiber.Handler {
	const bearerPrefix = "Bearer "

	return func(c fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)

		if header == "" || !strings.HasPrefix(header, bearerPrefix) {
			logger.Error("missing or malformed authorization header", slog.String("context", "auth middleware"))
			return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("invalid_token", usererror.ErrInvalidToken.Error()))
		}

		tokenString := strings.TrimPrefix(header, bearerPrefix)

		claims, err := tokens.ParseAccessToken(tokenString)
		if err != nil {
			logger.Error("failed to parse access token", slog.Any("error", err), slog.String("context", "auth middleware"))
			return c.Status(fiber.StatusUnauthorized).JSON(response.NewErrorResponse("invalid_token", usererror.ErrInvalidToken.Error()))
		}

		c.Locals(LocalsUserID, claims.UserID)
		c.Locals(LocalsTokenVersion, claims.TokenVersion)

		return c.Next()
	}
}