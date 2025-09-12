package core

import (
	"context"
	"fmt"
	"time"

	"news-aggregator/internal/collector/jobs"
	"news-aggregator/internal/collector/scheduling"
	"news-aggregator/internal/collector/sources"
	"news-aggregator/internal/collector/workers"
	"news-aggregator/internal/config"
	"news-aggregator/internal/datasources"
	"news-aggregator/pkg/queue"

	"github.com/rs/zerolog"
)

// collector implements the Collector interface
type collector struct {
	config        *config.Config
	logger        zerolog.Logger
	collectorConf config.CollectorConfig

	// Components
	sourceManager sources.SourceManager
	workerPool    workers.WorkerPool
	scheduler     scheduling.JobScheduler

	// State
	running bool
}

// New creates a new collector instance
func New(cfg *config.Config, logger zerolog.Logger) (Collector, error) {
	return NewWithConfig(cfg, logger, cfg.Collector)
}

// NewWithConfig creates a new collector instance with custom configuration
func NewWithConfig(cfg *config.Config, logger zerolog.Logger, collectorConf config.CollectorConfig) (Collector, error) {
	// Initialize message queue
	publisher, err := queue.NewRabbitMQPublisher(cfg.RabbitMQ.URL, cfg.RabbitMQ.Exchange)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue publisher: %w", err)
	}

	// Initialize components
	sourceManager := sources.NewSourceManager(logger)
	workerPool := workers.NewWorkerPool(collectorConf, logger, publisher)
	scheduler := scheduling.NewJobScheduler(logger)

	collector := &collector{
		config:        cfg,
		logger:        logger,
		collectorConf: collectorConf,
		sourceManager: sourceManager,
		workerPool:    workerPool,
		scheduler:     scheduler,
	}

	// Initialize data sources
	if err := collector.sourceManager.Initialize(cfg.Sources); err != nil {
		logger.Warn().Err(err).Msg("Some sources failed to initialize")
		// Continue despite source initialization errors
	}

	return collector, nil
}

// Start starts the collector service
func (c *collector) Start(ctx context.Context) error {
	if c.running {
		return fmt.Errorf("collector is already running")
	}

	c.logger.Info().Msg("Starting collector service")

	// Start components
	c.workerPool.Start(ctx)
	c.scheduler.Start()

	// Schedule all sources
	if err := c.scheduleAllSources(ctx); err != nil {
		c.logger.Error().Err(err).Msg("Failed to schedule sources")
		return fmt.Errorf("failed to schedule sources: %w", err)
	}

	c.running = true
	c.logger.Info().Msg("Collector service started successfully")

	// Wait for context cancellation
	<-ctx.Done()
	c.logger.Info().Msg("Collector service context cancelled")

	return nil
}

// scheduleAllSources schedules all sources for data collection
func (c *collector) scheduleAllSources(ctx context.Context) error {
	sources := c.sourceManager.GetAllSources()

	for name, source := range sources {
		handler := c.createCollectionHandler(ctx, name, source)

		if err := c.scheduler.ScheduleSource(name, source, handler); err != nil {
			c.logger.Error().
				Err(err).
				Str("source", name).
				Msg("Failed to schedule source")
			// Continue with other sources
		}
	}

	return nil
}

// Stop stops the collector service
func (c *collector) Stop() {
	if !c.running {
		c.logger.Warn().Msg("Collector is not running")
		return
	}

	c.logger.Info().Msg("Stopping collector service")

	// Stop components in reverse order
	c.scheduler.Stop()
	c.workerPool.Stop()

	c.running = false
	c.logger.Info().Msg("Collector service stopped")
}

// createCollectionHandler creates a handler function for source collection
func (c *collector) createCollectionHandler(ctx context.Context, sourceName string, source datasources.DataSource) func() {
	return func() {
		c.collectFromSource(ctx, sourceName, source)
	}
}

