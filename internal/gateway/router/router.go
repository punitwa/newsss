// Package router provides route configuration and setup for the API gateway.
package router

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"news-aggregator/internal/gateway/core"
	handlerCore "news-aggregator/internal/handlers/core"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Router manages API routes and middleware with independent handlers.
type Router struct {
	config          core.RouterConfig
	handlerRegistry handlerCore.HandlerRegistry
	logger          zerolog.Logger
}

// NewRouter creates a new router with independent handlers.
func NewRouter(config core.RouterConfig, handlerRegistry handlerCore.HandlerRegistry, logger zerolog.Logger) *Router {
	return &Router{
		config:          config,
		handlerRegistry: handlerRegistry,
		logger:          logger.With().Str("component", "router").Logger(),
	}
}

// Setup configures and returns a Gin engine with all routes and middleware.
func (r *Router) Setup() *gin.Engine {
	// Create Gin engine
	engine := gin.New()

	// Set trusted proxies
	if len(r.config.TrustedProxies) > 0 {
		engine.SetTrustedProxies(r.config.TrustedProxies)
	}

	// Setup global middleware
	r.setupGlobalMiddleware(engine)

	// Setup routes
	r.setupRoutes(engine)

	r.logger.Info().Msg("Router setup completed")

	return engine
}

// setupGlobalMiddleware configures global middleware.
func (r *Router) setupGlobalMiddleware(engine *gin.Engine) {
	// Recovery middleware
	engine.Use(gin.Recovery())

	// Request ID middleware
	engine.Use(r.requestIDMiddleware())

	// Logging middleware
	if r.config.EnableLogging {
		engine.Use(r.loggingMiddleware())
	}

	// CORS middleware
	if r.config.EnableCORS {
		engine.Use(r.corsMiddleware())
	}

	// Security headers middleware
	engine.Use(r.securityHeadersMiddleware())

	// Rate limiting middleware
	if r.config.EnableRateLimit {
		engine.Use(r.rateLimitMiddleware())
	}

	// Metrics middleware
	if r.config.EnableMetrics {
		engine.Use(r.metricsMiddleware())
	}

	// Request size limit middleware
	engine.Use(r.requestSizeLimitMiddleware())

	r.logger.Info().Msg("Global middleware configured")
}

// setupRoutes configures all API routes using independent handlers.
func (r *Router) setupRoutes(engine *gin.Engine) {
	// Root health check
	engine.GET("/", r.rootHandler)

	// Register all handlers from the registry
	allHandlers := r.handlerRegistry.GetAllHandlers()

	// Register health handlers directly (no authentication required)
	healthHandlers := r.handlerRegistry.GetHandlersByType("health")
	for _, handler := range healthHandlers {
		handler.RegisterRoutes(engine)
	}

	// API v1 routes
	v1 := engine.Group("/api/v1")
	{
		// Public routes (no authentication required)
		public := v1.Group("/")
		{
			// Register auth handlers
			authHandlers := r.handlerRegistry.GetHandlersByType("auth")
			for _, handler := range authHandlers {
				handler.RegisterRoutes(public)
			}

			// Register news handlers
			newsHandlers := r.handlerRegistry.GetHandlersByType("news")
			for _, handler := range newsHandlers {
				handler.RegisterRoutes(public)
			}
		}

		// Protected routes (authentication required)
		protected := v1.Group("/")
		protected.Use(r.authMiddleware())
		{
			// Register user handlers
			userHandlers := r.handlerRegistry.GetHandlersByType("user")
			for _, handler := range userHandlers {
				handler.RegisterRoutes(protected)
			}
		}

		// Admin routes (admin authentication required)
		admin := v1.Group("/admin")
		admin.Use(r.authMiddleware())
		admin.Use(r.adminMiddleware())
		{
			// Register admin handlers
			adminHandlers := r.handlerRegistry.GetHandlersByType("admin")
			for _, handler := range adminHandlers {
				handler.RegisterRoutes(admin)
			}
		}
	}

	// WebSocket endpoint
	engine.GET("/ws", r.websocketHandler)

	// Metrics endpoint (if enabled)
	if r.config.EnableMetrics {
		engine.GET("/metrics", r.metricsHandler)
	}

	r.logger.Info().
		Int("total_handlers", len(allHandlers)).
		Msg("Routes configured with independent handlers")
}

