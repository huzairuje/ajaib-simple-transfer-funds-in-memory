package main

import (
	"context"
	"log/slog"
	"os"

	"ajaib-testing-code/config"
	transferApp "ajaib-testing-code/internal/adapters/app/transfer"
	transferCore "ajaib-testing-code/internal/adapters/core/transfer"
	transferHandler "ajaib-testing-code/internal/adapters/framework/primary/rest_fiber/transfer"
	transferDB "ajaib-testing-code/internal/adapters/framework/secondary/repository/db/transfer"
	idempotencyCache "ajaib-testing-code/internal/adapters/framework/secondary/repository/cache/idempotency"
	"ajaib-testing-code/router"
)

func main() {
	ctx := context.Background()

	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	})
	logger := slog.New(logHandler)
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		slog.ErrorContext(ctx, "Failed to load config", "error", err)
		return
	}

	slog.InfoContext(ctx, "Starting Transfer Service", "port", cfg.App.Port)

	dbRepo := transferDB.New(transferDB.Config{})
	cacheRepo := idempotencyCache.New(idempotencyCache.Config{})

	coreTransfer := transferCore.New(transferCore.Config{
		DB:    dbRepo,
		Cache: cacheRepo,
	})

	appTransfer := transferApp.New(transferApp.Config{
		Core: coreTransfer,
	})

	transferHandlerInstance := transferHandler.NewHandler(transferHandler.Config{
		TransferApp: appTransfer,
	})

	httpRouter := router.NewRouter(router.Config{
		TransferHandler: transferHandlerInstance,
	})

	if err := httpRouter.Run(cfg.App.Port); err != nil {
		slog.ErrorContext(ctx, "Failed to start server", "error", err)
	}
}
