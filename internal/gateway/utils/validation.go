// Package utils provides request validation utilities.
package utils

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"news-aggregator/internal/gateway/core"
	"news-aggregator/internal/models"

	"github.com/rs/zerolog"
)

// RequestValidator implements request validation functionality.
type RequestValidator struct {
	logger zerolog.Logger
}

// NewRequestValidator creates a new request validator.
func NewRequestValidator(logger zerolog.Logger) core.RequestValidator {
	return &RequestValidator{
		logger: logger.With().Str("component", "request_validator").Logger(),
	}
}

// ValidateLogin validates login request.
func (v *RequestValidator) ValidateLogin(req interface{}) error {
	loginReq, ok := req.(*models.LoginRequest)
	if !ok {
		return core.ErrInvalidRequest
	}
	
	errors := make(map[string]string)
	
	// Validate email
	if err := v.validateEmail(loginReq.Email); err != nil {
		errors["email"] = err.Error()
	}
	
	// Validate password
	if err := v.validatePassword(loginReq.Password, false); err != nil {
		errors["password"] = err.Error()
	}
	
	if len(errors) > 0 {
		return core.NewValidationError(errors)
	}
	
	return nil
}

// ValidateRegistration validates registration request.
func (v *RequestValidator) ValidateRegistration(req interface{}) error {
	regReq, ok := req.(*models.RegisterRequest)
	if !ok {
		return core.ErrInvalidRequest
	}
	
	errors := make(map[string]string)
	
	// Validate email
	if err := v.validateEmail(regReq.Email); err != nil {
		errors["email"] = err.Error()
	}
	
	// Validate password with strength requirements
	if err := v.validatePassword(regReq.Password, true); err != nil {
		errors["password"] = err.Error()
	}
	
	// Validate first name
	if err := v.validateName(regReq.FirstName); err != nil {
		errors["first_name"] = err.Error()
	}
	
	// Validate last name
	if err := v.validateName(regReq.LastName); err != nil {
		errors["last_name"] = err.Error()
	}
	
	if len(errors) > 0 {
		return core.NewValidationError(errors)
	}
	
	return nil
}

// ValidatePagination validates pagination parameters.
func (v *RequestValidator) ValidatePagination(page, limit int) (int, int, error) {
	// Validate and normalize page
	if page < 1 {
		page = core.DefaultPage
	}
	
	// Validate and normalize limit
	if limit < 1 {
		limit = core.DefaultLimit
	}
	if limit > core.MaxLimit {
		limit = core.MaxLimit
	}
	
	return page, limit, nil
}

// ValidateNewsFilter validates news filter parameters.
func (v *RequestValidator) ValidateNewsFilter(filter interface{}) error {
	newsFilter, ok := filter.(*core.NewsFilter)
	if !ok {
		return core.ErrInvalidRequest
	}
	
	errors := make(map[string]string)
	
	// Validate category
	if newsFilter.Category != "" {
		if err := v.validateCategory(newsFilter.Category); err != nil {
			errors["category"] = err.Error()
		}
	}
	
	// Validate source
	if newsFilter.Source != "" {
		if err := v.validateSource(newsFilter.Source); err != nil {
			errors["source"] = err.Error()
		}
	}
	
	// Validate date range
	if !newsFilter.DateFrom.IsZero() && !newsFilter.DateTo.IsZero() {
		if newsFilter.DateFrom.After(newsFilter.DateTo) {
			errors["date_range"] = "date_from must be before date_to"
		}
	}
	
	// Validate tags
	if len(newsFilter.Tags) > 0 {
		for i, tag := range newsFilter.Tags {
			if err := v.validateTag(tag); err != nil {
				errors[fmt.Sprintf("tags[%d]", i)] = err.Error()
			}
		}
	}
	
	// Validate pagination
	var err error
	newsFilter.Page, newsFilter.Limit, err = v.ValidatePagination(newsFilter.Page, newsFilter.Limit)
	if err != nil {
		errors["pagination"] = err.Error()
	}
	
	if len(errors) > 0 {
		return core.NewValidationError(errors)
	}
	
	return nil
}

