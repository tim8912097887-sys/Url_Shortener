package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/tim8912097887-sys/url-shortener/cmd/api"
	"github.com/tim8912097887-sys/url-shortener/internal/configs"
	"github.com/tim8912097887-sys/url-shortener/internal/db"
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

    ctx := context.Background()

	pool,err := db.Init(logger,ctx,cfg.DbUrl)

	if err != nil {
		logger.Error("failed to connect to db", slog.Any("error", err))
		os.Exit(1)
	}

	// Close the connection pool
	defer pool.Close()

	app := api.Api{
		Addr: cfg.Addr,
	}

	if err := app.Run(context.Background(), slog.Default(), app.Mount(ctx,logger,pool), 8*time.Second); err != nil {
		logger.Error("failed to start server", slog.Any("error", err))
		os.Exit(1)
	}
}