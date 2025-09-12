# Gateway Module - Refactored Architecture

The gateway module has been completely refactored to achieve **modularity**, **readability**, **comprehensive functionality**, and **clean code design**. This document outlines the new architecture and how to use it.

## 🏗️ New Architecture Overview

```
gateway/
├── gateway.go                  # Main interface & backward compatibility
├── README.md                   # This documentation
├── core/                       # Core interfaces and types
│   ├── interfaces.go           # Primary interfaces (Gateway, Handler, etc.)
│   ├── types.go               # Common types and structures
│   └── errors.go              # Gateway-specific errors
├── handlers/                   # HTTP handlers by domain
│   ├── auth/                  # Authentication handlers
│   │   └── auth.go            # Login, register, token management
│   ├── news/                  # News-related handlers
│   │   └── news.go            # News retrieval, search, categories
│   ├── user/                  # User profile and bookmarks
│   │   └── user.go            # Profile, bookmarks, preferences
│   ├── admin/                 # Admin functionality
│   │   └── admin.go           # User management, system admin
│   ├── health/                # Health checks and monitoring
│   │   └── health.go          # Health, readiness, liveness checks
│   └── websocket/             # WebSocket handlers
│       └── websocket.go       # Real-time updates
├── router/                    # Route configuration
│   ├── router.go              # Main router setup and middleware
│   ├── routes.go              # Route definitions
│   └── middleware.go          # Gateway-specific middleware
├── utils/                     # Utilities and helpers
│   ├── response.go            # Standardized API responses
│   ├── validation.go          # Request validation
│   ├── context.go             # Context utilities
│   └── pagination.go          # Pagination helpers
└── metrics/                   # Metrics and monitoring
    ├── collector.go           # Metrics collection
    └── prometheus.go          # Prometheus integration
```

## ✨ Key Improvements

### 1. **Modularity**
- **Domain-Separated Handlers**: Each domain (auth, news, user) has its own handler module
- **Interface-Driven Design**: Clear contracts between components
- **Pluggable Architecture**: Easy to add new handlers and middleware
- **Independent Modules**: Components can be tested and developed independently

### 2. **Readability**
- **Clear Naming Conventions**: Descriptive names for all components
- **Comprehensive Documentation**: Every public function and interface documented
- **Logical Organization**: Related functionality grouped together
- **Consistent Patterns**: Uniform error handling, response formats, validation

### 3. **Comprehensive Functionality**
- **Standardized Responses**: Consistent API response format across all endpoints
- **Advanced Validation**: Comprehensive request validation with detailed error messages
- **Rich Error Handling**: Detailed error types with proper HTTP status codes
- **Security Features**: Built-in security headers, CORS, rate limiting
- **Health Monitoring**: Multiple levels of health checks (basic, readiness, liveness)
- **Metrics Collection**: Built-in metrics for monitoring and observability

### 4. **Clean Code Design**
- **SOLID Principles**: Single responsibility, dependency inversion
- **Design Patterns**: Handler pattern, middleware pattern, dependency injection
- **Type Safety**: Strong typing with validation
- **Resource Management**: Proper cleanup and graceful shutdown

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "news-aggregator/internal/config"
    "news-aggregator/internal/gateway"
    
    "github.com/rs/zerolog"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }
    
    // Create logger
    logger := zerolog.New(os.Stdout)
    
    // Create gateway
    gw, err := gateway.New(cfg, logger)
    if err != nil {
        log.Fatal(err)
    }
    
    // Start server
    ctx := context.Background()
    if err := gw.Start(ctx, ":8080"); err != nil {
        log.Fatal(err)
    }
}
```

### Environment-Specific Configurations

```go
// Production gateway
gw, err := gateway.CreateProductionGateway(cfg, logger)

// Development gateway
gw, err := gateway.CreateDevelopmentGateway(cfg, logger)

// Test gateway
gw, err := gateway.CreateTestGateway(cfg, logger)
```

### Custom Configuration

```go
routerConfig := core.RouterConfig{
    EnableCORS:        true,
    EnableRateLimit:   true,
    RateLimitRequests: 500,
    EnableMetrics:     true,
    EnableLogging:     true,
    TrustedProxies:    []string{"10.0.0.0/8"},
    MaxRequestSize:    5 << 20, // 5MB
}

