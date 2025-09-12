// Package utils provides response utilities for standardized API responses.
package utils

import (
	"net/http"
	"time"

	"news-aggregator/internal/gateway/core"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// ResponseWriter implements standardized API response writing.
type ResponseWriter struct {
	logger zerolog.Logger
}

// NewResponseWriter creates a new response writer.
func NewResponseWriter(logger zerolog.Logger) core.ResponseWriter {
	return &ResponseWriter{
		logger: logger.With().Str("component", "response_writer").Logger(),
	}
}

// Success writes a successful response.
func (rw *ResponseWriter) Success(c *gin.Context, data interface{}) {
	response := core.SuccessResponse{
		Data:      data,
		RequestID: rw.getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	
	c.JSON(http.StatusOK, response)
}

// SuccessWithPagination writes a successful response with pagination.
func (rw *ResponseWriter) SuccessWithPagination(c *gin.Context, data interface{}, pagination core.PaginationInfo) {
	meta := &core.Meta{
		Pagination: &pagination,
	}
	
	response := core.SuccessResponse{
		Data:      data,
		Meta:      meta,
		RequestID: rw.getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	
	c.JSON(http.StatusOK, response)
}

// Error writes an error response.
func (rw *ResponseWriter) Error(c *gin.Context, err error) {
	statusCode := core.MapErrorToHTTPStatus(err)
	errorCode := core.MapErrorToCode(err)
	message := core.SanitizeErrorMessage(err)
	
	// Log the error with context
	rw.logger.Error().
		Err(err).
		Str("request_id", rw.getRequestID(c)).
		Str("path", c.Request.URL.Path).
		Str("method", c.Request.Method).
		Int("status_code", statusCode).
		Msg("Request failed")
	
	apiError := core.APIError{
		Code:    errorCode,
		Message: message,
	}
	
	// Add details for specific error types
	if validationErr, ok := err.(*core.ValidationError); ok {
		apiError.Details = validationErr.Fields
	}
	
	response := core.ErrorResponse{
		Error:     apiError,
		RequestID: rw.getRequestID(c),
		Timestamp: time.Now().UTC(),
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
	}
	
	c.JSON(statusCode, response)
}

// ErrorWithCode writes an error response with specific status code.
func (rw *ResponseWriter) ErrorWithCode(c *gin.Context, code int, message string) {
	apiError := core.APIError{
		Code:    rw.mapStatusCodeToErrorCode(code),
		Message: message,
	}
	
	response := core.ErrorResponse{
		Error:     apiError,
		RequestID: rw.getRequestID(c),
		Timestamp: time.Now().UTC(),
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
	}
	
	rw.logger.Warn().
		Str("request_id", rw.getRequestID(c)).
		Str("path", c.Request.URL.Path).
		Str("method", c.Request.Method).
		Int("status_code", code).
		Str("message", message).
		Msg("Request failed with custom error")
	
	c.JSON(code, response)
}

// ValidationError writes a validation error response.
func (rw *ResponseWriter) ValidationError(c *gin.Context, errors map[string]string) {
	apiError := core.APIError{
		Code:    core.CodeValidationError,
		Message: "Validation failed",
		Details: errors,
	}
	
	response := core.ErrorResponse{
		Error:     apiError,
		RequestID: rw.getRequestID(c),
		Timestamp: time.Now().UTC(),
		Path:      c.Request.URL.Path,
		Method:    c.Request.Method,
	}
	
	rw.logger.Warn().
		Str("request_id", rw.getRequestID(c)).
		Str("path", c.Request.URL.Path).
		Str("method", c.Request.Method).
		Interface("validation_errors", errors).
		Msg("Validation failed")
	
	c.JSON(http.StatusBadRequest, response)
}

// SuccessWithMeta writes a successful response with custom metadata.
func (rw *ResponseWriter) SuccessWithMeta(c *gin.Context, data interface{}, meta *core.Meta) {
	response := core.SuccessResponse{
		Data:      data,
		Meta:      meta,
		RequestID: rw.getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	
	c.JSON(http.StatusOK, response)
}

// Created writes a successful creation response.
func (rw *ResponseWriter) Created(c *gin.Context, data interface{}) {
	response := core.SuccessResponse{
		Data:      data,
		RequestID: rw.getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	
	c.JSON(http.StatusCreated, response)
}

// NoContent writes a successful response with no content.
func (rw *ResponseWriter) NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Accepted writes an accepted response for async operations.
func (rw *ResponseWriter) Accepted(c *gin.Context, data interface{}) {
	response := core.SuccessResponse{
		Data:      data,
		RequestID: rw.getRequestID(c),
		Timestamp: time.Now().UTC(),
	}
	
	c.JSON(http.StatusAccepted, response)
}

// NotFound writes a not found error response.
func (rw *ResponseWriter) NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "Resource not found"
	}
	
	rw.ErrorWithCode(c, http.StatusNotFound, message)
}

// Unauthorized writes an unauthorized error response.
func (rw *ResponseWriter) Unauthorized(c *gin.Context, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	
	rw.ErrorWithCode(c, http.StatusUnauthorized, message)
}

// Forbidden writes a forbidden error response.
func (rw *ResponseWriter) Forbidden(c *gin.Context, message string) {
	if message == "" {
		message = "Forbidden"
	}
	
	rw.ErrorWithCode(c, http.StatusForbidden, message)
}

// BadRequest writes a bad request error response.
func (rw *ResponseWriter) BadRequest(c *gin.Context, message string) {
	if message == "" {
		message = "Bad request"
	}
	
	rw.ErrorWithCode(c, http.StatusBadRequest, message)
}

// InternalError writes an internal server error response.
func (rw *ResponseWriter) InternalError(c *gin.Context, err error) {
	// Log the actual error but don't expose it to the client
	rw.logger.Error().
		Err(err).
		Str("request_id", rw.getRequestID(c)).
		Str("path", c.Request.URL.Path).
		Str("method", c.Request.Method).
		Msg("Internal server error")
	
	rw.ErrorWithCode(c, http.StatusInternalServerError, "Internal server error")
}

// RateLimited writes a rate limited error response.
func (rw *ResponseWriter) RateLimited(c *gin.Context, retryAfter int) {
	c.Header("Retry-After", string(rune(retryAfter)))
	rw.ErrorWithCode(c, http.StatusTooManyRequests, "Rate limit exceeded")
}

// ServiceUnavailable writes a service unavailable error response.
func (rw *ResponseWriter) ServiceUnavailable(c *gin.Context, message string) {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	
	rw.ErrorWithCode(c, http.StatusServiceUnavailable, message)
}

// Batch writes a batch operation response.
func (rw *ResponseWriter) Batch(c *gin.Context, response core.BatchResponse) {
	statusCode := http.StatusOK
	if response.Failed > 0 {
		statusCode = http.StatusMultiStatus
	}
	
	c.JSON(statusCode, response)
}

// getRequestID extracts request ID from context.
func (rw *ResponseWriter) getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// mapStatusCodeToErrorCode maps HTTP status codes to error codes.
func (rw *ResponseWriter) mapStatusCodeToErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return core.CodeBadRequest
	case http.StatusUnauthorized:
		return core.CodeUnauthorized
	case http.StatusForbidden:
		return core.CodeForbidden
	case http.StatusNotFound:
		return core.CodeNotFound
	case http.StatusTooManyRequests:
		return core.CodeRateLimited
	case http.StatusInternalServerError:
		return core.CodeInternalError
	case http.StatusServiceUnavailable:
		return core.CodeServiceError
	default:
		return core.CodeInternalError
	}
}

// Helper functions for common response patterns

// WriteHealthCheck writes a health check response.
func WriteHealthCheck(c *gin.Context, status core.HealthStatus) {
	var statusCode int
	switch status.Status {
	case core.StatusHealthy:
		statusCode = http.StatusOK
	case core.StatusDegraded:
		statusCode = http.StatusOK // Still OK but degraded
	case core.StatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	default:
		statusCode = http.StatusInternalServerError
	}
	
	c.JSON(statusCode, status)
}

// WriteMetrics writes a metrics response.
func WriteMetrics(c *gin.Context, metrics core.MetricsSnapshot) {
	c.JSON(http.StatusOK, metrics)
}

// WritePaginatedResponse writes a paginated response.
func WritePaginatedResponse(c *gin.Context, data interface{}, page, limit int, total int64) {
	pagination := core.NewPaginationInfo(page, limit, total)
	
	meta := &core.Meta{
		Pagination: &pagination,
	}
	
	response := core.SuccessResponse{
		Data:      data,
		Meta:      meta,
		RequestID: getRequestIDFromContext(c),
		Timestamp: time.Now().UTC(),
	}
	
	c.JSON(http.StatusOK, response)
}

// getRequestIDFromContext is a helper to get request ID from context.
func getRequestIDFromContext(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
