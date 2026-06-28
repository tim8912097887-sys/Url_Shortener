package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/tim8912097887-sys/url-shortener/internal/configs"
)

func main() {
	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, handlerOpts))
	slog.SetDefault(logger)

	cfg, err := configs.InitConfigs()

	if err != nil {
		logger.Error("failed to load configs", slog.Any("error", err))
		os.Exit(1)
	}

	app := api{
		addr: cfg.Addr,
	}

	if err := app.run(context.Background(), slog.Default(), app.mount(), 8*time.Second); err != nil {
		logger.Error("failed to start server", slog.Any("error", err))
		os.Exit(1)
	}
}