// Package core provides base implementations for data sources.
package core

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

// BaseSource provides common functionality for all data sources.
// It implements basic metrics tracking, health monitoring, and configuration management.
type BaseSource struct {
	// Configuration
	config SourceConfig
	
	// Logger
	logger zerolog.Logger
	
	// Metrics (using atomic operations for thread safety)
	totalFetches      int64
	successfulFetches int64
	failedFetches     int64
	itemsCollected    int64
	totalResponseTime int64 // nanoseconds
	
	// State
	lastFetchTime time.Time
	lastError     error
	enabled       bool
	
	// Health status
	healthStatus HealthStatus
	healthMutex  sync.RWMutex
	
	// Metrics mutex for complex operations
	metricsMutex sync.RWMutex
}

// NewBaseSource creates a new BaseSource with the given configuration.
func NewBaseSource(config SourceConfig, logger zerolog.Logger) *BaseSource {
	return &BaseSource{
		config:  config,
		logger:  logger.With().Str("source", config.Name).Str("type", string(config.Type)).Logger(),
		enabled: config.Enabled,
		healthStatus: HealthStatus{
			IsHealthy:        true,
			LastChecked:      time.Now(),
			UptimePercentage: 100.0,
		},
	}
}

// GetName returns the source name.
func (bs *BaseSource) GetName() string {
	return bs.config.Name
}

// GetType returns the source type.
func (bs *BaseSource) GetType() string {
	return string(bs.config.Type)
}

// GetSchedule returns the polling schedule.
func (bs *BaseSource) GetSchedule() time.Duration {
	return bs.config.Schedule
}

// GetConfig returns a copy of the source configuration.
func (bs *BaseSource) GetConfig() SourceConfig {
	return bs.config
}

// IsEnabled returns true if the source is enabled.
func (bs *BaseSource) IsEnabled() bool {
	return bs.enabled
}

// SetEnabled enables or disables the source.
func (bs *BaseSource) SetEnabled(enabled bool) {
	bs.enabled = enabled
}

// Validate validates the source configuration.
func (bs *BaseSource) Validate() error {
	return bs.config.Validate()
}

// IsHealthy performs a basic health check.
func (bs *BaseSource) IsHealthy(ctx context.Context) bool {
	bs.healthMutex.RLock()
	defer bs.healthMutex.RUnlock()
	
	// Check if source is enabled
	if !bs.enabled {
		return false
	}
	
	// Check if last error was recent and critical
	if bs.lastError != nil && time.Since(bs.lastFetchTime) < bs.config.Schedule*2 {
		return false
	}
	
	return bs.healthStatus.IsHealthy
}

// GetHealthStatus returns the current health status.
func (bs *BaseSource) GetHealthStatus(ctx context.Context) HealthStatus {
	bs.healthMutex.RLock()
	defer bs.healthMutex.RUnlock()
	
	return bs.healthStatus
}

// UpdateHealthStatus updates the health status.
func (bs *BaseSource) UpdateHealthStatus(isHealthy bool, responseTime time.Duration, err error) {
	bs.healthMutex.Lock()
	defer bs.healthMutex.Unlock()
	
	bs.healthStatus.IsHealthy = isHealthy
	bs.healthStatus.LastChecked = time.Now()
	bs.healthStatus.ResponseTime = responseTime
	
	if err != nil {
		bs.healthStatus.ErrorMessage = err.Error()
		bs.lastError = err
	} else {
		bs.healthStatus.ErrorMessage = ""
	}
	
	// Update uptime percentage (simple moving average)
	if isHealthy {
		bs.healthStatus.UptimePercentage = (bs.healthStatus.UptimePercentage*0.9 + 10.0)
	} else {
		bs.healthStatus.UptimePercentage = bs.healthStatus.UptimePercentage * 0.9
	}
}

// GetLastError returns the last error encountered.
func (bs *BaseSource) GetLastError() error {
	return bs.lastError
}

// Metrics Implementation

// GetTotalFetches returns the total number of fetch operations.
func (bs *BaseSource) GetTotalFetches() int64 {
	return atomic.LoadInt64(&bs.totalFetches)
}

// GetSuccessfulFetches returns the number of successful fetch operations.
func (bs *BaseSource) GetSuccessfulFetches() int64 {
	return atomic.LoadInt64(&bs.successfulFetches)
}

