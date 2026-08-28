package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tim8912097887-sys/url-shortener/cmd/api"
	"github.com/tim8912097887-sys/url-shortener/internal/cache"
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

    ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	pool,err := db.Init(logger,ctx,cfg.DbUrl)

	if err != nil {
		logger.Error("failed to connect to db", slog.Any("error", err))
		os.Exit(1)
	}

	// Close the connection pool
	defer pool.Close()

	// Start the cleanup service in a goroutine
	go func() {
		if err := db.NewCleanupService(pool, logger).Start(ctx); err != nil {
			logger.Error("failed to start cleanup service", slog.Any("error", err))
		}
	}()

	rdb := cache.NewRedisClient(logger,cfg.RedisUrl)
	cache,err := cache.CacheInit(ctx, logger,rdb)

	if err != nil {
		logger.Error("failed to connect to redis", slog.Any("error", err))
		os.Exit(1)
	}

	// Close the connection pool
	defer cache.Close()

	app := api.NewApi(api.ApiConfig{
		Cfg: cfg,
		Logger: logger,
	})

	if err := app.Run(ctx, app.Mount(pool, cache, ctx), 8*time.Second); err != nil {
		logger.Error("failed to start server", slog.Any("error", err))
		os.Exit(1)
	}
}