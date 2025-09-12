package system

import (
	"time"
)

// HealthCheck represents system health status
type HealthCheck struct {
	Status     string            `json:"status"`     // healthy, degraded, unhealthy
	Timestamp  time.Time         `json:"timestamp"`
	Version    string            `json:"version"`
	Services   map[string]string `json:"services"`
	Uptime     time.Duration     `json:"uptime"`
	Environment string           `json:"environment"`
	BuildInfo  BuildInfo         `json:"build_info"`
}

// BuildInfo contains build information
type BuildInfo struct {
	Version   string    `json:"version"`
	Commit    string    `json:"commit"`
	Branch    string    `json:"branch"`
	BuildTime time.Time `json:"build_time"`
	GoVersion string    `json:"go_version"`
}

// ServiceHealth represents the health of an individual service
type ServiceHealth struct {
	Name         string        `json:"name"`
	Status       string        `json:"status"`       // healthy, degraded, unhealthy
	ResponseTime time.Duration `json:"response_time"`
	LastChecked  time.Time     `json:"last_checked"`
	Error        string        `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Metrics represents system metrics
type Metrics struct {
	RequestsTotal     int64   `json:"requests_total"`
	RequestDuration   float64 `json:"request_duration_avg"`
	ErrorRate         float64 `json:"error_rate"`
	ActiveConnections int64   `json:"active_connections"`
	MemoryUsage       int64   `json:"memory_usage"`
	CPUUsage          float64 `json:"cpu_usage"`
	DiskUsage         float64 `json:"disk_usage"`
	NetworkIO         NetworkIO `json:"network_io"`
	Timestamp         time.Time `json:"timestamp"`
}

// NetworkIO represents network I/O metrics
type NetworkIO struct {
	BytesReceived int64 `json:"bytes_received"`
	BytesSent     int64 `json:"bytes_sent"`
	PacketsReceived int64 `json:"packets_received"`
	PacketsSent   int64 `json:"packets_sent"`
}

// DatabaseMetrics represents database-specific metrics
type DatabaseMetrics struct {
	ConnectionsActive int64         `json:"connections_active"`
	ConnectionsIdle   int64         `json:"connections_idle"`
	ConnectionsTotal  int64         `json:"connections_total"`
	QueriesPerSecond  float64       `json:"queries_per_second"`
	AvgQueryTime      time.Duration `json:"avg_query_time"`
	SlowQueries       int64         `json:"slow_queries"`
	CacheHitRatio     float64       `json:"cache_hit_ratio"`
	TableSize         int64         `json:"table_size"`
	IndexSize         int64         `json:"index_size"`
	Timestamp         time.Time     `json:"timestamp"`
}

// QueueMetrics represents message queue metrics
type QueueMetrics struct {
	QueueName         string        `json:"queue_name"`
	MessageCount      int64         `json:"message_count"`
	ConsumerCount     int           `json:"consumer_count"`
	PublishRate       float64       `json:"publish_rate"`
	ConsumeRate       float64       `json:"consume_rate"`
	AvgProcessingTime time.Duration `json:"avg_processing_time"`
	DeadLetterCount   int64         `json:"dead_letter_count"`
	Timestamp         time.Time     `json:"timestamp"`
}

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	ID        string      `json:"id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
}

// WSNewsUpdate represents a WebSocket news update
type WSNewsUpdate struct {
	Action string      `json:"action"` // new, update, delete
	News   interface{} `json:"news"`   // Using interface{} to avoid circular dependency
}

