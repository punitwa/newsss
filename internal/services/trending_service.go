package services

import (
	"context"
	"sort"
	"strings"
	"time"

	"news-aggregator/internal/models"
	"news-aggregator/internal/repository"

	"github.com/rs/zerolog"
)

type TrendingTopic struct {
	Name           string    `json:"name"`
	ArticleCount   int       `json:"article_count"`
	TodayChange    int       `json:"today_change"`
	TrendDirection string    `json:"trend_direction"` // "up", "down", "stable"
	Percentage     float64   `json:"percentage"`
	Category       string    `json:"category,omitempty"`
	LastUpdated    time.Time `json:"last_updated"`
}

type TrendingService struct {
	newsRepo *repository.NewsRepository
	logger   zerolog.Logger
}

func NewTrendingService(newsRepo *repository.NewsRepository, logger zerolog.Logger) *TrendingService {
	return &TrendingService{
		newsRepo: newsRepo,
		logger:   logger.With().Str("service", "trending").Logger(),
	}
}

// GetTrendingTopics returns the top trending topics based on article tags and keywords
func (ts *TrendingService) GetTrendingTopics(ctx context.Context, limit int) ([]TrendingTopic, error) {
	ts.logger.Debug().Int("limit", limit).Msg("Getting trending topics")

	// Get recent articles (last 7 days for trend analysis)
	articles, err := ts.newsRepo.GetRecentArticles(ctx, 7*24*time.Hour)
	if err != nil {
		ts.logger.Error().Err(err).Msg("Failed to get recent articles")
		return nil, err
	}

	// Get articles from yesterday for comparison
	yesterdayArticles, err := ts.newsRepo.GetArticlesByDateRange(ctx, 
		time.Now().Add(-48*time.Hour), 
		time.Now().Add(-24*time.Hour))
	if err != nil {
		ts.logger.Warn().Err(err).Msg("Failed to get yesterday's articles for comparison")
		yesterdayArticles = []models.News{} // Continue without comparison
	}

	// Extract and count topics
	topicCounts := ts.extractTopics(articles)
	yesterdayTopicCounts := ts.extractTopics(yesterdayArticles)

	// Calculate trending topics
	trendingTopics := ts.calculateTrending(topicCounts, yesterdayTopicCounts, len(articles))

	// Sort by article count (descending)
	sort.Slice(trendingTopics, func(i, j int) bool {
		return trendingTopics[i].ArticleCount > trendingTopics[j].ArticleCount
	})

	// Limit results
	if len(trendingTopics) > limit {
		trendingTopics = trendingTopics[:limit]
	}

	ts.logger.Info().Int("topics_count", len(trendingTopics)).Msg("Generated trending topics")
	return trendingTopics, nil
}

// extractTopics extracts topics from article tags and titles
func (ts *TrendingService) extractTopics(articles []models.News) map[string]int {
	topicCounts := make(map[string]int)

	for _, article := range articles {
		// Extract from tags
		for _, tag := range article.Tags {
			if tag != "" && len(tag) > 2 {
				normalizedTag := ts.normalizeTag(tag)
				if normalizedTag != "" {
					topicCounts[normalizedTag]++
				}
			}
		}

		// Extract from title keywords
		titleKeywords := ts.extractKeywordsFromTitle(article.Title)
		for _, keyword := range titleKeywords {
			normalizedKeyword := ts.normalizeTag(keyword)
			if normalizedKeyword != "" {
				topicCounts[normalizedKeyword]++
			}
		}

		// Extract from category
		if article.Category != "" {
			normalizedCategory := ts.normalizeTag(article.Category)
			if normalizedCategory != "" {
				topicCounts[normalizedCategory]++
			}
		}
	}

	return topicCounts
}

// extractKeywordsFromTitle extracts important keywords from article titles
func (ts *TrendingService) extractKeywordsFromTitle(title string) []string {
	// Common important keywords that often indicate trending topics
	importantKeywords := []string{
		"AI", "Artificial Intelligence", "Machine Learning", "ChatGPT", "OpenAI",
		"Climate Change", "Global Warming", "Carbon", "Renewable",
		"Space", "NASA", "SpaceX", "Mars", "Satellite",
		"Cryptocurrency", "Bitcoin", "Ethereum", "Blockchain", "Crypto",
		"Healthcare", "Medicine", "Vaccine", "Health", "Medical",
		"Electric Vehicle", "EV", "Tesla", "Battery", "Autonomous",
		"Quantum", "5G", "Internet", "Cybersecurity", "Privacy",
		"Metaverse", "VR", "AR", "Virtual Reality", "Augmented Reality",
		"Economy", "Market", "Stock", "Investment", "Finance",
		"Politics", "Election", "Government", "Policy", "Law",
		"Technology", "Tech", "Innovation", "Startup", "Software",
	}

	var keywords []string
	titleLower := strings.ToLower(title)

	for _, keyword := range importantKeywords {
		if strings.Contains(titleLower, strings.ToLower(keyword)) {
			keywords = append(keywords, keyword)
		}
	}

	return keywords
}

