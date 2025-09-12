package models

import (
	"time"
)

// ArticleScore represents the comprehensive scoring for an article
type ArticleScore struct {
	ID               string    `json:"id" db:"id"`
	ArticleID        string    `json:"article_id" db:"article_id"`
	EngagementScore  float64   `json:"engagement_score" db:"engagement_score"`
	CredibilityScore float64   `json:"credibility_score" db:"credibility_score"`
	ContentScore     float64   `json:"content_score" db:"content_score"`
	SocialScore      float64   `json:"social_score" db:"social_score"`
	FinalScore       float64   `json:"final_score" db:"final_score"`
	LastUpdated      time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// EngagementMetrics tracks user engagement with articles
type EngagementMetrics struct {
	ID              string    `json:"id" db:"id"`
	ArticleID       string    `json:"article_id" db:"article_id"`
	ViewCount       int64     `json:"view_count" db:"view_count"`
	ClickCount      int64     `json:"click_count" db:"click_count"`
	ShareCount      int64     `json:"share_count" db:"share_count"`
	AverageReadTime int64     `json:"average_read_time" db:"average_read_time"` // in seconds
	BounceRate      float64   `json:"bounce_rate" db:"bounce_rate"`
	LastUpdated     time.Time `json:"last_updated" db:"last_updated"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// SourceCredibility defines credibility scores for news sources
type SourceCredibility struct {
	ID               string    `json:"id" db:"id"`
	SourceName       string    `json:"source_name" db:"source_name"`
	CredibilityScore float64   `json:"credibility_score" db:"credibility_score"` // 0.0 to 1.0
	ReliabilityScore float64   `json:"reliability_score" db:"reliability_score"` // 0.0 to 1.0
	BiasScore        float64   `json:"bias_score" db:"bias_score"`               // -1.0 (left) to 1.0 (right), 0 neutral
	FactualScore     float64   `json:"factual_score" db:"factual_score"`         // 0.0 to 1.0
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// ContentAnalysis stores NLP analysis results
type ContentAnalysis struct {
	ID                  string            `json:"id" db:"id"`
	ArticleID           string            `json:"article_id" db:"article_id"`
	SentimentScore      float64           `json:"sentiment_score" db:"sentiment_score"`     // -1.0 to 1.0
	ImportanceScore     float64           `json:"importance_score" db:"importance_score"`   // 0.0 to 1.0
	ReadabilityScore    float64           `json:"readability_score" db:"readability_score"` // 0.0 to 1.0
	KeywordsExtracted   []string          `json:"keywords_extracted" db:"keywords_extracted"`
	EntitiesExtracted   map[string]string `json:"entities_extracted" db:"entities_extracted"` // entity -> type
	TopicClassification string            `json:"topic_classification" db:"topic_classification"`
	LanguageDetected    string            `json:"language_detected" db:"language_detected"`
	ProcessedAt         time.Time         `json:"processed_at" db:"processed_at"`
	CreatedAt           time.Time         `json:"created_at" db:"created_at"`
}

// SocialMetrics tracks social media engagement
type SocialMetrics struct {
	ID             string             `json:"id" db:"id"`
	ArticleID      string             `json:"article_id" db:"article_id"`
	URL            string             `json:"url" db:"url"`
	TwitterShares  int64              `json:"twitter_shares" db:"twitter_shares"`
	FacebookShares int64              `json:"facebook_shares" db:"facebook_shares"`
	LinkedInShares int64              `json:"linkedin_shares" db:"linkedin_shares"`
	RedditScore    int64              `json:"reddit_score" db:"reddit_score"`
	TotalShares    int64              `json:"total_shares" db:"total_shares"`
	SocialMentions int64              `json:"social_mentions" db:"social_mentions"`
	SentimentData  map[string]float64 `json:"sentiment_data" db:"sentiment_data"` // platform -> sentiment
	LastFetched    time.Time          `json:"last_fetched" db:"last_fetched"`
	CreatedAt      time.Time          `json:"created_at" db:"created_at"`
}

// ScoringWeights defines the weights for different scoring components
type ScoringWeights struct {
	EngagementWeight  float64 `json:"engagement_weight" yaml:"engagement_weight"`
	CredibilityWeight float64 `json:"credibility_weight" yaml:"credibility_weight"`
	ContentWeight     float64 `json:"content_weight" yaml:"content_weight"`
	SocialWeight      float64 `json:"social_weight" yaml:"social_weight"`
	RecencyWeight     float64 `json:"recency_weight" yaml:"recency_weight"`
}

// CategoryBalance defines requirements for category diversity
type CategoryBalance struct {
	MinCategories      int                `json:"min_categories" yaml:"min_categories"`
	MaxPerCategory     int                `json:"max_per_category" yaml:"max_per_category"`
	CategoryWeights    map[string]float64 `json:"category_weights" yaml:"category_weights"`
	RequiredCategories []string           `json:"required_categories" yaml:"required_categories"`
}

// TopStoriesConfig defines configuration for the enhanced algorithm
type TopStoriesConfig struct {
	ScoringWeights  ScoringWeights  `json:"scoring_weights" yaml:"scoring_weights"`
	CategoryBalance CategoryBalance `json:"category_balance" yaml:"category_balance"`
	MinScore        float64         `json:"min_score" yaml:"min_score"`
	MaxAge          time.Duration   `json:"max_age" yaml:"max_age"`
	RefreshInterval time.Duration   `json:"refresh_interval" yaml:"refresh_interval"`
}