// collectFromSource performs data collection from a specific source
func (c *collector) collectFromSource(ctx context.Context, sourceName string, source datasources.DataSource) {
	c.logger.Debug().Str("source", sourceName).Msg("Starting collection from source")

	startTime := time.Now()

	// Fetch data from source
	items, err := source.Fetch(ctx)
	if err != nil {
		c.logger.Error().Err(err).Str("source", sourceName).Msg("Failed to fetch from source")
		return
	}

	if len(items) == 0 {
		c.logger.Debug().Str("source", sourceName).Msg("No new items from source")
		return
	}

	// Submit items to worker pool for processing
	successCount := 0
	for _, item := range items {
		job := jobs.NewCollectionJob(sourceName, item)

		if err := c.workerPool.SubmitJob(job); err != nil {
			c.logger.Warn().
				Err(err).
				Str("source", sourceName).
				Str("job_id", job.ID).
				Msg("Failed to submit job to worker pool")
			continue
		}
		successCount++
	}

	duration := time.Since(startTime)
	c.logger.Info().
		Str("source", sourceName).
		Int("total_items", len(items)).
		Int("submitted_jobs", successCount).
		Dur("duration", duration).
		Msg("Collection completed")
}

// AddSource adds a new source to the collector
func (c *collector) AddSource(sourceConfig config.SourceConfig) error {
	c.logger.Info().Str("source", sourceConfig.Name).Msg("Adding new source")

	// Add source to manager
	if err := c.sourceManager.AddSource(sourceConfig); err != nil {
		return fmt.Errorf("failed to add source: %w", err)
	}

	// If collector is running, schedule the new source
	if c.running {
		source, exists := c.sourceManager.GetSource(sourceConfig.Name)
		if !exists {
			return fmt.Errorf("source not found after adding: %s", sourceConfig.Name)
		}

		handler := c.createCollectionHandler(context.Background(), sourceConfig.Name, source)
		if err := c.scheduler.ScheduleSource(sourceConfig.Name, source, handler); err != nil {
			// Remove source if scheduling fails
			c.sourceManager.RemoveSource(sourceConfig.Name)
			return fmt.Errorf("failed to schedule source: %w", err)
		}
	}

	c.logger.Info().Str("source", sourceConfig.Name).Msg("Source added successfully")
	return nil
}

// RemoveSource removes a source from the collector
func (c *collector) RemoveSource(sourceName string) error {
	c.logger.Info().Str("source", sourceName).Msg("Removing source")

	// Remove from scheduler if running
	if c.running {
		if err := c.scheduler.RemoveSource(sourceName); err != nil {
			c.logger.Warn().Err(err).Str("source", sourceName).Msg("Failed to remove source from scheduler")
		}
	}

	// Remove from source manager
	if err := c.sourceManager.RemoveSource(sourceName); err != nil {
		return fmt.Errorf("failed to remove source: %w", err)
	}

	c.logger.Info().Str("source", sourceName).Msg("Source removed successfully")
	return nil
}

// GetSourceStatus returns the status of all sources
func (c *collector) GetSourceStatus() map[string]interface{} {
	status := c.sourceManager.GetStatus()

	// Add scheduler information if available
	if c.running {
		scheduledSources := c.scheduler.GetAllScheduleInfo()
		for sourceName, sourceStatus := range status {
			if statusMap, ok := sourceStatus.(map[string]interface{}); ok {
				if scheduleInfo, exists := scheduledSources[sourceName]; exists {
					statusMap["next_run"] = scheduleInfo.NextRun
					statusMap["last_run"] = scheduleInfo.LastRun
					statusMap["run_count"] = scheduleInfo.RunCount
				}
			}
		}
	}

	return status
}

// GetMetrics returns collector metrics
func (c *collector) GetMetrics() CollectorMetrics {
	workerStats := c.workerPool.GetStats()

	return CollectorMetrics{
		TotalJobs:          workerStats.TotalJobs,
		SuccessfulJobs:     workerStats.TotalJobs - workerStats.FailedJobs,
		FailedJobs:         workerStats.FailedJobs,
		AverageJobTime:     workerStats.AverageTime,
		ActiveSources:      c.sourceManager.GetSourceCount(),
		QueueUtilization:   float64(workerStats.QueueSize) / float64(c.collectorConf.QueueSize),
		LastCollectionTime: time.Now(), // This could be tracked more precisely
	}
}

// IsRunning returns true if the collector is currently running
func (c *collector) IsRunning() bool {
	return c.running
}

// GetConfig returns the current collector configuration
func (c *collector) GetConfig() config.CollectorConfig {
	return c.collectorConf
}
