package services

import (
	"context"
	"fmt"
	"time"

	"news-aggregator/internal/config"
	loggerPkg "news-aggregator/pkg/logger"

	"github.com/rs/zerolog"
)

type CleanupService struct {
	config      *config.Config
	logger      zerolog.Logger
	newsService *NewsService
	logRotator  *loggerPkg.LogRotator
	ticker      *time.Ticker
	done        chan bool
}

func NewCleanupService(cfg *config.Config, logger zerolog.Logger, newsService *NewsService) *CleanupService {
	// Get log files to rotate
	logFiles := loggerPkg.GetLogFiles()
	logRotator := loggerPkg.NewLogRotator(logger, logFiles)

	return &CleanupService{
		config:      cfg,
		logger:      logger.With().Str("service", "cleanup").Logger(),
		newsService: newsService,
		logRotator:  logRotator,
		ticker:      time.NewTicker(6 * time.Hour), // Run every 6 hours
		done:        make(chan bool),
	}
}

func (cs *CleanupService) Start(ctx context.Context) error {
	cs.logger.Info().Msg("Starting cleanup service")

	// Start log rotator
	cs.logRotator.Start()

	// Run initial cleanup
	cs.performCleanup(ctx)

	// Start periodic cleanup
	go func() {
		for {
			select {
			case <-cs.ticker.C:
				cs.performCleanup(ctx)
			case <-cs.done:
				cs.logger.Info().Msg("Cleanup service stopped")
				return
			case <-ctx.Done():
				cs.logger.Info().Msg("Cleanup service context cancelled")
				return
			}
		}
	}()

	return nil
}

func (cs *CleanupService) Stop() {
	cs.logger.Info().Msg("Stopping cleanup service")
	
	// Stop log rotator
	cs.logRotator.Stop()
	
	// Stop ticker
	cs.ticker.Stop()
	
	// Signal done
	cs.done <- true
}

func (cs *CleanupService) performCleanup(ctx context.Context) {
	cs.logger.Info().Msg("Starting periodic cleanup")
	
	// Cleanup old database articles (older than 2 days)
	if err := cs.newsService.CleanupOldArticles(ctx); err != nil {
		cs.logger.Error().Err(err).Msg("Failed to cleanup old articles from database")
	} else {
		cs.logger.Info().Msg("Database cleanup completed successfully")
	}
	
	cs.logger.Info().Msg("Periodic cleanup completed")
}

// ManualCleanup allows triggering cleanup manually
func (cs *CleanupService) ManualCleanup(ctx context.Context) error {
	cs.logger.Info().Msg("Manual cleanup triggered")
	
	// Perform database cleanup
	if err := cs.newsService.CleanupOldArticles(ctx); err != nil {
		return fmt.Errorf("failed to cleanup database: %w", err)
	}
	
	// Force log rotation check
	cs.logRotator.Stop()
	cs.logRotator.Start()
	
	cs.logger.Info().Msg("Manual cleanup completed")
	return nil
}
