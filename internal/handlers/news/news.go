// Package news provides news-related HTTP handlers that are independent of any gateway.
package news

import (
	"strconv"
	"time"

	"news-aggregator/internal/handlers/core"
	"news-aggregator/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Handler implements news-related operations independently.
type Handler struct {
	deps   *core.HandlerDependencies
	config core.HandlerConfig
	logger zerolog.Logger
}

// NewHandler creates a new independent news handler.
func NewHandler(deps *core.HandlerDependencies, config core.HandlerConfig) core.NewsHandler {
	return &Handler{
		deps:   deps,
		config: config,
		logger: deps.Logger.With().Str("handler", "news").Logger(),
	}
}

// RegisterRoutes registers news routes.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	news := router.Group(h.GetBasePath())
	{
		news.GET("", h.GetNews)
		news.GET("/:id", h.GetNewsByID)
		news.GET("/categories", h.GetCategories)
		news.GET("/sources", h.GetSources)
		news.GET("/trending", h.GetTrendingTopics)
		news.POST("/search", h.SearchNews)
		news.GET("/search", h.SearchNews) // Support both GET and POST for search
		news.GET("/feed/:category", h.GetNewsByCategory)
		news.GET("/feed/source/:source", h.GetNewsBySource)
		news.GET("/latest", h.GetLatestNews)
		news.GET("/popular", h.GetPopularNews)
		news.GET("/top-stories", h.GetTopStories)
	}
}

// GetBasePath returns the base path for news routes.
func (h *Handler) GetBasePath() string {
	return "/news"
}

// GetName returns a unique name for this handler.
func (h *Handler) GetName() string {
	return "news_handler"
}

// GetNews retrieves paginated news articles.
func (h *Handler) GetNews(c *gin.Context) {
	// Parse and validate query parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(h.config.DefaultPageSize)))
	if err != nil || limit < 1 {
		limit = h.config.DefaultPageSize
	}
	if limit > h.config.MaxPageSize {
		limit = h.config.MaxPageSize
	}

	// Validate pagination if validation is enabled
	if h.config.EnableValidation {
		page, limit, err = h.deps.Validator.ValidatePagination(page, limit)
		if err != nil {
			h.deps.ResponseWriter.BadRequest(c, "Invalid pagination parameters")
			return
		}
	}

	// Build filter
	filter := models.NewsFilter{
		Page:     page,
		Limit:    limit,
		Category: c.Query("category"),
		Source:   c.Query("source"),
		DateFrom: h.parseDateQuery(c.Query("date_from")),
	}

	// Apply default date filter (last 7 days)
	if filter.DateFrom.IsZero() {
		filter.DateFrom = time.Now().AddDate(0, 0, -7)
	}

	// Log request if logging is enabled
	if h.config.EnableLogging {
		h.logger.Info().
			Int("page", page).
			Int("limit", limit).
			Str("category", filter.Category).
			Str("source", filter.Source).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("News request")
	}

	// Fetch news
	news, total, err := h.deps.NewsService.GetNews(c.Request.Context(), filter)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Failed to get news")

		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	// Prepare pagination info
	pagination := core.NewPaginationInfo(page, limit, int64(total))

	h.deps.ResponseWriter.SuccessWithPagination(c, news, pagination)

	if h.config.EnableLogging {
		h.logger.Info().
			Int("page", page).
			Int("limit", limit).
			Int("total", total).
			Str("category", filter.Category).
			Str("source", filter.Source).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("News retrieved successfully")
	}
}

// GetNewsByID retrieves a specific news article.
func (h *Handler) GetNewsByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.deps.ResponseWriter.BadRequest(c, "News ID is required")
		return
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Str("id", id).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("News by ID request")
	}

	// Fetch news by ID
	news, err := h.deps.NewsService.GetNewsByID(c.Request.Context(), id)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("id", id).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Failed to get news by ID")

		h.deps.ResponseWriter.NotFound(c, "News article not found")
		return
	}

	h.deps.ResponseWriter.Success(c, news)

	if h.config.EnableLogging {
		h.logger.Info().
			Str("id", id).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("News article retrieved successfully")
	}
}

