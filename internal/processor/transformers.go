package processor

import (
	"context"
	"regexp"
	"strings"
	"time"

	"news-aggregator/internal/models"

	"github.com/rs/zerolog"
)

// Transformer interface for news processing transformations
type Transformer interface {
	Transform(ctx context.Context, news *models.News) (*models.News, error)
	GetName() string
}

// ContentCleanerTransformer cleans and normalizes news content
type ContentCleanerTransformer struct {
	logger zerolog.Logger
	htmlRegex *regexp.Regexp
	urlRegex  *regexp.Regexp
}

func NewContentCleanerTransformer(logger zerolog.Logger) *ContentCleanerTransformer {
	return &ContentCleanerTransformer{
		logger:    logger.With().Str("transformer", "content_cleaner").Logger(),
		htmlRegex: regexp.MustCompile(`<[^>]*>`),
		urlRegex:  regexp.MustCompile(`https?://[^\s]+`),
	}
}

func (c *ContentCleanerTransformer) GetName() string {
	return "content_cleaner"
}

func (c *ContentCleanerTransformer) Transform(ctx context.Context, news *models.News) (*models.News, error) {
	c.logger.Debug().Str("title", news.Title).Msg("Cleaning content")

	cleaned := *news

	// Clean title
	cleaned.Title = c.cleanText(news.Title)
	
	// Clean content
	cleaned.Content = c.cleanText(news.Content)
	
	// Clean summary
	cleaned.Summary = c.cleanText(news.Summary)
	
	// Generate summary if empty
	if cleaned.Summary == "" && cleaned.Content != "" {
		cleaned.Summary = c.generateSummary(cleaned.Content)
	}

	// Normalize author
	cleaned.Author = strings.TrimSpace(cleaned.Author)
	if cleaned.Author == "" {
		cleaned.Author = "Unknown"
	}

	cleaned.UpdatedAt = time.Now()

	return &cleaned, nil
}

func (c *ContentCleanerTransformer) cleanText(text string) string {
	// Remove HTML tags
	text = c.htmlRegex.ReplaceAllString(text, "")
	
	// Decode HTML entities
	text = c.decodeHTMLEntities(text)
	
	// Normalize whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	
	// Trim whitespace
	text = strings.TrimSpace(text)
	
	return text
}

func (c *ContentCleanerTransformer) decodeHTMLEntities(text string) string {
	entities := map[string]string{
		"&amp;":    "&",
		"&lt;":     "<",
		"&gt;":     ">",
		"&quot;":   "\"",
		"&apos;":   "'",
		"&nbsp;":   " ",
		"&hellip;": "...",
		"&mdash;":  "—",
		"&ndash;":  "–",
		"&rsquo;":  "'",
		"&lsquo;":  "'",
		"&rdquo;":  "\"",
		"&ldquo;":  "\"",
	}

	for entity, replacement := range entities {
		text = strings.ReplaceAll(text, entity, replacement)
	}

	return text
}

func (c *ContentCleanerTransformer) generateSummary(content string) string {
	words := strings.Fields(content)
	
	// Ensure summary has at least 80 words but not more than 120
	minWords := 80
	maxWords := 120
	
	if len(words) < minWords {
		// If content is too short, return what we have
		return content
	}
	
	var targetWords int
	if len(words) > maxWords {
		targetWords = maxWords
	} else {
		targetWords = len(words)
	}
	
	// Create summary with proper sentence ending
	summary := strings.Join(words[:targetWords], " ")
	
	// Try to end at a sentence boundary
	lastPeriod := strings.LastIndex(summary, ".")
	lastExclamation := strings.LastIndex(summary, "!")
	lastQuestion := strings.LastIndex(summary, "?")
	
	lastSentenceEnd := lastPeriod
	if lastExclamation > lastSentenceEnd {
		lastSentenceEnd = lastExclamation
	}
	if lastQuestion > lastSentenceEnd {
		lastSentenceEnd = lastQuestion
	}
	
	// If we found a sentence ending in the last 30 characters and it's past minimum, use it
	if lastSentenceEnd > len(summary)-30 && lastSentenceEnd > (minWords*5) { // ~5 chars per word
		return summary[:lastSentenceEnd+1]
	}
	
	return summary + "..."
}

// CategoryClassifierTransformer classifies news into categories
type CategoryClassifierTransformer struct {
	logger zerolog.Logger
	categoryKeywords map[string][]string
}