gw, err := gateway.NewWithConfig(cfg, logger, routerConfig)
```

## 🔧 Core Components

### Handler Interface

All handlers implement the core Handler interface:

```go
type Handler interface {
    RegisterRoutes(router gin.IRouter)
    GetBasePath() string
}
```

### Standardized Responses

All API responses follow a consistent format:

```go
// Success response
{
    "data": {...},
    "meta": {
        "pagination": {...},
        "count": 10,
        "updated_at": "2024-01-01T00:00:00Z"
    },
    "request_id": "uuid",
    "timestamp": "2024-01-01T00:00:00Z"
}

// Error response
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Validation failed",
        "details": {
            "email": "Invalid email format"
        }
    },
    "request_id": "uuid",
    "timestamp": "2024-01-01T00:00:00Z",
    "path": "/api/v1/auth/register",
    "method": "POST"
}
```

### Request Validation

Comprehensive validation with detailed error messages:

```go
// Validation error response
{
    "error": {
        "code": "VALIDATION_ERROR",
        "message": "Validation failed",
        "details": {
            "email": "Invalid email format",
            "password": "Password must contain at least one uppercase letter"
        }
    }
}
```

## 📊 Advanced Features

### Health Checks

Multiple levels of health monitoring:

```bash
# Basic health check
GET /health

# Readiness check (includes dependency checks)
GET /health/ready

# Liveness check (includes system resource checks)
GET /health/live

# Detailed status
GET /health/status
```

### Metrics and Monitoring

Built-in metrics collection:

```go
// Request metrics
gw.RecordRequest("GET", "/api/v1/news", 200, 0.125)

// Error metrics
gw.RecordError("news_service", "database_error")

// Custom counters
gw.IncrementCounter("user_registrations", map[string]string{
    "source": "web",
})
```

### Security Features

Comprehensive security measures:

- **Security Headers**: X-Content-Type-Options, X-Frame-Options, CSP
- **CORS Configuration**: Configurable origin policies
- **Rate Limiting**: Configurable per-endpoint rate limits
- **Request Size Limits**: Configurable maximum request sizes
- **JWT Authentication**: Secure token-based authentication
- **Input Validation**: Comprehensive input sanitization

## 🔄 API Endpoints

### Authentication Endpoints

```
POST   /api/v1/auth/login              # User login
POST   /api/v1/auth/register           # User registration
POST   /api/v1/auth/refresh            # Token refresh
POST   /api/v1/auth/logout             # User logout
POST   /api/v1/auth/forgot-password    # Password reset request
POST   /api/v1/auth/reset-password     # Password reset
GET    /api/v1/auth/verify-email/:token # Email verification
POST   /api/v1/auth/resend-verification # Resend verification email
```

### News Endpoints

```
GET    /api/v1/news                    # Get paginated news
GET    /api/v1/news/:id                # Get specific article
GET    /api/v1/news/categories         # Get categories
GET    /api/v1/news/sources            # Get sources
GET    /api/v1/news/trending           # Get trending topics
GET    /api/v1/news/latest             # Get latest news
GET    /api/v1/news/popular            # Get popular news
GET    /api/v1/news/feed/:category     # Get news by category
GET    /api/v1/news/feed/source/:source # Get news by source
GET    /api/v1/search                  # Search news (GET)
POST   /api/v1/search                  # Search news (POST)
```

### Health & Monitoring

```
GET    /health                         # Basic health check
GET    /health/ready                   # Readiness check
GET    /health/live                    # Liveness check
GET    /health/status                  # Detailed status
GET    /health/version                 # Version information
GET    /metrics                        # Prometheus metrics
```

## 🛠️ Handler Development

### Creating a New Handler

```go
package myhandler

import (
    "news-aggregator/internal/gateway/core"
    "github.com/gin-gonic/gin"
)

type Handler struct {
    context *core.HandlerContext
    logger  zerolog.Logger
}

func NewHandler(context *core.HandlerContext) *Handler {
    return &Handler{
        context: context,
        logger:  context.Services.Logger.With().Str("handler", "my_handler").Logger(),
    }
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
    group := router.Group(h.GetBasePath())
    {
        group.GET("/items", h.GetItems)
        group.POST("/items", h.CreateItem)
    }
}

func (h *Handler) GetBasePath() string {
    return "/my-handler"
}

func (h *Handler) GetItems(c *gin.Context) {
    // Implementation
    h.context.ResponseWriter.Success(c, items)
}
```

### Request Validation

```go
func (h *Handler) CreateItem(c *gin.Context) {
    var req MyRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.context.ResponseWriter.BadRequest(c, "Invalid request format")
        return
    }
    
    // Validate request
    if err := h.context.Validator.ValidateMyRequest(&req); err != nil {
        h.context.ResponseWriter.Error(c, err)
        return
    }
    
    // Process request...
}
```

### Error Handling

```go
// Return standardized error
if err != nil {
    h.context.ResponseWriter.Error(c, err)
    return
}

