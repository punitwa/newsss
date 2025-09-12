package shared

import (
	"time"
)

// PaginationRequest represents common pagination parameters
type PaginationRequest struct {
	Page  int `json:"page" form:"page"`
	Limit int `json:"limit" form:"limit"`
}

// PaginationResponse represents common pagination metadata
type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// DateRange represents a date range filter
type DateRange struct {
	From time.Time `json:"from" form:"from"`
	To   time.Time `json:"to" form:"to"`
}

// SortOptions represents sorting options
type SortOptions struct {
	Field string `json:"field" form:"field"`
	Order string `json:"order" form:"order"` // asc, desc
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
	Message    string      `json:"message,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
	RequestID  string      `json:"request_id,omitempty"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error     string    `json:"error"`
	Message   string    `json:"message"`
	Code      int       `json:"code"`
	Timestamp time.Time `json:"timestamp"`
	RequestID string    `json:"request_id,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Data      interface{} `json:"data"`
	Message   string      `json:"message,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
	Pagination *PaginationResponse `json:"pagination,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// Validation methods

// Validate validates the PaginationRequest
func (p *PaginationRequest) Validate() error {
	if p.Page < 0 {
		return ErrInvalidPage
	}
	if p.Limit < 0 || p.Limit > 1000 {
		return ErrInvalidLimit
	}
	return nil
}

// Validate validates the DateRange
func (d *DateRange) Validate() error {
	if !d.From.IsZero() && !d.To.IsZero() && d.From.After(d.To) {
		return ErrInvalidDateRange
	}
	return nil
}

// Validate validates the SortOptions
func (s *SortOptions) Validate() error {
	if s.Order != "" && s.Order != "asc" && s.Order != "desc" {
		return ErrInvalidSortOrder
	}
	return nil
}

// Helper methods

// SetDefaults sets default values for pagination
func (p *PaginationRequest) SetDefaults() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.Limit == 0 {
		p.Limit = 20
	}
}

// GetOffset returns the offset for database queries
func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.Limit
}

// SetDefaults sets default values for sort options
func (s *SortOptions) SetDefaults() {
	if s.Order == "" {
		s.Order = "desc"
	}
}

// IsEmpty returns true if the date range is empty
func (d *DateRange) IsEmpty() bool {
	return d.From.IsZero() && d.To.IsZero()
}

// Contains returns true if the given time is within the date range
func (d *DateRange) Contains(t time.Time) bool {
	if d.IsEmpty() {
		return true // Empty range contains everything
	}
	
	if !d.From.IsZero() && t.Before(d.From) {
		return false
	}
	
	if !d.To.IsZero() && t.After(d.To) {
		return false
	}
	
	return true
}

// NewAPIResponse creates a new API response
func NewAPIResponse(success bool, data interface{}) *APIResponse {
	return &APIResponse{
		Success:   success,
		Data:      data,
		Timestamp: time.Now(),
	}
}

// NewSuccessResponse creates a new success response
func NewSuccessResponse(data interface{}, message string) *SuccessResponse {
	return &SuccessResponse{
		Data:      data,
		Message:   message,
		Timestamp: time.Now(),
	}
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err string, code int, message string) *ErrorResponse {
	return &ErrorResponse{
		Error:     err,
		Message:   message,
		Code:      code,
		Timestamp: time.Now(),
	}
}

// WithPagination adds pagination metadata to the response
func (r *APIResponse) WithPagination(page, limit int, total int64) *APIResponse {
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	
	r.Pagination = &PaginationResponse{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
	
	return r
}

// WithRequestID adds a request ID to the response
func (r *APIResponse) WithRequestID(requestID string) *APIResponse {
	r.RequestID = requestID
	return r
}

// WithError adds an error to the response
func (r *APIResponse) WithError(err error) *APIResponse {
	r.Success = false
	if err != nil {
		r.Error = err.Error()
	}
	return r
}

// WithMessage adds a message to the response
func (r *APIResponse) WithMessage(message string) *APIResponse {
	r.Message = message
	return r
}

// AddError adds a validation error
func (v *ValidationErrors) AddError(field, message, code string) {
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// HasErrors returns true if there are validation errors
func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

// Error implements the error interface
func (v *ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return "no validation errors"
	}
	return v.Errors[0].Message
}
