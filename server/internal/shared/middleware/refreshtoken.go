package middleware

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	jwttoken "github.com/tim8912097887-sys/url-shortener/internal/shared/util/jwt_token"
)

// AuthMiddleware extracts and validates the bearer access token from the
// Authorization header, then attaches the user id and token version to the
// fiber context locals for downstream handlers. Every failure path — missing
// header, wrong scheme, malformed/expired token — reports the same generic
// invalid_token error so nothing about *why* it failed leaks to the client.
func RefreshTokenMiddleware(tokens jwttoken.TokenManager, logger *slog.Logger) fiber.Handler {

	return func(c fiber.Ctx) error {
		refreshToken := c.Cookies(RefreshCookieName)

		if refreshToken == "" {
			return c.Next()
		}

		claims, err := tokens.ParseRefreshToken(refreshToken)
		if err != nil {
			return c.Next()
		}

		c.Locals(LocalsUserID, claims.UserID)
		c.Locals(LocalsTokenVersion, claims.TokenVersion)

		return c.Next()
	}
}