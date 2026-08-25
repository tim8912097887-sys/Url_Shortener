package middleware

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	jwttoken "github.com/tim8912097887-sys/url-shortener/internal/shared/util/jwt_token"
)

// AuthMiddleware extracts and validates the bearer access token from the
// Authorization header, then attaches the user id and token version to the
// fiber context locals for downstream handlers. Every failure path — missing
// header, wrong scheme, malformed/expired token — reports the same generic
// invalid_token error so nothing about *why* it failed leaks to the client.
func AccessTokenMiddleware(tokens jwttoken.TokenManager, logger *slog.Logger) fiber.Handler {
	const bearerPrefix = "Bearer "

	return func(c fiber.Ctx) error {
		header := c.Get(fiber.HeaderAuthorization)

		// Allow unauthenticated users
		if header == "" || !strings.HasPrefix(header, bearerPrefix) {
			return c.Next()
		}
		tokenString := strings.TrimPrefix(header, bearerPrefix)

		claims, err := tokens.ParseAccessToken(tokenString)
		// Allow unauthenticated users
		if err != nil {
		    return c.Next()
		}

		c.Locals(LocalsUserID, claims.UserID)
		c.Locals(LocalsTokenVersion, claims.TokenVersion)

		return c.Next()
	}
}