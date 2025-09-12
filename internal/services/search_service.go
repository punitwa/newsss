package services

import (
	"context"
	"fmt"

	"news-aggregator/internal/config"
	"news-aggregator/internal/models"
	"news-aggregator/internal/repository"

	"github.com/rs/zerolog"
)

type SearchService struct {
	config     *config.Config
	logger     zerolog.Logger
	repository *repository.SearchRepository
}

func NewSearchService(cfg *config.Config, logger zerolog.Logger) (*SearchService, error) {
	repo, err := repository.NewSearchRepository(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create search repository: %w", err)
	}

	return &SearchService{
		config:     cfg,
		logger:     logger,
		repository: repo,
	}, nil
}

func (s *SearchService) Search(ctx context.Context, query string, page, limit int) ([]models.News, int64, error) {
	s.logger.Debug().
		Str("query", query).
		Int("page", page).
		Int("limit", limit).
		Msg("Performing search")

	results, total, err := s.repository.Search(ctx, query, page, limit)
	if err != nil {
		s.logger.Error().Err(err).Str("query", query).Msg("Search failed")
		return nil, 0, fmt.Errorf("search failed: %w", err)
	}

	return results, total, nil
}

func (s *SearchService) AdvancedSearch(ctx context.Context, searchQuery models.SearchQuery) (*models.SearchResult, error) {
	s.logger.Debug().
		Str("query", searchQuery.Query).
		Interface("categories", searchQuery.Categories).
		Interface("sources", searchQuery.Sources).
		Msg("Performing advanced search")

	results, err := s.repository.AdvancedSearch(ctx, searchQuery)
	if err != nil {
		s.logger.Error().Err(err).Str("query", searchQuery.Query).Msg("Advanced search failed")
		return nil, fmt.Errorf("advanced search failed: %w", err)
	}

	return results, nil
}

func (s *SearchService) IndexNews(ctx context.Context, news *models.News) error {
	s.logger.Debug().Str("id", news.ID).Str("title", news.Title).Msg("Indexing news")

	if err := s.repository.IndexNews(ctx, news); err != nil {
		s.logger.Error().Err(err).Str("id", news.ID).Msg("Failed to index news")
		return fmt.Errorf("failed to index news: %w", err)
	}

	return nil
}

func (s *SearchService) UpdateNewsIndex(ctx context.Context, news *models.News) error {
	s.logger.Debug().Str("id", news.ID).Str("title", news.Title).Msg("Updating news index")

	if err := s.repository.UpdateNewsIndex(ctx, news); err != nil {
		s.logger.Error().Err(err).Str("id", news.ID).Msg("Failed to update news index")
		return fmt.Errorf("failed to update news index: %w", err)
	}

	return nil
}

func (s *SearchService) DeleteFromIndex(ctx context.Context, newsID string) error {
	s.logger.Debug().Str("id", newsID).Msg("Deleting from index")

	if err := s.repository.DeleteFromIndex(ctx, newsID); err != nil {
		s.logger.Error().Err(err).Str("id", newsID).Msg("Failed to delete from index")
		return fmt.Errorf("failed to delete from index: %w", err)
	}

	return nil
}

func (s *SearchService) GetSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	s.logger.Debug().Str("query", query).Int("limit", limit).Msg("Getting search suggestions")

	suggestions, err := s.repository.GetSuggestions(ctx, query, limit)
	if err != nil {
		s.logger.Error().Err(err).Str("query", query).Msg("Failed to get suggestions")
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}

	return suggestions, nil
}