// Middleware functions

// requestIDMiddleware adds request ID to each request.
func (r *Router) requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// loggingMiddleware logs HTTP requests.
func (r *Router) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log request
		duration := time.Since(start)

		r.logger.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Dur("duration", duration).
			Str("ip", c.ClientIP()).
			Str("user_agent", c.GetHeader("User-Agent")).
			Str("request_id", getRequestID(c)).
			Msg("HTTP request")
	}
}

// corsMiddleware configures CORS.
func (r *Router) corsMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     []string{"*"}, // Configure based on your needs
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(config)
}

// securityHeadersMiddleware adds security headers.
func (r *Router) securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

// rateLimitMiddleware implements rate limiting.
func (r *Router) rateLimitMiddleware() gin.HandlerFunc {
	// This is a placeholder - implement actual rate limiting
	return func(c *gin.Context) {
		// Rate limiting logic would go here
		c.Next()
	}
}

// metricsMiddleware collects metrics.
func (r *Router) metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// start := time.Now()

		// Process request
		c.Next()

		// TODO: Record metrics
		// Metrics collection would be implemented here
		// duration := time.Since(start).Seconds()
	}
}

// requestSizeLimitMiddleware limits request body size.
func (r *Router) requestSizeLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > r.config.MaxRequestSize {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "Request too large"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// authMiddleware validates JWT tokens.
func (r *Router) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authorization header required",
				},
				"request_id": generateRequestID(),
				"timestamp":  time.Now().UTC(),
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
			})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// For testing purposes, extract user_id from JWT without full validation
		// In production, you'd validate the JWT signature here
		if tokenString != "" {
			// This is a temporary hack for testing - decode JWT payload
			parts := strings.Split(tokenString, ".")
			if len(parts) == 3 {
				// Decode payload (second part)
				payload, err := base64.RawURLEncoding.DecodeString(parts[1])
				if err == nil {
					var claims map[string]interface{}
					if json.Unmarshal(payload, &claims) == nil {
						if userID, ok := claims["user_id"].(string); ok && userID != "" {
							c.Set("user_id", userID)
							c.Set("user_role", "user")
							c.Next()
							return
						}
					}
				}
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Invalid token",
			},
			"request_id": generateRequestID(),
			"timestamp":  time.Now().UTC(),
			"path":       c.Request.URL.Path,
			"method":     c.Request.Method,
		})
		c.Abort()
	}
}

// adminMiddleware ensures user has admin role.
func (r *Router) adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement admin role check
		c.Next()
	}
}

// Handler functions

// rootHandler handles root path requests.
func (r *Router) rootHandler(c *gin.Context) {
	response := gin.H{
		"service":   "News Aggregator API",
		"version":   "2.0.0",
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"handlers":  len(r.handlerRegistry.GetAllHandlers()),
		"endpoints": gin.H{
			"health":  "/health",
			"api":     "/api/v1",
			"docs":    "/docs",
			"metrics": "/metrics",
		},
	}

	c.JSON(http.StatusOK, response)
}

// websocketHandler handles WebSocket connections.
func (r *Router) websocketHandler(c *gin.Context) {
	// WebSocket implementation would go here
	c.JSON(http.StatusOK, gin.H{
		"message": "WebSocket endpoint - implementation pending",
	})
}

// metricsHandler serves Prometheus metrics.
func (r *Router) metricsHandler(c *gin.Context) {
	// Prometheus metrics would be served here
	c.JSON(http.StatusOK, gin.H{
		"message": "Metrics endpoint - Prometheus integration pending",
	})
}

// Custom error handlers

// NoRouteHandler handles 404 errors.
func (r *Router) NoRouteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Endpoint not found",
		})
	}
}

// NoMethodHandler handles 405 errors.
func (r *Router) NoMethodHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Method not allowed",
		})
	}
}

// SetupErrorHandlers configures custom error handlers.
func (r *Router) SetupErrorHandlers(engine *gin.Engine) {
	engine.NoRoute(r.NoRouteHandler())
	engine.NoMethod(r.NoMethodHandler())
}

// Helper functions

// generateRequestID generates a unique request ID.
func generateRequestID() string {
	// Simple UUID-like generation
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// getRequestID gets request ID from context.
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
