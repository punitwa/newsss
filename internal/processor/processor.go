package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"news-aggregator/internal/config"
	"news-aggregator/internal/models"
	"news-aggregator/internal/services"
	"news-aggregator/pkg/queue"

	"github.com/rs/zerolog"
)

type Processor struct {
	config          *config.Config
	logger          zerolog.Logger
	consumer        queue.Consumer
	publisher       queue.Publisher
	newsService     *services.NewsService
	searchService   *services.SearchService
	transformers    []Transformer
	deduplicator    *Deduplicator
	workerPool      *ProcessorWorkerPool
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
}

func New(cfg *config.Config, logger zerolog.Logger) (*Processor, error) {
	// Initialize message queue consumer
	consumer, err := queue.NewRabbitMQConsumer(cfg.RabbitMQ.URL, cfg.RabbitMQ.Exchange, cfg.RabbitMQ.PrefetchCount)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue consumer: %w", err)
	}

	// Initialize message queue publisher
	publisher, err := queue.NewRabbitMQPublisher(cfg.RabbitMQ.URL, cfg.RabbitMQ.Exchange)
	if err != nil {
		return nil, fmt.Errorf("failed to create queue publisher: %w", err)
	}

	// Initialize services
	newsService, err := services.NewNewsService(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create news service: %w", err)
	}

	searchService, err := services.NewSearchService(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create search service: %w", err)
	}

	// Initialize transformers
	transformers := []Transformer{
		NewContentCleanerTransformer(logger),
		NewCategoryClassifierTransformer(logger),
		NewSentimentAnalyzerTransformer(logger),
		NewImageExtractorTransformer(logger),
	}

	// Initialize deduplicator
	deduplicator := NewDeduplicator(newsService, logger)

	// Initialize worker pool
	workerPool := NewProcessorWorkerPool(cfg, logger)

	return &Processor{
		config:        cfg,
		logger:        logger,
		consumer:      consumer,
		publisher:     publisher,
		newsService:   newsService,
		searchService: searchService,
		transformers:  transformers,
		deduplicator:  deduplicator,
		workerPool:    workerPool,
	}, nil
}

func (p *Processor) Start(ctx context.Context) error {
	p.logger.Info().Msg("Starting processor service")

	p.ctx, p.cancel = context.WithCancel(ctx)

	// Start worker pool
	p.workerPool.Start(p.ctx)

	// Start consuming messages
	err := p.consumer.Consume("news.raw", p.handleMessage)
	if err != nil {
		return fmt.Errorf("failed to start consuming messages: %w", err)
	}

	p.logger.Info().Msg("Processor service started")

	// Wait for context cancellation
	<-p.ctx.Done()
	p.logger.Info().Msg("Processor service context cancelled")

	return nil
}

func (p *Processor) Stop() {
	p.logger.Info().Msg("Stopping processor service")

	if p.cancel != nil {
		p.cancel()
	}

	// Stop worker pool
	p.workerPool.Stop()

	// Close connections
	if p.consumer != nil {
		p.consumer.Close()
	}
	if p.publisher != nil {
		p.publisher.Close()
	}

	p.wg.Wait()
	p.logger.Info().Msg("Processor service stopped")
}

