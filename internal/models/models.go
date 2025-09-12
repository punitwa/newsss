// Package models provides a compatibility layer for the refactored domain models.
// 
// DEPRECATED: This package is maintained for backward compatibility.
// New code should import domain-specific packages directly:
//   - news domain: "news-aggregator/internal/models/news"
//   - user domain: "news-aggregator/internal/models/user"
//   - source domain: "news-aggregator/internal/models/source"
//   - search domain: "news-aggregator/internal/models/search"
//   - messaging domain: "news-aggregator/internal/models/messaging"
//   - system domain: "news-aggregator/internal/models/system"
//   - shared utilities: "news-aggregator/internal/models/shared"
package models

import (
	"news-aggregator/internal/models/messaging"
	"news-aggregator/internal/models/news"
	"news-aggregator/internal/models/search"
	"news-aggregator/internal/models/shared"
	"news-aggregator/internal/models/source"
	"news-aggregator/internal/models/system"
	"news-aggregator/internal/models/user"
)

// =============================================================================
// NEWS DOMAIN - Re-exported types from news package
// =============================================================================

// News represents a news article
// DEPRECATED: Use news.News instead
type News = news.News

// Category represents a news category
// DEPRECATED: Use news.Category instead
type Category = news.Category

// NewsFilter represents filtering options for news
// DEPRECATED: Use news.Filter instead
type NewsFilter = news.Filter

// CategoryStats represents statistics for a category
// DEPRECATED: Use news.CategoryStats instead
type CategoryStats = news.CategoryStats

// SourceStats represents statistics for a source
// DEPRECATED: Use news.SourceStats instead
type SourceStats = news.SourceStats

// =============================================================================
// USER DOMAIN - Re-exported types from user package
// =============================================================================

// User represents a user in the system
// DEPRECATED: Use user.User instead
type User = user.User

// Bookmark represents a user's bookmarked article
// DEPRECATED: Use user.Bookmark instead
type Bookmark = user.Bookmark

// LoginRequest represents a user login request
// DEPRECATED: Use user.LoginRequest instead
type LoginRequest = user.LoginRequest

// RegisterRequest represents a user registration request
// DEPRECATED: Use user.RegisterRequest instead
type RegisterRequest = user.RegisterRequest

// UpdateProfileRequest represents a profile update request
// DEPRECATED: Use user.UpdateProfileRequest instead
type UpdateProfileRequest = user.UpdateProfileRequest

// BookmarkRequest represents a bookmark request
// DEPRECATED: Use user.BookmarkRequest instead
type BookmarkRequest = user.BookmarkRequest

// Preferences represents user preferences
// DEPRECATED: Use user.Preferences instead
type Preferences = user.Preferences

// PreferencesRequest represents a preferences update request
// DEPRECATED: Use user.UpdatePreferencesRequest instead
type PreferencesRequest = user.UpdatePreferencesRequest

// =============================================================================
// SOURCE DOMAIN - Re-exported types from source package
// =============================================================================

// Source represents a news source
// DEPRECATED: Use source.Source instead
type Source = source.Source

// SourceRequest represents a source request
// DEPRECATED: Use source.SourceRequest instead
type SourceRequest = source.SourceRequest

// =============================================================================
// SEARCH DOMAIN - Re-exported types from search package
// =============================================================================

// SearchResult represents search results
// DEPRECATED: Use search.Result instead
type SearchResult = search.Result

// SearchQuery represents a search query
// DEPRECATED: Use search.Query instead
type SearchQuery = search.Query

// =============================================================================
// MESSAGING DOMAIN - Re-exported types from messaging package
// =============================================================================

// NewsMessage represents a message in the processing pipeline
// DEPRECATED: Use messaging.NewsMessage instead
type NewsMessage = messaging.NewsMessage

// ProcessingResult represents processing results
// DEPRECATED: Use messaging.ProcessingResult instead
type ProcessingResult = messaging.ProcessingResult

// =============================================================================
// SYSTEM DOMAIN - Re-exported types from system package
// =============================================================================

// HealthCheck represents system health status
// DEPRECATED: Use system.HealthCheck instead
type HealthCheck = system.HealthCheck

// Metrics represents system metrics
// DEPRECATED: Use system.Metrics instead
type Metrics = system.Metrics

// WSMessage represents a WebSocket message
// DEPRECATED: Use system.WSMessage instead
type WSMessage = system.WSMessage

// WSNewsUpdate represents a WebSocket news update
// DEPRECATED: Use system.WSNewsUpdate instead
type WSNewsUpdate = system.WSNewsUpdate

// =============================================================================
// SHARED DOMAIN - Re-exported types from shared package
// =============================================================================

// Stats represents system statistics (combining news and system stats)
// DEPRECATED: Use system.SystemStats instead
type Stats struct {
	TotalArticles     int64             `json:"total_articles"`
	TotalUsers        int64             `json:"total_users"`
	TotalSources      int64             `json:"total_sources"`
	ArticlesToday     int64             `json:"articles_today"`
	ArticlesThisWeek  int64             `json:"articles_this_week"`
	ArticlesThisMonth int64             `json:"articles_this_month"`
	TopCategories     []CategoryStats   `json:"top_categories"`
	TopSources        []SourceStats     `json:"top_sources"`
}

// =============================================================================
// COMPATIBILITY ALIASES - For backward compatibility
// =============================================================================

// Type aliases for shared utilities
type (
	// PaginationRequest provides common pagination parameters
	PaginationRequest = shared.PaginationRequest
	
	// PaginationResponse provides pagination metadata
	PaginationResponse = shared.PaginationResponse
	
	// APIResponse provides standard API response structure
	APIResponse = shared.APIResponse
	
	// ErrorResponse provides standard error response structure
	ErrorResponse = shared.ErrorResponse
	
	// SuccessResponse provides standard success response structure
	SuccessResponse = shared.SuccessResponse
)
