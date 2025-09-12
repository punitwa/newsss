package processor

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"

	"news-aggregator/internal/models"
	"news-aggregator/internal/services"

	"github.com/rs/zerolog"
)

// Deduplicator handles duplicate detection for news articles
type Deduplicator struct {
	newsService *services.NewsService
	logger      zerolog.Logger
}

func NewDeduplicator(newsService *services.NewsService, logger zerolog.Logger) *Deduplicator {
	return &Deduplicator{
		newsService: newsService,
		logger:      logger.With().Str("component", "deduplicator").Logger(),
	}
}

// IsDuplicate checks if a news article is a duplicate
func (d *Deduplicator) IsDuplicate(ctx context.Context, news *models.News) (bool, error) {
	d.logger.Debug().Str("title", news.Title).Str("hash", news.Hash).Msg("Checking for duplicate")

	// Method 1: Check by content hash
	if news.Hash != "" {
		exists, err := d.newsService.CheckDuplicate(ctx, news.Hash)
		if err != nil {
			d.logger.Error().Err(err).Str("hash", news.Hash).Msg("Failed to check duplicate by hash")
			// Continue with other methods if hash check fails
		} else if exists {
			d.logger.Info().Str("hash", news.Hash).Msg("Duplicate found by content hash")
			return true, nil
		}
	}

	// Method 2: Check by URL
	if news.URL != "" {
		// Generate hash from URL for comparison
		urlHash := d.generateHash(news.URL)
		exists, err := d.newsService.CheckDuplicate(ctx, urlHash)
		if err != nil {
			d.logger.Error().Err(err).Str("url", news.URL).Msg("Failed to check duplicate by URL")
		} else if exists {
			d.logger.Info().Str("url", news.URL).Msg("Duplicate found by URL")
			return true, nil
		}
	}

	// Method 3: Check by title similarity
	if news.Title != "" {
		isDuplicate, err := d.checkTitleSimilarity(ctx, news.Title)
		if err != nil {
			d.logger.Error().Err(err).Str("title", news.Title).Msg("Failed to check title similarity")
		} else if isDuplicate {
			d.logger.Info().Str("title", news.Title).Msg("Duplicate found by title similarity")
			return true, nil
		}
	}

	// Method 4: Check by content similarity
	if news.Content != "" {
		isDuplicate, err := d.checkContentSimilarity(ctx, news.Content)
		if err != nil {
			d.logger.Error().Err(err).Msg("Failed to check content similarity")
		} else if isDuplicate {
			d.logger.Info().Str("title", news.Title).Msg("Duplicate found by content similarity")
			return true, nil
		}
	}

	// Ensure the news has a content hash for future duplicate checks
	if news.Hash == "" {
		news.Hash = d.generateContentHash(news)
	}

	d.logger.Debug().Str("title", news.Title).Msg("No duplicate found")
	return false, nil
}

// generateHash generates a hash from input string
func (d *Deduplicator) generateHash(input string) string {
	hash := md5.Sum([]byte(input))
	return fmt.Sprintf("%x", hash)
}

// generateContentHash generates a hash from news content for deduplication
func (d *Deduplicator) generateContentHash(news *models.News) string {
	// Combine title, content, and URL for a comprehensive hash
	content := strings.ToLower(strings.TrimSpace(news.Title)) + 
		strings.ToLower(strings.TrimSpace(news.Content)) + 
		strings.ToLower(strings.TrimSpace(news.URL))
	
	// Remove extra whitespace
	content = strings.Join(strings.Fields(content), " ")
	
	return d.generateHash(content)
}

// checkTitleSimilarity checks if a similar title already exists
func (d *Deduplicator) checkTitleSimilarity(ctx context.Context, title string) (bool, error) {
	// Normalize title for comparison
	normalizedTitle := d.normalizeTitle(title)
	
	// For now, we'll use a simple approach
	// In production, you might want to use more sophisticated similarity algorithms
	// like Levenshtein distance, Jaccard similarity, or semantic similarity
	
	// Check if a very similar title exists (exact match after normalization)
	titleHash := d.generateHash(normalizedTitle)
	exists, err := d.newsService.CheckDuplicate(ctx, titleHash)
	if err != nil {
		return false, fmt.Errorf("failed to check title hash: %w", err)
	}
	
	return exists, nil
}

