package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"news-aggregator/internal/models"
	"news-aggregator/internal/repository"

	"github.com/rs/zerolog"
)

// ScoringService handles comprehensive article scoring
type ScoringService struct {
	newsRepo     *repository.NewsRepository
	scoringRepo  *repository.ScoringRepository
	logger       zerolog.Logger
	config       models.TopStoriesConfig
	nlpClient    NLPClient
	socialClient SocialMetricsClient
}

// NLPClient interface for content analysis
type NLPClient interface {
	AnalyzeContent(ctx context.Context, title, content string) (*models.ContentAnalysis, error)
	ExtractKeywords(ctx context.Context, text string) ([]string, error)
	ClassifyTopic(ctx context.Context, text string) (string, error)
	CalculateImportance(ctx context.Context, title, content string) (float64, error)
}

// SocialMetricsClient interface for social media data
type SocialMetricsClient interface {
	GetSocialMetrics(ctx context.Context, url string) (*models.SocialMetrics, error)
	GetTwitterShares(ctx context.Context, url string) (int64, error)
	GetFacebookShares(ctx context.Context, url string) (int64, error)
	GetRedditScore(ctx context.Context, url string) (int64, error)
}

// NewScoringService creates a new scoring service
func NewScoringService(
	newsRepo *repository.NewsRepository,
	scoringRepo *repository.ScoringRepository,
	logger zerolog.Logger,
	config models.TopStoriesConfig,
	nlpClient NLPClient,
	socialClient SocialMetricsClient,
) *ScoringService {
	return &ScoringService{
		newsRepo:     newsRepo,
		scoringRepo:  scoringRepo,
		logger:       logger.With().Str("service", "scoring").Logger(),
		config:       config,
		nlpClient:    nlpClient,
		socialClient: socialClient,
	}
}

// CalculateTopStories returns top stories using enhanced algorithm
func (s *ScoringService) CalculateTopStories(ctx context.Context, limit int) ([]models.News, error) {
	s.logger.Info().Int("limit", limit).Msg("Calculating top stories with enhanced algorithm")

	// Get recent articles within max age
	articles, err := s.newsRepo.GetRecentArticles(ctx, s.config.MaxAge)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent articles: %w", err)
	}

	// Calculate scores for all articles
	scoredArticles, err := s.calculateArticleScores(ctx, articles)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate scores: %w", err)
	}

	// Apply category balancing
	balancedArticles := s.applyCategoryBalancing(scoredArticles, limit)

	// Sort by final score
	sort.Slice(balancedArticles, func(i, j int) bool {
		return balancedArticles[i].Score > balancedArticles[j].Score
	})

	// Extract news articles
	result := make([]models.News, 0, limit)
	for i, scoredArticle := range balancedArticles {
		if i >= limit {
			break
		}
		result = append(result, scoredArticle.Article)
	}

	s.logger.Info().
		Int("total_articles", len(articles)).
		Int("scored_articles", len(scoredArticles)).
		Int("final_count", len(result)).
		Msg("Top stories calculation completed")

	return result, nil
}

// ScoredArticle combines an article with its score
type ScoredArticle struct {
	Article models.News
	Score   float64
	Scores  models.ArticleScore
}

// calculateArticleScores calculates comprehensive scores for articles
func (s *ScoringService) calculateArticleScores(ctx context.Context, articles []models.News) ([]ScoredArticle, error) {
	var scoredArticles []ScoredArticle

	for _, article := range articles {
		score, err := s.calculateSingleArticleScore(ctx, article)
		if err != nil {
			s.logger.Warn().
				Str("article_id", article.ID).
				Err(err).
				Msg("Failed to calculate score for article, skipping")
			continue
		}

		if score.FinalScore >= s.config.MinScore {
			scoredArticles = append(scoredArticles, ScoredArticle{
				Article: article,
				Score:   score.FinalScore,
				Scores:  *score,
			})
		}
	}

	return scoredArticles, nil
}

