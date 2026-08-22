package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"

	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/tim8912097887-sys/url-shortener/internal/configs"
	"github.com/tim8912097887-sys/url-shortener/internal/oauth"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/middleware"
	"github.com/tim8912097887-sys/url-shortener/internal/shared/ratelimiter"
	jwttoken "github.com/tim8912097887-sys/url-shortener/internal/shared/util/jwt_token"
	"github.com/tim8912097887-sys/url-shortener/internal/url"
	"github.com/tim8912097887-sys/url-shortener/internal/user"
)

type ApiConfig struct {
    Logger *slog.Logger
    Cfg *configs.Configs
}

type Api struct{
	apiConfig ApiConfig
}

func NewApi(apiConfig ApiConfig) *Api {
	return &Api{
		apiConfig: apiConfig,
	}
}

func (a *Api) Mount(pool *pgxpool.Pool,cache *redis.Client) http.Handler {
	app := fiber.New()

	// Configure cors
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{a.apiConfig.Cfg.ClientOrigin},
		AllowMethods: []string{"POST","GET"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	// Utils
	tokenManager := jwttoken.NewTokenManager(a.apiConfig.Cfg.AccessTokenSecret, a.apiConfig.Cfg.RefreshTokenSecret)

	// Api Versioning
	api := app.Group("/api")      
    v1 := api.Group("/v1") 
	urlGroup := v1.Group("/urls")
    userGroup := v1.Group("/users")
	authGroup := v1.Group("/auth")

	// Rate Limiting
    rateLimiter := ratelimiter.NewRateLimiter(20, 10)
	urlGroup.Use(middleware.RateLimitMiddleware(rateLimiter))
	// Register Url handler
	urlCache := url.NewCache(cache)
	urlRepository := url.NewRepository(pool)
	urlService := url.NewService(&url.ServiceConfig{Repository: urlRepository, Cache: urlCache, Logger: a.apiConfig.Logger})
	urlHandler := url.NewHandler(url.HandlerConfig{
		Logger:  a.apiConfig.Logger,
		Service: urlService,
	})
	urlHandler.RegisterRoutes(urlGroup)
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Register user handler
    userRepository := user.NewRepository(pool)
	userService := user.NewService(&user.ServiceConfig{Repository: userRepository, Tokens: *tokenManager, Logger: a.apiConfig.Logger})
	userHandler := user.NewHandler(user.HandlerConfig{Logger: a.apiConfig.Logger, Service: userService, Tokens: *tokenManager, Cfg: a.apiConfig.Cfg})
	userHandler.RegisterRoutes(userGroup)

	// Register oauth handler
	oauthCofig := oauth.New(oauth.Config{
		GoogleClientID: a.apiConfig.Cfg.GoogleClientID,
		GoogleClientSecret: a.apiConfig.Cfg.GoogleClientSecret,
		BaseURL: a.apiConfig.Cfg.BaseURL,
	})
    oauthRepository := oauth.NewRepository(pool)
	oauthCache := oauth.NewCache(cache)
	oauthService := oauth.NewService(&oauth.ServiceConfig{
		Cache: oauthCache, 
		OAuthConfig: oauthCofig, 
		TokenManager: tokenManager, 
		Repository: oauthRepository,
	})
	oauthHandler := oauth.NewHandler(oauth.HandlerConfig{Logger: a.apiConfig.Logger, Service: oauthService, Cfg: a.apiConfig.Cfg})
	oauthHandler.RegisterRoutes(authGroup)

	return adaptor.FiberApp(app)
}

func (a *Api) Run(ctx context.Context, h http.Handler, shutdownTimeout time.Duration) error {
	server := &http.Server{
		Addr:    a.apiConfig.Cfg.Addr,
		Handler: h,
		ReadTimeout:       5 * time.Second,
        ReadHeaderTimeout: 2 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       120 * time.Second,
	}

	// Channel to notify when the server is initialized failure
	serverErrorCh := make(chan error, 1)
	// Start the server with goroutine
	go func() {
		a.apiConfig.Logger.Info("starting server",slog.String("address", a.apiConfig.Cfg.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.apiConfig.Logger.Error("failed to start server",slog.Any("error", err))
			serverErrorCh <- err
		}
	}()

	select {
		case <-ctx.Done():
			a.apiConfig.Logger.Info("shutting down the server",slog.String("reason", ctx.Err().Error()))
		case err := <-serverErrorCh:
			return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		a.apiConfig.Logger.Error("failed to shut down the server",slog.Any("error", err))
		if closeErr := server.Close(); closeErr != nil {
			a.apiConfig.Logger.Error("failed to close the server",slog.Any("error", err))
			return errors.Join(err,closeErr)
		}
		return err
	}

	a.apiConfig.Logger.Info("server shut down gracefully")
	return nil

}