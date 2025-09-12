package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"news-aggregator/internal/config"
	"news-aggregator/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type NewsRepository struct {
	db     *pgxpool.Pool
	logger zerolog.Logger
}

func NewNewsRepository(cfg *config.Config, logger zerolog.Logger) (*NewsRepository, error) {
	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Database,
		cfg.Database.SSLMode,
	)

	// Create connection pool
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.Database.MaxConns)
	poolConfig.MinConns = int32(cfg.Database.MaxIdleConns)
	poolConfig.MaxConnLifetime = time.Duration(cfg.Database.MaxLifetime) * time.Second

	db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	// Test connection
	if err := db.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &NewsRepository{
		db:     db,
		logger: logger.With().Str("component", "news_repository").Logger(),
	}

	// Initialize database schema
	if err := repo.initSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

func (r *NewsRepository) initSchema(ctx context.Context) error {
	r.logger.Info().Msg("Initializing database schema")

	// Create tables
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`CREATE TABLE IF NOT EXISTS news (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			title TEXT NOT NULL,
			content TEXT,
			summary TEXT,
			url TEXT UNIQUE,
			image_url TEXT,
			author TEXT,
			source TEXT NOT NULL,
			category TEXT DEFAULT 'general',
			tags JSONB DEFAULT '[]',
			published_at TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			content_hash TEXT UNIQUE
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			color TEXT,
			icon TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS sources (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			name TEXT UNIQUE NOT NULL,
			type TEXT NOT NULL,
			url TEXT NOT NULL,
			schedule TEXT NOT NULL,
			rate_limit INTEGER DEFAULT 10,
			headers JSONB DEFAULT '{}',
			enabled BOOLEAN DEFAULT true,
			last_fetched TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_news_published_at ON news(published_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_news_source ON news(source)`,
		`CREATE INDEX IF NOT EXISTS idx_news_category ON news(category)`,
		`CREATE INDEX IF NOT EXISTS idx_news_content_hash ON news(content_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_news_tags ON news USING GIN(tags)`,
		`CREATE INDEX IF NOT EXISTS idx_sources_enabled ON sources(enabled)`,
	}

	for _, query := range queries {
		if _, err := r.db.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	// Insert default categories
	if err := r.insertDefaultCategories(ctx); err != nil {
		return fmt.Errorf("failed to insert default categories: %w", err)
	}

	r.logger.Info().Msg("Database schema initialized successfully")
	return nil
}

func (r *NewsRepository) insertDefaultCategories(ctx context.Context) error {
	categories := []models.Category{
		{Name: "general", Description: "General news", Color: "#6B7280", Icon: "📰"},
		{Name: "technology", Description: "Technology and innovation", Color: "#3B82F6", Icon: "💻"},
		{Name: "business", Description: "Business and finance", Color: "#10B981", Icon: "💼"},
		{Name: "sports", Description: "Sports and athletics", Color: "#F59E0B", Icon: "⚽"},
		{Name: "politics", Description: "Politics and government", Color: "#EF4444", Icon: "🏛️"},
		{Name: "health", Description: "Health and medicine", Color: "#8B5CF6", Icon: "🏥"},
		{Name: "science", Description: "Science and research", Color: "#06B6D4", Icon: "🔬"},
		{Name: "entertainment", Description: "Entertainment and media", Color: "#F97316", Icon: "🎬"},
		{Name: "world", Description: "World and international news", Color: "#84CC16", Icon: "🌍"},
	}

	for _, category := range categories {
		_, err := r.db.Exec(ctx, `
			INSERT INTO categories (name, description, color, icon)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (name) DO NOTHING
		`, category.Name, category.Description, category.Color, category.Icon)
		
		if err != nil {
			return fmt.Errorf("failed to insert category %s: %w", category.Name, err)
		}
	}

	return nil
}

