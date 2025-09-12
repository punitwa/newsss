package sources

import (
	"fmt"
	"sync"
	"time"

	"news-aggregator/internal/config"
	"news-aggregator/internal/datasources"

	"github.com/rs/zerolog"
)

// sourceManager implements the SourceManager interface
type sourceManager struct {
	logger  zerolog.Logger
	sources map[string]datasources.DataSource
	mu      sync.RWMutex
}

// SourceManager interface defines the contract for managing data sources
type SourceManager interface {
	Initialize(sourceConfigs []config.SourceConfig) error
	AddSource(sourceConfig config.SourceConfig) error
	RemoveSource(sourceName string) error
	GetSource(sourceName string) (datasources.DataSource, bool)
	GetAllSources() map[string]datasources.DataSource
	GetStatus() map[string]interface{}
	GetSourceCount() int
	GetSourceNames() []string
	ValidateAllSources() map[string]error
}

// NewSourceManager creates a new source manager
func NewSourceManager(logger zerolog.Logger) SourceManager {
	return &sourceManager{
		logger:  logger,
		sources: make(map[string]datasources.DataSource),
	}
}

// Initialize initializes all sources from configuration
func (sm *sourceManager) Initialize(sourceConfigs []config.SourceConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.logger.Info().Int("count", len(sourceConfigs)).Msg("Initializing data sources")

	var errors []error
	successCount := 0

	for _, sourceConfig := range sourceConfigs {
		if !sourceConfig.Enabled {
			sm.logger.Info().
				Str("source", sourceConfig.Name).
				Msg("Source disabled, skipping")
			continue
		}

		if err := sm.initializeSource(sourceConfig); err != nil {
			sm.logger.Error().
				Err(err).
				Str("source", sourceConfig.Name).
				Str("type", sourceConfig.Type).
				Msg("Failed to initialize source")
			errors = append(errors, fmt.Errorf("source %s: %w", sourceConfig.Name, err))
			continue
		}

		successCount++
		sm.logger.Info().
			Str("source", sourceConfig.Name).
			Str("type", sourceConfig.Type).
			Msg("Source initialized successfully")
	}

	sm.logger.Info().
		Int("success", successCount).
		Int("failed", len(errors)).
		Int("total", len(sourceConfigs)).
		Msg("Source initialization completed")

	// Return error if any sources failed to initialize
	if len(errors) > 0 {
		return fmt.Errorf("failed to initialize %d sources: %v", len(errors), errors)
	}

	return nil
}

// initializeSource creates and initializes a single source
func (sm *sourceManager) initializeSource(sourceConfig config.SourceConfig) error {
	// Validate source configuration
	if err := sm.validateSourceConfig(sourceConfig); err != nil {
		return fmt.Errorf("invalid source configuration: %w", err)
	}

	// Check if source already exists
	if _, exists := sm.sources[sourceConfig.Name]; exists {
		return fmt.Errorf("source with name '%s' already exists", sourceConfig.Name)
	}

	// Create source based on type
	source, err := sm.createSource(sourceConfig)
	if err != nil {
		return fmt.Errorf("failed to create source: %w", err)
	}

	// Store the source
	sm.sources[sourceConfig.Name] = source

	return nil
}

// createSource creates a new data source based on configuration
func (sm *sourceManager) createSource(sourceConfig config.SourceConfig) (datasources.DataSource, error) {
	switch sourceConfig.Type {
	case "rss":
		return datasources.NewRSSSourceCompat(sourceConfig, sm.logger)
	case "api":
		return datasources.NewAPISource(sourceConfig, sm.logger)
	case "scraper":
		return datasources.NewScraperSource(sourceConfig, sm.logger)
	default:
		return nil, fmt.Errorf("unknown source type: %s", sourceConfig.Type)
	}
}

