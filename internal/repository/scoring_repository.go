package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"news-aggregator/internal/models"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
)

// ScoringRepository handles database operations for scoring-related data
type ScoringRepository struct {
	db     *pgxpool.Pool
	logger zerolog.Logger
}

// NewScoringRepository creates a new scoring repository
func NewScoringRepository(db *pgxpool.Pool, logger zerolog.Logger) *ScoringRepository {
	repo := &ScoringRepository{
		db:     db,
		logger: logger.With().Str("repository", "scoring").Logger(),
	}
	return repo
}

// InitSchema creates the necessary tables for scoring
func (r *ScoringRepository) InitSchema(ctx context.Context) error {
	r.logger.Info().Msg("Initializing scoring schema")

	queries := []string{
		// Article scores table
		`CREATE TABLE IF NOT EXISTS article_scores (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			article_id UUID NOT NULL REFERENCES news(id) ON DELETE CASCADE,
			engagement_score DECIMAL(5,4) DEFAULT 0.0,
			credibility_score DECIMAL(5,4) DEFAULT 0.0,
			content_score DECIMAL(5,4) DEFAULT 0.0,
			social_score DECIMAL(5,4) DEFAULT 0.0,
			final_score DECIMAL(5,4) DEFAULT 0.0,
			last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(article_id)
		)`,

		// Engagement metrics table
		`CREATE TABLE IF NOT EXISTS engagement_metrics (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			article_id UUID NOT NULL REFERENCES news(id) ON DELETE CASCADE,
			view_count BIGINT DEFAULT 0,
			click_count BIGINT DEFAULT 0,
			share_count BIGINT DEFAULT 0,
			average_read_time BIGINT DEFAULT 0,
			bounce_rate DECIMAL(5,4) DEFAULT 0.0,
			last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(article_id)
		)`,

		// Source credibility table
		`CREATE TABLE IF NOT EXISTS source_credibility (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			source_name TEXT NOT NULL UNIQUE,
			credibility_score DECIMAL(5,4) DEFAULT 0.5,
			reliability_score DECIMAL(5,4) DEFAULT 0.5,
			bias_score DECIMAL(5,4) DEFAULT 0.0,
			factual_score DECIMAL(5,4) DEFAULT 0.5,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Content analysis table
		`CREATE TABLE IF NOT EXISTS content_analysis (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			article_id UUID NOT NULL REFERENCES news(id) ON DELETE CASCADE,
			sentiment_score DECIMAL(5,4) DEFAULT 0.0,
			importance_score DECIMAL(5,4) DEFAULT 0.5,
			readability_score DECIMAL(5,4) DEFAULT 0.5,
			keywords_extracted JSONB DEFAULT '[]',
			entities_extracted JSONB DEFAULT '{}',
			topic_classification TEXT,
			language_detected TEXT DEFAULT 'en',
			processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(article_id)
		)`,

		// Social metrics table
		`CREATE TABLE IF NOT EXISTS social_metrics (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			article_id UUID NOT NULL REFERENCES news(id) ON DELETE CASCADE,
			url TEXT NOT NULL,
			twitter_shares BIGINT DEFAULT 0,
			facebook_shares BIGINT DEFAULT 0,
			linkedin_shares BIGINT DEFAULT 0,
			reddit_score BIGINT DEFAULT 0,
			total_shares BIGINT DEFAULT 0,
			social_mentions BIGINT DEFAULT 0,
			sentiment_data JSONB DEFAULT '{}',
			last_fetched TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(article_id)
		)`,

		// Indexes for performance
		`CREATE INDEX IF NOT EXISTS idx_article_scores_final_score ON article_scores(final_score DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_article_scores_article_id ON article_scores(article_id)`,
		`CREATE INDEX IF NOT EXISTS idx_engagement_metrics_article_id ON engagement_metrics(article_id)`,
		`CREATE INDEX IF NOT EXISTS idx_content_analysis_article_id ON content_analysis(article_id)`,
		`CREATE INDEX IF NOT EXISTS idx_social_metrics_article_id ON social_metrics(article_id)`,
		`CREATE INDEX IF NOT EXISTS idx_social_metrics_last_fetched ON social_metrics(last_fetched)`,
		`CREATE INDEX IF NOT EXISTS idx_source_credibility_name ON source_credibility(source_name)`,
	}

	for _, query := range queries {
		if _, err := r.db.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	// Insert default source credibility scores
	if err := r.insertDefaultSourceCredibility(ctx); err != nil {
		r.logger.Warn().Err(err).Msg("Failed to insert default source credibility")
	}

	r.logger.Info().Msg("Scoring schema initialized successfully")
	return nil
}

// insertDefaultSourceCredibility inserts default credibility scores
func (r *ScoringRepository) insertDefaultSourceCredibility(ctx context.Context) error {
	defaultSources := []models.SourceCredibility{
		{SourceName: "BBC News", CredibilityScore: 0.9, ReliabilityScore: 0.95, BiasScore: 0.0, FactualScore: 0.9},
		{SourceName: "Reuters", CredibilityScore: 0.9, ReliabilityScore: 0.95, BiasScore: 0.0, FactualScore: 0.95},
		{SourceName: "Associated Press", CredibilityScore: 0.9, ReliabilityScore: 0.95, BiasScore: 0.0, FactualScore: 0.9},
		{SourceName: "NPR", CredibilityScore: 0.85, ReliabilityScore: 0.9, BiasScore: -0.1, FactualScore: 0.85},
		{SourceName: "The Guardian", CredibilityScore: 0.8, ReliabilityScore: 0.85, BiasScore: -0.2, FactualScore: 0.8},
		{SourceName: "CNN", CredibilityScore: 0.75, ReliabilityScore: 0.8, BiasScore: -0.15, FactualScore: 0.75},
		{SourceName: "TechCrunch", CredibilityScore: 0.7, ReliabilityScore: 0.8, BiasScore: 0.0, FactualScore: 0.75},
		{SourceName: "NDTV", CredibilityScore: 0.75, ReliabilityScore: 0.8, BiasScore: 0.0, FactualScore: 0.75},
		{SourceName: "Times of India", CredibilityScore: 0.7, ReliabilityScore: 0.75, BiasScore: 0.0, FactualScore: 0.7},
		{SourceName: "The Hindu", CredibilityScore: 0.8, ReliabilityScore: 0.85, BiasScore: 0.0, FactualScore: 0.8},
		{SourceName: "Hindustan Times", CredibilityScore: 0.7, ReliabilityScore: 0.75, BiasScore: 0.0, FactualScore: 0.7},
	}

	for _, source := range defaultSources {
		query := `
			INSERT INTO source_credibility (source_name, credibility_score, reliability_score, bias_score, factual_score)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (source_name) DO NOTHING`

		_, err := r.db.Exec(ctx, query,
			source.SourceName,
			source.CredibilityScore,
			source.ReliabilityScore,
			source.BiasScore,
			source.FactualScore,
		)
		if err != nil {
			r.logger.Warn().Str("source", source.SourceName).Err(err).Msg("Failed to insert source credibility")
		}
	}

	return nil
}

// Article Scores
func (r *ScoringRepository) SaveArticleScore(ctx context.Context, score *models.ArticleScore) error {
	query := `
		INSERT INTO article_scores (article_id, engagement_score, credibility_score, content_score, social_score, final_score, last_updated)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (article_id) DO UPDATE SET
			engagement_score = EXCLUDED.engagement_score,
			credibility_score = EXCLUDED.credibility_score,
			content_score = EXCLUDED.content_score,
			social_score = EXCLUDED.social_score,
			final_score = EXCLUDED.final_score,
			last_updated = EXCLUDED.last_updated`

	_, err := r.db.Exec(ctx, query,
		score.ArticleID,
		score.EngagementScore,
		score.CredibilityScore,
		score.ContentScore,
		score.SocialScore,
		score.FinalScore,
		score.LastUpdated,
	)

	return err
}

func (r *ScoringRepository) GetArticleScore(ctx context.Context, articleID string) (*models.ArticleScore, error) {
	query := `
		SELECT id, article_id, engagement_score, credibility_score, content_score, social_score, final_score, last_updated, created_at
		FROM article_scores WHERE article_id = $1`

	var score models.ArticleScore
	err := r.db.QueryRow(ctx, query, articleID).Scan(
		&score.ID,
		&score.ArticleID,
		&score.EngagementScore,
		&score.CredibilityScore,
		&score.ContentScore,
		&score.SocialScore,
		&score.FinalScore,
		&score.LastUpdated,
		&score.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &score, nil
}

// Engagement Metrics
func (r *ScoringRepository) UpdateEngagementMetrics(ctx context.Context, articleID, engagementType string, value int64) error {
	// First, ensure the record exists
	_, err := r.db.Exec(ctx, `
		INSERT INTO engagement_metrics (article_id) 
		VALUES ($1) 
		ON CONFLICT (article_id) DO NOTHING`, articleID)
	if err != nil {
		return err
	}

	// Update the specific metric
	var query string
	switch engagementType {
	case "view":
		query = `UPDATE engagement_metrics SET view_count = view_count + $2, last_updated = NOW() WHERE article_id = $1`
	case "click":
		query = `UPDATE engagement_metrics SET click_count = click_count + $2, last_updated = NOW() WHERE article_id = $1`
	case "share":
		query = `UPDATE engagement_metrics SET share_count = share_count + $2, last_updated = NOW() WHERE article_id = $1`
	case "read_time":
		query = `UPDATE engagement_metrics SET average_read_time = (average_read_time + $2) / 2, last_updated = NOW() WHERE article_id = $1`
	case "bounce_rate":
		query = `UPDATE engagement_metrics SET bounce_rate = $2, last_updated = NOW() WHERE article_id = $1`
	default:
		return fmt.Errorf("unknown engagement type: %s", engagementType)
	}

	_, err = r.db.Exec(ctx, query, articleID, value)
	return err
}

func (r *ScoringRepository) GetEngagementMetrics(ctx context.Context, articleID string) (*models.EngagementMetrics, error) {
	query := `
		SELECT id, article_id, view_count, click_count, share_count, average_read_time, bounce_rate, last_updated, created_at
		FROM engagement_metrics WHERE article_id = $1`

	var metrics models.EngagementMetrics
	err := r.db.QueryRow(ctx, query, articleID).Scan(
		&metrics.ID,
		&metrics.ArticleID,
		&metrics.ViewCount,
		&metrics.ClickCount,
		&metrics.ShareCount,
		&metrics.AverageReadTime,
		&metrics.BounceRate,
		&metrics.LastUpdated,
		&metrics.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &metrics, nil
}

// Source Credibility
func (r *ScoringRepository) GetSourceCredibility(ctx context.Context, sourceName string) (*models.SourceCredibility, error) {
	query := `
		SELECT id, source_name, credibility_score, reliability_score, bias_score, factual_score, updated_at, created_at
		FROM source_credibility WHERE source_name = $1`

	var credibility models.SourceCredibility
	err := r.db.QueryRow(ctx, query, sourceName).Scan(
		&credibility.ID,
		&credibility.SourceName,
		&credibility.CredibilityScore,
		&credibility.ReliabilityScore,
		&credibility.BiasScore,
		&credibility.FactualScore,
		&credibility.UpdatedAt,
		&credibility.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &credibility, nil
}

func (r *ScoringRepository) UpdateSourceCredibility(ctx context.Context, credibility *models.SourceCredibility) error {
	query := `
		UPDATE source_credibility SET 
			credibility_score = $2, 
			reliability_score = $3, 
			bias_score = $4, 
			factual_score = $5, 
			updated_at = NOW()
		WHERE source_name = $1`

	_, err := r.db.Exec(ctx, query,
		credibility.SourceName,
		credibility.CredibilityScore,
		credibility.ReliabilityScore,
		credibility.BiasScore,
		credibility.FactualScore,
	)

	return err
}

// Content Analysis
func (r *ScoringRepository) SaveContentAnalysis(ctx context.Context, analysis *models.ContentAnalysis) error {
	keywordsJSON, _ := json.Marshal(analysis.KeywordsExtracted)
	entitiesJSON, _ := json.Marshal(analysis.EntitiesExtracted)

	query := `
		INSERT INTO content_analysis (
			article_id, sentiment_score, importance_score, readability_score,
			keywords_extracted, entities_extracted, topic_classification, language_detected, processed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (article_id) DO UPDATE SET
			sentiment_score = EXCLUDED.sentiment_score,
			importance_score = EXCLUDED.importance_score,
			readability_score = EXCLUDED.readability_score,
			keywords_extracted = EXCLUDED.keywords_extracted,
			entities_extracted = EXCLUDED.entities_extracted,
			topic_classification = EXCLUDED.topic_classification,
			language_detected = EXCLUDED.language_detected,
			processed_at = EXCLUDED.processed_at`

	_, err := r.db.Exec(ctx, query,
		analysis.ArticleID,
		analysis.SentimentScore,
		analysis.ImportanceScore,
		analysis.ReadabilityScore,
		keywordsJSON,
		entitiesJSON,
		analysis.TopicClassification,
		analysis.LanguageDetected,
		analysis.ProcessedAt,
	)

	return err
}

func (r *ScoringRepository) GetContentAnalysis(ctx context.Context, articleID string) (*models.ContentAnalysis, error) {
	query := `
		SELECT id, article_id, sentiment_score, importance_score, readability_score,
			   keywords_extracted, entities_extracted, topic_classification, language_detected, processed_at, created_at
		FROM content_analysis WHERE article_id = $1`

	var analysis models.ContentAnalysis
	var keywordsJSON, entitiesJSON []byte

	err := r.db.QueryRow(ctx, query, articleID).Scan(
		&analysis.ID,
		&analysis.ArticleID,
		&analysis.SentimentScore,
		&analysis.ImportanceScore,
		&analysis.ReadabilityScore,
		&keywordsJSON,
		&entitiesJSON,
		&analysis.TopicClassification,
		&analysis.LanguageDetected,
		&analysis.ProcessedAt,
		&analysis.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if len(keywordsJSON) > 0 {
		json.Unmarshal(keywordsJSON, &analysis.KeywordsExtracted)
	}
	if len(entitiesJSON) > 0 {
		json.Unmarshal(entitiesJSON, &analysis.EntitiesExtracted)
	}

	return &analysis, nil
}

// Social Metrics
func (r *ScoringRepository) SaveSocialMetrics(ctx context.Context, metrics *models.SocialMetrics) error {
	sentimentJSON, _ := json.Marshal(metrics.SentimentData)

	query := `
		INSERT INTO social_metrics (
			article_id, url, twitter_shares, facebook_shares, linkedin_shares, 
			reddit_score, total_shares, social_mentions, sentiment_data, last_fetched
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (article_id) DO UPDATE SET
			twitter_shares = EXCLUDED.twitter_shares,
			facebook_shares = EXCLUDED.facebook_shares,
			linkedin_shares = EXCLUDED.linkedin_shares,
			reddit_score = EXCLUDED.reddit_score,
			total_shares = EXCLUDED.total_shares,
			social_mentions = EXCLUDED.social_mentions,
			sentiment_data = EXCLUDED.sentiment_data,
			last_fetched = EXCLUDED.last_fetched`

	_, err := r.db.Exec(ctx, query,
		metrics.ArticleID,
		metrics.URL,
		metrics.TwitterShares,
		metrics.FacebookShares,
		metrics.LinkedInShares,
		metrics.RedditScore,
		metrics.TotalShares,
		metrics.SocialMentions,
		sentimentJSON,
		metrics.LastFetched,
	)

	return err
}

func (r *ScoringRepository) GetSocialMetrics(ctx context.Context, url string) (*models.SocialMetrics, error) {
	query := `
		SELECT id, article_id, url, twitter_shares, facebook_shares, linkedin_shares,
			   reddit_score, total_shares, social_mentions, sentiment_data, last_fetched, created_at
		FROM social_metrics WHERE url = $1`

	var metrics models.SocialMetrics
	var sentimentJSON []byte

	err := r.db.QueryRow(ctx, query, url).Scan(
		&metrics.ID,
		&metrics.ArticleID,
		&metrics.URL,
		&metrics.TwitterShares,
		&metrics.FacebookShares,
		&metrics.LinkedInShares,
		&metrics.RedditScore,
		&metrics.TotalShares,
		&metrics.SocialMentions,
		&sentimentJSON,
		&metrics.LastFetched,
		&metrics.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse sentiment JSON
	if len(sentimentJSON) > 0 {
		json.Unmarshal(sentimentJSON, &metrics.SentimentData)
	}

	return &metrics, nil
}

// Get top scored articles
func (r *ScoringRepository) GetTopScoredArticles(ctx context.Context, limit int, minScore float64) ([]string, error) {
	query := `
		SELECT article_id FROM article_scores 
		WHERE final_score >= $1 
		ORDER BY final_score DESC 
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, minScore, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articleIDs []string
	for rows.Next() {
		var articleID string
		if err := rows.Scan(&articleID); err != nil {
			continue
		}
		articleIDs = append(articleIDs, articleID)
	}

	return articleIDs, nil
}
