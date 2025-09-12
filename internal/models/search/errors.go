package search

import "errors"

// Search domain specific errors
var (
	// Validation errors
	ErrInvalidPage        = errors.New("page number must be positive")
	ErrInvalidLimit       = errors.New("limit must be between 1 and 1000")
	ErrInvalidDateRange   = errors.New("date from must be before date to")
	ErrInvalidSortBy      = errors.New("invalid sort by field")
	ErrInvalidSortOrder   = errors.New("invalid sort order (must be asc or desc)")
	ErrEmptyUserID        = errors.New("user ID cannot be empty")
	ErrEmptySearchName    = errors.New("search name cannot be empty")
	ErrEmptyQuery         = errors.New("search query cannot be empty")
	
	// Business logic errors
	ErrSearchNotFound     = errors.New("search not found")
	ErrSearchExists       = errors.New("search with this name already exists")
	ErrSearchTimeout      = errors.New("search operation timed out")
	ErrSearchFailed       = errors.New("search operation failed")
	ErrIndexNotAvailable  = errors.New("search index is not available")
	ErrTooManyResults     = errors.New("too many results, please refine your search")
	ErrInvalidSearchTerm  = errors.New("invalid search term")
	
	// Permission errors
	ErrUnauthorizedSearch = errors.New("unauthorized to perform this search")
	ErrSearchQuotaExceeded = errors.New("search quota exceeded")
)
