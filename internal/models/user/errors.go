package user

import "errors"

// User domain specific errors
var (
	// Validation errors
	ErrEmptyEmail             = errors.New("email cannot be empty")
	ErrEmptyUsername          = errors.New("username cannot be empty")
	ErrEmptyPassword          = errors.New("password cannot be empty")
	ErrEmptyFirstName         = errors.New("first name cannot be empty")
	ErrEmptyLastName          = errors.New("last name cannot be empty")
	ErrInvalidEmail           = errors.New("invalid email format")
	ErrInvalidUsername        = errors.New("invalid username format")
	ErrInvalidDigestFrequency = errors.New("invalid digest frequency")
	ErrInvalidTheme           = errors.New("invalid theme")
	
	// Password errors
	ErrPasswordTooShort = errors.New("password must be at least 6 characters long")
	ErrPasswordTooLong  = errors.New("password must be less than 128 characters long")
	ErrPasswordTooWeak  = errors.New("password must contain at least one uppercase letter, one lowercase letter, and one digit")
	ErrSamePassword     = errors.New("new password must be different from current password")
	ErrInvalidPassword  = errors.New("invalid password")
	
	// Business logic errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrUserInactive       = errors.New("user account is inactive")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	
	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token has expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenNotFound      = errors.New("token not found")
	
	// Bookmark errors
	ErrEmptyUserID       = errors.New("user ID cannot be empty")
	ErrEmptyNewsID       = errors.New("news ID cannot be empty")
	ErrBookmarkNotFound  = errors.New("bookmark not found")
	ErrBookmarkExists    = errors.New("bookmark already exists")
	ErrInvalidPage       = errors.New("page number must be positive")
	ErrInvalidLimit      = errors.New("limit must be between 1 and 1000")
	ErrInvalidDateRange  = errors.New("date from must be before date to")
)
