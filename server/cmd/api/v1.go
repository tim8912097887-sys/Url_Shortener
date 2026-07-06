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
	"github.com/tim8912097887-sys/url-shortener/internal/url"
)

type Api struct{
	Addr string
}

func (a *Api) Mount(logger *slog.Logger,pool *pgxpool.Pool,cache *redis.Client) http.Handler {
	app := fiber.New()

	// Configure cors
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173"},
		AllowMethods: []string{"POST","GET"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))
	// Api Versioning
	api := app.Group("/api")      
    v1 := api.Group("/v1") 
	urlGroup := v1.Group("/urls")
	// Register Url handler
	repository := url.NewRepository(pool)
	service := url.NewService(repository,cache,logger)
	handler := url.NewHandler(logger,service)
	handler.RegisterRoutes(urlGroup)
	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	return adaptor.FiberApp(app)
}

func (a *Api) Run(ctx context.Context, logger *slog.Logger, h http.Handler, shutdownTimeout time.Duration) error {
	server := &http.Server{
		Addr:    a.Addr,
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
		logger.Info("starting server",slog.String("address", a.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("failed to start server",slog.Any("error", err))
			serverErrorCh <- err
		}
	}()

	select {
		case <-ctx.Done():
			logger.Info("shutting down the server",slog.String("reason", ctx.Err().Error()))
		case err := <-serverErrorCh:
			return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("failed to shut down the server",slog.Any("error", err))
		if closeErr := server.Close(); closeErr != nil {
			logger.Error("failed to close the server",slog.Any("error", err))
			return errors.Join(err,closeErr)
		}
		return err
	}

	logger.Info("server shut down gracefully")
	return nil

}