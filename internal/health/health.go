package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"news-aggregator/internal/config"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

// Status represents the health status
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// Check represents a health check
type Check struct {
	Name        string        `json:"name"`
	Status      Status        `json:"status"`
	Message     string        `json:"message,omitempty"`
	Duration    time.Duration `json:"duration"`
	LastChecked time.Time     `json:"last_checked"`
}

// HealthChecker performs health checks
type HealthChecker struct {
	config   *config.Config
	logger   zerolog.Logger
	checks   map[string]CheckFunc
	results  map[string]Check
	mu       sync.RWMutex
	interval time.Duration
}

// CheckFunc is a function that performs a health check
type CheckFunc func(ctx context.Context) Check

// NewHealthChecker creates a new health checker
func NewHealthChecker(cfg *config.Config, logger zerolog.Logger) *HealthChecker {
	hc := &HealthChecker{
		config:   cfg,
		logger:   logger.With().Str("component", "health_checker").Logger(),
		checks:   make(map[string]CheckFunc),
		results:  make(map[string]Check),
		interval: 30 * time.Second,
	}

	// Register default checks
	hc.RegisterCheck("database", hc.checkDatabase)
	hc.RegisterCheck("redis", hc.checkRedis)
	hc.RegisterCheck("elasticsearch", hc.checkElasticsearch)
	hc.RegisterCheck("rabbitmq", hc.checkRabbitMQ)

	return hc
}

// RegisterCheck registers a new health check
func (hc *HealthChecker) RegisterCheck(name string, checkFunc CheckFunc) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[name] = checkFunc
}

// Start starts the health checker
func (hc *HealthChecker) Start(ctx context.Context) {
	hc.logger.Info().Msg("Starting health checker")

	// Run initial checks
	hc.runChecks(ctx)

	// Start periodic checks
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.runChecks(ctx)
		case <-ctx.Done():
			hc.logger.Info().Msg("Health checker stopped")
			return
		}
	}
}

// runChecks runs all registered health checks
func (hc *HealthChecker) runChecks(ctx context.Context) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	for name, checkFunc := range hc.checks {
		go func(name string, checkFunc CheckFunc) {
			start := time.Now()
			result := checkFunc(ctx)
			result.Duration = time.Since(start)
			result.LastChecked = time.Now()

			hc.mu.Lock()
			hc.results[name] = result
			hc.mu.Unlock()

			hc.logger.Debug().
				Str("check", name).
				Str("status", string(result.Status)).
				Dur("duration", result.Duration).
				Msg("Health check completed")
		}(name, checkFunc)
	}
}

// GetHealth returns the overall health status
func (hc *HealthChecker) GetHealth() map[string]interface{} {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	overallStatus := StatusHealthy
	checks := make(map[string]Check)

	for name, result := range hc.results {
		checks[name] = result
		if result.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
		} else if result.Status == StatusDegraded && overallStatus == StatusHealthy {
			overallStatus = StatusDegraded
		}
	}

	return map[string]interface{}{
		"status":    overallStatus,
		"timestamp": time.Now(),
		"checks":    checks,
		"version":   "1.0.0",
	}
}

// GetReadiness returns readiness status (simplified health check)
func (hc *HealthChecker) GetReadiness() bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	for _, result := range hc.results {
		if result.Status == StatusUnhealthy {
			return false
		}
	}
	return true
}

// GetLiveness returns liveness status (basic service availability)
func (hc *HealthChecker) GetLiveness() bool {
	// Simple liveness check - service is alive if we can respond
	return true
}

// Health check implementations

func (hc *HealthChecker) checkDatabase(ctx context.Context) Check {
	check := Check{Name: "database"}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		hc.config.Database.Host,
		hc.config.Database.Port,
		hc.config.Database.User,
		hc.config.Database.Password,
		hc.config.Database.Database,
		hc.config.Database.SSLMode,
	)

	// Create a connection with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, connStr)
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Failed to connect: %v", err)
		return check
	}
	defer db.Close()

	// Test the connection
	if err := db.Ping(ctx); err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Ping failed: %v", err)
		return check
	}

	// Test a simple query
	var result int
	err = db.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Query failed: %v", err)
		return check
	}

	check.Status = StatusHealthy
	check.Message = "Database connection successful"
	return check
}