func (r *NewsRepository) GetNews(ctx context.Context, filter models.NewsFilter) ([]models.News, int, error) {
	r.logger.Debug().Interface("filter", filter).Msg("Getting news with filter")

	// Build WHERE clause
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, filter.Category)
		argIndex++
	}

	if filter.Source != "" {
		conditions = append(conditions, fmt.Sprintf("source = $%d", argIndex))
		args = append(args, filter.Source)
		argIndex++
	}

	if len(filter.Tags) > 0 {
		tagsJson, _ := json.Marshal(filter.Tags)
		conditions = append(conditions, fmt.Sprintf("tags @> $%d", argIndex))
		args = append(args, string(tagsJson))
		argIndex++
	}

	if !filter.DateFrom.IsZero() {
		conditions = append(conditions, fmt.Sprintf("published_at >= $%d", argIndex))
		args = append(args, filter.DateFrom)
		argIndex++
	}

	if !filter.DateTo.IsZero() {
		conditions = append(conditions, fmt.Sprintf("published_at <= $%d", argIndex))
		args = append(args, filter.DateTo)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM news %s", whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get news count: %w", err)
	}

	// Get news with pagination - ensure page is at least 1
	page := filter.Page
	if page < 1 {
		page = 1
	}
	
	limit := filter.Limit
	if limit < 1 {
		limit = 20 // Default limit
	}
	
	offset := (page - 1) * limit
	query := fmt.Sprintf(`
		SELECT id, title, content, summary, url, image_url, author, source, 
			   category, tags, published_at, created_at, updated_at
		FROM news %s
		ORDER BY published_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query news: %w", err)
	}
	defer rows.Close()

	var news []models.News
	for rows.Next() {
		var n models.News
		var tagsJSON []byte

		err := rows.Scan(
			&n.ID, &n.Title, &n.Content, &n.Summary, &n.URL, &n.ImageURL,
			&n.Author, &n.Source, &n.Category, &tagsJSON, &n.PublishedAt,
			&n.CreatedAt, &n.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan news row: %w", err)
		}

		// Unmarshal tags
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &n.Tags); err != nil {
				r.logger.Warn().Err(err).Str("id", n.ID).Msg("Failed to unmarshal tags")
				n.Tags = []string{}
			}
		}

		news = append(news, n)
	}

	if rows.Err() != nil {
		return nil, 0, fmt.Errorf("error iterating news rows: %w", rows.Err())
	}

	return news, total, nil
}

func (r *NewsRepository) GetNewsByID(ctx context.Context, id string) (*models.News, error) {
	r.logger.Debug().Str("id", id).Msg("Getting news by ID")

	query := `
		SELECT id, title, content, summary, url, image_url, author, source, 
			   category, tags, published_at, created_at, updated_at, content_hash
		FROM news WHERE id = $1
	`

	var n models.News
	var tagsJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&n.ID, &n.Title, &n.Content, &n.Summary, &n.URL, &n.ImageURL,
		&n.Author, &n.Source, &n.Category, &tagsJSON, &n.PublishedAt,
		&n.CreatedAt, &n.UpdatedAt, &n.Hash,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("news not found")
		}
		return nil, fmt.Errorf("failed to get news by ID: %w", err)
	}

	// Unmarshal tags
	if len(tagsJSON) > 0 {
		if err := json.Unmarshal(tagsJSON, &n.Tags); err != nil {
			r.logger.Warn().Err(err).Str("id", id).Msg("Failed to unmarshal tags")
			n.Tags = []string{}
		}
	}

	return &n, nil
}

func (r *NewsRepository) CreateNews(ctx context.Context, news *models.News) error {
	r.logger.Debug().Str("title", news.Title).Msg("Creating news")

	// Marshal tags to JSON
	tagsJSON, err := json.Marshal(news.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		INSERT INTO news (title, content, summary, url, image_url, author, source, 
						 category, tags, published_at, content_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		news.Title, news.Content, news.Summary, news.URL, news.ImageURL,
		news.Author, news.Source, news.Category, tagsJSON, news.PublishedAt,
		news.Hash,
	).Scan(&news.ID, &news.CreatedAt, &news.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create news: %w", err)
	}

	return nil
}

func (r *NewsRepository) UpdateNews(ctx context.Context, news *models.News) error {
	r.logger.Debug().Str("id", news.ID).Msg("Updating news")

	// Marshal tags to JSON
	tagsJSON, err := json.Marshal(news.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		UPDATE news SET 
			title = $2, content = $3, summary = $4, url = $5, image_url = $6,
			author = $7, category = $8, tags = $9, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err = r.db.QueryRow(ctx, query,
		news.ID, news.Title, news.Content, news.Summary, news.URL,
		news.ImageURL, news.Author, news.Category, tagsJSON,
	).Scan(&news.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("news not found")
		}
		return fmt.Errorf("failed to update news: %w", err)
	}

	return nil
}

func (r *NewsRepository) DeleteNews(ctx context.Context, id string) error {
	r.logger.Debug().Str("id", id).Msg("Deleting news")

	query := `DELETE FROM news WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete news: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("news not found")
	}

	return nil
}

