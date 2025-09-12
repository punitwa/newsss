// Package factory provides a factory pattern for creating data sources.
package factory

import (
	"fmt"
	"time"

	"news-aggregator/internal/datasources/core"
	"news-aggregator/internal/datasources/sources/rss"

	"github.com/rs/zerolog"
)

// SourceFactory provides methods to create different types of data sources.
type SourceFactory struct {
	logger zerolog.Logger
}

// NewSourceFactory creates a new source factory.
func NewSourceFactory(logger zerolog.Logger) *SourceFactory {
	return &SourceFactory{
		logger: logger.With().Str("component", "source_factory").Logger(),
	}
}

// CreateSource creates a data source based on the provided configuration.
func (sf *SourceFactory) CreateSource(config core.SourceConfig) (core.DataSource, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid source configuration: %w", err)
	}
	
	sf.logger.Info().
		Str("name", config.Name).
		Str("type", string(config.Type)).
		Str("url", config.URL).
		Msg("Creating data source")
	
	switch config.Type {
	case core.SourceTypeRSS:
		return sf.createRSSSource(config)
	case core.SourceTypeAPI:
		return sf.createAPISource(config)
	case core.SourceTypeScraper:
		return sf.createScraperSource(config)
	default:
		return nil, core.NewValidationError("type", config.Type, "unsupported source type")
	}
}

// CreateSources creates multiple data sources from configurations.
func (sf *SourceFactory) CreateSources(configs []core.SourceConfig) ([]core.DataSource, []error) {
	var sources []core.DataSource
	var errors []error
	
	for i, config := range configs {
		source, err := sf.CreateSource(config)
		if err != nil {
			sf.logger.Error().
				Err(err).
				Int("config_index", i).
				Str("source_name", config.Name).
				Msg("Failed to create source")
			errors = append(errors, fmt.Errorf("source %s: %w", config.Name, err))
			continue
		}
		
		sources = append(sources, source)
		sf.logger.Info().
			Str("source_name", config.Name).
			Str("source_type", string(config.Type)).
			Msg("Source created successfully")
	}
	
	sf.logger.Info().
		Int("total_configs", len(configs)).
		Int("successful_sources", len(sources)).
		Int("failed_sources", len(errors)).
		Msg("Source creation completed")
	
	return sources, errors
}

// createRSSSource creates an RSS data source.
func (sf *SourceFactory) createRSSSource(config core.SourceConfig) (core.DataSource, error) {
	return rss.NewSource(config, sf.logger)
}

// createAPISource creates an API data source.
func (sf *SourceFactory) createAPISource(config core.SourceConfig) (core.DataSource, error) {
	// TODO: Implement API source creation
	// For now, return an error indicating it's not implemented
	return nil, fmt.Errorf("API source not yet implemented")
}

// createScraperSource creates a scraper data source.
func (sf *SourceFactory) createScraperSource(config core.SourceConfig) (core.DataSource, error) {
	// TODO: Implement scraper source creation
	// For now, return an error indicating it's not implemented
	return nil, fmt.Errorf("scraper source not yet implemented")
}

// ValidateSourceConfig validates a source configuration without creating the source.
func (sf *SourceFactory) ValidateSourceConfig(config core.SourceConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}
	
	// Additional validation based on source type
	switch config.Type {
	case core.SourceTypeRSS:
		return sf.validateRSSConfig(config)
	case core.SourceTypeAPI:
		return sf.validateAPIConfig(config)
	case core.SourceTypeScraper:
		return sf.validateScraperConfig(config)
	default:
		return core.NewValidationError("type", config.Type, "unsupported source type")
	}
}

// validateRSSConfig performs RSS-specific validation.
func (sf *SourceFactory) validateRSSConfig(config core.SourceConfig) error {
	// RSS feeds should typically have XML content type expectations
	// Additional RSS-specific validations can be added here
	return nil
}

