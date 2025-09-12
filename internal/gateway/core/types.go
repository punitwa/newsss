// Package core defines common types and structures for the API gateway.
package core

import (
	"time"
)

// PaginationInfo contains pagination metadata.
type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	Pages      int64 `json:"pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
	NextPage   *int  `json:"next_page,omitempty"`
	PrevPage   *int  `json:"prev_page,omitempty"`
}

// NewPaginationInfo creates pagination info from parameters.
func NewPaginationInfo(page, limit int, total int64) PaginationInfo {
	pages := (total + int64(limit) - 1) / int64(limit)
	
	info := PaginationInfo{
		Page:    page,
		Limit:   limit,
		Total:   total,
		Pages:   pages,
		HasNext: int64(page) < pages,
		HasPrev: page > 1,
	}
	
	if info.HasNext {
		nextPage := page + 1
		info.NextPage = &nextPage
	}
	
	if info.HasPrev {
		prevPage := page - 1
		info.PrevPage = &prevPage
	}
	
	return info
}

// APIResponse represents a standardized API response.
type APIResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *APIError   `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// APIError represents an API error.
type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Meta contains metadata for API responses.
type Meta struct {
	Pagination *PaginationInfo `json:"pagination,omitempty"`
	Count      *int            `json:"count,omitempty"`
	UpdatedAt  *time.Time      `json:"updated_at,omitempty"`
}

// ValidationErrors represents validation error details.
type ValidationErrors struct {
	Fields map[string]string `json:"fields"`
}

// RequestContext contains request-specific information.
type RequestContext struct {
	RequestID string
	UserID    string
	UserRole  string
	StartTime time.Time
	Method    string
	Path      string
	IP        string
	UserAgent string
}

// HealthStatus represents the health status of the service.
type HealthStatus struct {
	Status      string                 `json:"status"`
	Version     string                 `json:"version"`
	Timestamp   time.Time              `json:"timestamp"`
	Services    map[string]ServiceInfo `json:"services,omitempty"`
	System      *SystemInfo            `json:"system,omitempty"`
}

// ServiceInfo represents the status of a service dependency.
type ServiceInfo struct {
	Status      string        `json:"status"`
	ResponseTime time.Duration `json:"response_time,omitempty"`
	Error       string        `json:"error,omitempty"`
	LastChecked time.Time     `json:"last_checked"`
}

// SystemInfo represents system resource information.
type SystemInfo struct {
	MemoryUsage    uint64    `json:"memory_usage_bytes"`
	CPUUsage       float64   `json:"cpu_usage_percent"`
	GoroutineCount int       `json:"goroutine_count"`
	Uptime         time.Duration `json:"uptime"`
}

// MetricsSnapshot represents a snapshot of metrics.
type MetricsSnapshot struct {
	RequestCount    int64             `json:"request_count"`
	ErrorCount      int64             `json:"error_count"`
	AverageResponse time.Duration     `json:"average_response_time"`
	StatusCodes     map[int]int64     `json:"status_codes"`
	Endpoints       map[string]int64  `json:"endpoints"`
	Timestamp       time.Time         `json:"timestamp"`
}

// WebSocketMessage represents a WebSocket message.
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	ID        string      `json:"id,omitempty"`
}

// WebSocketClient represents a connected WebSocket client.
type WebSocketClient struct {
	ID         string
	UserID     string
	Connection interface{} // WebSocket connection
	LastPing   time.Time
	Subscriptions []string
}

// Constants for API response codes
const (
	// Success codes
	CodeSuccess = "SUCCESS"
	
	// Client error codes
	CodeBadRequest     = "BAD_REQUEST"
	CodeUnauthorized   = "UNAUTHORIZED"
	CodeForbidden      = "FORBIDDEN"
	CodeNotFound       = "NOT_FOUND"
	CodeValidationError = "VALIDATION_ERROR"
	CodeRateLimited    = "RATE_LIMITED"
	
	// Server error codes
	CodeInternalError  = "INTERNAL_ERROR"
	CodeServiceError   = "SERVICE_ERROR"
	CodeDatabaseError  = "DATABASE_ERROR"
	CodeExternalError  = "EXTERNAL_ERROR"
)

// Constants for health status
const (
	StatusHealthy   = "healthy"
	StatusDegraded  = "degraded"
	StatusUnhealthy = "unhealthy"
)

