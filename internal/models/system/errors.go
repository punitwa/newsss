package system

import "errors"

// System domain specific errors
var (
	// Validation errors
	ErrInvalidHealthStatus  = errors.New("invalid health status")
	ErrEmptyServiceName     = errors.New("service name cannot be empty")
	ErrInvalidServiceStatus = errors.New("invalid service status")
	ErrEmptyMessageType     = errors.New("message type cannot be empty")
	ErrInvalidMessageType   = errors.New("invalid message type")
	
	// Health check errors
	ErrHealthCheckFailed    = errors.New("health check failed")
	ErrServiceUnavailable   = errors.New("service is unavailable")
	ErrServiceTimeout       = errors.New("service health check timeout")
	ErrDependencyFailure    = errors.New("dependency health check failed")
	
	// Metrics errors
	ErrMetricsNotAvailable  = errors.New("metrics are not available")
	ErrMetricsCollectionFailed = errors.New("metrics collection failed")
	ErrInvalidMetricValue   = errors.New("invalid metric value")
	
	// WebSocket errors
	ErrWSConnectionFailed   = errors.New("WebSocket connection failed")
	ErrWSConnectionClosed   = errors.New("WebSocket connection closed")
	ErrWSMessageTooLarge    = errors.New("WebSocket message too large")
	ErrWSInvalidMessage     = errors.New("invalid WebSocket message")
	ErrWSUnauthorized       = errors.New("WebSocket connection unauthorized")
	ErrWSRateLimited        = errors.New("WebSocket rate limited")
	
	// System errors
	ErrSystemOverloaded     = errors.New("system is overloaded")
	ErrResourceExhausted    = errors.New("system resources exhausted")
	ErrMaintenanceMode      = errors.New("system is in maintenance mode")
	ErrConfigurationError   = errors.New("system configuration error")
)