func NewCategoryClassifierTransformer(logger zerolog.Logger) *CategoryClassifierTransformer {
	categoryKeywords := map[string][]string{
		"technology": {
			"tech", "software", "ai", "artificial intelligence", "machine learning",
			"computer", "digital", "startup", "innovation", "app", "mobile",
			"internet", "cyber", "data", "algorithm", "programming", "coding",
		},
		"business": {
			"business", "economy", "finance", "market", "stock", "investment",
			"company", "corporate", "revenue", "profit", "earnings", "trade",
			"commerce", "industry", "economic", "financial", "banking",
		},
		"sports": {
			"sports", "football", "basketball", "baseball", "soccer", "olympics",
			"championship", "game", "match", "tournament", "athlete", "team",
			"player", "coach", "league", "score", "victory", "defeat",
		},
		"politics": {
			"politics", "government", "election", "president", "congress", "senate",
			"policy", "vote", "campaign", "politician", "democracy", "republican",
			"democrat", "parliament", "minister", "law", "legislation",
		},
		"health": {
			"health", "medical", "doctor", "hospital", "disease", "treatment",
			"medicine", "healthcare", "patient", "virus", "vaccine", "drug",
			"therapy", "clinical", "diagnosis", "symptoms", "pandemic",
		},
		"science": {
			"science", "research", "study", "discovery", "experiment", "scientist",
			"laboratory", "analysis", "theory", "hypothesis", "evidence",
			"biology", "chemistry", "physics", "astronomy", "climate",
		},
		"entertainment": {
			"entertainment", "movie", "film", "music", "celebrity", "hollywood",
			"tv", "show", "actor", "actress", "director", "album", "concert",
			"theater", "comedy", "drama", "streaming", "netflix",
		},
		"world": {
			"world", "international", "global", "country", "nation", "foreign",
			"diplomatic", "embassy", "border", "immigration", "refugee",
			"conflict", "war", "peace", "treaty", "alliance",
		},
	}

	return &CategoryClassifierTransformer{
		logger:           logger.With().Str("transformer", "category_classifier").Logger(),
		categoryKeywords: categoryKeywords,
	}
}

func (c *CategoryClassifierTransformer) GetName() string {
	return "category_classifier"
}

func (c *CategoryClassifierTransformer) Transform(ctx context.Context, news *models.News) (*models.News, error) {
	c.logger.Debug().Str("title", news.Title).Msg("Classifying category")

	classified := *news

	// If category is already set and not "general", keep it
	if classified.Category != "" && classified.Category != "general" {
		return &classified, nil
	}

	// Combine title and content for classification
	text := strings.ToLower(classified.Title + " " + classified.Content)

	// Find best matching category
	bestCategory := "general"
	maxScore := 0

	for category, keywords := range c.categoryKeywords {
		score := 0
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				score++
			}
		}

		if score > maxScore {
			maxScore = score
			bestCategory = category
		}
	}

	classified.Category = bestCategory
	classified.UpdatedAt = time.Now()

	c.logger.Debug().
		Str("title", news.Title).
		Str("category", bestCategory).
		Int("score", maxScore).
		Msg("Category classified")

	return &classified, nil
}

// SentimentAnalyzerTransformer analyzes sentiment and adds tags
type SentimentAnalyzerTransformer struct {
	logger zerolog.Logger
	positiveWords []string
	negativeWords []string
}

func NewSentimentAnalyzerTransformer(logger zerolog.Logger) *SentimentAnalyzerTransformer {
	positiveWords := []string{
		"good", "great", "excellent", "amazing", "wonderful", "fantastic",
		"positive", "success", "win", "victory", "achievement", "progress",
		"improvement", "growth", "innovation", "breakthrough", "celebrate",
		"happy", "joy", "optimistic", "hope", "benefit", "advantage",
	}

	negativeWords := []string{
		"bad", "terrible", "awful", "horrible", "negative", "fail", "failure",
		"loss", "defeat", "problem", "issue", "crisis", "disaster", "concern",
		"worry", "fear", "decline", "drop", "fall", "crash", "collapse",
		"sad", "angry", "disappointed", "frustrated", "concerned", "alarmed",
	}

	return &SentimentAnalyzerTransformer{
		logger:        logger.With().Str("transformer", "sentiment_analyzer").Logger(),
		positiveWords: positiveWords,
		negativeWords: negativeWords,
	}
}

func (s *SentimentAnalyzerTransformer) GetName() string {
	return "sentiment_analyzer"
}

