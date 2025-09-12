package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"news-aggregator/internal/config"
	"news-aggregator/internal/collector"
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

	// Initialize collector service
	collectorService, err := collector.New(cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize collector service")
	}

	// Start collector service
	go func() {
		logger.Info().Msg("Starting data collector service")
		if err := collectorService.Start(ctx); err != nil {
			logger.Error().Err(err).Msg("Collector service error")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down collector service...")
	cancel()

	// Wait for graceful shutdown
	collectorService.Stop()
	logger.Info().Msg("Collector service stopped")
}
