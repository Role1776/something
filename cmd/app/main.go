package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"todoai/internal/app"
	"todoai/internal/config"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configuration: %v\n", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.Timeout)
	defer cancel()

	logger := setupLogger(cfg.Logging.Level)

	app, err := app.NewApp(cfg, logger)
	if err != nil {
		logger.Error("App initialization failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	if err = app.Run(); err != nil {
		logger.Error("App running failed")
		os.Exit(1)
	}

	if err := app.Close(ctx); err != nil {
		logger.Error("Error closing database connection", slog.Any("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("Server stopped gracefully")

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
		log.Warn("Unknown environment, defaulting to debug level logging")
	}

	return log
}
