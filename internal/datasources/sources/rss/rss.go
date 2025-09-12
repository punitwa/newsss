// Package rss provides RSS feed data source implementation.
package rss

import (
	"context"
	"fmt"
	"time"

	"news-aggregator/internal/datasources/core"
	"news-aggregator/internal/datasources/utils"
	"news-aggregator/internal/datasources/utils/image"
	"news-aggregator/internal/models"

	"github.com/rs/zerolog"
)

// Source implements the DataSource interface for RSS feeds.
type Source struct {
	*core.BaseSource

	// Configuration
	config core.SourceConfig

	// Components
	httpClient   core.HTTPClient
	rateLimiter  core.RateLimiter
	parser       *Parser
	imageScraper *image.Scraper

	// Logger
	logger zerolog.Logger
}

// NewSource creates a new RSS source with the given configuration.
func NewSource(config core.SourceConfig, logger zerolog.Logger) (*Source, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid RSS source configuration: %w", err)
	}

	// Ensure it's an RSS source
	if config.Type != core.SourceTypeRSS {
		return nil, core.NewValidationError("type", config.Type, "source type must be RSS")
	}

	// Create base source
	baseSource := core.NewBaseSource(config, logger)

	// Create HTTP client
	httpClient := utils.NewHTTPClient(
		config.Timeout,
		config.GetDefaultUserAgent(),
		logger,
	)

	// Create rate limiter
	rateLimiter := utils.NewRateLimiter(
		config.RateLimit,
		1, // burst of 1 for RSS feeds
		logger,
	)

	// Create parser with default options
	parsingOptions := DefaultParsingOptions()
	parser := NewParser(logger, parsingOptions)

	// Create image scraper
	imageScraper := image.NewScraper(
		10*time.Second, // shorter timeout for images
		config.GetDefaultUserAgent(),
		logger,
	)

	source := &Source{
		BaseSource:   baseSource,
		config:       config,
		httpClient:   httpClient,
		rateLimiter:  rateLimiter,
		parser:       parser,
		imageScraper: imageScraper,
		logger:       logger.With().Str("source_type", "rss").Str("source_name", config.Name).Logger(),
	}

	return source, nil
}

// Fetch retrieves and processes news items from the RSS feed.
func (s *Source) Fetch(ctx context.Context) ([]models.News, error) {
	if !s.IsEnabled() {
		return nil, core.ErrSourceDisabled
	}

	s.logger.Info().Str("url", s.config.URL).Msg("Starting RSS feed fetch")

	// Record fetch start
	s.RecordFetchStart()
	startTime := time.Now()

	// Apply rate limiting
	if err := s.rateLimiter.Wait(ctx); err != nil {
		responseTime := time.Since(startTime)
		s.RecordFetchFailure(responseTime, core.ErrRateLimitExceeded)
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Fetch RSS feed content
	content, err := s.httpClient.Get(ctx, s.config.URL, s.config.Headers)
	if err != nil {
		responseTime := time.Since(startTime)
		s.RecordFetchFailure(responseTime, err)
		return nil, core.NewSourceError(s.config.Name, s.config.Type, "fetch", err)
	}

	// Parse RSS content to news items
	newsItems, err := s.parser.ParseToNews(ctx, content, s.config.Name)
	if err != nil {
		responseTime := time.Since(startTime)
		s.RecordFetchFailure(responseTime, err)
		return nil, core.NewSourceError(s.config.Name, s.config.Type, "parse", err)
	}

	// Enhance items with additional processing
	newsItems = s.enhanceNewsItems(ctx, newsItems)

	// Record successful fetch
	responseTime := time.Since(startTime)
	s.RecordFetchSuccess(responseTime, int64(len(newsItems)))

	s.logger.Info().
		Int("items_fetched", len(newsItems)).
		Dur("response_time", responseTime).
		Msg("RSS feed fetch completed successfully")

	return newsItems, nil
}

// enhanceNewsItems performs additional processing on news items.
func (s *Source) enhanceNewsItems(ctx context.Context, items []models.News) []models.News {
	enhanced := make([]models.News, len(items))

	for i, item := range items {
		enhanced[i] = item

		// Extract image if not already present
		if item.ImageURL == "" && item.URL != "" {
			if imageURL, err := s.imageScraper.ExtractFromURL(ctx, item.URL); err == nil && imageURL != "" {
				enhanced[i].ImageURL = imageURL
				s.logger.Debug().
					Str("article_url", item.URL).
					Str("image_url", imageURL).
					Msg("Extracted image from article")
			}
		}

		// Ensure timestamps are set
		if enhanced[i].CreatedAt.IsZero() {
			enhanced[i].CreatedAt = time.Now()
		}
		if enhanced[i].UpdatedAt.IsZero() {
			enhanced[i].UpdatedAt = time.Now()
		}
	}

	return enhanced
}

// IsHealthy performs a health check on the RSS source.
func (s *Source) IsHealthy(ctx context.Context) bool {
	// Use base health check first
	if !s.BaseSource.IsHealthy(ctx) {
		return false
	}

	// Perform a GET request to check if the RSS feed is accessible
	if _, err := s.httpClient.Get(ctx, s.config.URL, s.config.Headers); err != nil {
		s.logger.Warn().Err(err).Msg("RSS source health check failed")
		return false
	}

	return true
}

// GetMetadata returns metadata about the RSS feed.
func (s *Source) GetMetadata(ctx context.Context) (*FeedMetadata, error) {
	// Fetch RSS content
	content, err := s.httpClient.Get(ctx, s.config.URL, s.config.Headers)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS feed for metadata: %w", err)
	}

	// Parse feed
	feed, err := s.parser.Parse(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed for metadata: %w", err)
	}

	// Extract metadata
	metadata := s.parser.GetFeedMetadata(feed)

	return metadata, nil
}

