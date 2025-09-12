# Models Package

The models package has been refactored to follow a domain-driven design approach, organizing models by use case and business domain. This improves maintainability, reduces coupling, and makes the codebase more scalable.

## Package Structure

```
models/
├── README.md
├── models.go (legacy compatibility layer)
├── news/
│   ├── news.go          # News articles, categories, filtering
│   └── errors.go        # News domain errors
├── user/
│   ├── user.go          # User entity and preferences
│   ├── requests.go      # Authentication and profile requests
│   ├── bookmark.go      # User bookmarks
│   └── errors.go        # User domain errors
├── source/
│   ├── source.go        # Data sources and management
│   └── errors.go        # Source domain errors
├── search/
│   ├── search.go        # Search queries and results
│   └── errors.go        # Search domain errors
├── messaging/
│   ├── messaging.go     # Message queue and processing
│   └── errors.go        # Messaging domain errors
├── system/
│   ├── system.go        # Health checks, metrics, WebSocket
│   └── errors.go        # System domain errors
└── shared/
    ├── types.go         # Common types and utilities
    └── errors.go        # Shared errors
```

## Domain Overview

### 📰 News Domain (`news/`)

Handles all news-related entities and operations:

- **News**: Core news article entity with validation
- **Category**: News categorization
- **Filter**: News filtering and pagination
- **Stats**: News statistics and analytics

**Key Features:**
- Content validation and sanitization
- Deduplication support via hash
- Age-based filtering helpers
- Category management

### 👤 User Domain (`user/`)

Manages user accounts, authentication, and preferences:

- **User**: Core user entity with profile management
- **Preferences**: User settings and customization
- **Requests**: Authentication and profile update requests
- **Bookmark**: User bookmarking functionality

**Key Features:**
- Password hashing and validation
- Email validation
- Preference management
- Bookmark organization with tags

### 🔗 Source Domain (`source/`)

Handles news sources and data collection:

- **Source**: News source configuration and management
- **SourceRequest**: Source creation and updates
- **HealthStatus**: Source health monitoring
- **SourceStats**: Performance metrics

**Key Features:**
- Multi-type source support (RSS, API, Scraper)
- Health monitoring and error tracking
- Rate limiting configuration
- Schedule management

### 🔍 Search Domain (`search/`)

Provides search functionality and saved searches:

- **Query**: Search queries with filters and facets
- **Result**: Search results with pagination
- **SavedSearch**: User-saved search queries
- **Facets**: Search result aggregations

**Key Features:**
- Advanced filtering and faceting
- Search history tracking
- Saved search management
- Trending query analysis

### 📨 Messaging Domain (`messaging/`)

Handles message queue operations and processing:

- **NewsMessage**: Pipeline message structure
- **ProcessingResult**: Processing outcomes
- **QueueStats**: Queue performance metrics
- **DeadLetterMessage**: Failed message handling

**Key Features:**
- Retry logic with exponential backoff
- Message expiration handling
- Processing pipeline support
- Dead letter queue management

### 🖥️ System Domain (`system/`)

Manages system health, metrics, and WebSocket communication:

- **HealthCheck**: System health monitoring
- **Metrics**: Performance and resource metrics
- **WSMessage**: WebSocket communication
- **SystemStats**: Overall system statistics

**Key Features:**
- Multi-service health checks
- Resource usage monitoring
- Real-time WebSocket communication
- System-wide statistics

### 🔧 Shared Domain (`shared/`)

Provides common utilities and types:

- **PaginationRequest/Response**: Common pagination
- **APIResponse**: Standard API responses
- **DateRange**: Date filtering utilities
- **ValidationErrors**: Structured validation errors

**Key Features:**
- Consistent API responses
- Reusable pagination logic
- Common validation patterns
- Error handling utilities

## Usage Examples

### News Operations

```go
import "news-aggregator/internal/models/news"

// Create and validate news article
article := &news.News{
    Title:   "Breaking News",
    Content: "Article content...",
    Source:  "example.com",
    URL:     "https://example.com/article",
}

if err := article.Validate(); err != nil {
    log.Printf("Validation error: %v", err)
}

// Check if article is recent
if article.IsRecent() {
    fmt.Println("This is a recent article")
}
```

### User Management

```go
import "news-aggregator/internal/models/user"

// Create user from registration request
req := &user.RegisterRequest{
    Email:     "user@example.com",
    Username:  "johndoe",
    Password:  "SecurePass123",
    FirstName: "John",
    LastName:  "Doe",
}

if err := req.Validate(); err != nil {
    return err
}

user := req.ToUser()
```

### Search Operations

```go
import "news-aggregator/internal/models/search"

// Create search query
query := &search.Query{
    Query:      "technology news",
    Categories: []string{"tech", "science"},
    DateFrom:   time.Now().AddDate(0, 0, -7),
    Page:       1,
    Limit:      20,
}

query.SetDefaults()
if err := query.Validate(); err != nil {
    return err
}
```

### System Health Checks

```go
import "news-aggregator/internal/models/system"

// Create health check
health := system.NewHealthCheck("v1.0.0")
health.AddService("database", "healthy")
health.AddService("redis", "healthy")
health.AddService("rabbitmq", "degraded")
health.SetOverallStatus()

if !health.IsHealthy() {
    log.Printf("System is %s", health.Status)
}
```

## Migration Guide

### From Legacy Models

The original `models.go` file has been kept as a compatibility layer. To migrate:

1. **Update imports**: Change from `models.ModelName` to `domain.ModelName`
2. **Use domain-specific errors**: Replace generic errors with domain-specific ones
3. **Leverage validation methods**: Use built-in validation instead of custom logic
4. **Adopt helper methods**: Use provided convenience methods

### Example Migration

**Before:**
```go
import "news-aggregator/internal/models"

user := &models.User{...}
if user.Email == "" {
    return errors.New("email is required")
}
```

**After:**
```go
import "news-aggregator/internal/models/user"

user := &user.User{...}
if err := user.Validate(); err != nil {
    return err
}
```

## Benefits

1. **Domain Separation**: Clear boundaries between business domains
2. **Reduced Coupling**: Models are self-contained within their domains
3. **Better Testability**: Domain-specific models can be tested in isolation
4. **Improved Maintainability**: Changes to one domain don't affect others
5. **Enhanced Validation**: Domain-specific validation rules
6. **Error Handling**: Contextual, domain-specific error messages
7. **Code Reusability**: Shared utilities reduce duplication

## Best Practices

1. **Use validation methods**: Always validate models before persistence
2. **Leverage helper methods**: Use provided convenience methods
3. **Handle domain errors**: Catch and handle domain-specific errors
4. **Follow naming conventions**: Use consistent naming across domains
5. **Document dependencies**: Clearly document cross-domain dependencies
6. **Test thoroughly**: Write comprehensive tests for each domain

This modular approach makes the codebase more maintainable, testable, and scalable while providing clear separation of concerns across different business domains.
