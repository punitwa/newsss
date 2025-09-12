package shared

import "errors"

// Common shared errors
var (
	// Validation errors
	ErrInvalidPage      = errors.New("page number must be positive")
	ErrInvalidLimit     = errors.New("limit must be between 1 and 1000")
	ErrInvalidDateRange = errors.New("date from must be before date to")
	ErrInvalidSortOrder = errors.New("sort order must be 'asc' or 'desc'")
	
	// General errors
	ErrNotFound         = errors.New("resource not found")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrUnauthorized     = errors.New("unauthorized access")
	ErrForbidden        = errors.New("access forbidden")
	ErrBadRequest       = errors.New("bad request")
	ErrInternalError    = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
	ErrTimeout          = errors.New("operation timed out")
	ErrRateLimited      = errors.New("rate limit exceeded")
	ErrInvalidInput     = errors.New("invalid input")
	ErrValidationFailed = errors.New("validation failed")
)
