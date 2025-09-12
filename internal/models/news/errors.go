package news

import "errors"

// News domain specific errors
var (
	ErrEmptyTitle        = errors.New("news title cannot be empty")
	ErrEmptyURL          = errors.New("news URL cannot be empty")
	ErrEmptySource       = errors.New("news source cannot be empty")
	ErrEmptyCategoryName = errors.New("category name cannot be empty")
	ErrInvalidPage       = errors.New("page number must be positive")
	ErrInvalidLimit      = errors.New("limit must be between 1 and 1000")
	ErrInvalidDateRange  = errors.New("date from must be before date to")
	ErrNewsNotFound      = errors.New("news article not found")
	ErrCategoryNotFound  = errors.New("category not found")
	ErrDuplicateNews     = errors.New("news article already exists")
)