// GetFailedFetches returns the number of failed fetch operations.
func (bs *BaseSource) GetFailedFetches() int64 {
	return atomic.LoadInt64(&bs.failedFetches)
}

// GetLastFetchTime returns the timestamp of the last fetch operation.
func (bs *BaseSource) GetLastFetchTime() time.Time {
	bs.metricsMutex.RLock()
	defer bs.metricsMutex.RUnlock()
	return bs.lastFetchTime
}

// GetAverageResponseTime returns the average response time for fetch operations.
func (bs *BaseSource) GetAverageResponseTime() time.Duration {
	totalFetches := atomic.LoadInt64(&bs.totalFetches)
	if totalFetches == 0 {
		return 0
	}
	
	totalTime := atomic.LoadInt64(&bs.totalResponseTime)
	return time.Duration(totalTime / totalFetches)
}

// GetItemsCollected returns the total number of items collected.
func (bs *BaseSource) GetItemsCollected() int64 {
	return atomic.LoadInt64(&bs.itemsCollected)
}

// ResetMetrics clears all metrics counters.
func (bs *BaseSource) ResetMetrics() {
	atomic.StoreInt64(&bs.totalFetches, 0)
	atomic.StoreInt64(&bs.successfulFetches, 0)
	atomic.StoreInt64(&bs.failedFetches, 0)
	atomic.StoreInt64(&bs.itemsCollected, 0)
	atomic.StoreInt64(&bs.totalResponseTime, 0)
	
	bs.metricsMutex.Lock()
	bs.lastFetchTime = time.Time{}
	bs.lastError = nil
	bs.metricsMutex.Unlock()
	
	bs.healthMutex.Lock()
	bs.healthStatus = HealthStatus{
		IsHealthy:        true,
		LastChecked:      time.Now(),
		UptimePercentage: 100.0,
	}
	bs.healthMutex.Unlock()
}

// GetStats returns comprehensive statistics for the source.
func (bs *BaseSource) GetStats() SourceStats {
	bs.metricsMutex.RLock()
	bs.healthMutex.RLock()
	defer bs.metricsMutex.RUnlock()
	defer bs.healthMutex.RUnlock()
	
	stats := SourceStats{
		Name:                bs.config.Name,
		Type:                bs.config.Type,
		TotalFetches:        atomic.LoadInt64(&bs.totalFetches),
		SuccessfulFetches:   atomic.LoadInt64(&bs.successfulFetches),
		FailedFetches:       atomic.LoadInt64(&bs.failedFetches),
		LastFetchTime:       bs.lastFetchTime,
		AverageResponseTime: bs.GetAverageResponseTime(),
		ItemsCollected:      atomic.LoadInt64(&bs.itemsCollected),
		Health:              bs.healthStatus,
	}
	
	if bs.lastError != nil {
		stats.LastError = bs.lastError.Error()
	}
	
	return stats
}

// RecordFetchStart records the start of a fetch operation.
func (bs *BaseSource) RecordFetchStart() {
	atomic.AddInt64(&bs.totalFetches, 1)
	
	bs.metricsMutex.Lock()
	bs.lastFetchTime = time.Now()
	bs.metricsMutex.Unlock()
}

// RecordFetchSuccess records a successful fetch operation.
func (bs *BaseSource) RecordFetchSuccess(responseTime time.Duration, itemCount int64) {
	atomic.AddInt64(&bs.successfulFetches, 1)
	atomic.AddInt64(&bs.itemsCollected, itemCount)
	atomic.AddInt64(&bs.totalResponseTime, responseTime.Nanoseconds())
	
	bs.UpdateHealthStatus(true, responseTime, nil)
	
	bs.logger.Debug().
		Dur("response_time", responseTime).
		Int64("items", itemCount).
		Msg("Fetch completed successfully")
}

// RecordFetchFailure records a failed fetch operation.
func (bs *BaseSource) RecordFetchFailure(responseTime time.Duration, err error) {
	atomic.AddInt64(&bs.failedFetches, 1)
	atomic.AddInt64(&bs.totalResponseTime, responseTime.Nanoseconds())
	
	bs.UpdateHealthStatus(false, responseTime, err)
	
	bs.logger.Error().
		Err(err).
		Dur("response_time", responseTime).
		Msg("Fetch failed")
}

// Logger returns the source logger.
func (bs *BaseSource) Logger() zerolog.Logger {
	return bs.logger
}
