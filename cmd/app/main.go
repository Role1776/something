package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	errCh := make(chan error, 1)
	go func() {
		logger.Info("Server is listening", slog.String("port", cfg.Server.Port))
		if err := app.Server.Run(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server run error: %w", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("Shutting down server...")
	case serverErr := <-errCh:
		logger.Error("Server run failed", slog.Any("error", serverErr))
	}

	if err = app.Server.Stop(ctx); err != nil {
		logger.Error("Server forced to shutdown", slog.Any("error", err))
	}

	if err := app.Close(); err != nil {
		logger.Error("Error closing database connection", slog.Any("error", err.Error()))
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