func (p *Processor) handleMessage(messageBody []byte) error {
	p.logger.Debug().Msg("Received message for processing")

	var message models.NewsMessage
	if err := json.Unmarshal(messageBody, &message); err != nil {
		p.logger.Error().Err(err).Msg("Failed to unmarshal message")
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Submit to worker pool
	job := &ProcessingJob{
		Message:   message,
		Processor: p,
	}

	select {
	case p.workerPool.jobQueue <- job:
		p.logger.Debug().Str("message_id", message.ID).Msg("Job submitted to worker pool")
		return nil
	case <-p.ctx.Done():
		return fmt.Errorf("processor context cancelled")
	default:
		p.logger.Warn().Str("message_id", message.ID).Msg("Worker pool queue full, rejecting message")
		return fmt.Errorf("worker pool queue full")
	}
}

func (p *Processor) processNews(ctx context.Context, message models.NewsMessage) error {
	startTime := time.Now()
	p.logger.Info().Str("message_id", message.ID).Str("title", message.Data.Title).Msg("Processing news article")

	// Check for duplicates
	isDuplicate, err := p.deduplicator.IsDuplicate(ctx, &message.Data)
	if err != nil {
		p.logger.Error().Err(err).Str("message_id", message.ID).Msg("Failed to check for duplicates")
		return fmt.Errorf("failed to check for duplicates: %w", err)
	}

	if isDuplicate {
		p.logger.Info().Str("message_id", message.ID).Str("hash", message.Data.Hash).Msg("Duplicate article detected, skipping")
		return nil
	}

	// Apply transformers
	processedNews := message.Data
	for _, transformer := range p.transformers {
		transformedNews, err := transformer.Transform(ctx, &processedNews)
		if err != nil {
			p.logger.Error().Err(err).Str("transformer", fmt.Sprintf("%T", transformer)).Msg("Transformer failed")
			continue // Continue with other transformers
		}
		processedNews = *transformedNews
	}

	// Save to database
	if err := p.newsService.CreateNews(ctx, &processedNews); err != nil {
		p.logger.Error().Err(err).Str("message_id", message.ID).Msg("Failed to save news to database")
		return fmt.Errorf("failed to save news: %w", err)
	}

	// Index for search
	if err := p.searchService.IndexNews(ctx, &processedNews); err != nil {
		p.logger.Error().Err(err).Str("message_id", message.ID).Msg("Failed to index news for search")
		// Don't return error as this is not critical
	}

	// Publish processed message
	processedMessage := models.NewsMessage{
		ID:        message.ID,
		Source:    message.Source,
		Type:      "processed",
		Data:      processedNews,
		Timestamp: time.Now(),
		Retry:     message.Retry,
	}

	if err := p.publisher.Publish("news.processed", processedMessage); err != nil {
		p.logger.Error().Err(err).Str("message_id", message.ID).Msg("Failed to publish processed message")
		// Don't return error as the main processing is complete
	}

	duration := time.Since(startTime)
	p.logger.Info().
		Str("message_id", message.ID).
		Str("title", processedNews.Title).
		Dur("duration", duration).
		Msg("News article processed successfully")

	return nil
}

// ProcessingJob represents a job for processing news
type ProcessingJob struct {
	Message   models.NewsMessage
	Processor *Processor
}

// ProcessorWorkerPool manages workers for processing news
type ProcessorWorkerPool struct {
	config   *config.Config
	logger   zerolog.Logger
	jobQueue chan *ProcessingJob
	workers  []*ProcessorWorker
	wg       sync.WaitGroup
}

func NewProcessorWorkerPool(cfg *config.Config, logger zerolog.Logger) *ProcessorWorkerPool {
	return &ProcessorWorkerPool{
		config:   cfg,
		logger:   logger,
		jobQueue: make(chan *ProcessingJob, 1000), // Buffer for 1000 jobs
	}
}

func (pwp *ProcessorWorkerPool) Start(ctx context.Context) {
	workerCount := 5 // TODO: Make this configurable
	pwp.workers = make([]*ProcessorWorker, workerCount)

	for i := 0; i < workerCount; i++ {
		worker := &ProcessorWorker{
			id:       i,
			logger:   pwp.logger,
			jobQueue: pwp.jobQueue,
		}

		pwp.workers[i] = worker
		pwp.wg.Add(1)

		go func(w *ProcessorWorker) {
			defer pwp.wg.Done()
			w.start(ctx)
		}(worker)
	}

	pwp.logger.Info().Int("workers", workerCount).Msg("Processor worker pool started")
}

func (pwp *ProcessorWorkerPool) Stop() {
	close(pwp.jobQueue)
	pwp.wg.Wait()
	pwp.logger.Info().Msg("Processor worker pool stopped")
}

// ProcessorWorker processes news articles
type ProcessorWorker struct {
	id       int
	logger   zerolog.Logger
	jobQueue <-chan *ProcessingJob
}

func (pw *ProcessorWorker) start(ctx context.Context) {
	pw.logger.Debug().Int("worker_id", pw.id).Msg("Processor worker started")

	for {
		select {
		case job, ok := <-pw.jobQueue:
			if !ok {
				pw.logger.Debug().Int("worker_id", pw.id).Msg("Job queue closed, worker stopping")
				return
			}

			pw.processJob(ctx, job)

		case <-ctx.Done():
			pw.logger.Debug().Int("worker_id", pw.id).Msg("Context cancelled, worker stopping")
			return
		}
	}
}

func (pw *ProcessorWorker) processJob(ctx context.Context, job *ProcessingJob) {
	pw.logger.Debug().
		Int("worker_id", pw.id).
		Str("message_id", job.Message.ID).
		Msg("Processing job")

	if err := job.Processor.processNews(ctx, job.Message); err != nil {
		pw.logger.Error().
			Err(err).
			Int("worker_id", pw.id).
			Str("message_id", job.Message.ID).
			Msg("Failed to process news")

		// Handle retry logic
		if job.Message.Retry < 3 { // Max 3 retries
			job.Message.Retry++
			
			// Publish to retry queue with delay
			retryMessage := job.Message
			retryMessage.Timestamp = time.Now().Add(time.Duration(job.Message.Retry) * time.Minute)
			
			if err := job.Processor.publisher.Publish("news.retry", retryMessage); err != nil {
				pw.logger.Error().Err(err).Str("message_id", job.Message.ID).Msg("Failed to publish retry message")
			}
		} else {
			// Max retries reached, send to failed queue
			failedMessage := job.Message
			failedMessage.Type = "failed"
			
			if err := job.Processor.publisher.Publish("news.failed", failedMessage); err != nil {
				pw.logger.Error().Err(err).Str("message_id", job.Message.ID).Msg("Failed to publish failed message")
			}
		}
		
		return
	}

	pw.logger.Debug().
		Int("worker_id", pw.id).
		Str("message_id", job.Message.ID).
		Msg("Job processed successfully")
}
