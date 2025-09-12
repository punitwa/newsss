// Package health provides health check HTTP handlers that are independent of any gateway.
package health

import (
	"runtime"
	"time"

	"news-aggregator/internal/handlers/core"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Handler implements health check operations independently.
type Handler struct {
	deps      *core.HandlerDependencies
	config    core.HandlerConfig
	logger    zerolog.Logger
	startTime time.Time
}

// NewHandler creates a new independent health handler.
func NewHandler(deps *core.HandlerDependencies, config core.HandlerConfig) core.HealthHandler {
	return &Handler{
		deps:      deps,
		config:    config,
		logger:    deps.Logger.With().Str("handler", "health").Logger(),
		startTime: time.Now(),
	}
}

// RegisterRoutes registers health check routes.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	health := router.Group(h.GetBasePath())
	{
		health.GET("", h.HealthCheck)
		health.GET("/ready", h.ReadinessCheck)
		health.GET("/live", h.LivenessCheck)
		health.GET("/status", h.GetDetailedStatus)
		health.GET("/version", h.GetVersion)
	}
}

// GetBasePath returns the base path for health routes.
func (h *Handler) GetBasePath() string {
	return "/health"
}

// GetName returns a unique name for this handler.
func (h *Handler) GetName() string {
	return "health_handler"
}

// HealthCheck performs basic health check.
func (h *Handler) HealthCheck(c *gin.Context) {
	status := gin.H{
		"status":    "healthy",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
	}
	
	h.deps.ResponseWriter.Success(c, status)
	
	if h.config.EnableLogging {
		h.logger.Debug().
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Health check performed")
	}
}

// ReadinessCheck performs readiness check.
func (h *Handler) ReadinessCheck(c *gin.Context) {
	status := gin.H{
		"status":    "healthy",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
		"services":  make(map[string]interface{}),
	}
	
	// Check database connectivity
	if err := h.checkDatabase(); err != nil {
		status["status"] = "unhealthy"
		status["services"].(map[string]interface{})["database"] = gin.H{
			"status":       "unhealthy",
			"error":        err.Error(),
			"last_checked": time.Now(),
		}
	} else {
		status["services"].(map[string]interface{})["database"] = gin.H{
			"status":       "healthy",
			"last_checked": time.Now(),
		}
	}
	
	// Check external services
	h.checkExternalServices(status)
	
	h.deps.ResponseWriter.Success(c, status)
	
	if h.config.EnableLogging {
		h.logger.Info().
			Str("status", status["status"].(string)).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Readiness check performed")
	}
}

// LivenessCheck performs liveness check.
func (h *Handler) LivenessCheck(c *gin.Context) {
	systemInfo := h.getSystemInfo()
	
	status := gin.H{
		"status":    "healthy",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
		"system":    systemInfo,
	}
	
	// Check if system resources are within acceptable limits
	if systemInfo["memory_usage"].(uint64) > 1024*1024*1024 { // 1GB
		status["status"] = "degraded"
	}
	
	if systemInfo["goroutine_count"].(int) > 10000 {
		status["status"] = "degraded"
	}
	
	h.deps.ResponseWriter.Success(c, status)
	
	if h.config.EnableLogging {
		h.logger.Debug().
			Str("status", status["status"].(string)).
			Uint64("memory_usage", systemInfo["memory_usage"].(uint64)).
			Int("goroutines", systemInfo["goroutine_count"].(int)).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Liveness check performed")
	}
}

// GetDetailedStatus returns detailed system status.
func (h *Handler) GetDetailedStatus(c *gin.Context) {
	status := gin.H{
		"status":    "healthy",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC(),
		"services":  make(map[string]interface{}),
		"system":    h.getSystemInfo(),
	}
	
	// Check all dependencies
	h.checkAllDependencies(status)
	
	h.deps.ResponseWriter.Success(c, status)
	
	if h.config.EnableLogging {
		servicesMap := status["services"].(map[string]interface{})
		h.logger.Info().
			Str("status", status["status"].(string)).
			Int("services_checked", len(servicesMap)).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Detailed status check performed")
	}
}

