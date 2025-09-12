// Package datasources provides a unified interface for accessing various news data sources.
// This package has been refactored for better modularity, readability, and maintainability.
package datasources

import (
	"context"
	"time"

	"news-aggregator/internal/config"
	"news-aggregator/internal/datasources/core"
	"news-aggregator/internal/datasources/factory"
	"news-aggregator/internal/datasources/sources/rss"
	"news-aggregator/internal/datasources/utils"
	"news-aggregator/internal/models"

	"github.com/rs/zerolog"
)

// Re-export core interfaces and types for backward compatibility
type (
	DataSource        = core.DataSource
	SourceMetrics     = core.SourceMetrics
	ProcessingOptions = core.ProcessingOptions
	HTTPClient        = core.HTTPClient
	RateLimiter       = core.RateLimiter

	// Error types
	ValidationError = core.ValidationError

	// Configuration types
	SourceConfig = core.SourceConfig
	SourceType   = core.SourceType
)

// Re-export constants for backward compatibility
const (
	SourceTypeRSS     = core.SourceTypeRSS
	SourceTypeAPI     = core.SourceTypeAPI
	SourceTypeScraper = core.SourceTypeScraper
)

// Re-export RSS-specific types for backward compatibility
type (
	ParsingOptions = rss.ParsingOptions
)

// DataSourceManager provides centralized management of data sources.
type DataSourceManager struct {
	factory *factory.SourceFactory
	sources map[string]core.DataSource
	logger  zerolog.Logger
}

// NewDataSourceManager creates a new data source manager.
func NewDataSourceManager(logger zerolog.Logger) *DataSourceManager {
	return &DataSourceManager{
		factory: factory.NewSourceFactory(logger),
		sources: make(map[string]core.DataSource),
		logger:  logger.With().Str("component", "datasource_manager").Logger(),
	}
}

// CreateSource creates a new data source from configuration.
func (dsm *DataSourceManager) CreateSource(config core.SourceConfig) (core.DataSource, error) {
	return dsm.factory.CreateSource(config)
}

// AddSource adds a new data source to the manager.
func (dsm *DataSourceManager) AddSource(name string, source core.DataSource) {
	dsm.sources[name] = source
	dsm.logger.Info().Str("source_name", name).Msg("Data source added to manager")
}

// GetSource retrieves a data source by name.
func (dsm *DataSourceManager) GetSource(name string) (core.DataSource, bool) {
	source, exists := dsm.sources[name]
	return source, exists
}

// RemoveSource removes a data source from the manager.
func (dsm *DataSourceManager) RemoveSource(name string) {
	delete(dsm.sources, name)
	dsm.logger.Info().Str("source_name", name).Msg("Data source removed from manager")
}

// GetAllSources returns all managed data sources.
func (dsm *DataSourceManager) GetAllSources() map[string]core.DataSource {
	return dsm.sources
}

// FetchAllSources fetches data from all managed sources.
func (dsm *DataSourceManager) FetchAllSources(ctx context.Context) ([]models.News, error) {
	var allNews []models.News

	for name, source := range dsm.sources {
		dsm.logger.Debug().Str("source_name", name).Msg("Fetching from source")

		news, err := source.Fetch(ctx)
		if err != nil {
			dsm.logger.Error().
				Err(err).
				Str("source_name", name).
				Msg("Failed to fetch from source")
			continue
		}

		allNews = append(allNews, news...)
		dsm.logger.Debug().
			Str("source_name", name).
			Int("items_count", len(news)).
			Msg("Successfully fetched from source")
	}

	return allNews, nil
}

// ValidateAllSources validates all managed sources.
func (dsm *DataSourceManager) ValidateAllSources() map[string]error {
	results := make(map[string]error)

	for name, source := range dsm.sources {
		if validator, ok := source.(interface{ Validate() error }); ok {
			if err := validator.Validate(); err != nil {
				results[name] = err
			}
		}
	}

	return results
}

// GetSourceMetrics returns metrics for all managed sources.
func (dsm *DataSourceManager) GetSourceMetrics() map[string]core.SourceMetrics {
	metrics := make(map[string]core.SourceMetrics)

	for name, source := range dsm.sources {
		if metricsProvider, ok := source.(interface{ GetMetrics() core.SourceMetrics }); ok {
			metrics[name] = metricsProvider.GetMetrics()
		}
	}

	return metrics
}

// Convenience functions for creating specific source types

// NewRSSSource creates a new RSS data source.
// func NewRSSSource(config core.SourceConfig, logger zerolog.Logger) (*rss.Source, error) {
// 	return rss.NewSource(config, logger)
// }

