package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"news-aggregator/internal/config"
	"news-aggregator/internal/gateway"
	"news-aggregator/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := logger.New(cfg.LogLevel)

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize gateway
	gw, err := gateway.New(cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize gateway")
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start gateway server using the new modular system
	go func() {
		if err := gw.Start(ctx, cfg.Server.Address); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start gateway server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("Shutting down server...")

	// Cancel context to trigger graceful shutdown
	cancel()

	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)

	logger.Info().Msg("Server exiting")
}
