package core

import (
	"context"
	"time"

	"news-aggregator/internal/collector/jobs"
	"news-aggregator/internal/config"
	"news-aggregator/internal/datasources"

	"github.com/rs/zerolog"
)

// Collector defines the main collector interface
type Collector interface {
	Start(ctx context.Context) error
	Stop()
	AddSource(sourceConfig config.SourceConfig) error
	RemoveSource(sourceName string) error
	GetSourceStatus() map[string]interface{}
}

// WorkerPool defines the interface for managing worker pools
type WorkerPool interface {
	Start(ctx context.Context)
	Stop()
	SubmitJob(job *jobs.CollectionJob) error
	GetStats() WorkerPoolStats
}

// SourceManager defines the interface for managing data sources
type SourceManager interface {
	Initialize(sourceConfigs []config.SourceConfig) error
	AddSource(sourceConfig config.SourceConfig) error
	RemoveSource(sourceName string) error
	GetSource(sourceName string) (datasources.DataSource, bool)
	GetAllSources() map[string]datasources.DataSource
	GetStatus() map[string]interface{}
}

// JobScheduler defines the interface for scheduling collection jobs
type JobScheduler interface {
	Start()
	Stop()
	ScheduleSource(sourceName string, source datasources.DataSource, handler func()) error
	RemoveSource(sourceName string) error
}

// CollectionJob is re-exported from jobs package for compatibility
type CollectionJob = jobs.CollectionJob

// WorkerPoolStats contains statistics about the worker pool
type WorkerPoolStats struct {
	ActiveWorkers int
	QueueSize     int
	TotalJobs     int64
	FailedJobs    int64
	AverageTime   time.Duration
}

// CollectorConfig is now imported from config package
// This type alias maintains compatibility
type CollectorConfig = config.CollectorConfig

// DefaultCollectorConfig returns default configuration values
func DefaultCollectorConfig() CollectorConfig {
	return CollectorConfig{
		WorkerCount:    10,
		QueueSize:      1000,
		JobTimeout:     30 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     5 * time.Second,
		MetricsEnabled: true,
	}
}

// CollectorMetrics contains metrics for monitoring
type CollectorMetrics struct {
	TotalJobs          int64
	SuccessfulJobs     int64
	FailedJobs         int64
	AverageJobTime     time.Duration
	ActiveSources      int
	QueueUtilization   float64
	LastCollectionTime time.Time
}

// Logger interface for dependency injection
type Logger interface {
	Debug() *zerolog.Event
	Info() *zerolog.Event
	Warn() *zerolog.Event
	Error() *zerolog.Event
	With() zerolog.Context
}