// calculateSingleArticleScore calculates score for a single article
func (s *ScoringService) calculateSingleArticleScore(ctx context.Context, article models.News) (*models.ArticleScore, error) {
	// Get or calculate engagement score
	engagementScore, err := s.calculateEngagementScore(ctx, article.ID)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to get engagement score, using default")
		engagementScore = 0.5 // Default neutral score
	}

	// Get or calculate credibility score
	credibilityScore, err := s.calculateCredibilityScore(ctx, article.Source)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to get credibility score, using default")
		credibilityScore = 0.7 // Default moderate credibility
	}

	// Get or calculate content score
	contentScore, err := s.calculateContentScore(ctx, article)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to get content score, using default")
		contentScore = 0.6 // Default moderate content score
	}

	// Get or calculate social score
	socialScore, err := s.calculateSocialScore(ctx, article.URL)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to get social score, using default")
		socialScore = 0.3 // Default low social score
	}

	// Calculate recency score
	recencyScore := s.calculateRecencyScore(article.PublishedAt)

	// Calculate weighted final score
	finalScore := s.calculateWeightedScore(
		engagementScore,
		credibilityScore,
		contentScore,
		socialScore,
		recencyScore,
	)

	return &models.ArticleScore{
		ArticleID:        article.ID,
		EngagementScore:  engagementScore,
		CredibilityScore: credibilityScore,
		ContentScore:     contentScore,
		SocialScore:      socialScore,
		FinalScore:       finalScore,
		LastUpdated:      time.Now(),
	}, nil
}

// calculateEngagementScore calculates engagement-based score
func (s *ScoringService) calculateEngagementScore(ctx context.Context, articleID string) (float64, error) {
	metrics, err := s.scoringRepo.GetEngagementMetrics(ctx, articleID)
	if err != nil {
		return 0.5, err // Return neutral score on error
	}

	// Normalize metrics (using logarithmic scaling to handle wide ranges)
	viewScore := math.Log10(float64(metrics.ViewCount+1)) / 6.0            // Normalize to ~0-1
	clickScore := math.Log10(float64(metrics.ClickCount+1)) / 5.0          // Normalize to ~0-1
	shareScore := math.Log10(float64(metrics.ShareCount+1)) / 4.0          // Normalize to ~0-1
	readTimeScore := math.Min(float64(metrics.AverageReadTime)/300.0, 1.0) // 5 minutes = 1.0
	bounceScore := 1.0 - metrics.BounceRate                                // Invert bounce rate

	// Weighted combination
	engagementScore := (viewScore*0.2 + clickScore*0.3 + shareScore*0.2 + readTimeScore*0.2 + bounceScore*0.1)

	return math.Min(engagementScore, 1.0), nil
}

// calculateCredibilityScore gets source credibility score
func (s *ScoringService) calculateCredibilityScore(ctx context.Context, sourceName string) (float64, error) {
	credibility, err := s.scoringRepo.GetSourceCredibility(ctx, sourceName)
	if err != nil {
		return s.getDefaultCredibilityScore(sourceName), err
	}

	// Combine different credibility factors
	score := (credibility.CredibilityScore*0.4 +
		credibility.ReliabilityScore*0.3 +
		credibility.FactualScore*0.3)

	return score, nil
}

// calculateContentScore analyzes content importance using NLP
func (s *ScoringService) calculateContentScore(ctx context.Context, article models.News) (float64, error) {
	// Check if analysis already exists
	analysis, err := s.scoringRepo.GetContentAnalysis(ctx, article.ID)
	if err == nil && time.Since(analysis.ProcessedAt) < 24*time.Hour {
		return analysis.ImportanceScore, nil
	}

	// Perform new analysis
	analysis, err = s.nlpClient.AnalyzeContent(ctx, article.Title, article.Content)
	if err != nil {
		return s.calculateBasicContentScore(article), err
	}

	// Store analysis results
	analysis.ArticleID = article.ID
	if err := s.scoringRepo.SaveContentAnalysis(ctx, analysis); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to save content analysis")
	}

	return analysis.ImportanceScore, nil
}