// ValidateSearchQuery validates search query parameters.
func (v *RequestValidator) ValidateSearchQuery(query *core.SearchQuery) error {
	errors := make(map[string]string)
	
	// Validate query string
	if strings.TrimSpace(query.Query) == "" {
		errors["query"] = "query is required"
	} else if len(query.Query) < 2 {
		errors["query"] = "query must be at least 2 characters long"
	} else if len(query.Query) > 200 {
		errors["query"] = "query must be less than 200 characters"
	}
	
	// Validate category
	if query.Category != "" {
		if err := v.validateCategory(query.Category); err != nil {
			errors["category"] = err.Error()
		}
	}
	
	// Validate source
	if query.Source != "" {
		if err := v.validateSource(query.Source); err != nil {
			errors["source"] = err.Error()
		}
	}
	
	// Validate date range
	if !query.DateFrom.IsZero() && !query.DateTo.IsZero() {
		if query.DateFrom.After(query.DateTo) {
			errors["date_range"] = "date_from must be before date_to"
		}
	}
	
	// Validate sort parameters
	if query.SortBy != "" {
		validSortFields := []string{"published_at", "relevance", "title"}
		if !v.contains(validSortFields, query.SortBy) {
			errors["sort_by"] = "invalid sort field"
		}
	}
	
	if query.SortOrder != "" {
		if query.SortOrder != "asc" && query.SortOrder != "desc" {
			errors["sort_order"] = "sort_order must be 'asc' or 'desc'"
		}
	}
	
	// Validate pagination
	var err error
	query.Page, query.Limit, err = v.ValidatePagination(query.Page, query.Limit)
	if err != nil {
		errors["pagination"] = err.Error()
	}
	
	if len(errors) > 0 {
		return core.NewValidationError(errors)
	}
	
	return nil
}

// ValidateBookmarkRequest validates bookmark request.
func (v *RequestValidator) ValidateBookmarkRequest(req *models.BookmarkRequest) error {
	errors := make(map[string]string)
	
	if strings.TrimSpace(req.NewsID) == "" {
		errors["news_id"] = "news_id is required"
	}
	
	if len(errors) > 0 {
		return core.NewValidationError(errors)
	}
	
	return nil
}

// ValidatePreferencesRequest validates preferences request.
func (v *RequestValidator) ValidatePreferencesRequest(req *models.PreferencesRequest) error {
	errors := make(map[string]string)
	
	// Validate categories
	if len(req.Categories) > 10 {
		errors["categories"] = "maximum 10 categories allowed"
	}
	
	for i, category := range req.Categories {
		if err := v.validateCategory(category); err != nil {
			errors[fmt.Sprintf("categories[%d]", i)] = err.Error()
		}
	}
	
	// Validate sources
	if len(req.Sources) > 20 {
		errors["sources"] = "maximum 20 sources allowed"
	}
	
	for i, source := range req.Sources {
		if err := v.validateSource(source); err != nil {
			errors[fmt.Sprintf("sources[%d]", i)] = err.Error()
		}
	}
	
	// Validate language
	if req.Language != "" {
		if err := v.validateLanguage(req.Language); err != nil {
			errors["language"] = err.Error()
		}
	}
	
	if len(errors) > 0 {
		return core.NewValidationError(errors)
	}
	
	return nil
}

// ValidateUpdateProfileRequest validates profile update request.
func (v *RequestValidator) ValidateUpdateProfileRequest(req *models.UpdateProfileRequest) error {
	errors := make(map[string]string)
	
	// Validate first name
	if req.FirstName != "" {
		if err := v.validateName(req.FirstName); err != nil {
			errors["first_name"] = err.Error()
		}
	}
	
	// Validate last name
	if req.LastName != "" {
		if err := v.validateName(req.LastName); err != nil {
			errors["last_name"] = err.Error()
		}
	}
	
	// Validate username
	if req.Username != "" {
		if err := v.validateName(req.Username); err != nil {
			errors["username"] = err.Error()
		}
	}
	
	if len(errors) > 0 {
		return core.NewValidationError(errors)
	}
	
	return nil
}

// Private validation methods

// validateEmail validates email format.
func (v *RequestValidator) validateEmail(email string) error {
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email is required")
	}
	
	if len(email) > 254 {
		return fmt.Errorf("email is too long")
	}
	
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format")
	}
	
	return nil
}

// validatePassword validates password strength.
func (v *RequestValidator) validatePassword(password string, enforceStrength bool) error {
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("password is required")
	}
	
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}
	
	if len(password) > 128 {
		return fmt.Errorf("password is too long")
	}
	
	if enforceStrength {
		// Check for at least one uppercase letter
		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
		if !hasUpper {
			return fmt.Errorf("password must contain at least one uppercase letter")
		}
		
		// Check for at least one lowercase letter
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
		if !hasLower {
			return fmt.Errorf("password must contain at least one lowercase letter")
		}
		
		// Check for at least one digit
		hasDigit := regexp.MustCompile(`\d`).MatchString(password)
		if !hasDigit {
			return fmt.Errorf("password must contain at least one digit")
		}
		
		// Check for at least one special character
		hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
		if !hasSpecial {
			return fmt.Errorf("password must contain at least one special character")
		}
	}
	
	return nil
}

