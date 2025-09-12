package source

import "errors"

// Source domain specific errors
var (
	// Validation errors
	ErrEmptySourceName    = errors.New("source name cannot be empty")
	ErrEmptySourceType    = errors.New("source type cannot be empty")
	ErrEmptySourceURL     = errors.New("source URL cannot be empty")
	ErrInvalidSourceType  = errors.New("invalid source type (must be rss, api, or scraper)")
	ErrInvalidSourceURL   = errors.New("invalid source URL format")
	ErrInvalidSchedule    = errors.New("invalid schedule format")
	ErrInvalidRateLimit   = errors.New("rate limit must be non-negative")
	ErrInvalidPage        = errors.New("page number must be positive")
	ErrInvalidLimit       = errors.New("limit must be between 1 and 1000")
	
	// Business logic errors
	ErrSourceNotFound       = errors.New("source not found")
	ErrSourceAlreadyExists  = errors.New("source already exists")
	ErrSourceDisabled       = errors.New("source is disabled")
	ErrSourceUnhealthy      = errors.New("source is unhealthy")
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")
	ErrFetchTimeout         = errors.New("fetch operation timed out")
	ErrFetchFailed          = errors.New("failed to fetch from source")
	ErrConnectionFailed     = errors.New("failed to connect to source")
	ErrInvalidResponse      = errors.New("invalid response from source")
	ErrParsingFailed        = errors.New("failed to parse source content")
)