// validateAPIConfig performs API-specific validation.
func (sf *SourceFactory) validateAPIConfig(config core.SourceConfig) error {
	// API sources might require authentication headers, specific formats, etc.
	// Additional API-specific validations can be added here
	return nil
}

// validateScraperConfig performs scraper-specific validation.
func (sf *SourceFactory) validateScraperConfig(config core.SourceConfig) error {
	// Scraper sources might need specific user agents, rate limits, etc.
	// Additional scraper-specific validations can be added here
	return nil
}

// GetSupportedTypes returns a list of supported source types.
func (sf *SourceFactory) GetSupportedTypes() []core.SourceType {
	return []core.SourceType{
		core.SourceTypeRSS,
		// TODO: Enable when implemented
		// core.SourceTypeAPI,
		// core.SourceTypeScraper,
	}
}

// IsTypeSupported checks if a source type is supported.
func (sf *SourceFactory) IsTypeSupported(sourceType core.SourceType) bool {
	supportedTypes := sf.GetSupportedTypes()
	for _, supported := range supportedTypes {
		if sourceType == supported {
			return true
		}
	}
	return false
}

// GetDefaultConfig returns a default configuration for a source type.
func (sf *SourceFactory) GetDefaultConfig(sourceType core.SourceType) (core.SourceConfig, error) {
	if !sf.IsTypeSupported(sourceType) {
		return core.SourceConfig{}, core.NewValidationError("type", sourceType, "unsupported source type")
	}
	
	baseConfig := core.SourceConfig{
		Type:         sourceType,
		Enabled:      true,
		Schedule:     DefaultScheduleForType(sourceType),
		RateLimit:    DefaultRateLimitForType(sourceType),
		Timeout:      DefaultTimeoutForType(sourceType),
		MaxRetries:   3,
		RetryDelay:   DefaultRetryDelayForType(sourceType),
		Headers:      make(map[string]string),
		Categories:   []string{},
		Keywords:     []string{},
	}
	
	return baseConfig, nil
}

// Helper functions for default values

// DefaultScheduleForType returns the default schedule for a source type.
func DefaultScheduleForType(sourceType core.SourceType) time.Duration {
	switch sourceType {
	case core.SourceTypeRSS:
		return 15 * time.Minute // RSS feeds updated every 15 minutes
	case core.SourceTypeAPI:
		return 10 * time.Minute // API sources more frequent
	case core.SourceTypeScraper:
		return 30 * time.Minute // Scrapers less frequent to be respectful
	default:
		return 30 * time.Minute
	}
}

// DefaultRateLimitForType returns the default rate limit for a source type.
func DefaultRateLimitForType(sourceType core.SourceType) float64 {
	switch sourceType {
	case core.SourceTypeRSS:
		return 2.0 // 2 requests per second for RSS
	case core.SourceTypeAPI:
		return 5.0 // 5 requests per second for APIs
	case core.SourceTypeScraper:
		return 0.5 // 0.5 requests per second for scrapers (more conservative)
	default:
		return 1.0
	}
}

// DefaultTimeoutForType returns the default timeout for a source type.
func DefaultTimeoutForType(sourceType core.SourceType) time.Duration {
	switch sourceType {
	case core.SourceTypeRSS:
		return 30 * time.Second
	case core.SourceTypeAPI:
		return 15 * time.Second // APIs should be faster
	case core.SourceTypeScraper:
		return 45 * time.Second // Scrapers might need more time
	default:
		return 30 * time.Second
	}
}

// DefaultRetryDelayForType returns the default retry delay for a source type.
func DefaultRetryDelayForType(sourceType core.SourceType) time.Duration {
	switch sourceType {
	case core.SourceTypeRSS:
		return 5 * time.Second
	case core.SourceTypeAPI:
		return 2 * time.Second // APIs might have shorter retry cycles
	case core.SourceTypeScraper:
		return 10 * time.Second // Scrapers should wait longer between retries
	default:
		return 5 * time.Second
	}
}
