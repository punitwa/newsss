package news

import (
	"strconv"
	"time"

	"news-aggregator/internal/handlers/core"
	"news-aggregator/internal/services"

	"github.com/gin-gonic/gin"
)

// EnhancedHandler extends the basic news handler with scoring capabilities
type EnhancedHandler struct {
	*Handler
	scoringService *services.ScoringService
}

// NewEnhancedHandler creates a new enhanced news handler with scoring
func NewEnhancedHandler(
	deps *core.HandlerDependencies,
	config core.HandlerConfig,
	scoringService *services.ScoringService,
) *EnhancedHandler {
	baseHandler := NewHandler(deps, config).(*Handler)

	return &EnhancedHandler{
		Handler:        baseHandler,
		scoringService: scoringService,
	}
}

// RegisterRoutes registers enhanced routes including scoring endpoints
func (h *EnhancedHandler) RegisterRoutes(router gin.IRouter) {
	// Register base routes
	h.Handler.RegisterRoutes(router)

	// Add enhanced routes
	news := router.Group(h.GetBasePath())
	{
		// Enhanced top stories with scoring
		news.GET("/top-stories", h.GetEnhancedTopStories)
		news.GET("/top-stories/refresh", h.RefreshTopStories)

		// Engagement tracking endpoints
		news.POST("/:id/track/view", h.TrackView)
		news.POST("/:id/track/click", h.TrackClick)
		news.POST("/:id/track/share", h.TrackShare)
		news.POST("/:id/track/read-time", h.TrackReadTime)

		// Scoring information endpoints
		news.GET("/:id/score", h.GetArticleScore)
		news.GET("/scores/top", h.GetTopScoredArticles)

		// Analytics endpoints
		news.GET("/analytics/engagement", h.GetEngagementAnalytics)
		news.GET("/analytics/sources", h.GetSourceAnalytics)
	}
}

// GetEnhancedTopStories returns top stories using the enhanced algorithm
func (h *EnhancedHandler) GetEnhancedTopStories(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	h.logger.Info().
		Int("limit", limit).
		Str("request_id", h.deps.ContextManager.GetRequestID(c)).
		Msg("Enhanced top stories request")

	// Get top stories using enhanced algorithm
	topStories, err := h.scoringService.CalculateTopStories(c.Request.Context(), limit)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Failed to get enhanced top stories")

		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	h.deps.ResponseWriter.Success(c, map[string]interface{}{
		"data": topStories,
		"meta": map[string]interface{}{
			"count":     len(topStories),
			"algorithm": "enhanced_scoring",
			"timestamp": time.Now(),
		},
	})

	h.logger.Info().
		Int("count", len(topStories)).
		Str("request_id", h.deps.ContextManager.GetRequestID(c)).
		Msg("Enhanced top stories response sent")
}

// RefreshTopStories triggers a refresh of all article scores
func (h *EnhancedHandler) RefreshTopStories(c *gin.Context) {
	h.logger.Info().
		Str("request_id", h.deps.ContextManager.GetRequestID(c)).
		Msg("Top stories refresh requested")

	// Trigger score refresh in background
	go func() {
		if err := h.scoringService.RefreshScores(c.Request.Context()); err != nil {
			h.logger.Error().Err(err).Msg("Failed to refresh scores")
		}
	}()

	h.deps.ResponseWriter.Success(c, map[string]interface{}{
		"message": "Score refresh initiated",
		"status":  "processing",
	})
}

// TrackView records a view event for an article
func (h *EnhancedHandler) TrackView(c *gin.Context) {
	articleID := c.Param("id")
	if articleID == "" {
		h.deps.ResponseWriter.BadRequest(c, "Article ID is required")
		return
	}

	err := h.scoringService.TrackEngagement(c.Request.Context(), articleID, "view", 1)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Str("article_id", articleID).
			Msg("Failed to track view")
		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	h.deps.ResponseWriter.Success(c, map[string]interface{}{
		"message": "View tracked successfully",
	})
}