// Return custom error with specific status
h.context.ResponseWriter.ErrorWithCode(c, http.StatusConflict, "Resource already exists")

// Return validation errors
errors := map[string]string{
    "email": "Invalid email format",
    "password": "Password too weak",
}
h.context.ResponseWriter.ValidationError(c, errors)
```

## 🧪 Testing

### Handler Testing

```go
func TestMyHandler(t *testing.T) {
    // Setup test context
    context := createTestHandlerContext()
    handler := myhandler.NewHandler(context)
    
    // Setup test router
    router := gin.New()
    handler.RegisterRoutes(router)
    
    // Test endpoint
    w := httptest.NewRecorder()
    req, _ := http.NewRequest("GET", "/my-handler/items", nil)
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
}
```

### Integration Testing

```go
func TestGatewayIntegration(t *testing.T) {
    // Create test gateway
    gw, err := gateway.CreateTestGateway(testConfig, testLogger)
    require.NoError(t, err)
    
    // Start test server
    server := httptest.NewServer(gw.router.Setup())
    defer server.Close()
    
    // Test endpoints
    resp, err := http.Get(server.URL + "/health")
    require.NoError(t, err)
    assert.Equal(t, http.StatusOK, resp.StatusCode)
}
```

## 📈 Performance Considerations

### Optimization Features

- **Connection Pooling**: HTTP clients use connection pooling
- **Middleware Caching**: Response caching for appropriate endpoints
- **Request Size Limits**: Prevents memory exhaustion
- **Rate Limiting**: Protects against abuse
- **Graceful Shutdown**: Proper cleanup on shutdown

### Best Practices

```go
// Use context for timeouts
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Set appropriate cache headers
h.context.ContextManager.SetCacheHeaders(c, 300) // 5 minutes

// Record metrics
h.context.MetricsCollector.IncrementCounter("api_calls", map[string]string{
    "endpoint": "/news",
    "method":   "GET",
})
```

## 🔧 Configuration

### Router Configuration

```go
type RouterConfig struct {
    EnableCORS        bool     // Enable CORS middleware
    EnableRateLimit   bool     // Enable rate limiting
    RateLimitRequests int      // Requests per minute
    EnableMetrics     bool     // Enable metrics collection
    EnableLogging     bool     // Enable request logging
    TrustedProxies    []string // Trusted proxy IPs
    MaxRequestSize    int64    // Maximum request body size
}
```

### Environment Variables

```bash
# Server configuration
PORT=8080
HOST=0.0.0.0

# Security
JWT_SECRET=your-secret-key
CORS_ORIGINS=http://localhost:3000,https://yourdomain.com

# Rate limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=1000

# Monitoring
METRICS_ENABLED=true
HEALTH_CHECK_INTERVAL=30s
```

## 🚧 Migration from Legacy Code

The refactored gateway maintains **backward compatibility**:

### Old Code (Still Works)
```go
// Legacy approach
gw, err := gateway.New(cfg, logger)
router := gin.New()
gw.SetupRoutes(router) // This still works but shows deprecation warning
```

### New Recommended Approach
```go
// New modular approach
gw, err := gateway.New(cfg, logger)
if err := gw.Start(ctx, ":8080"); err != nil {
    log.Fatal(err)
}
```

## 🚀 Future Extensions

The modular architecture makes it easy to add:

- **New Handler Modules**: Additional API domains
- **Advanced Middleware**: Custom authentication, logging, metrics
- **WebSocket Support**: Real-time features
- **GraphQL Support**: Alternative API format
- **API Versioning**: Multiple API versions
- **Caching Layers**: Redis integration
- **Message Queues**: Async processing

## 📝 Summary

This refactored gateway module provides:

✅ **Modular Design** - Clear separation of concerns by domain  
✅ **Readable Code** - Well-structured, documented interfaces  
✅ **Comprehensive Features** - Full-featured API gateway functionality  
✅ **Clean Architecture** - SOLID principles and design patterns  
✅ **Backward Compatibility** - Existing code continues to work  
✅ **Extensibility** - Easy to add new features and handlers  
✅ **Security** - Built-in security measures and best practices  
✅ **Monitoring** - Comprehensive health checks and metrics  
✅ **Performance** - Optimized for production use  
✅ **Maintainability** - Easy to understand and modify  

The new architecture provides a solid foundation for building robust, scalable API gateways while maintaining the simplicity and ease of use of the original design.