// validateSourceConfig validates a source configuration
func (sm *sourceManager) validateSourceConfig(config config.SourceConfig) error {
	if config.Name == "" {
		return fmt.Errorf("source name cannot be empty")
	}

	if config.Type == "" {
		return fmt.Errorf("source type cannot be empty")
	}

	if config.URL == "" {
		return fmt.Errorf("source URL cannot be empty")
	}

	// Validate source type
	validTypes := map[string]bool{
		"rss":     true,
		"api":     true,
		"scraper": true,
	}

	if !validTypes[config.Type] {
		return fmt.Errorf("invalid source type: %s (must be one of: rss, api, scraper)", config.Type)
	}

	// Validate schedule if provided
	if config.Schedule != "" {
		if _, err := time.ParseDuration(config.Schedule); err != nil {
			return fmt.Errorf("invalid schedule format: %w", err)
		}
	}

	// Validate rate limit
	if config.RateLimit < 0 {
		return fmt.Errorf("rate limit cannot be negative")
	}

	return nil
}

// AddSource adds a new source to the manager
func (sm *sourceManager) AddSource(sourceConfig config.SourceConfig) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.logger.Info().
		Str("source", sourceConfig.Name).
		Str("type", sourceConfig.Type).
		Msg("Adding new source")

	return sm.initializeSource(sourceConfig)
}

// RemoveSource removes a source from the manager
func (sm *sourceManager) RemoveSource(sourceName string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.logger.Info().Str("source", sourceName).Msg("Removing source")

	if _, exists := sm.sources[sourceName]; !exists {
		return fmt.Errorf("source '%s' not found", sourceName)
	}

	delete(sm.sources, sourceName)

	sm.logger.Info().Str("source", sourceName).Msg("Source removed successfully")
	return nil
}

// GetSource retrieves a specific source
func (sm *sourceManager) GetSource(sourceName string) (datasources.DataSource, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	source, exists := sm.sources[sourceName]
	return source, exists
}

// GetAllSources returns all sources
func (sm *sourceManager) GetAllSources() map[string]datasources.DataSource {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Create a copy to avoid concurrent access issues
	sourcesCopy := make(map[string]datasources.DataSource)
	for name, source := range sm.sources {
		sourcesCopy[name] = source
	}

	return sourcesCopy
}

// GetStatus returns the status of all sources
func (sm *sourceManager) GetStatus() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	status := make(map[string]interface{})

	for name, source := range sm.sources {
		sourceStatus := map[string]interface{}{
			"name":     name,
			"type":     fmt.Sprintf("%T", source),
			"schedule": source.GetSchedule().String(),
			"active":   true,
			"healthy":  sm.checkSourceHealth(source),
		}

		// Add additional source-specific information if available
		if healthChecker, ok := source.(interface{ Health() map[string]interface{} }); ok {
			sourceStatus["details"] = healthChecker.Health()
		}

		status[name] = sourceStatus
	}

	return status
}

// checkSourceHealth performs a basic health check on a source
func (sm *sourceManager) checkSourceHealth(source datasources.DataSource) bool {
	// This is a basic implementation
	// In a real-world scenario, you might want to implement more sophisticated health checks
	if source == nil {
		return false
	}

	// Check if the source implements a health check interface
	if healthChecker, ok := source.(interface{ IsHealthy() bool }); ok {
		return healthChecker.IsHealthy()
	}

	// Default to healthy if no specific health check is available
	return true
}

// GetSourceCount returns the number of active sources
func (sm *sourceManager) GetSourceCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.sources)
}

// GetSourceNames returns a list of all source names
func (sm *sourceManager) GetSourceNames() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	names := make([]string, 0, len(sm.sources))
	for name := range sm.sources {
		names = append(names, name)
	}

	return names
}

// ValidateAllSources validates the configuration of all sources
func (sm *sourceManager) ValidateAllSources() map[string]error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	results := make(map[string]error)

	for name, source := range sm.sources {
		// Basic validation - check if source is not nil
		if source == nil {
			results[name] = fmt.Errorf("source is nil")
			continue
		}

		// Check if source implements validation interface
		if validator, ok := source.(interface{ Validate() error }); ok {
			if err := validator.Validate(); err != nil {
				results[name] = err
			}
		}
	}

	return results
}
