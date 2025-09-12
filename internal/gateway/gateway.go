// Package gateway provides a modular, extensible API gateway for the news aggregation system.
// This package has been refactored for better modularity, readability, and maintainability.
// Handlers are now completely independent of the gateway implementation.
package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"news-aggregator/internal/config"
	"news-aggregator/internal/gateway/core"
	"news-aggregator/internal/gateway/router"
	"news-aggregator/internal/gateway/utils"

	"news-aggregator/internal/handlers/auth"
	handlerCore "news-aggregator/internal/handlers/core"
	"news-aggregator/internal/handlers/health"
	"news-aggregator/internal/handlers/news"
	"news-aggregator/internal/handlers/user"
	"news-aggregator/internal/models"
	"news-aggregator/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Gateway implements the main gateway functionality with modular architecture.
// It now uses independent handlers that are not tightly coupled to the gateway.
type Gateway struct {
	config *config.Config
	logger zerolog.Logger
	router *router.Router
	server *http.Server

	// Independent handler layer
	handlerRegistry handlerCore.HandlerRegistry
	handlerDeps     *handlerCore.HandlerDependencies

	// Services
	newsService     *services.NewsService
	userService     *services.UserService
	searchService   *services.SearchService
	trendingService *services.TrendingService
}

// New creates a new gateway instance with all dependencies.
func New(cfg *config.Config, logger zerolog.Logger) (*Gateway, error) {
	return NewWithConfig(cfg, logger, core.DefaultRouterConfig())
}

// NewWithConfig creates a new gateway instance with custom router configuration.
func NewWithConfig(cfg *config.Config, logger zerolog.Logger, routerConfig core.RouterConfig) (*Gateway, error) {
	// Initialize services
	newsService, err := services.NewNewsService(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create news service: %w", err)
	}

	userService, err := services.NewUserService(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create user service: %w", err)
	}

	searchService, err := services.NewSearchService(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create search service: %w", err)
	}

	// Initialize trending service
	trendingService := services.NewTrendingService(newsService.GetRepository(), logger)

	// Create utilities for handlers (independent of gateway)
	responseWriter := utils.NewResponseWriter(logger)
	validator := utils.NewRequestValidator(logger)
	contextManager := utils.NewContextManager(logger)

	// Create adapters to make gateway interfaces compatible with handler interfaces
	responseAdapter := &responseWriterAdapter{responseWriter}
	validatorAdapter := &requestValidatorAdapter{validator}
	contextAdapter := &contextManagerAdapter{contextManager}

	// Create independent handler dependencies
	handlerDeps := &handlerCore.HandlerDependencies{
		NewsService:     newsService,
		UserService:     userService,
		SearchService:   searchService,
		TrendingService: trendingService,
		Config:          cfg,
		Logger:          logger,
		ResponseWriter:  responseAdapter,
		Validator:       validatorAdapter,
		ContextManager:  contextAdapter,
	}

	// Create handler registry
	handlerRegistry := handlerCore.NewHandlerRegistry(logger)

	// Create and register independent handlers
	handlerConfig := handlerCore.DefaultHandlerConfig()

	// Create independent handlers
	authHandler := auth.NewHandler(handlerDeps, handlerConfig)
	newsHandler := news.NewHandler(handlerDeps, handlerConfig)
	userHandler := user.NewHandler(handlerDeps, handlerConfig)
	healthHandler := health.NewHandler(handlerDeps, handlerConfig)

	// Register handlers
	if err := handlerRegistry.RegisterHandler(authHandler); err != nil {
		return nil, fmt.Errorf("failed to register auth handler: %w", err)
	}
	if err := handlerRegistry.RegisterHandler(newsHandler); err != nil {
		return nil, fmt.Errorf("failed to register news handler: %w", err)
	}
	if err := handlerRegistry.RegisterHandler(userHandler); err != nil {
		return nil, fmt.Errorf("failed to register user handler: %w", err)
	}
	if err := handlerRegistry.RegisterHandler(healthHandler); err != nil {
		return nil, fmt.Errorf("failed to register health handler: %w", err)
	}

	// Create router with independent handlers
	gatewayRouter := router.NewRouter(routerConfig, handlerRegistry, logger)

	gateway := &Gateway{
		config:          cfg,
		logger:          logger.With().Str("component", "gateway").Logger(),
		router:          gatewayRouter,
		handlerRegistry: handlerRegistry,
		handlerDeps:     handlerDeps,
		newsService:     newsService,
		userService:     userService,
		searchService:   searchService,
		trendingService: trendingService,
	}

	return gateway, nil
}

