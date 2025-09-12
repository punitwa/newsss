// Package utils provides validation utilities for data sources.
package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"news-aggregator/internal/datasources/core"
)

// URL validation regex patterns
var (
	// urlRegex matches valid HTTP/HTTPS URLs
	urlRegex = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	
	// domainRegex matches valid domain names
	domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	
	// emailRegex matches valid email addresses
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

// ValidateSourceConfig validates a source configuration.
func ValidateSourceConfig(config *core.SourceConfig) []error {
	var errors []error
	
	// Validate name
	if err := ValidateSourceName(config.Name); err != nil {
		errors = append(errors, err)
	}
	
	// Validate type
	if err := ValidateSourceType(config.Type); err != nil {
		errors = append(errors, err)
	}
	
	// Validate URL
	if err := ValidateURL(config.URL); err != nil {
		errors = append(errors, err)
	}
	
	// Validate schedule
	if err := ValidateSchedule(config.Schedule); err != nil {
		errors = append(errors, err)
	}
	
	// Validate rate limit
	if err := ValidateRateLimit(config.RateLimit); err != nil {
		errors = append(errors, err)
	}
	
	// Validate timeout
	if err := ValidateTimeout(config.Timeout); err != nil {
		errors = append(errors, err)
	}
	
	// Validate retry settings
	if err := ValidateRetrySettings(config.MaxRetries, config.RetryDelay); err != nil {
		errors = append(errors, err)
	}
	
	// Validate headers
	if err := ValidateHeaders(config.Headers); err != nil {
		errors = append(errors, err)
	}
	
	return errors
}

// ValidateSourceName validates a source name.
func ValidateSourceName(name string) error {
	if name == "" {
		return core.NewValidationError("name", name, "source name cannot be empty")
	}
	
	if len(name) > 100 {
		return core.NewValidationError("name", name, "source name cannot exceed 100 characters")
	}
	
	// Check for valid characters (alphanumeric, hyphens, underscores)
	validNameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validNameRegex.MatchString(name) {
		return core.NewValidationError("name", name, "source name can only contain alphanumeric characters, hyphens, and underscores")
	}
	
	return nil
}

// ValidateSourceType validates a source type.
func ValidateSourceType(sourceType core.SourceType) error {
	if !sourceType.IsValid() {
		return core.NewValidationError("type", sourceType, "invalid source type")
	}
	return nil
}

// ValidateURL validates a URL.
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return core.NewValidationError("url", urlStr, "URL cannot be empty")
	}
	
	// Parse URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return core.NewValidationError("url", urlStr, fmt.Sprintf("invalid URL format: %v", err))
	}
	
	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return core.NewValidationError("url", urlStr, "URL must use HTTP or HTTPS scheme")
	}
	
	// Check host
	if parsedURL.Host == "" {
		return core.NewValidationError("url", urlStr, "URL must have a valid host")
	}
	
	// Validate domain
	if !domainRegex.MatchString(parsedURL.Hostname()) {
		return core.NewValidationError("url", urlStr, "URL contains invalid domain name")
	}
	
	return nil
}

// ValidateSchedule validates a schedule duration.
func ValidateSchedule(schedule time.Duration) error {
	if schedule <= 0 {
		return core.NewValidationError("schedule", schedule, "schedule must be positive")
	}
	
	// Minimum schedule interval (prevent too frequent polling)
	minInterval := 30 * time.Second
	if schedule < minInterval {
		return core.NewValidationError("schedule", schedule, 
			fmt.Sprintf("schedule interval too short (minimum: %v)", minInterval))
	}
	
	// Maximum schedule interval (prevent schedules that are too long)
	maxInterval := 24 * time.Hour
	if schedule > maxInterval {
		return core.NewValidationError("schedule", schedule,
			fmt.Sprintf("schedule interval too long (maximum: %v)", maxInterval))
	}
	
	return nil
}

// ValidateRateLimit validates a rate limit value.
func ValidateRateLimit(rateLimit float64) error {
	if rateLimit < 0 {
		return core.NewValidationError("rate_limit", rateLimit, "rate limit cannot be negative")
	}
	
	// Maximum rate limit (prevent abuse)
	maxRateLimit := 100.0
	if rateLimit > maxRateLimit {
		return core.NewValidationError("rate_limit", rateLimit,
			fmt.Sprintf("rate limit too high (maximum: %.1f)", maxRateLimit))
	}
	
	return nil
}

// ValidateTimeout validates a timeout duration.
func ValidateTimeout(timeout time.Duration) error {
	if timeout <= 0 {
		return core.NewValidationError("timeout", timeout, "timeout must be positive")
	}
	
	// Minimum timeout
	minTimeout := 1 * time.Second
	if timeout < minTimeout {
		return core.NewValidationError("timeout", timeout,
			fmt.Sprintf("timeout too short (minimum: %v)", minTimeout))
	}
	
	// Maximum timeout
	maxTimeout := 5 * time.Minute
	if timeout > maxTimeout {
		return core.NewValidationError("timeout", timeout,
			fmt.Sprintf("timeout too long (maximum: %v)", maxTimeout))
	}
	
	return nil
}

