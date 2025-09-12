// Package rss provides RSS feed parsing functionality.
package rss

import (
	"context"
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"news-aggregator/internal/datasources/core"
	"news-aggregator/internal/models"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// Parser provides RSS feed parsing functionality.
type Parser struct {
	logger  zerolog.Logger
	options ParsingOptions
}

// NewParser creates a new RSS parser with the specified options.
func NewParser(logger zerolog.Logger, options ParsingOptions) *Parser {
	return &Parser{
		logger:  logger.With().Str("component", "rss_parser").Logger(),
		options: options,
	}
}

// Parse parses RSS feed data into structured format.
func (p *Parser) Parse(ctx context.Context, data []byte) (*Feed, error) {
	if len(data) == 0 {
		return nil, core.NewParsingError("rss", "", fmt.Errorf("empty feed data"))
	}

	// Check feed size
	if len(data) > MaxFeedSize {
		return nil, core.NewParsingError("rss", "", fmt.Errorf("feed size exceeds maximum limit of %d bytes", MaxFeedSize))
	}

	p.logger.Debug().Int("data_size", len(data)).Msg("Parsing RSS feed")

	var feed Feed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, core.NewParsingError("rss", string(data), err)
	}

	// Validate feed structure
	if err := p.validateFeed(&feed); err != nil {
		return nil, err
	}

	p.logger.Debug().
		Str("title", feed.Channel.Title).
		Int("items", len(feed.Channel.Items)).
		Msg("RSS feed parsed successfully")

	return &feed, nil
}

// ParseToNews converts RSS feed data to news items.
func (p *Parser) ParseToNews(ctx context.Context, data []byte, sourceName string) ([]models.News, error) {
	feed, err := p.Parse(ctx, data)
	if err != nil {
		return nil, err
	}

	return p.ConvertToNews(ctx, feed, sourceName)
}

// ConvertToNews converts a parsed RSS feed to news items.
func (p *Parser) ConvertToNews(ctx context.Context, feed *Feed, sourceName string) ([]models.News, error) {
	if feed == nil {
		return nil, fmt.Errorf("feed cannot be nil")
	}

	startTime := time.Now()

	var newsItems []models.News
	var stats FeedStats
	stats.TotalItems = len(feed.Channel.Items)

	// Track duplicates if filtering is enabled
	var seenItems map[string]bool
	if p.options.FilterDuplicates {
		seenItems = make(map[string]bool)
	}

	// Process items
	maxItems := p.options.MaxItems
	if maxItems <= 0 || maxItems > MaxItemsPerFeed {
		maxItems = MaxItemsPerFeed
	}

	processed := 0
	for i, item := range feed.Channel.Items {
		if processed >= maxItems {
			break
		}

		// Parse item
		newsItem, err := p.parseItem(&item, &feed.Channel, sourceName)
		if err != nil {
			p.logger.Warn().
				Err(err).
				Int("item_index", i).
				Str("item_title", item.Title).
				Msg("Failed to parse RSS item")
			stats.SkippedItems++
			continue
		}

		// Check for duplicates
		if p.options.FilterDuplicates {
			itemKey := p.generateItemKey(newsItem)
			if seenItems[itemKey] {
				stats.DuplicateItems++
				continue
			}
			seenItems[itemKey] = true
		}

		// Apply content length filter
		if p.options.MinContentLength > 0 && len(newsItem.Content) < p.options.MinContentLength {
			stats.SkippedItems++
			continue
		}

		newsItems = append(newsItems, *newsItem)
		stats.ValidItems++

		// Update statistics
		if len(newsItem.Content) > 0 {
			stats.AverageLength += len(newsItem.Content)
		}
		if newsItem.ImageURL != "" {
			stats.HasImages++
		}

		processed++
	}

	// Calculate final statistics
	stats.ProcessingTime = time.Since(startTime)
	if stats.ValidItems > 0 {
		stats.AverageLength /= stats.ValidItems
	}

	p.logger.Info().
		Str("source", sourceName).
		Int("total_items", stats.TotalItems).
		Int("valid_items", stats.ValidItems).
		Int("skipped_items", stats.SkippedItems).
		Int("duplicate_items", stats.DuplicateItems).
		Dur("processing_time", stats.ProcessingTime).
		Msg("RSS feed processing completed")

	return newsItems, nil
}

