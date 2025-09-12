// Package core defines gateway-specific errors and error handling utilities.
package core

import (
	"errors"
	"fmt"
	"net/http"
)

// Common gateway errors
var (
	// Authentication errors
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInvalidToken     = errors.New("invalid token")
	ErrTokenExpired     = errors.New("token expired")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound     = errors.New("user not found")
	ErrUserExists       = errors.New("user already exists")
	
	// Authorization errors
	ErrForbidden        = errors.New("forbidden")
	ErrInsufficientRole = errors.New("insufficient role")
	ErrAccessDenied     = errors.New("access denied")
	
	// Validation errors
	ErrInvalidRequest   = errors.New("invalid request")
	ErrMissingParameter = errors.New("missing required parameter")
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrValidationFailed = errors.New("validation failed")
	
	// Resource errors
	ErrResourceNotFound = errors.New("resource not found")
	ErrResourceExists   = errors.New("resource already exists")
	ErrResourceLocked   = errors.New("resource locked")
	
	// Rate limiting errors
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrTooManyRequests   = errors.New("too many requests")
	
	// Service errors
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrServiceTimeout     = errors.New("service timeout")
	ErrExternalService    = errors.New("external service error")
	
	// Data errors
	ErrInvalidData      = errors.New("invalid data")
	ErrDataCorrupted    = errors.New("data corrupted")
	ErrDatabaseError    = errors.New("database error")
	
	// System errors
	ErrInternalError    = errors.New("internal server error")
	ErrConfigurationError = errors.New("configuration error")
	ErrSystemOverload   = errors.New("system overload")
)

// GatewayError represents a gateway-specific error with additional context.
type GatewayError struct {
	Code       string
	Message    string
	Details    map[string]string
	Cause      error
	HTTPStatus int
	Retryable  bool
	UserFacing bool
}

// Error implements the error interface.
func (e *GatewayError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *GatewayError) Unwrap() error {
	return e.Cause
}

// IsRetryable returns true if the error is retryable.
func (e *GatewayError) IsRetryable() bool {
	return e.Retryable
}

// IsUserFacing returns true if the error should be shown to users.
func (e *GatewayError) IsUserFacing() bool {
	return e.UserFacing
}

// GetHTTPStatus returns the appropriate HTTP status code for this error.
func (e *GatewayError) GetHTTPStatus() int {
	if e.HTTPStatus != 0 {
		return e.HTTPStatus
	}
	return http.StatusInternalServerError
}

// ToAPIError converts the gateway error to an API error response.
func (e *GatewayError) ToAPIError() APIError {
	return APIError{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Details,
	}
}

// NewGatewayError creates a new gateway error.
func NewGatewayError(code, message string, httpStatus int) *GatewayError {
	return &GatewayError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		UserFacing: true,
	}
}

// NewGatewayErrorWithCause creates a new gateway error with a cause.
func NewGatewayErrorWithCause(code, message string, httpStatus int, cause error) *GatewayError {
	return &GatewayError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Cause:      cause,
		UserFacing: true,
	}
}

// NewInternalError creates a new internal error (not user-facing).
func NewInternalError(message string, cause error) *GatewayError {
	return &GatewayError{
		Code:       CodeInternalError,
		Message:    "Internal server error",
		HTTPStatus: http.StatusInternalServerError,
		Cause:      cause,
		UserFacing: false,
	}
}

// ValidationError represents a validation error with field-specific details.
type ValidationError struct {
	*GatewayError
	Fields map[string]string
}

// NewValidationError creates a new validation error.
func NewValidationError(fields map[string]string) *ValidationError {
	return &ValidationError{
		GatewayError: &GatewayError{
			Code:       CodeValidationError,
			Message:    "Validation failed",
			HTTPStatus: http.StatusBadRequest,
			UserFacing: true,
		},
		Fields: fields,
	}
}

// AddField adds a field error to the validation error.
func (e *ValidationError) AddField(field, message string) {
	if e.Fields == nil {
		e.Fields = make(map[string]string)
	}
	e.Fields[field] = message
}

// HasErrors returns true if there are validation errors.
func (e *ValidationError) HasErrors() bool {
	return len(e.Fields) > 0
}

// ToAPIError converts validation error to API error.
func (e *ValidationError) ToAPIError() APIError {
	return APIError{
		Code:    e.Code,
		Message: e.Message,
		Details: e.Fields,
	}
}

// AuthenticationError represents an authentication error.
type AuthenticationError struct {
	*GatewayError
	TokenType string
	Realm     string
}

// NewAuthenticationError creates a new authentication error.
func NewAuthenticationError(message string) *AuthenticationError {
	return &AuthenticationError{
		GatewayError: &GatewayError{
			Code:       CodeUnauthorized,
			Message:    message,
			HTTPStatus: http.StatusUnauthorized,
			UserFacing: true,
		},
		TokenType: "Bearer",
		Realm:     "api",
	}
}

// AuthorizationError represents an authorization error.
type AuthorizationError struct {
	*GatewayError
	RequiredRole   string
	RequiredScopes []string
}

// NewAuthorizationError creates a new authorization error.
func NewAuthorizationError(message string) *AuthorizationError {
	return &AuthorizationError{
		GatewayError: &GatewayError{
			Code:       CodeForbidden,
			Message:    message,
			HTTPStatus: http.StatusForbidden,
			UserFacing: true,
		},
	}
}

// RateLimitError represents a rate limiting error.
type RateLimitError struct {
	*GatewayError
	Limit      int
	Remaining  int
	ResetTime  int64
	RetryAfter int
}

