// Package core defines common types and constants used across all data sources.
package core

import (
	"time"
)

// SourceType represents the different types of data sources supported.
type SourceType string

const (
	// SourceTypeRSS represents RSS feed sources
	SourceTypeRSS SourceType = "rss"
	
	// SourceTypeAPI represents REST API sources
	SourceTypeAPI SourceType = "api"
	
	// SourceTypeScraper represents web scraper sources
	SourceTypeScraper SourceType = "scraper"
)

// String returns the string representation of SourceType.
func (st SourceType) String() string {
	return string(st)
}

// IsValid checks if the SourceType is valid.
func (st SourceType) IsValid() bool {
	switch st {
	case SourceTypeRSS, SourceTypeAPI, SourceTypeScraper:
		return true
	default:
		return false
	}
}

// HealthStatus represents the health status of a data source.
type HealthStatus struct {
	// IsHealthy indicates if the source is healthy
	IsHealthy bool `json:"is_healthy"`
	
	// LastChecked is the timestamp of the last health check
	LastChecked time.Time `json:"last_checked"`
	
	// ResponseTime is the response time of the last health check
	ResponseTime time.Duration `json:"response_time"`
	
	// ErrorMessage contains the error message if not healthy
	ErrorMessage string `json:"error_message,omitempty"`
	
	// StatusCode contains the HTTP status code if applicable
	StatusCode int `json:"status_code,omitempty"`
	
	// Uptime percentage over the last period
	UptimePercentage float64 `json:"uptime_percentage"`
}

// SourceConfig represents the configuration for a data source.
type SourceConfig struct {
	// Name is the unique identifier for the source
	Name string `json:"name" yaml:"name"`
	
	// Type is the type of source (rss, api, scraper)
	Type SourceType `json:"type" yaml:"type"`
	
	// URL is the endpoint URL for the source
	URL string `json:"url" yaml:"url"`
	
	// Schedule defines how often to fetch from this source
	Schedule time.Duration `json:"schedule" yaml:"schedule"`
	
	// RateLimit defines the maximum requests per second
	RateLimit float64 `json:"rate_limit" yaml:"rate_limit"`
	
	// Headers contains custom HTTP headers
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	
	// Enabled indicates if the source is active
	Enabled bool `json:"enabled" yaml:"enabled"`
	
	// Timeout for HTTP requests
	Timeout time.Duration `json:"timeout" yaml:"timeout"`
	
	// MaxRetries for failed requests
	MaxRetries int `json:"max_retries" yaml:"max_retries"`
	
	// RetryDelay between retry attempts
	RetryDelay time.Duration `json:"retry_delay" yaml:"retry_delay"`
	
	// UserAgent for HTTP requests
	UserAgent string `json:"user_agent,omitempty" yaml:"user_agent,omitempty"`
	
	// Categories to filter content
	Categories []string `json:"categories,omitempty" yaml:"categories,omitempty"`
	
	// Keywords to filter content
	Keywords []string `json:"keywords,omitempty" yaml:"keywords,omitempty"`
	
	// Language preference
	Language string `json:"language,omitempty" yaml:"language,omitempty"`
	
	// Country preference
	Country string `json:"country,omitempty" yaml:"country,omitempty"`
}

// Validate checks if the SourceConfig is valid.
func (sc *SourceConfig) Validate() error {
	if sc.Name == "" {
		return ErrInvalidSourceName
	}
	
	if !sc.Type.IsValid() {
		return ErrInvalidSourceType
	}
	
	if sc.URL == "" {
		return ErrInvalidSourceURL
	}
	
	if sc.Schedule <= 0 {
		return ErrInvalidSchedule
	}
	
	if sc.RateLimit < 0 {
		return ErrInvalidRateLimit
	}
	
	if sc.Timeout <= 0 {
		sc.Timeout = 30 * time.Second // Default timeout
	}
	
	if sc.MaxRetries < 0 {
		sc.MaxRetries = 3 // Default retries
	}
	
	if sc.RetryDelay <= 0 {
		sc.RetryDelay = 1 * time.Second // Default retry delay
	}
	
	return nil
}

// GetDefaultUserAgent returns a default user agent string.
func (sc *SourceConfig) GetDefaultUserAgent() string {
	if sc.UserAgent != "" {
		return sc.UserAgent
	}
	return "NewsAggregator/1.0 (compatible; news collector)"
}

// SourceStats represents statistics for a data source.
type SourceStats struct {
	// Name of the source
	Name string `json:"name"`
	
	// Type of the source
	Type SourceType `json:"type"`
	
	// TotalFetches is the total number of fetch operations
	TotalFetches int64 `json:"total_fetches"`
	
	// SuccessfulFetches is the number of successful operations
	SuccessfulFetches int64 `json:"successful_fetches"`
	
	// FailedFetches is the number of failed operations
	FailedFetches int64 `json:"failed_fetches"`
	
	// LastFetchTime is the timestamp of the last fetch
	LastFetchTime time.Time `json:"last_fetch_time"`
	
	// AverageResponseTime is the average response time
	AverageResponseTime time.Duration `json:"average_response_time"`
	
	// ItemsCollected is the total number of items collected
	ItemsCollected int64 `json:"items_collected"`
	
	// LastError contains the last error message
	LastError string `json:"last_error,omitempty"`
	
	// Health status
	Health HealthStatus `json:"health"`
}

// GetSuccessRate returns the success rate as a percentage.
func (ss *SourceStats) GetSuccessRate() float64 {
	if ss.TotalFetches == 0 {
		return 0.0
	}
	return float64(ss.SuccessfulFetches) / float64(ss.TotalFetches) * 100.0
}

// ContentType represents different types of content that can be processed.
type ContentType string

const (
	// ContentTypeHTML represents HTML content
	ContentTypeHTML ContentType = "text/html"
	
	// ContentTypeXML represents XML content
	ContentTypeXML ContentType = "application/xml"
	
	// ContentTypeJSON represents JSON content
	ContentTypeJSON ContentType = "application/json"
	
	// ContentTypeText represents plain text content
	ContentTypeText ContentType = "text/plain"
)

// String returns the string representation of ContentType.
func (ct ContentType) String() string {
	return string(ct)
}

// ProcessingOptions contains options for content processing.
type ProcessingOptions struct {
	// ExtractImages indicates whether to extract images from content
	ExtractImages bool `json:"extract_images"`
	
	// SanitizeHTML indicates whether to sanitize HTML content
	SanitizeHTML bool `json:"sanitize_html"`
	
	// MaxContentLength limits the content length
	MaxContentLength int `json:"max_content_length"`
	
	// IncludeMetadata indicates whether to include metadata
	IncludeMetadata bool `json:"include_metadata"`
	
	// FilterDuplicates indicates whether to filter duplicate content
	FilterDuplicates bool `json:"filter_duplicates"`
}

// DefaultProcessingOptions returns default processing options.
func DefaultProcessingOptions() ProcessingOptions {
	return ProcessingOptions{
		ExtractImages:    true,
		SanitizeHTML:     true,
		MaxContentLength: 10000,
		IncludeMetadata:  true,
		FilterDuplicates: true,
	}
}