// ValidateFeed validates an RSS feed without processing all items.
func (s *Source) ValidateFeed(ctx context.Context) (*ValidationResult, error) {
	result := &ValidationResult{
		IsValid: false,
	}

	// Fetch RSS content
	content, err := s.httpClient.Get(ctx, s.config.URL, s.config.Headers)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to fetch feed: %v", err))
		return result, nil
	}

	// Try to parse the feed
	feed, err := s.parser.Parse(ctx, content)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse feed: %v", err))
		return result, nil
	}

	// Basic validation
	if feed.Channel.Title == "" {
		result.Warnings = append(result.Warnings, "Feed has no title")
	}

	if feed.Channel.Description == "" {
		result.Warnings = append(result.Warnings, "Feed has no description")
	}

	result.ItemCount = len(feed.Channel.Items)
	if result.ItemCount == 0 {
		result.Warnings = append(result.Warnings, "Feed contains no items")
	}

	// Check if items have content
	hasContent := false
	for _, item := range feed.Channel.Items {
		if item.Content != "" || item.Description != "" {
			hasContent = true
			break
		}
	}
	result.HasContent = hasContent

	if !hasContent {
		result.Warnings = append(result.Warnings, "Feed items have no content")
	}

	// Check for recent updates
	if feed.Channel.LastBuildDate != "" {
		result.LastModified = feed.Channel.LastBuildDate
	} else if feed.Channel.PubDate != "" {
		result.LastModified = feed.Channel.PubDate
	}

	// If we got this far, the feed is valid
	result.IsValid = true

	return result, nil
}

// UpdateConfiguration updates the source configuration.
func (s *Source) UpdateConfiguration(config core.SourceConfig) error {
	// Validate new configuration
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	if config.Type != core.SourceTypeRSS {
		return core.NewValidationError("type", config.Type, "source type must be RSS")
	}

	// Update configuration
	s.config = config

	// Update HTTP client settings
	s.httpClient.SetTimeout(config.Timeout)
	s.httpClient.SetUserAgent(config.GetDefaultUserAgent())

	// Update rate limiter
	s.rateLimiter.SetLimit(config.RateLimit)

	// Update base source
	*s.BaseSource = *core.NewBaseSource(config, s.logger)

	s.logger.Info().Msg("RSS source configuration updated")

	return nil
}

// GetParsingOptions returns the current parsing options.
func (s *Source) GetParsingOptions() ParsingOptions {
	return s.parser.options
}

// SetParsingOptions updates the parsing options.
func (s *Source) SetParsingOptions(options ParsingOptions) {
	s.parser.options = options
	s.logger.Info().
		Int("max_items", options.MaxItems).
		Bool("extract_images", options.ExtractImages).
		Bool("sanitize_html", options.SanitizeHTML).
		Msg("RSS parsing options updated")
}

// GetSourceInfo returns information about this RSS source.
func (s *Source) GetSourceInfo() map[string]interface{} {
	info := map[string]interface{}{
		"type":       s.GetType(),
		"name":       s.GetName(),
		"url":        s.config.URL,
		"schedule":   s.GetSchedule().String(),
		"enabled":    s.IsEnabled(),
		"rate_limit": s.config.RateLimit,
		"timeout":    s.config.Timeout.String(),
		"user_agent": s.config.GetDefaultUserAgent(),
		"headers":    s.config.Headers,
		"categories": s.config.Categories,
		"keywords":   s.config.Keywords,
		"language":   s.config.Language,
		"country":    s.config.Country,
	}

	// Add statistics
	stats := s.GetStats()
	info["statistics"] = stats

	// Add parsing options
	info["parsing_options"] = s.GetParsingOptions()

	return info
}

// Close cleans up resources used by the RSS source.
func (s *Source) Close() error {
	s.logger.Info().Msg("Closing RSS source")

	// Reset metrics
	s.ResetMetrics()

	return nil
}
