package url

import (
	"context"
	"log/slog"
	"time"
)

type ClickSyncWorkerConfig struct {
	Cache       UrlCache
	Repository  UrlRepository
	Logger      *slog.Logger
}

type ClickSyncWorker struct {
	cache      UrlCache
	repository UrlRepository
	logger     *slog.Logger
	interval    time.Duration
	syncTimeout time.Duration
}

func NewClickSyncWorker(config ClickSyncWorkerConfig) *ClickSyncWorker {
	return &ClickSyncWorker{
		cache:       config.Cache,
		repository:  config.Repository,
		logger:      config.Logger,
		interval:    WorkerInterval,
		syncTimeout: WorkerTimeout,
	}
}

func (w *ClickSyncWorker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.Sync(ctx); err != nil {
				w.logger.Error("failed to sync clicks", slog.Any("error", err))
			}
		}
	}
}

func (w *ClickSyncWorker) Sync(ctx context.Context) error {
	
	syncCtx, cancel := context.WithTimeout(
		ctx,
		w.syncTimeout,
	)
	defer cancel()

	// Get all pending clicks
	pendingClicks, err := w.cache.GetPendingClicks(syncCtx)
	if err != nil {
		w.logger.Error(
			"failed to get pending click URLs",
			slog.Any("error", err),
		)
		return err
	}

	// Sync clicks
	for _, click := range pendingClicks {
		if err := w.syncClicks(ctx, click) ; err != nil {
			w.logger.Error("failed to sync clicks", slog.String("shortUrl", click), slog.Any("error", err))
			continue
		}
	}
	return nil
}

func (w *ClickSyncWorker) syncClicks(ctx context.Context, shortUrl string) error {
	
	clicks, err := w.cache.GetAndReset(ctx, UrlClickKey(shortUrl))
	if err != nil {
		w.logger.Error("failed to get clicks", slog.String("shortUrl", shortUrl), slog.Any("error", err))
		return err
	}
	
	if clicks == 0 {
		return w.cache.RemovePendingClick(ctx, shortUrl)
	}
	// Update clicks
	if err := w.repository.UpdateUrlClicks(ctx, shortUrl, int(clicks)); err != nil {
		w.logger.Error("failed to update clicks", slog.String("shortUrl", shortUrl), slog.Any("error", err))
		return err
	}

	// Remove Pending Click
	return w.cache.RemovePendingClick(ctx, shortUrl)
}