// parseItem converts an RSS item to a news item.
func (p *Parser) parseItem(item *Item, channel *Channel, sourceName string) (*models.News, error) {
	// Generate unique ID
	id := p.generateItemID(item)

	// Parse publication date
	pubDate := p.parseDate(item.PubDate)
	if pubDate.IsZero() && item.DCDate != "" {
		pubDate = p.parseDate(item.DCDate)
	}

	// Extract and clean content
	content := p.extractContent(item)
	if p.options.SanitizeHTML {
		content = p.sanitizeHTML(content)
	}

	// Extract description
	description := p.extractDescription(item)
	if p.options.SanitizeHTML {
		description = p.sanitizeHTML(description)
	}

	// Extract image URL
	var imageURL string
	if p.options.ExtractImages {
		imageURL = p.extractImageURL(item, channel)
	}

	// Extract categories
	categories := p.extractCategories(item)

	// Extract author
	author := p.extractAuthor(item)

	// Create news item
	newsItem := &models.News{
		ID:          id,
		Title:       strings.TrimSpace(html.UnescapeString(item.Title)),
		Content:     content,
		Summary:     description,
		URL:         strings.TrimSpace(item.Link),
		Author:      author,
		PublishedAt: pubDate,
		Category:    strings.Join(categories, ", "), // Convert slice to string
		ImageURL:    imageURL,
		Source:      sourceName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Validate the news item
	if err := p.validateNewsItem(newsItem); err != nil {
		return nil, err
	}

	return newsItem, nil
}

// extractContent extracts the main content from an RSS item.
func (p *Parser) extractContent(item *Item) string {
	// Priority order: content:encoded > description > title
	if item.Content != "" {
		return strings.TrimSpace(item.Content)
	}

	if item.Description != "" {
		return strings.TrimSpace(item.Description)
	}

	return strings.TrimSpace(item.Title)
}

// extractDescription extracts description from an RSS item.
func (p *Parser) extractDescription(item *Item) string {
	if item.Description != "" {
		desc := strings.TrimSpace(item.Description)

		// If description is too long, truncate it
		if len(desc) > 500 {
			// Find a good breaking point
			if idx := strings.LastIndex(desc[:500], ". "); idx > 100 {
				return desc[:idx+1]
			}
			return desc[:497] + "..."
		}

		return desc
	}

	// Fallback to content if no description
	if item.Content != "" {
		content := p.sanitizeHTML(item.Content)
		if len(content) > 200 {
			if idx := strings.LastIndex(content[:200], ". "); idx > 50 {
				return content[:idx+1]
			}
			return content[:197] + "..."
		}
		return content
	}

	return ""
}

// extractImageURL extracts image URL from RSS item.
func (p *Parser) extractImageURL(item *Item, channel *Channel) string {
	// Check enclosure for images
	if item.Enclosure != nil && p.isImageType(item.Enclosure.Type) {
		return item.Enclosure.URL
	}

	// Check Media RSS content
	for _, media := range item.MediaContent {
		if p.isImageType(media.Type) || media.Medium == "image" {
			return media.URL
		}
	}

	// Check Media RSS thumbnails
	for _, thumb := range item.MediaThumbnail {
		if thumb.URL != "" {
			return thumb.URL
		}
	}

	// Extract from content/description HTML
	content := item.Content
	if content == "" {
		content = item.Description
	}

	if content != "" {
		if imgURL := p.extractImageFromHTML(content); imgURL != "" {
			return imgURL
		}
	}

	// Fallback to channel image
	if channel.Image != nil && channel.Image.URL != "" {
		return channel.Image.URL
	}

	return ""
}

// extractImageFromHTML extracts the first image URL from HTML content.
func (p *Parser) extractImageFromHTML(htmlContent string) string {
	// Look for img tags
	imgRegex := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["'][^>]*>`)
	matches := imgRegex.FindStringSubmatch(htmlContent)

	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	return ""
}

// extractCategories extracts categories from RSS item.
func (p *Parser) extractCategories(item *Item) []string {
	var categories []string

	for _, cat := range item.Category {
		if cat.Value != "" {
			categories = append(categories, strings.TrimSpace(cat.Value))
		}
	}

	// Also check Dublin Core subject
	if item.DCSubject != "" {
		subjects := strings.Split(item.DCSubject, ",")
		for _, subject := range subjects {
			subject = strings.TrimSpace(subject)
			if subject != "" {
				categories = append(categories, subject)
			}
		}
	}

	// Remove duplicates
	return p.removeDuplicateStrings(categories)
}

// extractAuthor extracts author information from RSS item.
func (p *Parser) extractAuthor(item *Item) string {
	// Priority: author > dc:creator
	if item.Author != "" {
		return strings.TrimSpace(item.Author)
	}

	if item.DCCreator != "" {
		return strings.TrimSpace(item.DCCreator)
	}

	return ""
}

// parseDate parses various date formats commonly used in RSS feeds.
func (p *Parser) parseDate(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}

	dateStr = strings.TrimSpace(dateStr)

	// Common RSS date formats to try
	formats := []string{
		RFC822,
		RFC822Z,
		RFC3339,
		ISO8601,
		"Mon, 2 Jan 2006 15:04:05 MST",
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	p.logger.Warn().Str("date_string", dateStr).Msg("Failed to parse date")
	return time.Time{}
}

// sanitizeHTML removes or escapes potentially harmful HTML content.
func (p *Parser) sanitizeHTML(content string) string {
	if content == "" {
		return ""
	}

	// Remove script and style tags completely
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	content = scriptRegex.ReplaceAllString(content, "")

	styleRegex := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	content = styleRegex.ReplaceAllString(content, "")

	// Remove dangerous attributes
	onEventRegex := regexp.MustCompile(`(?i)\s+on\w+\s*=\s*["'][^"']*["']`)
	content = onEventRegex.ReplaceAllString(content, "")

	// Clean up extra whitespace
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
	content = strings.TrimSpace(content)

	return content
}

// generateItemID generates a unique ID for an RSS item.
func (p *Parser) generateItemID(item *Item) string {
	// Create a deterministic UUID based on the content
	// This ensures the same article gets the same ID across multiple fetches
	var sourceContent string

	// Use GUID if available
	if item.GUID != nil && item.GUID.Value != "" {
		sourceContent = item.GUID.Value
	} else if item.Link != "" {
		// Use link if available
		sourceContent = item.Link
	} else {
		// Generate from title and date
		sourceContent = item.Title + item.PubDate + item.Description
	}

	// Generate a deterministic UUID from the source content
	// Create a UUID from the hash (using version 5 namespace approach)
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(sourceContent)).String()
}

// generateItemKey generates a key for duplicate detection.
func (p *Parser) generateItemKey(item *models.News) string {
	// Use URL if available
	if item.URL != "" {
		return item.URL
	}

	// Generate hash from title and content
	content := item.Title + item.Content
	hash := md5.Sum([]byte(content))
	return fmt.Sprintf("%x", hash)
}

// validateFeed validates the basic structure of an RSS feed.
func (p *Parser) validateFeed(feed *Feed) error {
	if feed.Channel.Title == "" {
		return core.NewValidationError("channel_title", "", "RSS feed must have a channel title")
	}

	if len(feed.Channel.Items) == 0 {
		return core.NewValidationError("items", 0, "RSS feed must contain at least one item")
	}

	return nil
}

// validateNewsItem validates a converted news item.
func (p *Parser) validateNewsItem(item *models.News) error {
	if item.Title == "" {
		return core.NewValidationError("title", "", "news item must have a title")
	}

	if item.Content == "" && item.Summary == "" {
		return core.NewValidationError("content", "", "news item must have content or summary")
	}

	return nil
}

// isImageType checks if a MIME type represents an image.
func (p *Parser) isImageType(mimeType string) bool {
	if mimeType == "" {
		return false
	}

	return strings.HasPrefix(strings.ToLower(mimeType), "image/")
}

// removeDuplicateStrings removes duplicate strings from a slice.
func (p *Parser) removeDuplicateStrings(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}

// Validate validates the parser configuration.
func (p *Parser) Validate(data *Feed) error {
	if data == nil {
		return fmt.Errorf("feed data cannot be nil")
	}

	return p.validateFeed(data)
}

// GetFeedMetadata extracts metadata from an RSS feed.
func (p *Parser) GetFeedMetadata(feed *Feed) *FeedMetadata {
	if feed == nil {
		return nil
	}

	metadata := &FeedMetadata{
		Title:       feed.Channel.Title,
		Description: feed.Channel.Description,
		Link:        feed.Channel.Link,
		Language:    feed.Channel.Language,
		Copyright:   feed.Channel.Copyright,
		Editor:      feed.Channel.ManagingEditor,
		WebMaster:   feed.Channel.WebMaster,
		Generator:   feed.Channel.Generator,
		UpdateFreq:  feed.Channel.TTL,
		ItemCount:   len(feed.Channel.Items),
	}

	// Parse dates
	if feed.Channel.PubDate != "" {
		metadata.PublishedAt = p.parseDate(feed.Channel.PubDate)
	}

	if feed.Channel.LastBuildDate != "" {
		metadata.LastBuiltAt = p.parseDate(feed.Channel.LastBuildDate)
	}

	// Extract image URL
	if feed.Channel.Image != nil && feed.Channel.Image.URL != "" {
		metadata.ImageURL = feed.Channel.Image.URL
	}

	return metadata
}
