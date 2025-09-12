package scheduling

import (
	"fmt"
	"sync"
	"time"

	"news-aggregator/internal/datasources"

	"github.com/go-co-op/gocron"
	"github.com/rs/zerolog"
)

// JobScheduler interface defines the contract for job scheduling
type JobScheduler interface {
	Start()
	Stop()
	ScheduleSource(sourceName string, source datasources.DataSource, handler func()) error
	RemoveSource(sourceName string) error
	GetAllScheduleInfo() map[string]ScheduleInfo
	GetSourceScheduleInfo(sourceName string) (ScheduleInfo, error)
	IsRunning() bool
	GetSchedulerStats() SchedulerStats
	RescheduleSource(sourceName string, source datasources.DataSource, handler func()) error
}

// jobScheduler implements the JobScheduler interface
type jobScheduler struct {
	logger    zerolog.Logger
	scheduler *gocron.Scheduler
	jobs      map[string]*gocron.Job
	mu        sync.RWMutex
	running   bool
}

// NewJobScheduler creates a new job scheduler
func NewJobScheduler(logger zerolog.Logger) JobScheduler {
	return &jobScheduler{
		logger:    logger,
		scheduler: gocron.NewScheduler(time.UTC),
		jobs:      make(map[string]*gocron.Job),
	}
}

// Start starts the scheduler
func (js *jobScheduler) Start() {
	js.mu.Lock()
	defer js.mu.Unlock()

	if js.running {
		js.logger.Warn().Msg("Scheduler is already running")
		return
	}

	js.scheduler.StartAsync()
	js.running = true

	js.logger.Info().Msg("Job scheduler started")
}

// Stop stops the scheduler
func (js *jobScheduler) Stop() {
	js.mu.Lock()
	defer js.mu.Unlock()

	if !js.running {
		js.logger.Warn().Msg("Scheduler is not running")
		return
	}

	js.scheduler.Stop()
	js.running = false

	// Clear all jobs
	js.jobs = make(map[string]*gocron.Job)

	js.logger.Info().Msg("Job scheduler stopped")
}

// ScheduleSource schedules a data source for regular collection
func (js *jobScheduler) ScheduleSource(sourceName string, source datasources.DataSource, handler func()) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	if !js.running {
		return fmt.Errorf("scheduler is not running")
	}

	// Remove existing job if it exists
	if existingJob, exists := js.jobs[sourceName]; exists {
		js.scheduler.RemoveByReference(existingJob)
		delete(js.jobs, sourceName)
		js.logger.Debug().Str("source", sourceName).Msg("Removed existing scheduled job")
	}

	// Get schedule from source
	schedule := source.GetSchedule()

	// Validate schedule
	if err := js.validateSchedule(schedule); err != nil {
		return fmt.Errorf("invalid schedule for source %s: %w", sourceName, err)
	}

	// Create job with error handling wrapper
	wrappedHandler := js.createJobHandler(sourceName, handler)

	// Schedule the job
	job, err := js.scheduler.Every(schedule).Do(wrappedHandler)
	if err != nil {
		return fmt.Errorf("failed to schedule source %s: %w", sourceName, err)
	}

	// Store job reference
	js.jobs[sourceName] = job

	js.logger.Info().
		Str("source", sourceName).
		Str("schedule", schedule.String()).
		Time("next_run", job.NextRun()).
		Msg("Source scheduled successfully")

	return nil
}

// RemoveSource removes a scheduled source
func (js *jobScheduler) RemoveSource(sourceName string) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	job, exists := js.jobs[sourceName]
	if !exists {
		return fmt.Errorf("no scheduled job found for source: %s", sourceName)
	}

	// Remove job from scheduler
	js.scheduler.RemoveByReference(job)

	// Remove from our tracking
	delete(js.jobs, sourceName)

	js.logger.Info().Str("source", sourceName).Msg("Scheduled source removed")
	return nil
}

// validateSchedule validates that a schedule duration is reasonable
func (js *jobScheduler) validateSchedule(schedule time.Duration) error {
	// Minimum schedule interval (prevent too frequent polling)
	minInterval := 30 * time.Second
	if schedule < minInterval {
		return fmt.Errorf("schedule interval too short: %v (minimum: %v)", schedule, minInterval)
	}

	// Maximum schedule interval (prevent schedules that are too long)
	maxInterval := 24 * time.Hour
	if schedule > maxInterval {
		return fmt.Errorf("schedule interval too long: %v (maximum: %v)", schedule, maxInterval)
	}

	return nil
}

