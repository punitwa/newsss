# Datasources Module - Refactored Architecture

The datasources module has been completely refactored to achieve **modularity**, **readability**, **comprehensive functionality**, and **clean code design**. This document outlines the new architecture and how to use it.

## 🏗️ New Architecture Overview

```
datasources/
├── datasources.go              # Main package interface & backward compatibility
├── README.md                   # This documentation
├── core/                       # Core interfaces and base implementations
│   ├── interfaces.go           # Primary interfaces (DataSource, Parser, etc.)
│   ├── types.go               # Common types and constants
│   ├── errors.go              # Error types and definitions
│   └── base.go                # BaseSource implementation with metrics
├── sources/                    # Source-specific implementations
│   └── rss/                   # RSS feed source
│       ├── rss.go             # Main RSS source implementation
│       ├── parser.go          # RSS parsing logic
│       └── types.go           # RSS-specific types
├── utils/                      # Shared utilities
│   ├── http.go                # HTTP client utilities
│   ├── rate_limit.go          # Rate limiting implementations
│   ├── validation.go          # Input validation utilities
│   └── image/                 # Image processing utilities
│       ├── scraper.go         # Image extraction from web pages
│       └── processor.go       # Image validation and processing
└── factory/                   # Factory pattern for source creation
    └── factory.go             # Source factory implementation
```

## ✨ Key Improvements

### 1. **Modularity**
- **Separated Concerns**: Each component has a single responsibility
- **Interface-Driven Design**: Clear contracts between components
- **Pluggable Architecture**: Easy to add new source types
- **Independent Modules**: Components can be used independently

### 2. **Readability**
- **Clear Naming Conventions**: Descriptive names for all components
- **Comprehensive Documentation**: Every public function and type documented
- **Logical Organization**: Related functionality grouped together
- **Consistent Code Style**: Uniform patterns across the codebase

### 3. **Comprehensive Functionality**
- **Rich Error Handling**: Detailed error types with context
- **Metrics and Monitoring**: Built-in statistics and health checks
- **Flexible Configuration**: Extensive configuration options
- **Advanced Features**: Rate limiting, retry logic, image extraction

### 4. **Clean Code Design**
- **SOLID Principles**: Single responsibility, open/closed, dependency inversion
- **Design Patterns**: Factory, Strategy, Template Method patterns
- **Type Safety**: Strong typing with validation
- **Resource Management**: Proper cleanup and resource handling

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "time"

    "news-aggregator/internal/datasources"
    "news-aggregator/internal/datasources/core"
    
    "github.com/rs/zerolog"
)

func main() {
    logger := zerolog.New(os.Stdout)
    
    // Create a data source manager
    manager := datasources.NewManager(logger)
    
    // Configure an RSS source
    config := core.SourceConfig{
        Name:      "example-rss",
        Type:      core.SourceTypeRSS,
        URL:       "https://example.com/feed.xml",
        Schedule:  15 * time.Minute,
        RateLimit: 2.0,
        Enabled:   true,
        Timeout:   30 * time.Second,
    }
    
    // Add the source to the manager
    if err := manager.AddSource(config); err != nil {
        log.Fatal(err)
    }
    
    // Fetch news items
    ctx := context.Background()
    items, err := manager.FetchFromSource(ctx, "example-rss")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Fetched %d news items\n", len(items))
}
```

### Direct Source Creation

```go
// Create an RSS source directly
rssConfig := core.SourceConfig{
    Name:      "tech-news",
    Type:      core.SourceTypeRSS,
    URL:       "https://technews.com/feed.xml",
    Schedule:  10 * time.Minute,
    RateLimit: 3.0,
    Enabled:   true,
}

source, err := datasources.NewRSSSource(rssConfig, logger)
if err != nil {
    log.Fatal(err)
}

// Fetch items
items, err := source.Fetch(context.Background())
if err != nil {
    log.Fatal(err)
}
```

## 🔧 Core Components

### DataSource Interface

The main interface that all data sources implement:

```go
type DataSource interface {
    Fetch(ctx context.Context) ([]models.News, error)
    GetSchedule() time.Duration
    GetName() string
    GetType() string
    IsHealthy(ctx context.Context) bool
    Validate() error
}
```

### BaseSource

Provides common functionality for all sources:
- **Metrics Tracking**: Automatic collection of fetch statistics
- **Health Monitoring**: Built-in health checks and status tracking
- **Configuration Management**: Centralized config handling
- **Thread Safety**: Safe for concurrent use

### Source Factory

Creates sources using the factory pattern:

```go
factory := factory.NewSourceFactory(logger)

// Create any type of source
source, err := factory.CreateSource(config)

// Validate configuration before creation
err := factory.ValidateSourceConfig(config)

// Get supported types
types := factory.GetSupportedTypes()
```

## 📊 Advanced Features

### Metrics and Monitoring

```go
// Get source statistics
stats := source.GetStats()
fmt.Printf("Success rate: %.2f%%\n", stats.GetSuccessRate())

// Check health status
if source.IsHealthy(ctx) {
    fmt.Println("Source is healthy")
}

// Get detailed health information
healthStatus := source.GetHealthStatus(ctx)
```

### Rate Limiting

```go
// Built-in rate limiting
rateLimiter := datasources.NewRateLimiter(5.0, 1, logger) // 5 req/sec, burst of 1

// Wait for rate limit
err := rateLimiter.Wait(ctx)

