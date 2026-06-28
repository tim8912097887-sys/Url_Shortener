package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
	"github.com/tim8912097887-sys/url-shortener/internal/url"
)

type Api struct{
	Addr string
}

func (a *Api) Mount(logger *slog.Logger) http.Handler {
	app := fiber.New()

	// Api Versioning
	api := app.Group("/api")      
    v1 := api.Group("/v1") 
	urlGroup := v1.Group("/urls")
	// Register Url handler
	handler := url.NewHandler(logger,url.NewService())
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
	shutdownCh := make(chan os.Signal, 1)
	// Notify the serverErrorCh when the shutdown signal is received
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)

	defer signal.Stop(shutdownCh)

	select {
		case <-ctx.Done():
			logger.Info("shutting down the server",slog.String("reason", ctx.Err().Error()))
		case err := <-serverErrorCh:
			return err
		case <-shutdownCh:
			logger.Info("shutting down the server",slog.String("reason", "signal received"))
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