// createJobHandler creates a wrapped job handler with error handling and logging
func (js *jobScheduler) createJobHandler(sourceName string, handler func()) func() {
	return func() {
		startTime := time.Now()

		js.logger.Debug().
			Str("source", sourceName).
			Time("start_time", startTime).
			Msg("Starting scheduled collection")

		// Execute the handler with panic recovery
		func() {
			defer func() {
				if r := recover(); r != nil {
					js.logger.Error().
						Str("source", sourceName).
						Interface("panic", r).
						Msg("Panic occurred during scheduled collection")
				}
			}()

			handler()
		}()

		duration := time.Since(startTime)
		js.logger.Debug().
			Str("source", sourceName).
			Dur("duration", duration).
			Msg("Scheduled collection completed")
	}
}

// GetScheduledSources returns information about all scheduled sources
func (js *jobScheduler) GetAllScheduleInfo() map[string]ScheduleInfo {
	js.mu.RLock()
	defer js.mu.RUnlock()

	info := make(map[string]ScheduleInfo)

	for sourceName, job := range js.jobs {
		info[sourceName] = ScheduleInfo{
			SourceName:  sourceName,
			NextRun:     job.NextRun(),
			LastRun:     job.LastRun(),
			RunCount:    uint64(job.RunCount()),
			IsScheduled: true,
		}
	}

	return info
}

// GetSourceScheduleInfo returns schedule information for a specific source
func (js *jobScheduler) GetSourceScheduleInfo(sourceName string) (ScheduleInfo, error) {
	js.mu.RLock()
	defer js.mu.RUnlock()

	job, exists := js.jobs[sourceName]
	if !exists {
		return ScheduleInfo{}, fmt.Errorf("no scheduled job found for source: %s", sourceName)
	}

	return ScheduleInfo{
		SourceName:  sourceName,
		NextRun:     job.NextRun(),
		LastRun:     job.LastRun(),
		RunCount:    uint64(job.RunCount()),
		IsScheduled: true,
	}, nil
}

// IsRunning returns true if the scheduler is currently running
func (js *jobScheduler) IsRunning() bool {
	js.mu.RLock()
	defer js.mu.RUnlock()
	return js.running
}

// GetJobCount returns the number of scheduled jobs
func (js *jobScheduler) GetJobCount() int {
	js.mu.RLock()
	defer js.mu.RUnlock()
	return len(js.jobs)
}

// ScheduleInfo contains information about a scheduled source
type ScheduleInfo struct {
	SourceName  string    `json:"source_name"`
	NextRun     time.Time `json:"next_run"`
	LastRun     time.Time `json:"last_run"`
	RunCount    uint64    `json:"run_count"`
	IsScheduled bool      `json:"is_scheduled"`
}

// SchedulerStats contains statistics about the scheduler
type SchedulerStats struct {
	IsRunning     bool   `json:"is_running"`
	ScheduledJobs int    `json:"scheduled_jobs"`
	TotalRuns     uint64 `json:"total_runs"`
}

// RescheduleSource reschedules an existing source with a new schedule
func (js *jobScheduler) RescheduleSource(sourceName string, source datasources.DataSource, handler func()) error {
	// Remove existing schedule
	if err := js.RemoveSource(sourceName); err != nil {
		// Log the error but continue with scheduling
		js.logger.Warn().
			Err(err).
			Str("source", sourceName).
			Msg("Failed to remove existing schedule during reschedule")
	}

	// Add new schedule
	return js.ScheduleSource(sourceName, source, handler)
}

// PauseSource temporarily pauses a scheduled source
func (js *jobScheduler) PauseSource(sourceName string) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	job, exists := js.jobs[sourceName]
	if !exists {
		return fmt.Errorf("no scheduled job found for source: %s", sourceName)
	}

	// Note: gocron doesn't have a built-in pause functionality
	// We'll remove the job and keep track that it was paused
	js.scheduler.RemoveByReference(job)

	js.logger.Info().Str("source", sourceName).Msg("Source schedule paused")
	return nil
}

// GetSchedulerStats returns statistics about the scheduler
func (js *jobScheduler) GetSchedulerStats() SchedulerStats {
	js.mu.RLock()
	defer js.mu.RUnlock()

	totalRuns := uint64(0)
	for _, job := range js.jobs {
		totalRuns += uint64(job.RunCount())
	}

	return SchedulerStats{
		IsRunning:     js.running,
		ScheduledJobs: len(js.jobs),
		TotalRuns:     totalRuns,
	}
}

// getNextJobTime returns the time of the next scheduled job
func (js *jobScheduler) getNextJobTime() time.Time {
	var nextTime time.Time
	first := true

	for _, job := range js.jobs {
		nextRun := job.NextRun()
		if first || nextRun.Before(nextTime) {
			nextTime = nextRun
			first = false
		}
	}

	return nextTime
}
