package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"news-aggregator/internal/config"
	"news-aggregator/internal/services"
	"news-aggregator/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logger.New(cfg.LogLevel)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize news service (needed for database cleanup)
	newsService, err := services.NewNewsService(cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize news service")
	}

	// Initialize cleanup service
	cleanupService := services.NewCleanupService(cfg, logger, newsService)

	// Start cleanup service
	go func() {
		logger.Info().Msg("Starting cleanup service")
		if err := cleanupService.Start(ctx); err != nil {
			logger.Error().Err(err).Msg("Cleanup service error")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down cleanup service...")
	cancel()

	// Wait for graceful shutdown
	cleanupService.Stop()
	logger.Info().Msg("Cleanup service stopped")
}
