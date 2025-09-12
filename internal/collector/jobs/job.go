package jobs

import (
	"context"
	"fmt"
	"time"

	"news-aggregator/internal/models"
	"news-aggregator/pkg/queue"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// JobPriority defines the priority levels for collection jobs
type JobPriority int

const (
	PriorityLow JobPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityUrgent
)

// JobStatus represents the current status of a job
type JobStatus int

const (
	JobStatusPending JobStatus = iota
	JobStatusProcessing
	JobStatusCompleted
	JobStatusFailed
	JobStatusRetrying
)

// CollectionJob represents a data collection job
type CollectionJob struct {
	ID       string
	Source   string
	Item     models.News
	Priority int
	Created  time.Time
	runCount int
}

// RunCount returns the number of times this job has been run
func (j *CollectionJob) RunCount() int {
	return j.runCount
}

// IncrementRunCount increments the run count
func (j *CollectionJob) IncrementRunCount() {
	j.runCount++
}

// JobResult contains the result of job processing
type JobResult struct {
	JobID     string
	Success   bool
	Error     error
	Duration  time.Duration
	Timestamp time.Time
}

// NewCollectionJob creates a new collection job with default values
func NewCollectionJob(source string, item models.News) *CollectionJob {
	return &CollectionJob{
		ID:       uuid.New().String(),
		Source:   source,
		Item:     item,
		Priority: int(PriorityNormal),
		Created:  time.Now(),
	}
}

// NewCollectionJobWithPriority creates a new collection job with specified priority
func NewCollectionJobWithPriority(source string, item models.News, priority JobPriority) *CollectionJob {
	return &CollectionJob{
		ID:       uuid.New().String(),
		Source:   source,
		Item:     item,
		Priority: int(priority),
		Created:  time.Now(),
	}
}

// Age returns how long ago the job was created
func (j *CollectionJob) Age() time.Duration {
	return time.Since(j.Created)
}

// IsExpired checks if the job has exceeded the maximum age
func (j *CollectionJob) IsExpired(maxAge time.Duration) bool {
	return j.Age() > maxAge
}

// Validate checks if the job is valid for processing
func (j *CollectionJob) Validate() error {
	if j.ID == "" {
		return fmt.Errorf("job ID cannot be empty")
	}
	if j.Source == "" {
		return fmt.Errorf("job source cannot be empty")
	}
	if j.Item.ID == "" {
		return fmt.Errorf("news item ID cannot be empty")
	}
	if j.Item.Title == "" {
		return fmt.Errorf("news item title cannot be empty")
	}
	return nil
}

// JobProcessor handles the processing of collection jobs
type JobProcessor struct {
	logger    zerolog.Logger
	queue     queue.Publisher
	timeout   time.Duration
	retryConf RetryConfig
}

// RetryConfig contains retry configuration
type RetryConfig struct {
	MaxAttempts int
	Delay       time.Duration
	Backoff     float64 // Multiplier for exponential backoff
}

// NewJobProcessor creates a new job processor
func NewJobProcessor(logger zerolog.Logger, publisher queue.Publisher, timeout time.Duration, retryConf RetryConfig) *JobProcessor {
	return &JobProcessor{
		logger:    logger,
		queue:     publisher,
		timeout:   timeout,
		retryConf: retryConf,
	}
}

// ProcessJob processes a single collection job
func (jp *JobProcessor) ProcessJob(ctx context.Context, job *CollectionJob) *JobResult {
	startTime := time.Now()

	result := &JobResult{
		JobID:     job.ID,
		Timestamp: startTime,
	}

	// Validate job
	if err := job.Validate(); err != nil {
		result.Success = false
		result.Error = fmt.Errorf("job validation failed: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, jp.timeout)
	defer cancel()

	// Process with retry logic
	var lastErr error
	for attempt := 0; attempt < jp.retryConf.MaxAttempts; attempt++ {
		if attempt > 0 {
			// Calculate backoff delay
			delay := time.Duration(float64(jp.retryConf.Delay) * jp.calculateBackoff(attempt))

			jp.logger.Debug().
				Str("job_id", job.ID).
				Int("attempt", attempt).
				Dur("delay", delay).
				Msg("Retrying job processing")

			select {
			case <-time.After(delay):
				// Continue with retry
			case <-timeoutCtx.Done():
				result.Success = false
				result.Error = fmt.Errorf("job timeout during retry: %w", timeoutCtx.Err())
				result.Duration = time.Since(startTime)
				return result
			}
		}

		// Attempt to process the job
		if err := jp.processJobAttempt(timeoutCtx, job); err != nil {
			lastErr = err
			jp.logger.Warn().
				Err(err).
				Str("job_id", job.ID).
				Int("attempt", attempt+1).
				Msg("Job processing attempt failed")
			continue
		}

		// Success
		result.Success = true
		result.Duration = time.Since(startTime)

		jp.logger.Debug().
			Str("job_id", job.ID).
			Str("source", job.Source).
			Dur("duration", result.Duration).
			Int("attempts", attempt+1).
			Msg("Job processed successfully")

		return result
	}

	// All attempts failed
	result.Success = false
	result.Error = fmt.Errorf("job failed after %d attempts: %w", jp.retryConf.MaxAttempts, lastErr)
	result.Duration = time.Since(startTime)

	jp.logger.Error().
		Err(result.Error).
		Str("job_id", job.ID).
		Str("source", job.Source).
		Msg("Job processing failed permanently")

	return result
}

// processJobAttempt performs a single attempt to process the job
func (jp *JobProcessor) processJobAttempt(ctx context.Context, job *CollectionJob) error {
	// Create message for processing pipeline
	message := models.NewsMessage{
		ID:        job.Item.ID,
		Source:    job.Source,
		Type:      "raw",
		Data:      job.Item,
		Timestamp: time.Now(),
		Retry:     0,
	}

	// Publish to message queue
	if err := jp.queue.Publish("news.raw", message); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// calculateBackoff calculates the backoff delay for retry attempts
func (jp *JobProcessor) calculateBackoff(attempt int) float64 {
	if jp.retryConf.Backoff <= 1.0 {
		return 1.0
	}

	backoff := 1.0
	for i := 0; i < attempt; i++ {
		backoff *= jp.retryConf.Backoff
	}

	// Cap the backoff to prevent excessive delays
	const maxBackoff = 60.0 // 60x the base delay
	if backoff > maxBackoff {
		return maxBackoff
	}

	return backoff
}