// Start starts the gateway server.
func (g *Gateway) Start(ctx context.Context, addr string) error {
	// Setup Gin engine
	engine := g.router.Setup()

	// Setup error handlers
	g.router.SetupErrorHandlers(engine)

	// Create HTTP server
	g.server = &http.Server{
		Addr:           addr,
		Handler:        engine,
		ReadTimeout:    30 * time.Second,
		WriteTimeout:   30 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	g.logger.Info().
		Str("addr", addr).
		Msg("Starting gateway server")

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := g.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server failed to start: %w", err)
		}
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		g.logger.Info().Msg("Gateway server context cancelled")
		return g.Stop(context.Background())
	case err := <-errChan:
		return err
	}
}

// Stop gracefully shuts down the gateway server.
func (g *Gateway) Stop(ctx context.Context) error {
	if g.server == nil {
		return nil
	}

	g.logger.Info().Msg("Shutting down gateway server")

	// Create shutdown context with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := g.server.Shutdown(shutdownCtx); err != nil {
		g.logger.Error().Err(err).Msg("Server shutdown failed")
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	g.logger.Info().Msg("Gateway server stopped")
	return nil
}

// GetConfig returns the gateway configuration.
func (g *Gateway) GetConfig() *config.Config {
	return g.config
}

// GetLogger returns the gateway logger.
func (g *Gateway) GetLogger() zerolog.Logger {
	return g.logger
}

// SetupRoutes configures all API routes (backward compatibility).
func (g *Gateway) SetupRoutes(router *gin.Engine) {
	// This method is kept for backward compatibility
	// The actual route setup is now handled by the router module
	g.logger.Warn().Msg("SetupRoutes called - this method is deprecated, routes are now auto-configured")

	// If someone calls this method, we'll setup basic routes for compatibility
	g.setupLegacyRoutes(router)
}

// setupLegacyRoutes sets up routes in the old style for backward compatibility.
func (g *Gateway) setupLegacyRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", g.legacyHealthCheck)
	router.GET("/metrics", g.legacyMetrics)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		public := v1.Group("/")
		{
			public.POST("/auth/login", g.legacyLogin)
			public.POST("/auth/register", g.legacyRegister)
			public.GET("/news", g.legacyGetNews)
			public.GET("/news/:id", g.legacyGetNewsById)
			public.GET("/search", g.legacySearchNews)
			public.GET("/categories", g.legacyGetCategories)
			public.GET("/trending", g.legacyGetTrendingTopics)
		}
	}

	// WebSocket endpoint
	router.GET("/ws", g.legacyWebsocketHandler)
}

// Legacy handler methods for backward compatibility
func (g *Gateway) legacyHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

func (g *Gateway) legacyMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Metrics endpoint - integrate with Prometheus",
	})
}

func (g *Gateway) legacyLogin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Use the new modular handlers"})
}

func (g *Gateway) legacyRegister(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Use the new modular handlers"})
}

func (g *Gateway) legacyGetNews(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Use the new modular handlers"})
}

func (g *Gateway) legacyGetNewsById(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Use the new modular handlers"})
}

func (g *Gateway) legacySearchNews(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Use the new modular handlers"})
}

func (g *Gateway) legacyGetCategories(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Use the new modular handlers"})
}

func (g *Gateway) legacyGetTrendingTopics(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Use the new modular handlers"})
}

func (g *Gateway) legacyWebsocketHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "WebSocket endpoint - implement real-time updates",
	})
}

// NoOpMetricsCollector is a placeholder metrics collector.
type NoOpMetricsCollector struct{}

func (n *NoOpMetricsCollector) RecordRequest(method, path string, statusCode int, duration float64) {
	// No-op implementation
}

func (n *NoOpMetricsCollector) RecordError(operation string, errorType string) {
	// No-op implementation
}

func (n *NoOpMetricsCollector) IncrementCounter(name string, labels map[string]string) {
	// No-op implementation
}

func (n *NoOpMetricsCollector) SetGauge(name string, value float64, labels map[string]string) {
	// No-op implementation
}

func (n *NoOpMetricsCollector) RecordHistogram(name string, value float64, labels map[string]string) {
	// No-op implementation
}

// Utility functions for creating gateway instances

// CreateProductionGateway creates a gateway configured for production use.
func CreateProductionGateway(cfg *config.Config, logger zerolog.Logger) (*Gateway, error) {
	routerConfig := core.RouterConfig{
		EnableCORS:        true,
		EnableRateLimit:   true,
		RateLimitRequests: 1000, // Higher limit for production
		EnableMetrics:     true,
		EnableLogging:     true,
		TrustedProxies:    []string{"127.0.0.1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"},
		MaxRequestSize:    10 << 20, // 10MB
	}

	return NewWithConfig(cfg, logger, routerConfig)
}

