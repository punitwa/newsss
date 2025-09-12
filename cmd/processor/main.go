package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"news-aggregator/internal/config"
	"news-aggregator/internal/processor"
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

	// Initialize processor service
	processorService, err := processor.New(cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize processor service")
	}

	// Start processor service
	go func() {
		logger.Info().Msg("Starting data processor service")
		if err := processorService.Start(ctx); err != nil {
			logger.Error().Err(err).Msg("Processor service error")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down processor service...")
	cancel()

	// Wait for graceful shutdown
	processorService.Stop()
	logger.Info().Msg("Processor service stopped")
}