// normalizeTag normalizes tags for consistency
func (ts *TrendingService) normalizeTag(tag string) string {
	tag = strings.TrimSpace(tag)
	tag = strings.ToLower(tag)

	// Skip very short or common words
	if len(tag) < 3 {
		return ""
	}

	// Skip common stop words
	stopWords := []string{"the", "and", "for", "are", "but", "not", "you", "all", "can", "had", "her", "was", "one", "our", "out", "day", "get", "has", "him", "his", "how", "man", "new", "now", "old", "see", "two", "way", "who", "boy", "did", "its", "let", "put", "say", "she", "too", "use"}
	for _, stopWord := range stopWords {
		if tag == stopWord {
			return ""
		}
	}

	// Normalize common variations
	normalizations := map[string]string{
		"artificial intelligence": "AI",
		"machine learning":       "AI",
		"chatgpt":               "AI",
		"openai":                "AI",
		"climate change":        "Climate Change",
		"global warming":        "Climate Change",
		"renewable energy":      "Climate Change",
		"electric vehicle":      "Electric Vehicles",
		"electric vehicles":     "Electric Vehicles",
		"ev":                   "Electric Vehicles",
		"tesla":                "Electric Vehicles",
		"autonomous vehicle":    "Electric Vehicles",
		"cryptocurrency":       "Cryptocurrency",
		"bitcoin":              "Cryptocurrency",
		"ethereum":             "Cryptocurrency",
		"blockchain":           "Cryptocurrency",
		"crypto":               "Cryptocurrency",
		"space exploration":    "Space Exploration",
		"nasa":                 "Space Exploration",
		"spacex":               "Space Exploration",
		"mars":                 "Space Exploration",
		"satellite":            "Space Exploration",
		"healthcare":           "Healthcare",
		"medicine":             "Healthcare",
		"medical":              "Healthcare",
		"health":               "Healthcare",
		"vaccine":              "Healthcare",
		"quantum computing":    "Quantum Computing",
		"quantum":              "Quantum Computing",
		"5g":                   "5G Technology",
		"cybersecurity":        "Cybersecurity",
		"privacy":              "Cybersecurity",
		"metaverse":            "Metaverse",
		"virtual reality":      "Metaverse",
		"vr":                   "Metaverse",
		"augmented reality":    "Metaverse",
		"ar":                   "Metaverse",
		"technology":           "Technology",
		"tech":                 "Technology",
		"innovation":           "Technology",
		"startup":              "Technology",
		"software":             "Technology",
	}

	if normalized, exists := normalizations[tag]; exists {
		return normalized
	}

	// Capitalize first letter
	if len(tag) > 0 {
		return strings.ToUpper(string(tag[0])) + tag[1:]
	}

	return tag
}

// calculateTrending calculates trending metrics for topics
func (ts *TrendingService) calculateTrending(currentCounts, yesterdayCounts map[string]int, totalArticles int) []TrendingTopic {
	var topics []TrendingTopic

	for topic, count := range currentCounts {
		// Skip topics with very few articles
		if count < 3 {
			continue
		}

		yesterdayCount := yesterdayCounts[topic]
		todayChange := count - yesterdayCount

		// Determine trend direction
		var trendDirection string
		if todayChange > 0 {
			trendDirection = "up"
		} else if todayChange < 0 {
			trendDirection = "down"
		} else {
			trendDirection = "stable"
		}

		// Calculate percentage relative to total articles
		percentage := float64(count) / float64(totalArticles) * 100

		// Determine category based on topic
		category := ts.categorizeTopics(topic)

		topics = append(topics, TrendingTopic{
			Name:           topic,
			ArticleCount:   count,
			TodayChange:    todayChange,
			TrendDirection: trendDirection,
			Percentage:     percentage,
			Category:       category,
			LastUpdated:    time.Now(),
		})
	}

	return topics
}

// categorizeTopics assigns categories to topics
func (ts *TrendingService) categorizeTopics(topic string) string {
	topicLower := strings.ToLower(topic)

	categories := map[string][]string{
		"Technology": {"ai", "technology", "tech", "software", "quantum", "5g", "cybersecurity", "metaverse"},
		"Science":    {"space", "climate", "healthcare", "medicine", "research"},
		"Finance":    {"cryptocurrency", "bitcoin", "market", "economy", "investment"},
		"Transport":  {"electric vehicles", "autonomous", "tesla"},
		"Politics":   {"politics", "election", "government", "policy", "law"},
		"Business":   {"startup", "business", "company", "innovation"},
	}

	for category, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(topicLower, keyword) {
				return category
			}
		}
	}

	return "General"
}
