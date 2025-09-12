package source

import (
	"net/url"
	"time"
)

// Source represents a news source
type Source struct {
	ID          string            `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Type        string            `json:"type" db:"type"` // rss, api, scraper
	URL         string            `json:"url" db:"url"`
	Schedule    string            `json:"schedule" db:"schedule"`
	RateLimit   int               `json:"rate_limit" db:"rate_limit"`
	Headers     map[string]string `json:"headers" db:"headers"`
	Enabled     bool              `json:"enabled" db:"enabled"`
	LastFetched time.Time         `json:"last_fetched" db:"last_fetched"`
	LastError   string            `json:"last_error" db:"last_error"`
	ErrorCount  int               `json:"error_count" db:"error_count"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

// SourceRequest represents a request to create or update a source
type SourceRequest struct {
	Name      string            `json:"name" binding:"required"`
	Type      string            `json:"type" binding:"required"`
	URL       string            `json:"url" binding:"required,url"`
	Schedule  string            `json:"schedule" binding:"required"`
	RateLimit int               `json:"rate_limit"`
	Headers   map[string]string `json:"headers"`
	Enabled   bool              `json:"enabled"`
}

// SourceFilter represents filtering options for sources
type SourceFilter struct {
	Type     string `json:"type"`
	Enabled  *bool  `json:"enabled"`
	Page     int    `json:"page"`
	Limit    int    `json:"limit"`
}

// SourceStats represents statistics for a source
type SourceStats struct {
	SourceID        string        `json:"source_id"`
	SourceName      string        `json:"source_name"`
	TotalFetches    int64         `json:"total_fetches"`
	SuccessfulFetches int64       `json:"successful_fetches"`
	FailedFetches   int64         `json:"failed_fetches"`
	LastFetched     time.Time     `json:"last_fetched"`
	AverageFetchTime time.Duration `json:"average_fetch_time"`
	ArticlesCollected int64        `json:"articles_collected"`
	ErrorRate       float64       `json:"error_rate"`
}

// HealthStatus represents the health status of a source
type HealthStatus struct {
	SourceID    string    `json:"source_id"`
	SourceName  string    `json:"source_name"`
	IsHealthy   bool      `json:"is_healthy"`
	LastChecked time.Time `json:"last_checked"`
	LastError   string    `json:"last_error,omitempty"`
	ErrorCount  int       `json:"error_count"`
	Uptime      float64   `json:"uptime"` // Percentage
}

// Validation methods

// Validate validates the Source struct
func (s *Source) Validate() error {
	if s.Name == "" {
		return ErrEmptySourceName
	}
	if s.Type == "" {
		return ErrEmptySourceType
	}
	if s.URL == "" {
		return ErrEmptySourceURL
	}
	
	// Validate source type
	validTypes := map[string]bool{
		"rss":     true,
		"api":     true,
		"scraper": true,
	}
	if !validTypes[s.Type] {
		return ErrInvalidSourceType
	}
	
	// Validate URL
	if _, err := url.Parse(s.URL); err != nil {
		return ErrInvalidSourceURL
	}
	
	// Validate schedule
	if s.Schedule != "" {
		if _, err := time.ParseDuration(s.Schedule); err != nil {
			return ErrInvalidSchedule
		}
	}
	
	// Validate rate limit
	if s.RateLimit < 0 {
		return ErrInvalidRateLimit
	}
	
	return nil
}

// Validate validates the SourceRequest
func (r *SourceRequest) Validate() error {
	if r.Name == "" {
		return ErrEmptySourceName
	}
	if r.Type == "" {
		return ErrEmptySourceType
	}
	if r.URL == "" {
		return ErrEmptySourceURL
	}
	
	// Validate source type
	validTypes := map[string]bool{
		"rss":     true,
		"api":     true,
		"scraper": true,
	}
	if !validTypes[r.Type] {
		return ErrInvalidSourceType
	}
	
	// Validate URL
	if _, err := url.Parse(r.URL); err != nil {
		return ErrInvalidSourceURL
	}
	
	// Validate schedule
	if r.Schedule != "" {
		if _, err := time.ParseDuration(r.Schedule); err != nil {
			return ErrInvalidSchedule
		}
	}
	
	// Validate rate limit
	if r.RateLimit < 0 {
		return ErrInvalidRateLimit
	}
	
	return nil
}

// Validate validates the SourceFilter
func (f *SourceFilter) Validate() error {
	if f.Page < 0 {
		return ErrInvalidPage
	}
	if f.Limit < 0 || f.Limit > 1000 {
		return ErrInvalidLimit
	}
	if f.Type != "" {
		validTypes := map[string]bool{
			"rss":     true,
			"api":     true,
			"scraper": true,
		}
		if !validTypes[f.Type] {
			return ErrInvalidSourceType
		}
	}
	return nil
}

// Helper methods

// IsHealthy returns true if the source is considered healthy
func (s *Source) IsHealthy() bool {
	// Consider a source unhealthy if it has more than 5 consecutive errors
	// or hasn't been successfully fetched in the last 24 hours
	if s.ErrorCount > 5 {
		return false
	}
	
	if s.Enabled && !s.LastFetched.IsZero() {
		return time.Since(s.LastFetched) < 24*time.Hour
	}
	
	return s.Enabled
}

// GetScheduleDuration returns the schedule as a time.Duration
func (s *Source) GetScheduleDuration() (time.Duration, error) {
	if s.Schedule == "" {
		return 5 * time.Minute, nil // Default schedule
	}
	return time.ParseDuration(s.Schedule)
}

// RecordError records an error for the source
func (s *Source) RecordError(err string) {
	s.LastError = err
	s.ErrorCount++
	s.UpdatedAt = time.Now()
}

// RecordSuccess records a successful fetch
func (s *Source) RecordSuccess() {
	s.LastFetched = time.Now()
	s.LastError = ""
	s.ErrorCount = 0
	s.UpdatedAt = time.Now()
}

// SetDefaults sets default values for the SourceFilter
func (f *SourceFilter) SetDefaults() {
	if f.Page == 0 {
		f.Page = 1
	}
	if f.Limit == 0 {
		f.Limit = 20
	}
}

// ToSource converts SourceRequest to Source
func (r *SourceRequest) ToSource() *Source {
	return &Source{
		Name:      r.Name,
		Type:      r.Type,
		URL:       r.URL,
		Schedule:  r.Schedule,
		RateLimit: r.RateLimit,
		Headers:   r.Headers,
		Enabled:   r.Enabled,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// ApplyToSource applies the request to an existing source
func (r *SourceRequest) ApplyToSource(source *Source) {
	source.Name = r.Name
	source.Type = r.Type
	source.URL = r.URL
	source.Schedule = r.Schedule
	source.RateLimit = r.RateLimit
	source.Headers = r.Headers
	source.Enabled = r.Enabled
	source.UpdatedAt = time.Now()
}

// CalculateErrorRate calculates the error rate for source stats
func (s *SourceStats) CalculateErrorRate() {
	if s.TotalFetches > 0 {
		s.ErrorRate = float64(s.FailedFetches) / float64(s.TotalFetches) * 100
	}
}

// CalculateUptime calculates the uptime percentage for health status
func (h *HealthStatus) CalculateUptime(totalChecks, successfulChecks int64) {
	if totalChecks > 0 {
		h.Uptime = float64(successfulChecks) / float64(totalChecks) * 100
	}
}
