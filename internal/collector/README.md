# News Collector Module

The collector module is responsible for gathering news articles from various data sources and processing them through a robust, scalable pipeline. This module has been refactored to follow clean code principles and maintain high modularity.

## Architecture Overview

The collector module is composed of several key components:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Collector     │────│ Source Manager  │────│  Data Sources   │
│   (Main)        │    │                 │    │  (RSS/API/Web)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │
         ├─────────────────┐
         │                 │
┌─────────────────┐    ┌─────────────────┐
│  Job Scheduler  │    │  Worker Pool    │
│                 │    │                 │
└─────────────────┘    └─────────────────┘
                              │
                    ┌─────────────────┐
                    │  Message Queue  │
                    │   (RabbitMQ)    │
                    └─────────────────┘
```

## Components

### 1. Collector (`collector.go`)

The main collector orchestrates all other components and provides the primary interface for the data collection service.

**Key Features:**
- Manages component lifecycle (start/stop)
- Coordinates source scheduling
- Provides status and metrics
- Handles graceful shutdown

**Interface:**
```go
type Collector interface {
    Start(ctx context.Context) error
    Stop()
    AddSource(sourceConfig config.SourceConfig) error
    RemoveSource(sourceName string) error
    GetSourceStatus() map[string]interface{}
}
```

### 2. Source Manager (`source_manager.go`)

Manages all data sources including initialization, validation, and health checking.

**Key Features:**
- Source lifecycle management
- Configuration validation
- Health monitoring
- Thread-safe operations

**Interface:**
```go
type SourceManager interface {
    Initialize(sourceConfigs []config.SourceConfig) error
    AddSource(sourceConfig config.SourceConfig) error
    RemoveSource(sourceName string) error
    GetSource(sourceName string) (datasources.DataSource, bool)
    GetAllSources() map[string]datasources.DataSource
    GetStatus() map[string]interface{}
}
```

### 3. Job Scheduler (`scheduler.go`)

Handles the scheduling of data collection jobs from various sources.

**Key Features:**
- Flexible scheduling per source
- Job execution with error handling
- Schedule validation and management
- Panic recovery

**Interface:**
```go
type JobScheduler interface {
    Start()
    Stop()
    ScheduleSource(sourceName string, source datasources.DataSource, handler func()) error
    RemoveSource(sourceName string) error
}
```

### 4. Worker Pool (`worker_pool.go`)

Manages a pool of workers for concurrent job processing.

**Key Features:**
- Configurable worker count
- Job queuing with backpressure
- Metrics collection
- Graceful shutdown

**Interface:**
```go
type WorkerPool interface {
    Start(ctx context.Context)
    Stop()
    SubmitJob(job *CollectionJob) error
    GetStats() WorkerPoolStats
}
```

### 5. Job Processing (`job.go`)

Defines job types and processing logic with retry mechanisms.

**Key Features:**
- Job validation
- Retry logic with exponential backoff
- Timeout handling
- Result tracking

## Configuration

The collector is highly configurable through the configuration system:

```yaml
collector:
  worker_count: 10        # Number of worker goroutines
  queue_size: 1000        # Job queue buffer size
  job_timeout: 30s        # Maximum time for job processing
  retry_attempts: 3       # Number of retry attempts for failed jobs
  retry_delay: 5s         # Base delay between retries
  metrics_enabled: true   # Enable metrics collection
```

## Usage

### Basic Usage

```go
// Load configuration
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// Create collector
collector, err := collector.New(cfg, logger)
if err != nil {
    log.Fatal(err)
}

// Start collector
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    if err := collector.Start(ctx); err != nil {
        log.Printf("Collector error: %v", err)
    }
}()

// Graceful shutdown
collector.Stop()
```

### Adding Sources Dynamically

```go
sourceConfig := config.SourceConfig{
    Name:     "new-rss-feed",
    Type:     "rss",
    URL:      "https://example.com/feed.xml",
    Schedule: "5m",
    Enabled:  true,
}

err := collector.AddSource(sourceConfig)
if err != nil {
    log.Printf("Failed to add source: %v", err)
}
```

### Monitoring

```go
// Get source status
status := collector.GetSourceStatus()

// Get metrics
metrics := collector.GetMetrics()
fmt.Printf("Total jobs: %d, Success rate: %.2f%%", 
    metrics.TotalJobs, 
    float64(metrics.SuccessfulJobs)/float64(metrics.TotalJobs)*100)
```

## Error Handling

The collector implements comprehensive error handling:

1. **Source Initialization Errors**: Non-fatal, continues with other sources
2. **Scheduling Errors**: Logged and reported in status
3. **Job Processing Errors**: Retried with exponential backoff
4. **Queue Full**: Jobs are dropped with warning logs
5. **Panic Recovery**: Workers recover from panics and continue

## Metrics and Monitoring

The collector provides detailed metrics:

- **Job Statistics**: Total, successful, and failed job counts
- **Performance Metrics**: Average job processing time
- **Queue Metrics**: Queue utilization and active workers
- **Source Health**: Individual source status and health checks

## Testing

The modular architecture makes testing straightforward:

```go
func TestCollector(t *testing.T) {
    // Mock dependencies
    mockSourceManager := &MockSourceManager{}
    mockWorkerPool := &MockWorkerPool{}
    mockScheduler := &MockScheduler{}
    
    // Test collector behavior
    collector := &collector{
        sourceManager: mockSourceManager,
        workerPool:    mockWorkerPool,
        scheduler:     mockScheduler,
    }
    
    // Run tests...
}
```

## Performance Considerations

1. **Worker Pool Sizing**: Configure based on I/O characteristics of sources
2. **Queue Size**: Balance memory usage with throughput requirements
3. **Retry Strategy**: Tune retry attempts and delays for your sources
4. **Timeout Settings**: Set appropriate timeouts for different source types

## Migration from Legacy Code

The refactored collector maintains API compatibility while providing:

- Better separation of concerns
- Improved testability
- Enhanced error handling
- Configurable parameters
- Comprehensive monitoring

Existing code should work with minimal changes, requiring only configuration updates to take advantage of new features.