// calculateSocialScore gets social media engagement score
func (s *ScoringService) calculateSocialScore(ctx context.Context, url string) (float64, error) {
	// Check if metrics already exist and are recent
	metrics, err := s.scoringRepo.GetSocialMetrics(ctx, url)
	if err == nil && time.Since(metrics.LastFetched) < 6*time.Hour {
		return s.normalizeSocialScore(metrics), nil
	}

	// Fetch new social metrics
	metrics, err = s.socialClient.GetSocialMetrics(ctx, url)
	if err != nil {
		return 0.3, err // Default low social score
	}

	// Store metrics
	if err := s.scoringRepo.SaveSocialMetrics(ctx, metrics); err != nil {
		s.logger.Warn().Err(err).Msg("Failed to save social metrics")
	}

	return s.normalizeSocialScore(metrics), nil
}

// calculateRecencyScore calculates score based on article age
func (s *ScoringService) calculateRecencyScore(publishedAt time.Time) float64 {
	age := time.Since(publishedAt)
	maxAge := s.config.MaxAge

	if age >= maxAge {
		return 0.0
	}

	// Exponential decay: newer articles get higher scores
	decayRate := 0.1 // Adjust for faster/slower decay
	normalizedAge := age.Seconds() / maxAge.Seconds()
	score := math.Exp(-decayRate * normalizedAge)

	return score
}

// calculateWeightedScore combines all scores with configured weights
func (s *ScoringService) calculateWeightedScore(engagement, credibility, content, social, recency float64) float64 {
	weights := s.config.ScoringWeights

	score := (engagement*weights.EngagementWeight +
		credibility*weights.CredibilityWeight +
		content*weights.ContentWeight +
		social*weights.SocialWeight +
		recency*weights.RecencyWeight)

	// Normalize to 0-1 range
	totalWeight := weights.EngagementWeight + weights.CredibilityWeight +
		weights.ContentWeight + weights.SocialWeight + weights.RecencyWeight

	return score / totalWeight
}

// applyCategoryBalancing ensures diverse categories in top stories
func (s *ScoringService) applyCategoryBalancing(articles []ScoredArticle, limit int) []ScoredArticle {
	if len(articles) <= limit {
		return articles
	}

	balance := s.config.CategoryBalance
	categoryCount := make(map[string]int)
	var result []ScoredArticle

	// Sort by score first
	sort.Slice(articles, func(i, j int) bool {
		return articles[i].Score > articles[j].Score
	})

	// Apply balancing rules
	for _, article := range articles {
		if len(result) >= limit {
			break
		}

		category := s.normalizeCategory(article.Article.Category)

		// Check category limits
		if categoryCount[category] >= balance.MaxPerCategory {
			continue
		}

		// Add article and update count
		result = append(result, article)
		categoryCount[category]++
	}

	// Ensure minimum category diversity
	if len(s.getUniqueCategories(result)) < balance.MinCategories {
		result = s.enforceMinimumCategoryDiversity(articles, result, limit)
	}

	return result
}

// Helper methods

func (s *ScoringService) getDefaultCredibilityScore(sourceName string) float64 {
	// Default credibility scores for known sources
	defaultScores := map[string]float64{
		"BBC News":         0.9,
		"Reuters":          0.9,
		"Associated Press": 0.9,
		"NPR":              0.85,
		"The Guardian":     0.8,
		"CNN":              0.75,
		"TechCrunch":       0.7,
		"NDTV":             0.75,
		"Times of India":   0.7,
		"The Hindu":        0.8,
	}

	if score, exists := defaultScores[sourceName]; exists {
		return score
	}
	return 0.6 // Default moderate credibility
}