// WSConnectionInfo represents WebSocket connection information
type WSConnectionInfo struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id,omitempty"`
	ConnectedAt   time.Time `json:"connected_at"`
	LastActivity  time.Time `json:"last_activity"`
	MessagesSent  int64     `json:"messages_sent"`
	MessagesReceived int64  `json:"messages_received"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
}

// SystemStats represents overall system statistics
type SystemStats struct {
	TotalUsers        int64     `json:"total_users"`
	ActiveUsers       int64     `json:"active_users"`
	TotalArticles     int64     `json:"total_articles"`
	TotalSources      int64     `json:"total_sources"`
	ActiveSources     int64     `json:"active_sources"`
	WSConnections     int64     `json:"websocket_connections"`
	RequestsPerMinute float64   `json:"requests_per_minute"`
	ErrorsPerMinute   float64   `json:"errors_per_minute"`
	Timestamp         time.Time `json:"timestamp"`
}

// Validation methods

// Validate validates the HealthCheck struct
func (h *HealthCheck) Validate() error {
	validStatuses := map[string]bool{
		"healthy":   true,
		"degraded":  true,
		"unhealthy": true,
	}
	
	if !validStatuses[h.Status] {
		return ErrInvalidHealthStatus
	}
	
	return nil
}

// Validate validates the ServiceHealth struct
func (s *ServiceHealth) Validate() error {
	if s.Name == "" {
		return ErrEmptyServiceName
	}
	
	validStatuses := map[string]bool{
		"healthy":   true,
		"degraded":  true,
		"unhealthy": true,
	}
	
	if !validStatuses[s.Status] {
		return ErrInvalidServiceStatus
	}
	
	return nil
}

// Validate validates the WSMessage struct
func (m *WSMessage) Validate() error {
	if m.Type == "" {
		return ErrEmptyMessageType
	}
	
	validTypes := map[string]bool{
		"news_update":     true,
		"user_activity":   true,
		"system_alert":    true,
		"notification":    true,
		"heartbeat":       true,
	}
	
	if !validTypes[m.Type] {
		return ErrInvalidMessageType
	}
	
	return nil
}

// Helper methods

// IsHealthy returns true if the system is healthy
func (h *HealthCheck) IsHealthy() bool {
	return h.Status == "healthy"
}

// AddService adds a service health status
func (h *HealthCheck) AddService(name, status string) {
	if h.Services == nil {
		h.Services = make(map[string]string)
	}
	h.Services[name] = status
}

// SetOverallStatus sets the overall health status based on service statuses
func (h *HealthCheck) SetOverallStatus() {
	if h.Services == nil || len(h.Services) == 0 {
		h.Status = "unknown"
		return
	}
	
	healthyCount := 0
	unhealthyCount := 0
	
	for _, status := range h.Services {
		switch status {
		case "healthy":
			healthyCount++
		case "unhealthy":
			unhealthyCount++
		}
	}
	
	totalServices := len(h.Services)
	
	if unhealthyCount == 0 {
		h.Status = "healthy"
	} else if unhealthyCount < totalServices/2 {
		h.Status = "degraded"
	} else {
		h.Status = "unhealthy"
	}
}

// IsHealthy returns true if the service is healthy
func (s *ServiceHealth) IsHealthy() bool {
	return s.Status == "healthy"
}

// CalculateErrorRate calculates the error rate for metrics
func (m *Metrics) CalculateErrorRate(totalRequests, errorRequests int64) {
	if totalRequests > 0 {
		m.ErrorRate = float64(errorRequests) / float64(totalRequests) * 100
	}
}

// NewWSMessage creates a new WebSocket message
func NewWSMessage(msgType string, data interface{}) *WSMessage {
	return &WSMessage{
		Type:      msgType,
		Data:      data,
		Timestamp: time.Now(),
		ID:        generateMessageID(),
	}
}

// NewHealthCheck creates a new health check
func NewHealthCheck(version string) *HealthCheck {
	return &HealthCheck{
		Status:    "healthy",
		Timestamp: time.Now(),
		Version:   version,
		Services:  make(map[string]string),
		Uptime:    0,
	}
}

// NewServiceHealth creates a new service health status
func NewServiceHealth(name string) *ServiceHealth {
	return &ServiceHealth{
		Name:        name,
		Status:      "healthy",
		LastChecked: time.Now(),
		Metadata:    make(map[string]interface{}),
	}
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return "ws_" + time.Now().Format("20060102150405") + "_" + randString(6)
}

// randString generates a random string
func randString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