func (r *NewsRepository) CheckDuplicate(ctx context.Context, hash string) (bool, error) {
	r.logger.Debug().Str("hash", hash).Msg("Checking for duplicate")

	query := `SELECT EXISTS(SELECT 1 FROM news WHERE content_hash = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, hash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check duplicate: %w", err)
	}

	return exists, nil
}

func (r *NewsRepository) GetCategories(ctx context.Context) ([]models.Category, error) {
	r.logger.Debug().Msg("Getting categories")

	query := `SELECT id, name, description, color, icon FROM categories ORDER BY name`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Color, &c.Icon)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category row: %w", err)
		}
		categories = append(categories, c)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating category rows: %w", rows.Err())
	}

	return categories, nil
}

func (r *NewsRepository) GetStats(ctx context.Context) (*models.Stats, error) {
	r.logger.Debug().Msg("Getting stats")

	stats := &models.Stats{}

	// Get total articles
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM news").Scan(&stats.TotalArticles)
	if err != nil {
		return nil, fmt.Errorf("failed to get total articles: %w", err)
	}

	// Get articles today
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM news 
		WHERE published_at >= CURRENT_DATE
	`).Scan(&stats.ArticlesToday)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles today: %w", err)
	}

	// Get articles this week
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM news 
		WHERE published_at >= DATE_TRUNC('week', CURRENT_DATE)
	`).Scan(&stats.ArticlesThisWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles this week: %w", err)
	}

	// Get articles this month
	err = r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM news 
		WHERE published_at >= DATE_TRUNC('month', CURRENT_DATE)
	`).Scan(&stats.ArticlesThisMonth)
	if err != nil {
		return nil, fmt.Errorf("failed to get articles this month: %w", err)
	}

	// Get top categories
	rows, err := r.db.Query(ctx, `
		SELECT category, COUNT(*) as count 
		FROM news 
		GROUP BY category 
		ORDER BY count DESC 
		LIMIT 10
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get top categories: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var categoryStats models.CategoryStats
		err := rows.Scan(&categoryStats.Category, &categoryStats.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category stats: %w", err)
		}
		stats.TopCategories = append(stats.TopCategories, categoryStats)
	}

	// Get top sources
	rows, err = r.db.Query(ctx, `
		SELECT source, COUNT(*) as count 
		FROM news 
		GROUP BY source 
		ORDER BY count DESC 
		LIMIT 10
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get top sources: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sourceStats models.SourceStats
		err := rows.Scan(&sourceStats.Source, &sourceStats.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source stats: %w", err)
		}
		stats.TopSources = append(stats.TopSources, sourceStats)
	}

	return stats, nil
}

func (r *NewsRepository) GetSources(ctx context.Context) ([]models.Source, error) {
	r.logger.Debug().Msg("Getting sources")

	query := `
		SELECT id, name, type, url, schedule, rate_limit, headers, enabled, 
			   last_fetched, created_at, updated_at
		FROM sources ORDER BY name
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sources: %w", err)
	}
	defer rows.Close()

	var sources []models.Source
	for rows.Next() {
		var s models.Source
		var headersJSON []byte

		err := rows.Scan(
			&s.ID, &s.Name, &s.Type, &s.URL, &s.Schedule, &s.RateLimit,
			&headersJSON, &s.Enabled, &s.LastFetched, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source row: %w", err)
		}

		// Unmarshal headers
		if len(headersJSON) > 0 {
			if err := json.Unmarshal(headersJSON, &s.Headers); err != nil {
				r.logger.Warn().Err(err).Str("id", s.ID).Msg("Failed to unmarshal headers")
				s.Headers = make(map[string]string)
			}
		}

		sources = append(sources, s)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating source rows: %w", rows.Err())
	}

	return sources, nil
}

func (r *NewsRepository) CreateSource(ctx context.Context, source *models.Source) error {
	r.logger.Debug().Str("name", source.Name).Msg("Creating source")

	// Marshal headers to JSON
	headersJSON, err := json.Marshal(source.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	query := `
		INSERT INTO sources (name, type, url, schedule, rate_limit, headers, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(ctx, query,
		source.Name, source.Type, source.URL, source.Schedule,
		source.RateLimit, headersJSON, source.Enabled,
	).Scan(&source.ID, &source.CreatedAt, &source.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create source: %w", err)
	}

	return nil
}

func (r *NewsRepository) UpdateSource(ctx context.Context, source *models.Source) error {
	r.logger.Debug().Str("id", source.ID).Msg("Updating source")

	// Marshal headers to JSON
	headersJSON, err := json.Marshal(source.Headers)
	if err != nil {
		return fmt.Errorf("failed to marshal headers: %w", err)
	}

	query := `
		UPDATE sources SET 
			name = $2, type = $3, url = $4, schedule = $5, rate_limit = $6,
			headers = $7, enabled = $8, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	err = r.db.QueryRow(ctx, query,
		source.ID, source.Name, source.Type, source.URL, source.Schedule,
		source.RateLimit, headersJSON, source.Enabled,
	).Scan(&source.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("source not found")
		}
		return fmt.Errorf("failed to update source: %w", err)
	}

	return nil
}

func (r *NewsRepository) DeleteSource(ctx context.Context, id string) error {
	r.logger.Debug().Str("id", id).Msg("Deleting source")

	query := `DELETE FROM sources WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete source: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("source not found")
	}

	return nil
}

