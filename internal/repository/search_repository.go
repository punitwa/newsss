package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"news-aggregator/internal/config"
	"news-aggregator/internal/models"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/rs/zerolog"
)

type SearchRepository struct {
	client *elasticsearch.Client
	logger zerolog.Logger
	index  string
}

func NewSearchRepository(cfg *config.Config, logger zerolog.Logger) (*SearchRepository, error) {
	// Create Elasticsearch client
	esConfig := elasticsearch.Config{
		Addresses: cfg.Elasticsearch.Addresses,
	}

	if cfg.Elasticsearch.Username != "" {
		esConfig.Username = cfg.Elasticsearch.Username
		esConfig.Password = cfg.Elasticsearch.Password
	}

	client, err := elasticsearch.NewClient(esConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	repo := &SearchRepository{
		client: client,
		logger: logger.With().Str("component", "search_repository").Logger(),
		index:  cfg.Elasticsearch.Index,
	}

	// Initialize index
	if err := repo.initIndex(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize index: %w", err)
	}

	return repo, nil
}

func (r *SearchRepository) initIndex(ctx context.Context) error {
	r.logger.Info().Str("index", r.index).Msg("Initializing Elasticsearch index")

	// Check if index exists
	req := esapi.IndicesExistsRequest{
		Index: []string{r.index},
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	// If index exists, return
	if res.StatusCode == 200 {
		r.logger.Info().Str("index", r.index).Msg("Index already exists")
		return nil
	}

	// Create index with mapping
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"title": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
					"fields": map[string]interface{}{
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"content": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
				},
				"summary": map[string]interface{}{
					"type":     "text",
					"analyzer": "standard",
				},
				"author": map[string]interface{}{
					"type": "keyword",
				},
				"source": map[string]interface{}{
					"type": "keyword",
				},
				"category": map[string]interface{}{
					"type": "keyword",
				},
				"tags": map[string]interface{}{
					"type": "keyword",
				},
				"url": map[string]interface{}{
					"type":  "keyword",
					"index": false,
				},
				"image_url": map[string]interface{}{
					"type":  "keyword",
					"index": false,
				},
				"published_at": map[string]interface{}{
					"type": "date",
				},
				"created_at": map[string]interface{}{
					"type": "date",
				},
			},
		},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
			"analysis": map[string]interface{}{
				"analyzer": map[string]interface{}{
					"news_analyzer": map[string]interface{}{
						"type":      "custom",
						"tokenizer": "standard",
						"filter": []string{
							"lowercase",
							"stop",
							"snowball",
						},
					},
				},
			},
		},
	}

	mappingJSON, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping: %w", err)
	}

	createReq := esapi.IndicesCreateRequest{
		Index: r.index,
		Body:  bytes.NewReader(mappingJSON),
	}

	res, err = createReq.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to create index: %s", res.String())
	}

	r.logger.Info().Str("index", r.index).Msg("Index created successfully")
	return nil
}

func (r *SearchRepository) IndexNews(ctx context.Context, news *models.News) error {
	r.logger.Debug().Str("id", news.ID).Str("title", news.Title).Msg("Indexing news")

	// Prepare document for indexing
	doc := map[string]interface{}{
		"title":        news.Title,
		"content":      news.Content,
		"summary":      news.Summary,
		"author":       news.Author,
		"source":       news.Source,
		"category":     news.Category,
		"tags":         news.Tags,
		"url":          news.URL,
		"image_url":    news.ImageURL,
		"published_at": news.PublishedAt,
		"created_at":   news.CreatedAt,
	}

	docJSON, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      r.index,
		DocumentID: news.ID,
		Body:       bytes.NewReader(docJSON),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index document: %s", res.String())
	}

	return nil
}

func (r *SearchRepository) UpdateNewsIndex(ctx context.Context, news *models.News) error {
	r.logger.Debug().Str("id", news.ID).Str("title", news.Title).Msg("Updating news index")

	// Use the same method as indexing since Elasticsearch handles updates automatically
	return r.IndexNews(ctx, news)
}

func (r *SearchRepository) DeleteFromIndex(ctx context.Context, newsID string) error {
	r.logger.Debug().Str("id", newsID).Msg("Deleting from index")

	req := esapi.DeleteRequest{
		Index:      r.index,
		DocumentID: newsID,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("failed to delete document: %s", res.String())
	}

	return nil
}

func (r *SearchRepository) Search(ctx context.Context, query string, page, limit int) ([]models.News, int64, error) {
	r.logger.Debug().Str("query", query).Int("page", page).Int("limit", limit).Msg("Performing search")

	from := (page - 1) * limit

	// Build search query with 7-day filter
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query":  query,
						"fields": []string{"title^3", "content^2", "summary^2", "author", "category", "tags"},
						"type":   "best_fields",
					},
				},
				"filter": map[string]interface{}{
					"range": map[string]interface{}{
						"published_at": map[string]interface{}{
							"gte": sevenDaysAgo,
						},
					},
				},
			},
		},
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"title":   map[string]interface{}{},
				"content": map[string]interface{}{},
				"summary": map[string]interface{}{},
			},
		},
		"sort": []map[string]interface{}{
			{
				"published_at": map[string]interface{}{
					"order": "desc",
				},
			},
		},
		"from": from,
		"size": limit,
	}

	queryJSON, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to marshal search query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{r.index},
		Body:  bytes.NewReader(queryJSON),
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("search error: %s", res.String())
	}

	var searchResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, 0, fmt.Errorf("failed to decode search result: %w", err)
	}

	return r.parseSearchResult(searchResult)
}

