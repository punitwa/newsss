// Package core defines the fundamental interfaces and contracts for the API gateway.
package core

import (
	"context"

	"news-aggregator/internal/config"
	"news-aggregator/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Gateway defines the main gateway interface.
type Gateway interface {
	// SetupRoutes configures all API routes
	SetupRoutes(router *gin.Engine)

	// Start starts the gateway server
	Start(ctx context.Context, addr string) error

	// Stop gracefully shuts down the gateway
	Stop(ctx context.Context) error

	// GetConfig returns the gateway configuration
	GetConfig() *config.Config

	// GetLogger returns the gateway logger
	GetLogger() zerolog.Logger
}

// Handler defines the interface for HTTP handlers.
type Handler interface {
	// RegisterRoutes registers handler routes with the router
	RegisterRoutes(router gin.IRouter)

	// GetBasePath returns the base path for this handler
	GetBasePath() string
}

// AuthHandler defines authentication-related operations.
type AuthHandler interface {
	Handler

	// Login handles user login
	Login(c *gin.Context)

	// Register handles user registration
	Register(c *gin.Context)

	// RefreshToken handles token refresh
	RefreshToken(c *gin.Context)

	// Logout handles user logout
	Logout(c *gin.Context)
}

// NewsHandler defines news-related operations.
type NewsHandler interface {
	Handler

	// GetNews retrieves paginated news articles
	GetNews(c *gin.Context)

	// GetNewsByID retrieves a specific news article
	GetNewsByID(c *gin.Context)

	// GetCategories retrieves available news categories
	GetCategories(c *gin.Context)

	// SearchNews searches for news articles
	SearchNews(c *gin.Context)
}

// UserHandler defines user-related operations.
type UserHandler interface {
	Handler

	// GetProfile retrieves user profile
	GetProfile(c *gin.Context)

	// UpdateProfile updates user profile
	UpdateProfile(c *gin.Context)

	// AddBookmark adds a news bookmark
	AddBookmark(c *gin.Context)

	// GetBookmarks retrieves user bookmarks
	GetBookmarks(c *gin.Context)

	// RemoveBookmark removes a bookmark
	RemoveBookmark(c *gin.Context)

	// UpdatePreferences updates user preferences
	UpdatePreferences(c *gin.Context)
}

// AdminHandler defines admin-related operations.
type AdminHandler interface {
	Handler

	// GetUsers retrieves all users (admin only)
	GetUsers(c *gin.Context)

	// GetStats retrieves system statistics
	GetStats(c *gin.Context)

	// AddSource adds a new news source
	AddSource(c *gin.Context)

	// UpdateSource updates a news source
	UpdateSource(c *gin.Context)

	// DeleteSource deletes a news source
	DeleteSource(c *gin.Context)

	// CleanupOldArticles triggers cleanup of old articles
	CleanupOldArticles(c *gin.Context)

	// CleanupLogs triggers log cleanup
	CleanupLogs(c *gin.Context)
}

// HealthHandler defines health check operations.
type HealthHandler interface {
	Handler

	// HealthCheck performs basic health check
	HealthCheck(c *gin.Context)

	// ReadinessCheck performs readiness check
	ReadinessCheck(c *gin.Context)

	// LivenessCheck performs liveness check
	LivenessCheck(c *gin.Context)
}

// MetricsHandler defines metrics operations.
type MetricsHandler interface {
	Handler

	// GetMetrics returns Prometheus metrics
	GetMetrics(c *gin.Context)

	// GetStats returns application statistics
	GetStats(c *gin.Context)
}

// WebSocketHandler defines WebSocket operations.
type WebSocketHandler interface {
	Handler

	// HandleConnection handles WebSocket connections
	HandleConnection(c *gin.Context)

	// BroadcastNews broadcasts news updates to connected clients
	BroadcastNews(news interface{}) error

	// GetConnectedClients returns the number of connected clients
	GetConnectedClients() int
}

// TrendingHandler defines trending topics operations.
type TrendingHandler interface {
	Handler

	// GetTrendingTopics retrieves trending topics
	GetTrendingTopics(c *gin.Context)
}

// ResponseWriter defines the interface for standardized API responses.
type ResponseWriter interface {
	// Success writes a successful response
	Success(c *gin.Context, data interface{})

	// SuccessWithPagination writes a successful response with pagination
	SuccessWithPagination(c *gin.Context, data interface{}, pagination PaginationInfo)

	// Error writes an error response
	Error(c *gin.Context, err error)

	// ErrorWithCode writes an error response with specific status code
	ErrorWithCode(c *gin.Context, code int, message string)

	// ValidationError writes a validation error response
	ValidationError(c *gin.Context, errors map[string]string)

	// BadRequest writes a bad request error
	BadRequest(c *gin.Context, message string)

	// Unauthorized writes an unauthorized error
	Unauthorized(c *gin.Context, message string)

	// Forbidden writes a forbidden error
	Forbidden(c *gin.Context, message string)

	// NotFound writes a not found error
	NotFound(c *gin.Context, message string)

	// InternalError writes an internal server error
	InternalError(c *gin.Context, err error)

	// RateLimited writes a rate limited error
	RateLimited(c *gin.Context, retryAfter int)

	// ServiceUnavailable writes a service unavailable error
	ServiceUnavailable(c *gin.Context, message string)

	// Batch writes a batch operation response
	Batch(c *gin.Context, response BatchResponse)

	// SuccessWithMeta writes a successful response with custom metadata
	SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta)

	// Created writes a successful creation response
	Created(c *gin.Context, data interface{})

	// NoContent writes a successful response with no content
	NoContent(c *gin.Context)

	// Accepted writes an accepted response for async operations
	Accepted(c *gin.Context, data interface{})
}