// NewRateLimitError creates a new rate limit error.
func NewRateLimitError(limit, remaining int, resetTime int64, retryAfter int) *RateLimitError {
	return &RateLimitError{
		GatewayError: &GatewayError{
			Code:       CodeRateLimited,
			Message:    "Rate limit exceeded",
			HTTPStatus: http.StatusTooManyRequests,
			UserFacing: true,
			Retryable:  true,
		},
		Limit:      limit,
		Remaining:  remaining,
		ResetTime:  resetTime,
		RetryAfter: retryAfter,
	}
}

// ServiceError represents a service-level error.
type ServiceError struct {
	*GatewayError
	Service   string
	Operation string
}

// NewServiceError creates a new service error.
func NewServiceError(service, operation, message string, cause error) *ServiceError {
	return &ServiceError{
		GatewayError: &GatewayError{
			Code:       CodeServiceError,
			Message:    message,
			HTTPStatus: http.StatusServiceUnavailable,
			Cause:      cause,
			UserFacing: false,
			Retryable:  true,
		},
		Service:   service,
		Operation: operation,
	}
}

// Error mapping functions

// MapErrorToHTTPStatus maps common errors to HTTP status codes.
func MapErrorToHTTPStatus(err error) int {
	if gatewayErr, ok := err.(*GatewayError); ok {
		return gatewayErr.GetHTTPStatus()
	}
	
	switch {
	case errors.Is(err, ErrUnauthorized), errors.Is(err, ErrInvalidToken), 
		 errors.Is(err, ErrTokenExpired), errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized
		
	case errors.Is(err, ErrForbidden), errors.Is(err, ErrInsufficientRole), 
		 errors.Is(err, ErrAccessDenied):
		return http.StatusForbidden
		
	case errors.Is(err, ErrResourceNotFound), errors.Is(err, ErrUserNotFound):
		return http.StatusNotFound
		
	case errors.Is(err, ErrInvalidRequest), errors.Is(err, ErrMissingParameter),
		 errors.Is(err, ErrInvalidParameter), errors.Is(err, ErrValidationFailed),
		 errors.Is(err, ErrUserExists), errors.Is(err, ErrResourceExists):
		return http.StatusBadRequest
		
	case errors.Is(err, ErrRateLimitExceeded), errors.Is(err, ErrTooManyRequests):
		return http.StatusTooManyRequests
		
	case errors.Is(err, ErrServiceUnavailable), errors.Is(err, ErrServiceTimeout),
		 errors.Is(err, ErrExternalService):
		return http.StatusServiceUnavailable
		
	case errors.Is(err, ErrResourceLocked):
		return http.StatusLocked
		
	default:
		return http.StatusInternalServerError
	}
}

// MapErrorToCode maps common errors to error codes.
func MapErrorToCode(err error) string {
	if gatewayErr, ok := err.(*GatewayError); ok {
		return gatewayErr.Code
	}
	
	switch {
	case errors.Is(err, ErrUnauthorized), errors.Is(err, ErrInvalidToken), 
		 errors.Is(err, ErrTokenExpired), errors.Is(err, ErrInvalidCredentials):
		return CodeUnauthorized
		
	case errors.Is(err, ErrForbidden), errors.Is(err, ErrInsufficientRole), 
		 errors.Is(err, ErrAccessDenied):
		return CodeForbidden
		
	case errors.Is(err, ErrResourceNotFound), errors.Is(err, ErrUserNotFound):
		return CodeNotFound
		
	case errors.Is(err, ErrInvalidRequest), errors.Is(err, ErrMissingParameter),
		 errors.Is(err, ErrInvalidParameter), errors.Is(err, ErrValidationFailed):
		return CodeValidationError
		
	case errors.Is(err, ErrRateLimitExceeded), errors.Is(err, ErrTooManyRequests):
		return CodeRateLimited
		
	case errors.Is(err, ErrServiceUnavailable), errors.Is(err, ErrServiceTimeout),
		 errors.Is(err, ErrExternalService):
		return CodeServiceError
		
	case errors.Is(err, ErrDatabaseError):
		return CodeDatabaseError
		
	default:
		return CodeInternalError
	}
}

// IsRetryableError checks if an error is retryable.
func IsRetryableError(err error) bool {
	if gatewayErr, ok := err.(*GatewayError); ok {
		return gatewayErr.IsRetryable()
	}
	
	// By default, consider service and external errors as retryable
	switch {
	case errors.Is(err, ErrServiceUnavailable), errors.Is(err, ErrServiceTimeout),
		 errors.Is(err, ErrExternalService), errors.Is(err, ErrRateLimitExceeded):
		return true
	default:
		return false
	}
}

// IsUserFacingError checks if an error should be shown to users.
func IsUserFacingError(err error) bool {
	if gatewayErr, ok := err.(*GatewayError); ok {
		return gatewayErr.IsUserFacing()
	}
	
	// By default, don't show internal errors to users
	switch {
	case errors.Is(err, ErrInternalError), errors.Is(err, ErrDatabaseError),
		 errors.Is(err, ErrConfigurationError), errors.Is(err, ErrSystemOverload):
		return false
	default:
		return true
	}
}

// SanitizeErrorMessage returns a user-safe error message.
func SanitizeErrorMessage(err error) string {
	if !IsUserFacingError(err) {
		return "An internal error occurred"
	}
	
	if gatewayErr, ok := err.(*GatewayError); ok {
		return gatewayErr.Message
	}
	
	return err.Error()
}
