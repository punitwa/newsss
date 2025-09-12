package services

import (
	"context"
	"fmt"
	"strings"

	"news-aggregator/internal/config"
	"news-aggregator/internal/models"
	"news-aggregator/internal/repository"

	"github.com/rs/zerolog"
)

type NewsService struct {
	config     *config.Config
	logger     zerolog.Logger
	repository *repository.NewsRepository
}

func NewNewsService(cfg *config.Config, logger zerolog.Logger) (*NewsService, error) {
	repo, err := repository.NewNewsRepository(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create news repository: %w", err)
	}

	return &NewsService{
		config:     cfg,
		logger:     logger,
		repository: repo,
	}, nil
}

func (s *NewsService) GetNews(ctx context.Context, filter models.NewsFilter) ([]models.News, int, error) {
	s.logger.Debug().
		Int("page", filter.Page).
		Int("limit", filter.Limit).
		Str("category", filter.Category).
		Str("source", filter.Source).
		Msg("Getting news with filter")

	news, total, err := s.repository.GetNews(ctx, filter)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get news from repository")
		return nil, 0, fmt.Errorf("failed to get news: %w", err)
	}

	return news, total, nil
}

func (s *NewsService) GetNewsByID(ctx context.Context, id string) (*models.News, error) {
	s.logger.Debug().Str("id", id).Msg("Getting news by ID")

	news, err := s.repository.GetNewsByID(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Str("id", id).Msg("Failed to get news by ID")
		return nil, fmt.Errorf("failed to get news by ID: %w", err)
	}

	return news, nil
}

func (s *NewsService) CreateNews(ctx context.Context, news *models.News) error {
	s.logger.Debug().Str("title", news.Title).Str("source", news.Source).Msg("Creating news")

	if err := s.repository.CreateNews(ctx, news); err != nil {
		// Check if this is a duplicate URL error
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") && 
		   strings.Contains(err.Error(), "news_url_key") {
			s.logger.Debug().Str("title", news.Title).Str("url", news.URL).Msg("Duplicate article URL detected, skipping")
			return nil // Don't treat duplicates as errors
		}
		
		// Check if this is a duplicate content hash error
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") && 
		   strings.Contains(err.Error(), "content_hash") {
			s.logger.Debug().Str("title", news.Title).Str("hash", news.Hash).Msg("Duplicate article content detected, skipping")
			return nil // Don't treat duplicates as errors
		}
		
		s.logger.Error().Err(err).Str("title", news.Title).Msg("Failed to create news")
		return fmt.Errorf("failed to create news: %w", err)
	}

	s.logger.Info().Str("title", news.Title).Str("source", news.Source).Msg("News article created successfully")
	return nil
}

func (s *NewsService) UpdateNews(ctx context.Context, news *models.News) error {
	s.logger.Debug().Str("id", news.ID).Str("title", news.Title).Msg("Updating news")

	if err := s.repository.UpdateNews(ctx, news); err != nil {
		s.logger.Error().Err(err).Str("id", news.ID).Msg("Failed to update news")
		return fmt.Errorf("failed to update news: %w", err)
	}

	return nil
}

func (s *NewsService) DeleteNews(ctx context.Context, id string) error {
	s.logger.Debug().Str("id", id).Msg("Deleting news")

	if err := s.repository.DeleteNews(ctx, id); err != nil {
		s.logger.Error().Err(err).Str("id", id).Msg("Failed to delete news")
		return fmt.Errorf("failed to delete news: %w", err)
	}

	return nil
}

func (s *NewsService) GetCategories(ctx context.Context) ([]models.Category, error) {
	s.logger.Debug().Msg("Getting categories")

	categories, err := s.repository.GetCategories(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get categories")
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

func (s *NewsService) GetStats(ctx context.Context) (*models.Stats, error) {
	s.logger.Debug().Msg("Getting stats")

	stats, err := s.repository.GetStats(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get stats")
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	return stats, nil
}

func (s *NewsService) AddSource(ctx context.Context, req *models.SourceRequest) (*models.Source, error) {
	s.logger.Debug().Str("name", req.Name).Str("url", req.URL).Msg("Adding source")

	source := &models.Source{
		Name:      req.Name,
		Type:      req.Type,
		URL:       req.URL,
		Schedule:  req.Schedule,
		RateLimit: req.RateLimit,
		Headers:   req.Headers,
		Enabled:   req.Enabled,
	}

	if err := s.repository.CreateSource(ctx, source); err != nil {
		s.logger.Error().Err(err).Str("name", req.Name).Msg("Failed to add source")
		return nil, fmt.Errorf("failed to add source: %w", err)
	}

	return source, nil
}

func (s *NewsService) UpdateSource(ctx context.Context, id string, req *models.SourceRequest) error {
	s.logger.Debug().Str("id", id).Str("name", req.Name).Msg("Updating source")

	source := &models.Source{
		ID:        id,
		Name:      req.Name,
		Type:      req.Type,
		URL:       req.URL,
		Schedule:  req.Schedule,
		RateLimit: req.RateLimit,
		Headers:   req.Headers,
		Enabled:   req.Enabled,
	}

	if err := s.repository.UpdateSource(ctx, source); err != nil {
		s.logger.Error().Err(err).Str("id", id).Msg("Failed to update source")
		return fmt.Errorf("failed to update source: %w", err)
	}

	return nil
}

func (s *NewsService) DeleteSource(ctx context.Context, id string) error {
	s.logger.Debug().Str("id", id).Msg("Deleting source")

	if err := s.repository.DeleteSource(ctx, id); err != nil {
		s.logger.Error().Err(err).Str("id", id).Msg("Failed to delete source")
		return fmt.Errorf("failed to delete source: %w", err)
	}

	return nil
}

func (s *NewsService) GetSources(ctx context.Context) ([]models.Source, error) {
	s.logger.Debug().Msg("Getting sources")

	sources, err := s.repository.GetSources(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get sources")
		return nil, fmt.Errorf("failed to get sources: %w", err)
	}

	return sources, nil
}

func (s *NewsService) CheckDuplicate(ctx context.Context, hash string) (bool, error) {
	s.logger.Debug().Str("hash", hash).Msg("Checking for duplicate")

	exists, err := s.repository.CheckDuplicate(ctx, hash)
	if err != nil {
		s.logger.Error().Err(err).Str("hash", hash).Msg("Failed to check duplicate")
		return false, fmt.Errorf("failed to check duplicate: %w", err)
	}

	return exists, nil
}

// CleanupOldArticles removes articles older than 2 days
func (s *NewsService) CleanupOldArticles(ctx context.Context) error {
	s.logger.Info().Msg("Cleaning up old articles (older than 2 days)")
	
	if err := s.repository.CleanupOldArticles(ctx); err != nil {
		s.logger.Error().Err(err).Msg("Failed to cleanup old articles")
		return fmt.Errorf("failed to cleanup old articles: %w", err)
	}
	
	return nil
}

// GetRepository returns the news repository for use by other services
func (s *NewsService) GetRepository() *repository.NewsRepository {
	return s.repository
}