// Constants for WebSocket message types
const (
	WSMessageTypeNews       = "news"
	WSMessageTypeTrending   = "trending"
	WSMessageTypeBookmark   = "bookmark"
	WSMessageTypeError      = "error"
	WSMessageTypeHeartbeat  = "heartbeat"
	WSMessageTypeSubscribe  = "subscribe"
	WSMessageTypeUnsubscribe = "unsubscribe"
)

// Default values
const (
	DefaultPage      = 1
	DefaultLimit     = 20
	MaxLimit         = 100
	DefaultTimeout   = 30 * time.Second
	MaxRequestSize   = 10 << 20 // 10MB
)

// RateLimitInfo contains rate limiting information.
type RateLimitInfo struct {
	Limit     int           `json:"limit"`
	Remaining int           `json:"remaining"`
	Reset     time.Time     `json:"reset"`
	RetryAfter time.Duration `json:"retry_after,omitempty"`
}

// AuthInfo contains authentication information.
type AuthInfo struct {
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	TokenType string    `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
	Scope     []string  `json:"scope,omitempty"`
}

// SearchQuery represents search parameters.
type SearchQuery struct {
	Query     string    `json:"query"`
	Category  string    `json:"category,omitempty"`
	Source    string    `json:"source,omitempty"`
	DateFrom  time.Time `json:"date_from,omitempty"`
	DateTo    time.Time `json:"date_to,omitempty"`
	SortBy    string    `json:"sort_by,omitempty"`
	SortOrder string    `json:"sort_order,omitempty"`
	Page      int       `json:"page"`
	Limit     int       `json:"limit"`
}

// NewsFilter represents news filtering parameters.
type NewsFilter struct {
	Category  string    `json:"category,omitempty"`
	Source    string    `json:"source,omitempty"`
	DateFrom  time.Time `json:"date_from,omitempty"`
	DateTo    time.Time `json:"date_to,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	Page      int       `json:"page"`
	Limit     int       `json:"limit"`
}

// AdminStats represents admin dashboard statistics.
type AdminStats struct {
	TotalUsers     int64     `json:"total_users"`
	ActiveUsers    int64     `json:"active_users"`
	TotalArticles  int64     `json:"total_articles"`
	TodayArticles  int64     `json:"today_articles"`
	TotalSources   int64     `json:"total_sources"`
	ActiveSources  int64     `json:"active_sources"`
	SystemUptime   time.Duration `json:"system_uptime"`
	LastUpdated    time.Time `json:"last_updated"`
}

// UserStats represents user-specific statistics.
type UserStats struct {
	BookmarkCount  int64     `json:"bookmark_count"`
	ReadArticles   int64     `json:"read_articles"`
	LastActivity   time.Time `json:"last_activity"`
	PreferredCategories []string `json:"preferred_categories"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error     APIError  `json:"error"`
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
	Path      string    `json:"path,omitempty"`
	Method    string    `json:"method,omitempty"`
}

// SuccessResponse represents a success response.
type SuccessResponse struct {
	Data      interface{} `json:"data"`
	Meta      *Meta       `json:"meta,omitempty"`
	RequestID string      `json:"request_id"`
	Timestamp time.Time   `json:"timestamp"`
}

// BatchResponse represents a batch operation response.
type BatchResponse struct {
	Success   []interface{} `json:"success"`
	Errors    []APIError    `json:"errors"`
	Total     int           `json:"total"`
	Processed int           `json:"processed"`
	Failed    int           `json:"failed"`
}

// CacheInfo contains caching information.
type CacheInfo struct {
	Cached    bool          `json:"cached"`
	TTL       time.Duration `json:"ttl,omitempty"`
	ExpiresAt time.Time     `json:"expires_at,omitempty"`
	Key       string        `json:"key,omitempty"`
}

// RequestLog represents a request log entry.
type RequestLog struct {
	RequestID    string        `json:"request_id"`
	Method       string        `json:"method"`
	Path         string        `json:"path"`
	StatusCode   int           `json:"status_code"`
	ResponseTime time.Duration `json:"response_time"`
	UserID       string        `json:"user_id,omitempty"`
	IP           string        `json:"ip"`
	UserAgent    string        `json:"user_agent"`
	Timestamp    time.Time     `json:"timestamp"`
	Error        string        `json:"error,omitempty"`
}