// Check if request is allowed
if rateLimiter.Allow() {
    // Make request
}
```

### Image Processing

```go
// Extract images from web pages
imageScraper := datasources.NewImageScraper(10*time.Second, userAgent, logger)
imageURL, err := imageScraper.ExtractFromURL(ctx, articleURL)

// Validate and process images
imageProcessor := datasources.NewImageProcessor(5*time.Second, logger)
imageInfo, err := imageProcessor.ValidateImage(ctx, imageURL)
```

## 🔄 Migration from Legacy Code

The refactored module maintains **full backward compatibility**:

### Old Code (Still Works)
```go
// Legacy function calls still work
source, err := datasources.NewRSSSource(config, logger)
```

### New Recommended Approach
```go
// Use the new manager for better organization
manager := datasources.NewManager(logger)
err := manager.AddSource(config)
items, err := manager.FetchFromSource(ctx, "source-name")
```

## 🎯 RSS Source Features

### Advanced Parsing Options

```go
options := datasources.DefaultRSSParsingOptions()
options.MaxItems = 50
options.ExtractImages = true
options.SanitizeHTML = true
options.FilterDuplicates = true

// Apply options to RSS source
rssSource.SetParsingOptions(options)
```

### RSS-Specific Types

```go
// Access RSS-specific metadata
metadata, err := rssSource.GetMetadata(ctx)
fmt.Printf("Feed: %s (%d items)\n", metadata.Title, metadata.ItemCount)

// Validate RSS feed
validation, err := rssSource.ValidateFeed(ctx)
if validation.IsValid {
    fmt.Println("RSS feed is valid")
}
```

## 🛠️ Configuration Options

### Complete Source Configuration

```go
config := core.SourceConfig{
    Name:             "my-source",
    Type:             core.SourceTypeRSS,
    URL:              "https://example.com/feed.xml",
    Schedule:         15 * time.Minute,
    RateLimit:        2.0,
    Timeout:          30 * time.Second,
    MaxRetries:       3,
    RetryDelay:       5 * time.Second,
    Enabled:          true,
    UserAgent:        "MyApp/1.0",
    Headers: map[string]string{
        "Authorization": "Bearer token",
        "Accept":        "application/rss+xml",
    },
    Categories:       []string{"technology", "science"},
    Keywords:         []string{"AI", "machine learning"},
    Language:         "en",
    Country:          "US",
}
```

### Validation

```go
// Validate configuration
if err := config.Validate(); err != nil {
    log.Fatal("Invalid configuration:", err)
}

// Use validation utilities
if err := utils.ValidateURL(config.URL); err != nil {
    log.Fatal("Invalid URL:", err)
}
```

## 🔍 Error Handling

### Comprehensive Error Types

```go
// Specific error types
if errors.Is(err, core.ErrRateLimitExceeded) {
    // Handle rate limit
}

if errors.Is(err, core.ErrSourceDisabled) {
    // Handle disabled source
}

// Source-specific errors
if sourceErr, ok := err.(*core.SourceError); ok {
    fmt.Printf("Source %s failed: %v\n", sourceErr.SourceName, sourceErr.Err)
    if sourceErr.IsRetryable() {
        // Retry the operation
    }
}
```

## 📈 Performance Considerations

### Optimization Features

- **Connection Pooling**: HTTP clients use connection pooling
- **Rate Limiting**: Prevents overwhelming external services
- **Caching**: Built-in deduplication and caching mechanisms
- **Concurrent Processing**: Safe for concurrent operations
- **Resource Management**: Proper cleanup and resource handling

### Best Practices

```go
// Use context for timeouts
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
defer cancel()

// Monitor source health
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    for range ticker.C {
        if !source.IsHealthy(ctx) {
            log.Warn("Source is unhealthy")
        }
    }
}()

// Clean up resources
defer manager.Close()
```

## 🧪 Testing

### Test Structure (Planned)

```
tests/
├── unit/                  # Unit tests for individual components
├── integration/           # Integration tests
├── benchmarks/           # Performance benchmarks
└── fixtures/             # Test data and fixtures
```

### Example Test

```go
func TestRSSSource(t *testing.T) {
    logger := zerolog.Nop()
    config := core.SourceConfig{
        Name: "test-rss",
        Type: core.SourceTypeRSS,
        URL:  "https://example.com/feed.xml",
    }
    
    source, err := rss.NewSource(config, logger)
    require.NoError(t, err)
    
    // Test fetch
    items, err := source.Fetch(context.Background())
    require.NoError(t, err)
    assert.NotEmpty(t, items)
}
```

## 🚧 Future Extensions

The modular architecture makes it easy to add:

- **New Source Types**: API sources, web scrapers, social media feeds
- **Advanced Parsers**: JSON-LD, Microdata, custom formats
- **Enhanced Processing**: AI-powered content analysis, sentiment analysis
- **Caching Layers**: Redis, in-memory caching
- **Monitoring**: Prometheus metrics, health dashboards

## 📝 Summary

This refactored datasources module provides:

✅ **Modular Design** - Clean separation of concerns  
✅ **Readable Code** - Clear, well-documented interfaces  
✅ **Comprehensive Features** - Rich functionality with advanced options  
✅ **Clean Architecture** - SOLID principles and design patterns  
✅ **Backward Compatibility** - Existing code continues to work  
✅ **Extensibility** - Easy to add new features and source types  
✅ **Performance** - Optimized for production use  
✅ **Maintainability** - Easy to understand and modify  

The new architecture provides a solid foundation for building robust, scalable news aggregation systems while maintaining the simplicity and ease of use of the original design.
