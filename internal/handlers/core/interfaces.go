// Package core defines the fundamental interfaces for the independent handler layer.
package core

import (
	"news-aggregator/internal/config"
	"news-aggregator/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Handler defines the base interface for all HTTP handlers.
// Handlers are independent of the gateway and can be used in any HTTP framework.
type Handler interface {
	// RegisterRoutes registers handler routes with the given router
	RegisterRoutes(router gin.IRouter)

	// GetBasePath returns the base path for this handler
	GetBasePath() string

	// GetName returns a unique name for this handler
	GetName() string
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

	// ForgotPassword handles forgot password requests
	ForgotPassword(c *gin.Context)

	// ResetPassword handles password reset
	ResetPassword(c *gin.Context)
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

	// GetTrendingTopics retrieves trending topics
	GetTrendingTopics(c *gin.Context)
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

// HandlerDependencies contains all dependencies needed by handlers.
// This replaces the gateway-specific HandlerContext.
type HandlerDependencies struct {
	// Services
	NewsService     *services.NewsService
	UserService     *services.UserService
	SearchService   *services.SearchService
	TrendingService *services.TrendingService

	// Configuration
	Config *config.Config

	// Logger
	Logger zerolog.Logger

	// Utilities (these will be injected, not gateway-specific)
	ResponseWriter ResponseWriter
	Validator      RequestValidator
	ContextManager ContextManager
}

// ResponseWriter defines the interface for standardized API responses.
// This is independent of any specific gateway implementation.
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

	// ValidateSearchQuery validates search query parameters
	ValidateSearchQuery(query interface{}) error

	// ValidateUpdateProfileRequest validates profile update request
	ValidateUpdateProfileRequest(req interface{}) error

	// ValidatePreferencesRequest validates preferences request
	ValidatePreferencesRequest(req interface{}) error
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

	// RequireAuth ensures user is authenticated
	RequireAuth(c *gin.Context) error

	// RequireAdmin ensures user is admin
	RequireAdmin(c *gin.Context) error
}

// PaginationInfo contains pagination metadata.
type PaginationInfo struct {
	Page     int   `json:"page"`
	Limit    int   `json:"limit"`
	Total    int64 `json:"total"`
	Pages    int64 `json:"pages"`
	HasNext  bool  `json:"has_next"`
	HasPrev  bool  `json:"has_prev"`
	NextPage *int  `json:"next_page,omitempty"`
	PrevPage *int  `json:"prev_page,omitempty"`
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

// HandlerRegistry manages handler registration and discovery.
type HandlerRegistry interface {
	// RegisterHandler registers a handler
	RegisterHandler(handler Handler) error

	// GetHandler retrieves a handler by name
	GetHandler(name string) (Handler, error)

	// GetAllHandlers returns all registered handlers
	GetAllHandlers() []Handler

	// GetHandlersByType returns handlers of a specific type
	GetHandlersByType(handlerType string) []Handler
}

// HandlerFactory creates handlers with their dependencies.
type HandlerFactory interface {
	// CreateAuthHandler creates an authentication handler
	CreateAuthHandler(deps *HandlerDependencies) AuthHandler

	// CreateNewsHandler creates a news handler
	CreateNewsHandler(deps *HandlerDependencies) NewsHandler

	// CreateUserHandler creates a user handler
	CreateUserHandler(deps *HandlerDependencies) UserHandler

	// CreateAdminHandler creates an admin handler
	CreateAdminHandler(deps *HandlerDependencies) AdminHandler

	// CreateHealthHandler creates a health handler
	CreateHealthHandler(deps *HandlerDependencies) HealthHandler
}

// HandlerConfig contains configuration for handlers.
type HandlerConfig struct {
	// EnableMetrics enables metrics collection for handlers
	EnableMetrics bool

	// EnableValidation enables request validation
	EnableValidation bool

	// EnableLogging enables handler-level logging
	EnableLogging bool

	// DefaultPageSize default page size for pagination
	DefaultPageSize int

	// MaxPageSize maximum page size for pagination
	MaxPageSize int

	// RequestTimeout timeout for handler operations
	RequestTimeout int
}

// DefaultHandlerConfig returns default handler configuration.
func DefaultHandlerConfig() HandlerConfig {
	return HandlerConfig{
		EnableMetrics:    true,
		EnableValidation: true,
		EnableLogging:    true,
		DefaultPageSize:  20,
		MaxPageSize:      100,
		RequestTimeout:   30,
	}
}