// GetCategories retrieves available news categories.
func (h *Handler) GetCategories(c *gin.Context) {
	if h.config.EnableLogging {
		h.logger.Info().
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Categories request")
	}

	// Fetch categories
	categories, err := h.deps.NewsService.GetCategories(c.Request.Context())
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Failed to get categories")

		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	h.deps.ResponseWriter.Success(c, categories)
}

// SearchNews searches for news articles.
func (h *Handler) SearchNews(c *gin.Context) {
	var query string
	var page, limit int
	var err error

	// Handle both GET and POST requests
	if c.Request.Method == "POST" {
		var searchReq struct {
			Query    string `json:"query" binding:"required"`
			Category string `json:"category"`
			Source   string `json:"source"`
			Page     int    `json:"page"`
			Limit    int    `json:"limit"`
		}

		if err := c.ShouldBindJSON(&searchReq); err != nil {
			h.deps.ResponseWriter.BadRequest(c, "Invalid request format")
			return
		}

		query = searchReq.Query
		page = searchReq.Page
		limit = searchReq.Limit
	} else {
		// Parse query parameters for GET request
		query = c.Query("q")
		if query == "" {
			h.deps.ResponseWriter.BadRequest(c, "Query parameter 'q' is required")
			return
		}

		page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ = strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(h.config.DefaultPageSize)))
	}

	// Set defaults and validate
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = h.config.DefaultPageSize
	}
	if limit > h.config.MaxPageSize {
		limit = h.config.MaxPageSize
	}

	// Validate pagination if validation is enabled
	if h.config.EnableValidation {
		page, limit, err = h.deps.Validator.ValidatePagination(page, limit)
		if err != nil {
			h.deps.ResponseWriter.BadRequest(c, "Invalid pagination parameters")
			return
		}
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Str("query", query).
			Int("page", page).
			Int("limit", limit).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Search request")
	}

	// Perform search
	results, total, err := h.deps.SearchService.Search(
		c.Request.Context(),
		query,
		page,
		limit,
	)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("query", query).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Search failed")

		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	// Prepare pagination info
	pagination := core.NewPaginationInfo(page, limit, int64(total))

	h.deps.ResponseWriter.SuccessWithPagination(c, results, pagination)

	if h.config.EnableLogging {
		h.logger.Info().
			Str("query", query).
			Int("results", len(results)).
			Int("total", int(total)).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Search completed successfully")
	}
}

// GetTrendingTopics retrieves trending topics.
func (h *Handler) GetTrendingTopics(c *gin.Context) {
	// Parse limit parameter
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 50 {
		limit = 10
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Int("limit", limit).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Trending topics request")
	}

	// Get trending topics
	topics, err := h.deps.TrendingService.GetTrendingTopics(c.Request.Context(), limit)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Failed to get trending topics")

		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	h.deps.ResponseWriter.Success(c, gin.H{
		"data": topics,
		"meta": gin.H{
			"count":      len(topics),
			"limit":      limit,
			"updated_at": time.Now(),
		},
	})
}

// GetNewsByCategory retrieves news by category.
func (h *Handler) GetNewsByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		h.deps.ResponseWriter.BadRequest(c, "Category is required")
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(h.config.DefaultPageSize)))

	if h.config.EnableValidation {
		page, limit, _ = h.deps.Validator.ValidatePagination(page, limit)
	}

	// Build filter
	filter := models.NewsFilter{
		Page:     page,
		Limit:    limit,
		Category: category,
		DateFrom: time.Now().AddDate(0, 0, -7), // Last 7 days
	}

	// Fetch news
	news, total, err := h.deps.NewsService.GetNews(c.Request.Context(), filter)
	if err != nil {
		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	pagination := core.NewPaginationInfo(page, limit, int64(total))
	h.deps.ResponseWriter.SuccessWithPagination(c, news, pagination)
}