// validateName validates name field.
func (v *RequestValidator) validateName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("name is required")
	}
	
	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters long")
	}
	
	if len(name) > 100 {
		return fmt.Errorf("name must be less than 100 characters")
	}
	
	// Check for valid characters (letters, spaces, hyphens, apostrophes)
	validName := regexp.MustCompile(`^[a-zA-Z\s\-']+$`).MatchString(name)
	if !validName {
		return fmt.Errorf("name contains invalid characters")
	}
	
	return nil
}

// validateCategory validates category name.
func (v *RequestValidator) validateCategory(category string) error {
	category = strings.TrimSpace(category)
	if category == "" {
		return fmt.Errorf("category cannot be empty")
	}
	
	if len(category) > 50 {
		return fmt.Errorf("category name is too long")
	}
	
	// Check for valid characters (alphanumeric, spaces, hyphens, underscores)
	validCategory := regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`).MatchString(category)
	if !validCategory {
		return fmt.Errorf("category contains invalid characters")
	}
	
	return nil
}

// validateSource validates source name.
func (v *RequestValidator) validateSource(source string) error {
	source = strings.TrimSpace(source)
	if source == "" {
		return fmt.Errorf("source cannot be empty")
	}
	
	if len(source) > 100 {
		return fmt.Errorf("source name is too long")
	}
	
	// Check for valid characters (alphanumeric, spaces, hyphens, underscores, dots)
	validSource := regexp.MustCompile(`^[a-zA-Z0-9\s\-_.]+$`).MatchString(source)
	if !validSource {
		return fmt.Errorf("source contains invalid characters")
	}
	
	return nil
}

// validateTag validates tag.
func (v *RequestValidator) validateTag(tag string) error {
	tag = strings.TrimSpace(tag)
	if tag == "" {
		return fmt.Errorf("tag cannot be empty")
	}
	
	if len(tag) > 30 {
		return fmt.Errorf("tag is too long")
	}
	
	// Check for valid characters (alphanumeric, hyphens, underscores)
	validTag := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`).MatchString(tag)
	if !validTag {
		return fmt.Errorf("tag contains invalid characters")
	}
	
	return nil
}

// validateLanguage validates language code.
func (v *RequestValidator) validateLanguage(language string) error {
	language = strings.TrimSpace(language)
	if language == "" {
		return fmt.Errorf("language cannot be empty")
	}
	
	// Check for valid ISO 639-1 language code format (2 letters)
	validLanguage := regexp.MustCompile(`^[a-z]{2}$`).MatchString(language)
	if !validLanguage {
		return fmt.Errorf("invalid language code format")
	}
	
	// List of supported languages
	supportedLanguages := []string{
		"en", "es", "fr", "de", "it", "pt", "ru", "zh", "ja", "ko",
		"ar", "hi", "tr", "nl", "sv", "no", "da", "fi", "pl", "cs",
	}
	
	if !v.contains(supportedLanguages, language) {
		return fmt.Errorf("unsupported language")
	}
	
	return nil
}

// validateTimezone validates timezone.
func (v *RequestValidator) validateTimezone(timezone string) error {
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		return fmt.Errorf("timezone cannot be empty")
	}
	
	// Try to load the timezone
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("invalid timezone")
	}
	
	return nil
}

// contains checks if a slice contains a string.
func (v *RequestValidator) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// SanitizeString removes potentially harmful characters from input.
func SanitizeString(input string) string {
	// Remove control characters except tab and newline
	sanitized := strings.Map(func(r rune) rune {
		if r < 32 && r != 9 && r != 10 && r != 13 {
			return -1
		}
		return r
	}, input)
	
	// Trim whitespace
	return strings.TrimSpace(sanitized)
}

// ValidateID validates that an ID is a valid format.
func ValidateID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	
	if len(id) > 100 {
		return fmt.Errorf("ID is too long")
	}
	
	// Check for valid ID characters (alphanumeric, hyphens, underscores)
	validID := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`).MatchString(id)
	if !validID {
		return fmt.Errorf("ID contains invalid characters")
	}
	
	return nil
}