// CreateDevelopmentGateway creates a gateway configured for development use.
func CreateDevelopmentGateway(cfg *config.Config, logger zerolog.Logger) (*Gateway, error) {
	routerConfig := core.RouterConfig{
		EnableCORS:        true,
		EnableRateLimit:   false, // Disabled for development
		RateLimitRequests: 100,
		EnableMetrics:     true,
		EnableLogging:     true,
		TrustedProxies:    []string{"*"}, // Allow all for development
		MaxRequestSize:    50 << 20,      // 50MB for development
	}

	return NewWithConfig(cfg, logger, routerConfig)
}

// CreateTestGateway creates a gateway configured for testing.
func CreateTestGateway(cfg *config.Config, logger zerolog.Logger) (*Gateway, error) {
	routerConfig := core.RouterConfig{
		EnableCORS:        false,
		EnableRateLimit:   false,
		RateLimitRequests: 1000,
		EnableMetrics:     false,
		EnableLogging:     false,
		TrustedProxies:    []string{"127.0.0.1"},
		MaxRequestSize:    1 << 20, // 1MB for testing
	}

	return NewWithConfig(cfg, logger, routerConfig)
}

// Package information
const (
	// Version of the gateway package
	Version = "2.0.0"

	// Description of the package
	Description = "Modular, extensible API gateway for news aggregation"
)

// Adapter types to make gateway interfaces compatible with handler interfaces

type responseWriterAdapter struct {
	core.ResponseWriter
}

func (rw *responseWriterAdapter) BadRequest(c *gin.Context, message string) {
	rw.ResponseWriter.BadRequest(c, message)
}

func (rw *responseWriterAdapter) Unauthorized(c *gin.Context, message string) {
	rw.ResponseWriter.Unauthorized(c, message)
}

func (rw *responseWriterAdapter) Forbidden(c *gin.Context, message string) {
	rw.ResponseWriter.Forbidden(c, message)
}

func (rw *responseWriterAdapter) NotFound(c *gin.Context, message string) {
	rw.ResponseWriter.NotFound(c, message)
}

func (rw *responseWriterAdapter) InternalError(c *gin.Context, err error) {
	rw.ResponseWriter.InternalError(c, err)
}

func (rw *responseWriterAdapter) Success(c *gin.Context, data interface{}) {
	rw.ResponseWriter.Success(c, data)
}

func (rw *responseWriterAdapter) SuccessWithPagination(c *gin.Context, data interface{}, pagination handlerCore.PaginationInfo) {
	// Convert handler PaginationInfo to gateway PaginationInfo
	gatewayPagination := core.PaginationInfo{
		Page:    pagination.Page,
		Limit:   pagination.Limit,
		Total:   pagination.Total,
		Pages:   pagination.Pages,
		HasNext: pagination.HasNext,
		HasPrev: pagination.HasPrev,
	}
	rw.ResponseWriter.SuccessWithPagination(c, data, gatewayPagination)
}

func (rw *responseWriterAdapter) ErrorWithCode(c *gin.Context, code int, message string) {
	rw.ResponseWriter.ErrorWithCode(c, code, message)
}

type requestValidatorAdapter struct {
	core.RequestValidator
}

func (v *requestValidatorAdapter) ValidateSearchQuery(query interface{}) error {
	// Add implementation or delegate to existing method
	return nil // TODO: Implement
}

func (v *requestValidatorAdapter) ValidateUpdateProfileRequest(req interface{}) error {
	// Delegate to the gateway's validation implementation
	if validator, ok := v.RequestValidator.(*utils.RequestValidator); ok {
		return validator.ValidateUpdateProfileRequest(req.(*models.UpdateProfileRequest))
	}
	return nil
}

func (v *requestValidatorAdapter) ValidatePreferencesRequest(req interface{}) error {
	// Delegate to the gateway's validation implementation
	if validator, ok := v.RequestValidator.(*utils.RequestValidator); ok {
		return validator.ValidatePreferencesRequest(req.(*models.PreferencesRequest))
	}
	return nil
}

type contextManagerAdapter struct {
	core.ContextManager
}

func (cm *contextManagerAdapter) RequireAuth(c *gin.Context) error {
	_, err := cm.ContextManager.GetUserID(c)
	return err
}

func (cm *contextManagerAdapter) RequireAdmin(c *gin.Context) error {
	if err := cm.RequireAuth(c); err != nil {
		return err
	}
	return nil // TODO: Implement admin check
}