// RequestValidator defines the interface for request validation.
type RequestValidator interface {
	// ValidateLogin validates login request
	ValidateLogin(req interface{}) error

	// ValidateRegistration validates registration request
	ValidateRegistration(req interface{}) error

	// ValidatePagination validates pagination parameters
	ValidatePagination(page, limit int) (int, int, error)

	// ValidateNewsFilter validates news filter parameters
	ValidateNewsFilter(filter interface{}) error
}

// ContextManager defines the interface for context management.
type ContextManager interface {
	// GetUserID extracts user ID from context
	GetUserID(c *gin.Context) (string, error)

	// SetUserID sets user ID in context
	SetUserID(c *gin.Context, userID string)

	// GetUserRole extracts user role from context
	GetUserRole(c *gin.Context) (string, error)

	// IsAdmin checks if user is admin
	IsAdmin(c *gin.Context) bool

	// GetRequestID gets or generates request ID
	GetRequestID(c *gin.Context) string
}

// MetricsCollector defines the interface for metrics collection.
type MetricsCollector interface {
	// RecordRequest records an HTTP request
	RecordRequest(method, path string, statusCode int, duration float64)

	// RecordError records an error
	RecordError(operation string, errorType string)

	// IncrementCounter increments a counter metric
	IncrementCounter(name string, labels map[string]string)

	// SetGauge sets a gauge metric
	SetGauge(name string, value float64, labels map[string]string)

	// RecordHistogram records a histogram metric
	RecordHistogram(name string, value float64, labels map[string]string)
}

// ServiceContainer holds all the services used by handlers.
type ServiceContainer struct {
	NewsService     *services.NewsService
	UserService     *services.UserService
	SearchService   *services.SearchService
	TrendingService *services.TrendingService
	Config          *config.Config
	Logger          zerolog.Logger
}

// HandlerContext provides context for handlers.
type HandlerContext struct {
	Services         *ServiceContainer
	ResponseWriter   ResponseWriter
	Validator        RequestValidator
	ContextManager   ContextManager
	MetricsCollector MetricsCollector
}

// RouterConfig defines configuration for the router.
type RouterConfig struct {
	// EnableCORS enables CORS middleware
	EnableCORS bool

	// EnableRateLimit enables rate limiting
	EnableRateLimit bool

	// RateLimitRequests requests per minute for rate limiting
	RateLimitRequests int

	// EnableMetrics enables metrics collection
	EnableMetrics bool

	// EnableLogging enables request logging
	EnableLogging bool

	// TrustedProxies list of trusted proxy IPs
	TrustedProxies []string

	// MaxRequestSize maximum request body size
	MaxRequestSize int64
}

// DefaultRouterConfig returns default router configuration.
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		EnableCORS:        true,
		EnableRateLimit:   true,
		RateLimitRequests: 100,
		EnableMetrics:     true,
		EnableLogging:     true,
		TrustedProxies:    []string{"127.0.0.1"},
		MaxRequestSize:    10 << 20, // 10MB
	}
}

// MiddlewareConfig defines middleware configuration.
type MiddlewareConfig struct {
	JWTSecretKey      string
	JWTExpirationTime int64
	EnableCORS        bool
	EnableRateLimit   bool
	RateLimit         int
	EnableMetrics     bool
	EnableLogging     bool
}

// WebSocketConfig defines WebSocket configuration.
type WebSocketConfig struct {
	// ReadBufferSize WebSocket read buffer size
	ReadBufferSize int

	// WriteBufferSize WebSocket write buffer size
	WriteBufferSize int

	// MaxMessageSize maximum message size
	MaxMessageSize int64

	// PingPeriod period for ping messages
	PingPeriod int64

	// PongWait timeout for pong messages
	PongWait int64

	// EnableCompression enables WebSocket compression
	EnableCompression bool
}

// DefaultWebSocketConfig returns default WebSocket configuration.
func DefaultWebSocketConfig() WebSocketConfig {
	return WebSocketConfig{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		MaxMessageSize:    512,
		PingPeriod:        54,
		PongWait:          60,
		EnableCompression: true,
	}
}