// GetNewsBySource retrieves news by source.
func (h *Handler) GetNewsBySource(c *gin.Context) {
	source := c.Param("source")
	if source == "" {
		h.deps.ResponseWriter.BadRequest(c, "Source is required")
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(h.config.DefaultPageSize)))

	if h.config.EnableValidation {
		page, limit, _ = h.deps.Validator.ValidatePagination(page, limit)
	}

	// Build filter
	filter := models.NewsFilter{
		Page:     page,
		Limit:    limit,
		Source:   source,
		DateFrom: time.Now().AddDate(0, 0, -7), // Last 7 days
	}

	// Fetch news
	news, total, err := h.deps.NewsService.GetNews(c.Request.Context(), filter)
	if err != nil {
		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	pagination := core.NewPaginationInfo(page, limit, int64(total))
	h.deps.ResponseWriter.SuccessWithPagination(c, news, pagination)
}

// GetLatestNews retrieves the latest news articles.
func (h *Handler) GetLatestNews(c *gin.Context) {
	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(h.config.DefaultPageSize)))

	if h.config.EnableValidation {
		page, limit, _ = h.deps.Validator.ValidatePagination(page, limit)
	}

	// Build filter for latest news
	filter := models.NewsFilter{
		Page:     page,
		Limit:    limit,
		DateFrom: time.Now().AddDate(0, 0, -1), // Last 24 hours
	}

	// Fetch latest news
	news, total, err := h.deps.NewsService.GetNews(c.Request.Context(), filter)
	if err != nil {
		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	pagination := core.NewPaginationInfo(page, limit, int64(total))
	h.deps.ResponseWriter.SuccessWithPagination(c, news, pagination)
}

// GetPopularNews retrieves popular news articles.
func (h *Handler) GetPopularNews(c *gin.Context) {
	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(h.config.DefaultPageSize)))

	if h.config.EnableValidation {
		page, limit, _ = h.deps.Validator.ValidatePagination(page, limit)
	}

	// Build filter for popular news
	filter := models.NewsFilter{
		Page:     page,
		Limit:    limit,
		DateFrom: time.Now().AddDate(0, 0, -3), // Last 3 days
	}

	news, total, err := h.deps.NewsService.GetNews(c.Request.Context(), filter)
	if err != nil {
		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	pagination := core.NewPaginationInfo(page, limit, int64(total))
	h.deps.ResponseWriter.SuccessWithPagination(c, news, pagination)
}

// GetTopStories retrieves top stories (simplified version)
func (h *Handler) GetTopStories(c *gin.Context) {
	// Parse pagination
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// For now, use the same logic as GetNews but with a smaller limit
	// This is a temporary implementation until the full scoring service is integrated
	filter := models.NewsFilter{
		Page:     1,
		Limit:    limit,
		DateFrom: time.Now().AddDate(0, 0, -1), // Last 24 hours for top stories
	}

	news, total, err := h.deps.NewsService.GetNews(c.Request.Context(), filter)
	if err != nil {
		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	// Return with enhanced metadata
	h.deps.ResponseWriter.Success(c, map[string]interface{}{
		"data": news,
		"meta": map[string]interface{}{
			"count":     len(news),
			"total":     total,
			"algorithm": "time_based_temporary",
			"timestamp": time.Now(),
		},
	})
}

// GetSources retrieves available news sources.
func (h *Handler) GetSources(c *gin.Context) {
	// This would need to be implemented in the service
	sources := []gin.H{
		{"id": "bbc", "name": "BBC News", "category": "general"},
		{"id": "cnn", "name": "CNN", "category": "general"},
		{"id": "techcrunch", "name": "TechCrunch", "category": "technology"},
	}

	h.deps.ResponseWriter.Success(c, sources)
}

// parseDateQuery parses date query parameter.
func (h *Handler) parseDateQuery(dateStr string) time.Time {
	if dateStr == "" {
		return time.Time{}
	}

	// Try different date formats
	formats := []string{
		"2006-01-02",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	return time.Time{}
}