// GetRecentArticles returns articles from the last specified duration
func (nr *NewsRepository) GetRecentArticles(ctx context.Context, duration time.Duration) ([]models.News, error) {
	since := time.Now().Add(-duration)
	
	query := `
		SELECT id, title, content, summary, url, image_url, author, source, category, tags, 
		       published_at, created_at, updated_at
		FROM news 
		WHERE created_at >= $1
		ORDER BY created_at DESC
	`
	
	rows, err := nr.db.Query(ctx, query, since)
	if err != nil {
		nr.logger.Error().Err(err).Msg("Failed to get recent articles")
		return nil, err
	}
	defer rows.Close()

	var articles []models.News
	for rows.Next() {
		var article models.News
		var tagsJSON []byte

		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Content,
			&article.Summary,
			&article.URL,
			&article.ImageURL,
			&article.Author,
			&article.Source,
			&article.Category,
			&tagsJSON,
			&article.PublishedAt,
			&article.CreatedAt,
			&article.UpdatedAt,
		)
		if err != nil {
			nr.logger.Error().Err(err).Msg("Failed to scan article row")
			continue
		}

		// Parse tags JSON
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &article.Tags); err != nil {
				nr.logger.Warn().Err(err).Msg("Failed to parse tags JSON")
				article.Tags = []string{}
			}
		}

		articles = append(articles, article)
	}

	nr.logger.Debug().Int("count", len(articles)).Dur("duration", duration).Msg("Retrieved recent articles")
	return articles, nil
}

// GetArticlesByDateRange returns articles within a specific date range
func (nr *NewsRepository) GetArticlesByDateRange(ctx context.Context, start, end time.Time) ([]models.News, error) {
	query := `
		SELECT id, title, content, summary, url, image_url, author, source, category, tags, 
		       published_at, created_at, updated_at
		FROM news 
		WHERE created_at >= $1 AND created_at < $2
		ORDER BY created_at DESC
	`
	
	rows, err := nr.db.Query(ctx, query, start, end)
	if err != nil {
		nr.logger.Error().Err(err).Msg("Failed to get articles by date range")
		return nil, err
	}
	defer rows.Close()

	var articles []models.News
	for rows.Next() {
		var article models.News
		var tagsJSON []byte

		err := rows.Scan(
			&article.ID,
			&article.Title,
			&article.Content,
			&article.Summary,
			&article.URL,
			&article.ImageURL,
			&article.Author,
			&article.Source,
			&article.Category,
			&tagsJSON,
			&article.PublishedAt,
			&article.CreatedAt,
			&article.UpdatedAt,
		)
		if err != nil {
			nr.logger.Error().Err(err).Msg("Failed to scan article row")
			continue
		}

		// Parse tags JSON
		if len(tagsJSON) > 0 {
			if err := json.Unmarshal(tagsJSON, &article.Tags); err != nil {
				nr.logger.Warn().Err(err).Msg("Failed to parse tags JSON")
				article.Tags = []string{}
			}
		}

		articles = append(articles, article)
	}

	nr.logger.Debug().Int("count", len(articles)).Time("start", start).Time("end", end).Msg("Retrieved articles by date range")
	return articles, nil
}

// CleanupOldArticles removes articles older than 2 days from the database
func (r *NewsRepository) CleanupOldArticles(ctx context.Context) error {
	r.logger.Info().Msg("Starting cleanup of articles older than 2 days")
	
	twoDaysAgo := time.Now().AddDate(0, 0, -2)
	
	query := `DELETE FROM news WHERE published_at < $1`
	
	result, err := r.db.Exec(ctx, query, twoDaysAgo)
	if err != nil {
		r.logger.Error().Err(err).Msg("Failed to cleanup old articles")
		return fmt.Errorf("failed to cleanup old articles: %w", err)
	}
	
	deletedCount := result.RowsAffected()
	r.logger.Info().Int64("deleted_count", deletedCount).Time("cutoff_date", twoDaysAgo).Msg("Cleanup completed")
	
	return nil
}

func (r *NewsRepository) Close() error {
	r.db.Close()
	return nil
}
