// Package core defines the fundamental interfaces and contracts for all data sources.
package core

import (
	"context"
	"time"

	"news-aggregator/internal/models"
)

// DataSource defines the contract that all data sources must implement.
// This interface ensures consistency across different source types (RSS, API, Scraper).
type DataSource interface {
	// Fetch retrieves news items from the data source
	Fetch(ctx context.Context) ([]models.News, error)
	
	// GetSchedule returns the polling interval for this source
	GetSchedule() time.Duration
	
	// GetName returns the unique name identifier for this source
	GetName() string
	
	// GetType returns the type of this source (rss, api, scraper)
	GetType() string
	
	// IsHealthy performs a health check on the data source
	IsHealthy(ctx context.Context) bool
	
	// Validate ensures the source configuration is valid
	Validate() error
}

// SourceMetrics provides metrics and statistics about a data source.
type SourceMetrics interface {
	// GetTotalFetches returns the total number of fetch operations
	GetTotalFetches() int64
	
	// GetSuccessfulFetches returns the number of successful fetch operations
	GetSuccessfulFetches() int64
	
	// GetFailedFetches returns the number of failed fetch operations
	GetFailedFetches() int64
	
	// GetLastFetchTime returns the timestamp of the last fetch operation
	GetLastFetchTime() time.Time
	
	// GetAverageResponseTime returns the average response time for fetch operations
	GetAverageResponseTime() time.Duration
	
	// ResetMetrics clears all metrics counters
	ResetMetrics()
}

// SourceHealth provides detailed health information about a data source.
type SourceHealth interface {
	// IsHealthy performs a basic health check
	IsHealthy(ctx context.Context) bool
	
	// GetHealthStatus returns detailed health information
	GetHealthStatus(ctx context.Context) HealthStatus
	
	// GetLastError returns the last error encountered
	GetLastError() error
}

// Parser defines the interface for parsing different data formats.
type Parser[T any] interface {
	// Parse converts raw data into structured format
	Parse(ctx context.Context, data []byte) (T, error)
	
	// Validate checks if the parsed data is valid
	Validate(data T) error
}

// HTTPClient defines the interface for HTTP operations.
type HTTPClient interface {
	// Get performs a GET request with context
	Get(ctx context.Context, url string, headers map[string]string) ([]byte, error)
	
	// Post performs a POST request with context
	Post(ctx context.Context, url string, body []byte, headers map[string]string) ([]byte, error)
	
	// SetTimeout sets the request timeout
	SetTimeout(timeout time.Duration)
	
	// SetUserAgent sets the User-Agent header
	SetUserAgent(userAgent string)
}

// RateLimiter defines the interface for rate limiting operations.
type RateLimiter interface {
	// Wait blocks until the rate limiter allows the operation
	Wait(ctx context.Context) error
	
	// Allow checks if an operation is allowed without blocking
	Allow() bool
	
	// SetLimit updates the rate limit
	SetLimit(limit float64)
}

// ContentProcessor defines the interface for processing fetched content.
type ContentProcessor interface {
	// Process transforms raw content into news items
	Process(ctx context.Context, content []byte, sourceURL string) ([]models.News, error)
	
	// ExtractImages extracts image URLs from content
	ExtractImages(ctx context.Context, content string) ([]string, error)
	
	// SanitizeContent cleans and sanitizes content
	SanitizeContent(content string) string
}