// checkContentSimilarity checks if similar content already exists
func (d *Deduplicator) checkContentSimilarity(ctx context.Context, content string) (bool, error) {
	// Normalize content for comparison
	normalizedContent := d.normalizeContent(content)
	
	// Generate hash from normalized content
	contentHash := d.generateHash(normalizedContent)
	exists, err := d.newsService.CheckDuplicate(ctx, contentHash)
	if err != nil {
		return false, fmt.Errorf("failed to check content hash: %w", err)
	}
	
	return exists, nil
}

// normalizeTitle normalizes title for comparison
func (d *Deduplicator) normalizeTitle(title string) string {
	// Convert to lowercase
	normalized := strings.ToLower(title)
	
	// Remove common prefixes and suffixes
	prefixes := []string{
		"breaking:", "urgent:", "update:", "exclusive:", "news:",
		"report:", "analysis:", "opinion:", "editorial:",
	}
	
	for _, prefix := range prefixes {
		if strings.HasPrefix(normalized, prefix) {
			normalized = strings.TrimPrefix(normalized, prefix)
			normalized = strings.TrimSpace(normalized)
			break
		}
	}
	
	// Remove common suffixes
	suffixes := []string{
		"- cnn", "- bbc", "- reuters", "- ap", "- bloomberg",
		"| reuters", "| cnn", "| bbc", "| bloomberg",
	}
	
	for _, suffix := range suffixes {
		if strings.HasSuffix(normalized, suffix) {
			normalized = strings.TrimSuffix(normalized, suffix)
			normalized = strings.TrimSpace(normalized)
			break
		}
	}
	
	// Remove extra whitespace
	normalized = strings.Join(strings.Fields(normalized), " ")
	
	return normalized
}

// normalizeContent normalizes content for comparison
func (d *Deduplicator) normalizeContent(content string) string {
	// Convert to lowercase
	normalized := strings.ToLower(content)
	
	// Remove URLs
	// Simple regex to remove URLs
	urlPattern := `https?://[^\s]+`
	normalized = strings.ReplaceAll(normalized, urlPattern, "")
	
	// Remove email addresses
	emailPattern := `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`
	normalized = strings.ReplaceAll(normalized, emailPattern, "")
	
	// Remove extra whitespace and normalize
	normalized = strings.Join(strings.Fields(normalized), " ")
	
	// Take first 1000 characters for comparison (to avoid very long content)
	if len(normalized) > 1000 {
		normalized = normalized[:1000]
	}
	
	return normalized
}

// SimilarityThreshold represents different similarity thresholds
type SimilarityThreshold struct {
	Title   float64 // Threshold for title similarity (0.0 to 1.0)
	Content float64 // Threshold for content similarity (0.0 to 1.0)
}

// DefaultSimilarityThreshold returns default similarity thresholds
func DefaultSimilarityThreshold() SimilarityThreshold {
	return SimilarityThreshold{
		Title:   0.85, // 85% similarity for titles
		Content: 0.80, // 80% similarity for content
	}
}

// Advanced similarity methods (for future implementation)

// calculateJaccardSimilarity calculates Jaccard similarity between two strings
func (d *Deduplicator) calculateJaccardSimilarity(str1, str2 string) float64 {
	words1 := strings.Fields(strings.ToLower(str1))
	words2 := strings.Fields(strings.ToLower(str2))
	
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)
	
	for _, word := range words1 {
		set1[word] = true
	}
	
	for _, word := range words2 {
		set2[word] = true
	}
	
	intersection := 0
	for word := range set1 {
		if set2[word] {
			intersection++
		}
	}
	
	union := len(set1) + len(set2) - intersection
	if union == 0 {
		return 0.0
	}
	
	return float64(intersection) / float64(union)
}

// calculateLevenshteinDistance calculates Levenshtein distance between two strings
func (d *Deduplicator) calculateLevenshteinDistance(str1, str2 string) int {
	len1, len2 := len(str1), len(str2)
	matrix := make([][]int, len1+1)
	
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
		matrix[i][0] = i
	}
	
	for j := 1; j <= len2; j++ {
		matrix[0][j] = j
	}
	
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			cost := 0
			if str1[i-1] != str2[j-1] {
				cost = 1
			}
			
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}
	
	return matrix[len1][len2]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}