// ValidateRetrySettings validates retry configuration.
func ValidateRetrySettings(maxRetries int, retryDelay time.Duration) error {
	if maxRetries < 0 {
		return core.NewValidationError("max_retries", maxRetries, "max retries cannot be negative")
	}
	
	if maxRetries > 10 {
		return core.NewValidationError("max_retries", maxRetries, "max retries cannot exceed 10")
	}
	
	if retryDelay < 0 {
		return core.NewValidationError("retry_delay", retryDelay, "retry delay cannot be negative")
	}
	
	if retryDelay > 1*time.Minute {
		return core.NewValidationError("retry_delay", retryDelay, "retry delay cannot exceed 1 minute")
	}
	
	return nil
}

// ValidateHeaders validates HTTP headers.
func ValidateHeaders(headers map[string]string) error {
	for key, value := range headers {
		if err := ValidateHeaderName(key); err != nil {
			return err
		}
		
		if err := ValidateHeaderValue(value); err != nil {
			return err
		}
	}
	return nil
}

// ValidateHeaderName validates an HTTP header name.
func ValidateHeaderName(name string) error {
	if name == "" {
		return core.NewValidationError("header_name", name, "header name cannot be empty")
	}
	
	// HTTP header names should only contain token characters
	validHeaderNameRegex := regexp.MustCompile(`^[a-zA-Z0-9!#$%&'*+\-.^_` + "`" + `|~]+$`)
	if !validHeaderNameRegex.MatchString(name) {
		return core.NewValidationError("header_name", name, "invalid header name format")
	}
	
	return nil
}

// ValidateHeaderValue validates an HTTP header value.
func ValidateHeaderValue(value string) error {
	// Header values can contain most characters, but not control characters
	for _, r := range value {
		if r < 32 && r != 9 { // Allow tab (9) but not other control characters
			return core.NewValidationError("header_value", value, "header value contains invalid control characters")
		}
	}
	
	return nil
}

// ValidateEmail validates an email address.
func ValidateEmail(email string) error {
	if email == "" {
		return core.NewValidationError("email", email, "email cannot be empty")
	}
	
	if !emailRegex.MatchString(email) {
		return core.NewValidationError("email", email, "invalid email format")
	}
	
	return nil
}

// ValidateLanguageCode validates an ISO 639-1 language code.
func ValidateLanguageCode(code string) error {
	if code == "" {
		return nil // Language code is optional
	}
	
	// Simple validation for 2-letter language codes
	if len(code) != 2 {
		return core.NewValidationError("language", code, "language code must be 2 characters")
	}
	
	validLanguageRegex := regexp.MustCompile(`^[a-z]{2}$`)
	if !validLanguageRegex.MatchString(code) {
		return core.NewValidationError("language", code, "invalid language code format")
	}
	
	return nil
}

// ValidateCountryCode validates an ISO 3166-1 alpha-2 country code.
func ValidateCountryCode(code string) error {
	if code == "" {
		return nil // Country code is optional
	}
	
	// Simple validation for 2-letter country codes
	if len(code) != 2 {
		return core.NewValidationError("country", code, "country code must be 2 characters")
	}
	
	validCountryRegex := regexp.MustCompile(`^[A-Z]{2}$`)
	if !validCountryRegex.MatchString(code) {
		return core.NewValidationError("country", code, "invalid country code format")
	}
	
	return nil
}

// SanitizeString removes potentially harmful characters from a string.
func SanitizeString(s string) string {
	// Remove control characters except tab and newline
	var result strings.Builder
	for _, r := range s {
		if r >= 32 || r == 9 || r == 10 || r == 13 {
			result.WriteRune(r)
		}
	}
	
	return strings.TrimSpace(result.String())
}

// ValidateContentLength validates content length limits.
func ValidateContentLength(content string, maxLength int) error {
	if maxLength <= 0 {
		return nil // No limit
	}
	
	if len(content) > maxLength {
		return core.NewValidationError("content_length", len(content),
			fmt.Sprintf("content exceeds maximum length of %d characters", maxLength))
	}
	
	return nil
}

// IsValidDomain checks if a string is a valid domain name.
func IsValidDomain(domain string) bool {
	return domainRegex.MatchString(domain)
}

// IsValidEmail checks if a string is a valid email address.
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsValidHTTPURL checks if a string is a valid HTTP/HTTPS URL.
func IsValidHTTPURL(urlStr string) bool {
	return urlRegex.MatchString(urlStr)
}