func (s *SentimentAnalyzerTransformer) Transform(ctx context.Context, news *models.News) (*models.News, error) {
	s.logger.Debug().Str("title", news.Title).Msg("Analyzing sentiment")

	analyzed := *news

	// Combine title and content for analysis
	text := strings.ToLower(analyzed.Title + " " + analyzed.Content)

	// Count positive and negative words
	positiveScore := 0
	negativeScore := 0

	for _, word := range s.positiveWords {
		if strings.Contains(text, word) {
			positiveScore++
		}
	}

	for _, word := range s.negativeWords {
		if strings.Contains(text, word) {
			negativeScore++
		}
	}

	// Add sentiment tags
	if analyzed.Tags == nil {
		analyzed.Tags = make([]string, 0)
	}

	if positiveScore > negativeScore {
		analyzed.Tags = append(analyzed.Tags, "positive")
	} else if negativeScore > positiveScore {
		analyzed.Tags = append(analyzed.Tags, "negative")
	} else {
		analyzed.Tags = append(analyzed.Tags, "neutral")
	}

	// Add urgency tags based on keywords
	urgencyKeywords := []string{"breaking", "urgent", "alert", "emergency", "crisis"}
	for _, keyword := range urgencyKeywords {
		if strings.Contains(text, keyword) {
			analyzed.Tags = append(analyzed.Tags, "urgent")
			break
		}
	}

	analyzed.UpdatedAt = time.Now()

	s.logger.Debug().
		Str("title", news.Title).
		Int("positive_score", positiveScore).
		Int("negative_score", negativeScore).
		Interface("tags", analyzed.Tags).
		Msg("Sentiment analyzed")

	return &analyzed, nil
}

// ImageExtractorTransformer extracts and validates images
type ImageExtractorTransformer struct {
	logger zerolog.Logger
}

func NewImageExtractorTransformer(logger zerolog.Logger) *ImageExtractorTransformer {
	return &ImageExtractorTransformer{
		logger: logger.With().Str("transformer", "image_extractor").Logger(),
	}
}

func (i *ImageExtractorTransformer) GetName() string {
	return "image_extractor"
}

func (i *ImageExtractorTransformer) Transform(ctx context.Context, news *models.News) (*models.News, error) {
	i.logger.Debug().Str("title", news.Title).Msg("Extracting images")

	enhanced := *news

	// If image URL is already present, validate it
	if enhanced.ImageURL != "" {
		if i.isValidImageURL(enhanced.ImageURL) {
			return &enhanced, nil
		} else {
			// Invalid image URL, clear it
			enhanced.ImageURL = ""
		}
	}

	// Extract image URLs from multiple sources
	imageURL := i.extractImageFromContent(enhanced.Content)
	if imageURL == "" {
		imageURL = i.extractImageFromContent(enhanced.Summary)
	}
	if imageURL == "" {
		imageURL = i.extractImageFromContent(enhanced.Title)
	}
	
	if imageURL != "" {
		enhanced.ImageURL = imageURL
	}

	enhanced.UpdatedAt = time.Now()

	return &enhanced, nil
}

func (i *ImageExtractorTransformer) isValidImageURL(url string) bool {
	// Basic validation - check if URL has image extension
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg"}
	
	urlLower := strings.ToLower(url)
	for _, ext := range imageExtensions {
		if strings.Contains(urlLower, ext) {
			return true
		}
	}

	// Check for common image hosting patterns
	imageHosts := []string{"imgur.com", "flickr.com", "unsplash.com", "pixabay.com"}
	for _, host := range imageHosts {
		if strings.Contains(urlLower, host) {
			return true
		}
	}

	return false
}

func (i *ImageExtractorTransformer) extractImageFromContent(content string) string {
	// 1. Try to find HTML img tags first
	imgTagRegex := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["'][^>]*>`)
	imgMatches := imgTagRegex.FindStringSubmatch(content)
	
	if len(imgMatches) > 1 {
		imageURL := imgMatches[1]
		if i.isValidImageURL(imageURL) {
			return imageURL
		}
	}
	
	// 2. Try to find direct image URLs
	imgRegex := regexp.MustCompile(`https?://[^\s]+\.(jpg|jpeg|png|gif|webp|bmp)`)
	matches := imgRegex.FindStringSubmatch(content)
	
	if len(matches) > 0 {
		return matches[0]
	}
	
	// 3. Try to find images from known news media domains
	mediaRegex := regexp.MustCompile(`https?://[^\s]*(?:media\.cnn\.com|ichef\.bbci\.co\.uk|techcrunch\.com/wp-content)[^\s]*\.(jpg|jpeg|png|gif|webp)`)
	mediaMatches := mediaRegex.FindStringSubmatch(content)
	
	if len(mediaMatches) > 0 {
		return mediaMatches[0]
	}

	return ""
}