// NewDataSource creates a data source using the factory.
func NewDataSource(config core.SourceConfig, logger zerolog.Logger) (core.DataSource, error) {
	factory := factory.NewSourceFactory(logger)
	return factory.CreateSource(config)
}

// NewSourceFactory creates a new source factory.
func NewSourceFactory(logger zerolog.Logger) *factory.SourceFactory {
	return factory.NewSourceFactory(logger)
}

// NewHTTPClient creates a new HTTP client.
func NewHTTPClient(timeout time.Duration, userAgent string, logger zerolog.Logger) core.HTTPClient {
	return utils.NewHTTPClient(timeout, userAgent, logger)
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(rateLimit float64, burst int, logger zerolog.Logger) core.RateLimiter {
	return utils.NewRateLimiter(rateLimit, burst, logger)
}

// Helper functions

// DefaultProcessingOptions returns default processing options.
func DefaultProcessingOptions() core.ProcessingOptions {
	return core.DefaultProcessingOptions()
}

// DefaultRSSParsingOptions returns default RSS parsing options.
func DefaultRSSParsingOptions() rss.ParsingOptions {
	return rss.DefaultParsingOptions()
}

// IsValidSourceType checks if a source type is valid.
func IsValidSourceType(sourceType core.SourceType) bool {
	return sourceType.IsValid()
}

// Compatibility wrapper functions for the old config system

// NewRSSSource creates a new RSS data source (compatibility wrapper)
func NewRSSSourceCompat(sourceConfig config.SourceConfig, logger zerolog.Logger) (core.DataSource, error) {
	// Parse schedule duration
	scheduleDuration := 15 * time.Minute // default
	if sourceConfig.Schedule != "" {
		if dur, err := time.ParseDuration(sourceConfig.Schedule); err == nil {
			scheduleDuration = dur
		}
	}

	// Convert config.SourceConfig to core.SourceConfig
	coreConfig := core.SourceConfig{
		Name:      sourceConfig.Name,
		Type:      core.SourceType(sourceConfig.Type),
		URL:       sourceConfig.URL,
		Schedule:  scheduleDuration,
		RateLimit: float64(sourceConfig.RateLimit),
		Headers:   sourceConfig.Headers,
		Enabled:   sourceConfig.Enabled,
	}

	factory := factory.NewSourceFactory(logger)
	return factory.CreateSource(coreConfig)
}

// NewAPISource creates a new API data source (compatibility wrapper)
func NewAPISource(sourceConfig config.SourceConfig, logger zerolog.Logger) (core.DataSource, error) {
	// Parse schedule duration
	scheduleDuration := 15 * time.Minute // default
	if sourceConfig.Schedule != "" {
		if dur, err := time.ParseDuration(sourceConfig.Schedule); err == nil {
			scheduleDuration = dur
		}
	}

	// Convert config.SourceConfig to core.SourceConfig
	coreConfig := core.SourceConfig{
		Name:      sourceConfig.Name,
		Type:      core.SourceType(sourceConfig.Type),
		URL:       sourceConfig.URL,
		Schedule:  scheduleDuration,
		RateLimit: float64(sourceConfig.RateLimit),
		Headers:   sourceConfig.Headers,
		Enabled:   sourceConfig.Enabled,
	}

	factory := factory.NewSourceFactory(logger)
	return factory.CreateSource(coreConfig)
}

// NewScraperSource creates a new scraper data source (compatibility wrapper)
func NewScraperSource(sourceConfig config.SourceConfig, logger zerolog.Logger) (core.DataSource, error) {
	// Parse schedule duration
	scheduleDuration := 15 * time.Minute // default
	if sourceConfig.Schedule != "" {
		if dur, err := time.ParseDuration(sourceConfig.Schedule); err == nil {
			scheduleDuration = dur
		}
	}

	// Convert config.SourceConfig to core.SourceConfig
	coreConfig := core.SourceConfig{
		Name:      sourceConfig.Name,
		Type:      core.SourceType(sourceConfig.Type),
		URL:       sourceConfig.URL,
		Schedule:  scheduleDuration,
		RateLimit: float64(sourceConfig.RateLimit),
		Headers:   sourceConfig.Headers,
		Enabled:   sourceConfig.Enabled,
	}

	factory := factory.NewSourceFactory(logger)
	return factory.CreateSource(coreConfig)
}

// Package information
const (
	// Version of the datasources package
	Version = "2.0.0"

	// Description of the package
	Description = "Modular, extensible data source management for news aggregation"
)
