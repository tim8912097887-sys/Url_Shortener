package middleware

import (
	"github.com/gofiber/fiber/v3"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/response/envelope"
)

type RateLimiter interface {
	Allow() bool
	AllowN(n int) bool
}

func RateLimitMiddleware(limiter RateLimiter) fiber.Handler {
	return func(c fiber.Ctx) error {
		if !limiter.Allow() {
			return c.Status(fiber.StatusTooManyRequests).JSON(envelope.NewErrorResponse(envelope.Error{Code: "TOO_MANY_REQUESTS", Message: "too many requests"})) 
		}
		return c.Next()
	}
}