func (hc *HealthChecker) checkRedis(ctx context.Context) Check {
	check := Check{Name: "redis"}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     hc.config.Redis.Address,
		Password: hc.config.Redis.Password,
		DB:       hc.config.Redis.DB,
	})
	defer client.Close()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Test ping
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Redis ping failed: %v", err)
		return check
	}

	if pong != "PONG" {
		check.Status = StatusUnhealthy
		check.Message = "Redis ping returned unexpected response"
		return check
	}

	// Test set/get
	key := fmt.Sprintf("health_check_%d", time.Now().Unix())
	err = client.Set(ctx, key, "test", time.Minute).Err()
	if err != nil {
		check.Status = StatusDegraded
		check.Message = fmt.Sprintf("Redis set failed: %v", err)
		return check
	}

	_, err = client.Get(ctx, key).Result()
	if err != nil {
		check.Status = StatusDegraded
		check.Message = fmt.Sprintf("Redis get failed: %v", err)
		return check
	}

	// Clean up
	client.Del(ctx, key)

	check.Status = StatusHealthy
	check.Message = "Redis connection successful"
	return check
}

func (hc *HealthChecker) checkElasticsearch(ctx context.Context) Check {
	check := Check{Name: "elasticsearch"}

	// Create Elasticsearch client
	cfg := elasticsearch.Config{
		Addresses: hc.config.Elasticsearch.Addresses,
	}

	if hc.config.Elasticsearch.Username != "" {
		cfg.Username = hc.config.Elasticsearch.Username
		cfg.Password = hc.config.Elasticsearch.Password
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Failed to create client: %v", err)
		return check
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Test cluster health
	res, err := client.Cluster.Health(
		client.Cluster.Health.WithContext(ctx),
		client.Cluster.Health.WithWaitForStatus("yellow"),
		client.Cluster.Health.WithTimeout(time.Second*5),
	)
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Cluster health check failed: %v", err)
		return check
	}
	defer res.Body.Close()

	if res.IsError() {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Elasticsearch error: %s", res.String())
		return check
	}

	check.Status = StatusHealthy
	check.Message = "Elasticsearch connection successful"
	return check
}

func (hc *HealthChecker) checkRabbitMQ(ctx context.Context) Check {
	check := Check{Name: "rabbitmq"}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Connect to RabbitMQ
	conn, err := amqp.Dial(hc.config.RabbitMQ.URL)
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Failed to connect: %v", err)
		return check
	}
	defer conn.Close()

	// Create channel
	ch, err := conn.Channel()
	if err != nil {
		check.Status = StatusUnhealthy
		check.Message = fmt.Sprintf("Failed to create channel: %v", err)
		return check
	}
	defer ch.Close()

	// Test queue declaration (temporary queue)
	_, err = ch.QueueDeclare(
		"health_check", // name
		false,          // durable
		true,           // delete when unused
		true,           // exclusive
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		check.Status = StatusDegraded
		check.Message = fmt.Sprintf("Queue declaration failed: %v", err)
		return check
	}

	check.Status = StatusHealthy
	check.Message = "RabbitMQ connection successful"
	return check
}

// HTTPHandler returns an HTTP handler for health checks
func (hc *HealthChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := hc.GetHealth()
		
		w.Header().Set("Content-Type", "application/json")
		
		status := health["status"].(Status)
		switch status {
		case StatusHealthy:
			w.WriteHeader(http.StatusOK)
		case StatusDegraded:
			w.WriteHeader(http.StatusOK) // Still OK, but degraded
		case StatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		// Write JSON response
		if err := writeJSON(w, health); err != nil {
			hc.logger.Error().Err(err).Msg("Failed to write health check response")
		}
	}
}

// ReadinessHandler returns an HTTP handler for readiness checks
func (hc *HealthChecker) ReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if hc.GetReadiness() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ready"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"not ready"}`))
		}
	}
}

// LivenessHandler returns an HTTP handler for liveness checks
func (hc *HealthChecker) LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if hc.GetLiveness() {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"alive"}`))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"status":"dead"}`))
		}
	}
}

// Helper function to write JSON response
func writeJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}