// GetVersion returns application version information.
func (h *Handler) GetVersion(c *gin.Context) {
	version := gin.H{
		"version":    "1.0.0",
		"build_time": "2024-01-01T00:00:00Z",
		"git_commit": "unknown",
		"go_version": runtime.Version(),
		"uptime":     time.Since(h.startTime).String(),
	}
	
	h.deps.ResponseWriter.Success(c, version)
}

// checkDatabase checks database connectivity.
func (h *Handler) checkDatabase() error {
	// This would typically ping the database
	// For now, we'll assume it's healthy
	return nil
}

// checkExternalServices checks external service dependencies.
func (h *Handler) checkExternalServices(status gin.H) {
	servicesMap := status["services"].(map[string]interface{})
	
	// Check news service
	if err := h.checkNewsService(); err != nil {
		servicesMap["news_service"] = gin.H{
			"status":       "unhealthy",
			"error":        err.Error(),
			"last_checked": time.Now(),
		}
		if status["status"].(string) == "healthy" {
			status["status"] = "degraded"
		}
	} else {
		servicesMap["news_service"] = gin.H{
			"status":       "healthy",
			"last_checked": time.Now(),
		}
	}
	
	// Check search service
	if err := h.checkSearchService(); err != nil {
		servicesMap["search_service"] = gin.H{
			"status":       "unhealthy",
			"error":        err.Error(),
			"last_checked": time.Now(),
		}
		if status["status"].(string) == "healthy" {
			status["status"] = "degraded"
		}
	} else {
		servicesMap["search_service"] = gin.H{
			"status":       "healthy",
			"last_checked": time.Now(),
		}
	}
}

// checkAllDependencies performs comprehensive dependency checks.
func (h *Handler) checkAllDependencies(status gin.H) {
	servicesMap := status["services"].(map[string]interface{})
	
	// Check database
	if err := h.checkDatabase(); err != nil {
		servicesMap["database"] = gin.H{
			"status":       "unhealthy",
			"error":        err.Error(),
			"last_checked": time.Now(),
		}
		status["status"] = "unhealthy"
	} else {
		servicesMap["database"] = gin.H{
			"status":       "healthy",
			"last_checked": time.Now(),
		}
	}
	
	// Check external services
	h.checkExternalServices(status)
	
	// Check cache (Redis, if applicable)
	if err := h.checkCache(); err != nil {
		servicesMap["cache"] = gin.H{
			"status":       "unhealthy",
			"error":        err.Error(),
			"last_checked": time.Now(),
		}
		// Cache failure is not critical, so only degrade
		if status["status"].(string) == "healthy" {
			status["status"] = "degraded"
		}
	} else {
		servicesMap["cache"] = gin.H{
			"status":       "healthy",
			"last_checked": time.Now(),
		}
	}
	
	// Check message queue (if applicable)
	if err := h.checkMessageQueue(); err != nil {
		servicesMap["message_queue"] = gin.H{
			"status":       "unhealthy",
			"error":        err.Error(),
			"last_checked": time.Now(),
		}
		if status["status"].(string) == "healthy" {
			status["status"] = "degraded"
		}
	} else {
		servicesMap["message_queue"] = gin.H{
			"status":       "healthy",
			"last_checked": time.Now(),
		}
	}
}

// checkNewsService checks news service health.
func (h *Handler) checkNewsService() error {
	// This would typically make a health check call to the news service
	// For now, we'll assume it's healthy
	return nil
}

// checkSearchService checks search service health.
func (h *Handler) checkSearchService() error {
	// This would typically make a health check call to the search service
	// For now, we'll assume it's healthy
	return nil
}

// checkCache checks cache service health.
func (h *Handler) checkCache() error {
	// This would typically ping Redis or other cache service
	// For now, we'll assume it's healthy
	return nil
}

// checkMessageQueue checks message queue health.
func (h *Handler) checkMessageQueue() error {
	// This would typically check RabbitMQ or other message queue
	// For now, we'll assume it's healthy
	return nil
}

// getSystemInfo returns current system information.
func (h *Handler) getSystemInfo() gin.H {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	return gin.H{
		"memory_usage":    memStats.Alloc,
		"goroutine_count": runtime.NumGoroutine(),
		"uptime":          time.Since(h.startTime),
	}
}
