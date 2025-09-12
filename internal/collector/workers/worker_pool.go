package workers

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"news-aggregator/internal/collector/jobs"
	"news-aggregator/internal/config"
	"news-aggregator/pkg/queue"

	"github.com/rs/zerolog"
)

// WorkerPool defines the interface for managing worker pools
type WorkerPool interface {
	Start(ctx context.Context)
	Stop()
	SubmitJob(job *jobs.CollectionJob) error
	GetStats() WorkerPoolStats
}

// WorkerPoolStats contains statistics about the worker pool
type WorkerPoolStats struct {
	ActiveWorkers int
	QueueSize     int
	TotalJobs     int64
	FailedJobs    int64
	AverageTime   time.Duration
}

// worker represents a single worker in the pool
type worker struct {
	id        int
	logger    zerolog.Logger
	processor *jobs.JobProcessor
	jobQueue  <-chan *jobs.CollectionJob
	busy      int32 // atomic flag for busy state
}

// newWorker creates a new worker
func newWorker(id int, logger zerolog.Logger, processor *jobs.JobProcessor, jobQueue <-chan *jobs.CollectionJob) *worker {
	return &worker{
		id:        id,
		logger:    logger.With().Int("worker_id", id).Logger(),
		processor: processor,
		jobQueue:  jobQueue,
	}
}

// start starts the worker processing loop
func (w *worker) start(ctx context.Context, onComplete func(*jobs.JobResult)) {
	w.logger.Debug().Msg("Worker started")
	defer w.logger.Debug().Msg("Worker stopped")

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-w.jobQueue:
			if !ok {
				return
			}

			// Mark worker as busy
			atomic.StoreInt32(&w.busy, 1)

			w.logger.Debug().Str("job_id", job.ID).Msg("Processing job")
			result := w.processor.ProcessJob(ctx, job)

			if onComplete != nil {
				onComplete(result)
			}

			// Mark worker as available
			atomic.StoreInt32(&w.busy, 0)
		}
	}
}

// isBusy returns true if the worker is currently processing a job
func (w *worker) isBusy() bool {
	return atomic.LoadInt32(&w.busy) == 1
}

// workerPool implements the WorkerPool interface
type workerPool struct {
	config    config.CollectorConfig
	logger    zerolog.Logger
	processor *jobs.JobProcessor
	jobQueue  chan *jobs.CollectionJob
	workers   []*worker
	wg        sync.WaitGroup
	mu        sync.RWMutex

	// Metrics
	totalJobs  int64
	failedJobs int64
	totalTime  int64 // nanoseconds

	// State
	running bool
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(config config.CollectorConfig, logger zerolog.Logger, publisher queue.Publisher) WorkerPool {
	retryConfig := jobs.RetryConfig{
		MaxAttempts: config.RetryAttempts,
		Delay:       config.RetryDelay,
		Backoff:     1.5, // 50% increase per retry
	}

	processor := jobs.NewJobProcessor(logger, publisher, config.JobTimeout, retryConfig)

	return &workerPool{
		config:    config,
		logger:    logger,
		processor: processor,
		jobQueue:  make(chan *jobs.CollectionJob, config.QueueSize),
	}
}

// Start starts the worker pool
func (wp *workerPool) Start(ctx context.Context) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if wp.running {
		wp.logger.Warn().Msg("Worker pool is already running")
		return
	}

	wp.ctx, wp.cancel = context.WithCancel(ctx)
	wp.running = true

	// Initialize workers
	wp.workers = make([]*worker, wp.config.WorkerCount)

	for i := 0; i < wp.config.WorkerCount; i++ {
		w := newWorker(i, wp.logger, wp.processor, wp.jobQueue)
		wp.workers[i] = w

		wp.wg.Add(1)
		go func(worker *worker) {
			defer wp.wg.Done()
			worker.start(wp.ctx, wp.onJobComplete)
		}(w)
	}

	wp.logger.Info().
		Int("worker_count", wp.config.WorkerCount).
		Int("queue_size", wp.config.QueueSize).
		Msg("Worker pool started")
}

// Stop stops the worker pool
func (wp *workerPool) Stop() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if !wp.running {
		wp.logger.Warn().Msg("Worker pool is not running")
		return
	}

	wp.logger.Info().Msg("Stopping worker pool")

	// Cancel context to signal workers to stop
	if wp.cancel != nil {
		wp.cancel()
	}

	// Close job queue to prevent new jobs
	close(wp.jobQueue)

	// Wait for all workers to finish
	wp.wg.Wait()

	wp.running = false
	wp.logger.Info().Msg("Worker pool stopped")
}

// SubmitJob submits a job to the worker pool
func (wp *workerPool) SubmitJob(job *jobs.CollectionJob) error {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	if !wp.running {
		return fmt.Errorf("worker pool is not running")
	}

	select {
	case wp.jobQueue <- job:
		wp.logger.Debug().
			Str("job_id", job.ID).
			Str("source", job.Source).
			Msg("Job submitted to queue")
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}

// GetStats returns statistics about the worker pool
func (wp *workerPool) GetStats() WorkerPoolStats {
	wp.mu.RLock()
	defer wp.mu.RUnlock()

	activeWorkers := 0
	for _, worker := range wp.workers {
		if worker.isBusy() {
			activeWorkers++
		}
	}

	totalJobs := atomic.LoadInt64(&wp.totalJobs)
	totalTime := atomic.LoadInt64(&wp.totalTime)

	var avgTime time.Duration
	if totalJobs > 0 {
		avgTime = time.Duration(totalTime / totalJobs)
	}

	return WorkerPoolStats{
		ActiveWorkers: activeWorkers,
		QueueSize:     len(wp.jobQueue),
		TotalJobs:     totalJobs,
		FailedJobs:    atomic.LoadInt64(&wp.failedJobs),
		AverageTime:   avgTime,
	}
}

// onJobComplete handles job completion
func (wp *workerPool) onJobComplete(result *jobs.JobResult) {
	// Update metrics
	atomic.AddInt64(&wp.totalJobs, 1)
	if !result.Success {
		atomic.AddInt64(&wp.failedJobs, 1)
	}
	atomic.AddInt64(&wp.totalTime, result.Duration.Nanoseconds())

	// Log result
	if result.Success {
		wp.logger.Debug().
			Str("job_id", result.JobID).
			Dur("duration", result.Duration).
			Msg("Job completed successfully")
	} else {
		wp.logger.Error().
			Err(result.Error).
			Str("job_id", result.JobID).
			Dur("duration", result.Duration).
			Msg("Job failed")
	}
}