func (s *ScoringService) calculateBasicContentScore(article models.News) float64 {
	score := 0.5 // Base score

	// Title length (optimal around 60 characters)
	titleLen := len(article.Title)
	if titleLen >= 30 && titleLen <= 80 {
		score += 0.1
	}

	// Content length
	contentLen := len(article.Content)
	if contentLen >= 500 && contentLen <= 5000 {
		score += 0.2
	}

	// Has image
	if article.ImageURL != "" {
		score += 0.1
	}

	// Has summary
	if article.Summary != "" {
		score += 0.1
	}

	return math.Min(score, 1.0)
}

func (s *ScoringService) normalizeSocialScore(metrics *models.SocialMetrics) float64 {
	// Logarithmic scaling for social metrics
	twitterScore := math.Log10(float64(metrics.TwitterShares+1)) / 4.0
	facebookScore := math.Log10(float64(metrics.FacebookShares+1)) / 4.0
	linkedinScore := math.Log10(float64(metrics.LinkedInShares+1)) / 3.0
	redditScore := math.Log10(float64(metrics.RedditScore+1)) / 3.0

	totalScore := (twitterScore + facebookScore + linkedinScore + redditScore) / 4.0
	return math.Min(totalScore, 1.0)
}

func (s *ScoringService) normalizeCategory(category string) string {
	category = strings.ToLower(strings.TrimSpace(category))

	// Normalize similar categories
	categoryMap := map[string]string{
		"tech":          "technology",
		"business":      "business",
		"sports":        "sports",
		"politics":      "politics",
		"health":        "health",
		"science":       "science",
		"entertainment": "entertainment",
		"world":         "world",
	}

	if normalized, exists := categoryMap[category]; exists {
		return normalized
	}
	return "general"
}

func (s *ScoringService) getUniqueCategories(articles []ScoredArticle) []string {
	categories := make(map[string]bool)
	for _, article := range articles {
		categories[s.normalizeCategory(article.Article.Category)] = true
	}

	var result []string
	for category := range categories {
		result = append(result, category)
	}
	return result
}

func (s *ScoringService) enforceMinimumCategoryDiversity(
	allArticles, currentResult []ScoredArticle, limit int) []ScoredArticle {

	// Implementation for ensuring minimum category diversity
	// This is a simplified version - could be more sophisticated
	currentCategories := make(map[string]bool)
	for _, article := range currentResult {
		currentCategories[s.normalizeCategory(article.Article.Category)] = true
	}

	// Find articles from missing categories
	for _, article := range allArticles {
		if len(currentResult) >= limit {
			break
		}

		category := s.normalizeCategory(article.Article.Category)
		if !currentCategories[category] {
			currentResult = append(currentResult, article)
			currentCategories[category] = true
		}
	}

	return currentResult
}

// TrackEngagement records user engagement with an article
func (s *ScoringService) TrackEngagement(ctx context.Context, articleID string, engagementType string, value int64) error {
	return s.scoringRepo.UpdateEngagementMetrics(ctx, articleID, engagementType, value)
}

// RefreshScores recalculates scores for all recent articles
func (s *ScoringService) RefreshScores(ctx context.Context) error {
	s.logger.Info().Msg("Starting score refresh for all articles")

	articles, err := s.newsRepo.GetRecentArticles(ctx, s.config.MaxAge)
	if err != nil {
		return fmt.Errorf("failed to get recent articles: %w", err)
	}

	for _, article := range articles {
		score, err := s.calculateSingleArticleScore(ctx, article)
		if err != nil {
			s.logger.Warn().Str("article_id", article.ID).Err(err).Msg("Failed to calculate score")
			continue
		}

		if err := s.scoringRepo.SaveArticleScore(ctx, score); err != nil {
			s.logger.Warn().Str("article_id", article.ID).Err(err).Msg("Failed to save score")
		}
	}

	s.logger.Info().Int("articles_processed", len(articles)).Msg("Score refresh completed")
	return nil
}