func (r *SearchRepository) AdvancedSearch(ctx context.Context, searchQuery models.SearchQuery) (*models.SearchResult, error) {
	r.logger.Debug().Interface("query", searchQuery).Msg("Performing advanced search")

	from := (searchQuery.Page - 1) * searchQuery.Limit

	// Build advanced search query
	mustQueries := []map[string]interface{}{}

	// Text query
	if searchQuery.Query != "" {
		mustQueries = append(mustQueries, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  searchQuery.Query,
				"fields": []string{"title^3", "content^2", "summary^2"},
				"type":   "best_fields",
			},
		})
	}

	// Category filter
	if len(searchQuery.Categories) > 0 {
		mustQueries = append(mustQueries, map[string]interface{}{
			"terms": map[string]interface{}{
				"category": searchQuery.Categories,
			},
		})
	}

	// Source filter
	if len(searchQuery.Sources) > 0 {
		mustQueries = append(mustQueries, map[string]interface{}{
			"terms": map[string]interface{}{
				"source": searchQuery.Sources,
			},
		})
	}

	// Date range filter
	if !searchQuery.DateFrom.IsZero() || !searchQuery.DateTo.IsZero() {
		dateRange := map[string]interface{}{}
		if !searchQuery.DateFrom.IsZero() {
			dateRange["gte"] = searchQuery.DateFrom
		}
		if !searchQuery.DateTo.IsZero() {
			dateRange["lte"] = searchQuery.DateTo
		}

		mustQueries = append(mustQueries, map[string]interface{}{
			"range": map[string]interface{}{
				"published_at": dateRange,
			},
		})
	}

	// Build final query
	var finalQuery map[string]interface{}
	if len(mustQueries) == 0 {
		finalQuery = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	} else if len(mustQueries) == 1 {
		finalQuery = mustQueries[0]
	} else {
		finalQuery = map[string]interface{}{
			"bool": map[string]interface{}{
				"must": mustQueries,
			},
		}
	}

	esQuery := map[string]interface{}{
		"query": finalQuery,
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"title":   map[string]interface{}{},
				"content": map[string]interface{}{},
				"summary": map[string]interface{}{},
			},
		},
		"sort": []map[string]interface{}{
			{
				"published_at": map[string]interface{}{
					"order": "desc",
				},
			},
		},
		"from": from,
		"size": searchQuery.Limit,
	}

	queryJSON, err := json.Marshal(esQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal advanced search query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{r.index},
		Body:  bytes.NewReader(queryJSON),
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute advanced search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("advanced search error: %s", res.String())
	}

	var searchResult map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode advanced search result: %w", err)
	}

	news, total, err := r.parseSearchResult(searchResult)
	if err != nil {
		return nil, err
	}

	return &models.SearchResult{
		News:  news,
		Total: total,
	}, nil
}

func (r *SearchRepository) GetSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	r.logger.Debug().Str("query", query).Int("limit", limit).Msg("Getting search suggestions")

	// Build suggestion query
	suggestQuery := map[string]interface{}{
		"suggest": map[string]interface{}{
			"title_suggest": map[string]interface{}{
				"prefix": query,
				"completion": map[string]interface{}{
					"field": "title.keyword",
					"size":  limit,
				},
			},
		},
	}

	queryJSON, err := json.Marshal(suggestQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal suggestion query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{r.index},
		Body:  bytes.NewReader(queryJSON),
	}

	res, err := req.Do(ctx, r.client)
	if err != nil {
		return nil, fmt.Errorf("failed to execute suggestion query: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("suggestion error: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode suggestion result: %w", err)
	}

	// For now, return a simple implementation
	// In production, you would parse the suggestion response properly
	suggestions := []string{}
	
	// Simple prefix matching fallback
	if strings.TrimSpace(query) != "" {
		suggestions = append(suggestions, query+" news")
		suggestions = append(suggestions, query+" latest")
		suggestions = append(suggestions, query+" update")
	}

	return suggestions, nil
}

func (r *SearchRepository) parseSearchResult(result map[string]interface{}) ([]models.News, int64, error) {
	hits, ok := result["hits"].(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("invalid search result format")
	}

	total, ok := hits["total"].(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("invalid total format")
	}

	totalValue, ok := total["value"].(float64)
	if !ok {
		return nil, 0, fmt.Errorf("invalid total value format")
	}

	documents, ok := hits["hits"].([]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("invalid hits format")
	}

	var news []models.News
	for _, doc := range documents {
		docMap, ok := doc.(map[string]interface{})
		if !ok {
			continue
		}

		source, ok := docMap["_source"].(map[string]interface{})
		if !ok {
			continue
		}

		var n models.News
		n.ID = docMap["_id"].(string)

		if title, ok := source["title"].(string); ok {
			n.Title = title
		}
		if content, ok := source["content"].(string); ok {
			n.Content = content
		}
		if summary, ok := source["summary"].(string); ok {
			n.Summary = summary
		}
		if author, ok := source["author"].(string); ok {
			n.Author = author
		}
		if sourceStr, ok := source["source"].(string); ok {
			n.Source = sourceStr
		}
		if category, ok := source["category"].(string); ok {
			n.Category = category
		}
		if url, ok := source["url"].(string); ok {
			n.URL = url
		}
		if imageURL, ok := source["image_url"].(string); ok {
			n.ImageURL = imageURL
		}

		// Parse tags
		if tags, ok := source["tags"].([]interface{}); ok {
			for _, tag := range tags {
				if tagStr, ok := tag.(string); ok {
					n.Tags = append(n.Tags, tagStr)
				}
			}
		}

		news = append(news, n)
	}

	return news, int64(totalValue), nil
}