// TrackClick records a click event for an article
func (h *EnhancedHandler) TrackClick(c *gin.Context) {
	articleID := c.Param("id")
	if articleID == "" {
		h.deps.ResponseWriter.BadRequest(c, "Article ID is required")
		return
	}

	err := h.scoringService.TrackEngagement(c.Request.Context(), articleID, "click", 1)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Str("article_id", articleID).
			Msg("Failed to track click")
		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	h.deps.ResponseWriter.Success(c, map[string]interface{}{
		"message": "Click tracked successfully",
	})
}

// TrackShare records a share event for an article
func (h *EnhancedHandler) TrackShare(c *gin.Context) {
	articleID := c.Param("id")
	if articleID == "" {
		h.deps.ResponseWriter.BadRequest(c, "Article ID is required")
		return
	}

	err := h.scoringService.TrackEngagement(c.Request.Context(), articleID, "share", 1)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Str("article_id", articleID).
			Msg("Failed to track share")
		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	h.deps.ResponseWriter.Success(c, map[string]interface{}{
		"message": "Share tracked successfully",
	})
}

// TrackReadTime records reading time for an article
func (h *EnhancedHandler) TrackReadTime(c *gin.Context) {
	articleID := c.Param("id")
	if articleID == "" {
		h.deps.ResponseWriter.BadRequest(c, "Article ID is required")
		return
	}

	var request struct {
		ReadTime int64 `json:"read_time"` // in seconds
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.deps.ResponseWriter.BadRequest(c, "Invalid request format")
		return
	}

	if request.ReadTime <= 0 {
		h.deps.ResponseWriter.BadRequest(c, "Read time must be positive")
		return
	}

	err := h.scoringService.TrackEngagement(c.Request.Context(), articleID, "read_time", request.ReadTime)
	if err != nil {
		h.logger.Warn().
			Err(err).
			Str("article_id", articleID).
			Int64("read_time", request.ReadTime).
			Msg("Failed to track read time")
		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	h.deps.ResponseWriter.Success(c, map[string]interface{}{
		"message": "Read time tracked successfully",
	})
}

// GetArticleScore returns the comprehensive score for an article
func (h *EnhancedHandler) GetArticleScore(c *gin.Context) {
	articleID := c.Param("id")
	if articleID == "" {
		h.deps.ResponseWriter.BadRequest(c, "Article ID is required")
		return
	}

	// This would require implementing GetArticleScore in the scoring service
	// For now, return a placeholder response
	h.deps.ResponseWriter.Success(c, map[string]interface{}{
		"article_id": articleID,
		"message":    "Score retrieval not yet implemented",
	})
}

// GetTopScoredArticles returns articles with the highest scores
func (h *EnhancedHandler) GetTopScoredArticles(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	minScore := 0.5
	if scoreStr := c.Query("min_score"); scoreStr != "" {
		if parsed, err := strconv.ParseFloat(scoreStr, 64); err == nil {
			minScore = parsed
		}
	}

	// This would require implementing GetTopScoredArticles in the scoring service
	h.deps.ResponseWriter.Success(c, map[string]interface{}{
		"limit":     limit,
		"min_score": minScore,
		"message":   "Top scored articles retrieval not yet implemented",
	})
}

// GetEngagementAnalytics returns engagement analytics
func (h *EnhancedHandler) GetEngagementAnalytics(c *gin.Context) {
	period := c.DefaultQuery("period", "7d") // 1d, 7d, 30d

	h.deps.ResponseWriter.Success(c, map[string]interface{}{
		"period":  period,
		"message": "Engagement analytics not yet implemented",
		"data": map[string]interface{}{
			"total_views":       0,
			"total_clicks":      0,
			"total_shares":      0,
			"average_read_time": 0,
			"top_articles":      []interface{}{},
		},
	})
}

// GetSourceAnalytics returns source credibility analytics
func (h *EnhancedHandler) GetSourceAnalytics(c *gin.Context) {
	h.deps.ResponseWriter.Success(c, map[string]interface{}{
		"message": "Source analytics not yet implemented",
		"data": map[string]interface{}{
			"sources": []interface{}{},
			"credibility_distribution": map[string]int{
				"high":   0,
				"medium": 0,
				"low":    0,
			},
		},
	})
}
