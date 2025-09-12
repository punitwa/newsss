// Package core defines common errors used across all data sources.
package core

import (
	"errors"
	"fmt"
)

// Common errors for data sources
var (
	// ErrInvalidSourceName indicates an invalid or empty source name
	ErrInvalidSourceName = errors.New("invalid or empty source name")
	
	// ErrInvalidSourceType indicates an unsupported source type
	ErrInvalidSourceType = errors.New("invalid or unsupported source type")
	
	// ErrInvalidSourceURL indicates an invalid or empty source URL
	ErrInvalidSourceURL = errors.New("invalid or empty source URL")
	
	// ErrInvalidSchedule indicates an invalid schedule duration
	ErrInvalidSchedule = errors.New("invalid schedule duration")
	
	// ErrInvalidRateLimit indicates an invalid rate limit value
	ErrInvalidRateLimit = errors.New("invalid rate limit value")
	
	// ErrSourceNotFound indicates a source was not found
	ErrSourceNotFound = errors.New("source not found")
	
	// ErrSourceDisabled indicates a source is disabled
	ErrSourceDisabled = errors.New("source is disabled")
	
	// ErrFetchTimeout indicates a fetch operation timed out
	ErrFetchTimeout = errors.New("fetch operation timed out")
	
	// ErrRateLimitExceeded indicates rate limit was exceeded
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	
	// ErrInvalidContent indicates content could not be parsed
	ErrInvalidContent = errors.New("invalid or unparseable content")
	
	// ErrNoContent indicates no content was found
	ErrNoContent = errors.New("no content found")
	
	// ErrNetworkError indicates a network-related error
	ErrNetworkError = errors.New("network error")
	
	// ErrAuthenticationFailed indicates authentication failure
	ErrAuthenticationFailed = errors.New("authentication failed")
	
	// ErrQuotaExceeded indicates API quota was exceeded
	ErrQuotaExceeded = errors.New("API quota exceeded")
)

// SourceError represents an error from a specific data source.
type SourceError struct {
	// SourceName is the name of the source that generated the error
	SourceName string
	
	// SourceType is the type of the source
	SourceType SourceType
	
	// Operation is the operation that failed
	Operation string
	
	// Err is the underlying error
	Err error
	
	// Retryable indicates if the error is retryable
	Retryable bool
	
	// StatusCode contains HTTP status code if applicable
	StatusCode int
}

// Error implements the error interface.
func (se *SourceError) Error() string {
	return fmt.Sprintf("source %s (%s) %s: %v", se.SourceName, se.SourceType, se.Operation, se.Err)
}

// Unwrap returns the underlying error.
func (se *SourceError) Unwrap() error {
	return se.Err
}

// IsRetryable returns true if the error is retryable.
func (se *SourceError) IsRetryable() bool {
	return se.Retryable
}

// NewSourceError creates a new SourceError.
func NewSourceError(sourceName string, sourceType SourceType, operation string, err error) *SourceError {
	return &SourceError{
		SourceName: sourceName,
		SourceType: sourceType,
		Operation:  operation,
		Err:        err,
		Retryable:  isRetryableError(err),
	}
}

// NewSourceErrorWithCode creates a new SourceError with HTTP status code.
func NewSourceErrorWithCode(sourceName string, sourceType SourceType, operation string, err error, statusCode int) *SourceError {
	return &SourceError{
		SourceName: sourceName,
		SourceType: sourceType,
		Operation:  operation,
		Err:        err,
		Retryable:  isRetryableHTTPError(statusCode),
		StatusCode: statusCode,
	}
}

// isRetryableError determines if an error is retryable.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	// Check for specific error types that are retryable
	switch {
	case errors.Is(err, ErrFetchTimeout):
		return true
	case errors.Is(err, ErrNetworkError):
		return true
	case errors.Is(err, ErrRateLimitExceeded):
		return true
	default:
		return false
	}
}

// isRetryableHTTPError determines if an HTTP status code indicates a retryable error.
func isRetryableHTTPError(statusCode int) bool {
	switch statusCode {
	case 429: // Too Many Requests
		return true
	case 500, 502, 503, 504: // Server errors
		return true
	case 408: // Request Timeout
		return true
	default:
		return false
	}
}

// ValidationError represents a validation error.
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

// Error implements the error interface.
func (ve *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s' with value '%v': %s", ve.Field, ve.Value, ve.Message)
}

// NewValidationError creates a new ValidationError.
func NewValidationError(field string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// ParsingError represents a content parsing error.
type ParsingError struct {
	ContentType string
	Content     string
	Err         error
}

// Error implements the error interface.
func (pe *ParsingError) Error() string {
	return fmt.Sprintf("failed to parse %s content: %v", pe.ContentType, pe.Err)
}

// Unwrap returns the underlying error.
func (pe *ParsingError) Unwrap() error {
	return pe.Err
}

// NewParsingError creates a new ParsingError.
func NewParsingError(contentType string, content string, err error) *ParsingError {
	return &ParsingError{
		ContentType: contentType,
		Content:     content,
		Err:         err,
	}
}
