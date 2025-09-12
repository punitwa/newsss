// Package user provides user-related HTTP handlers that are independent of any gateway.
package user

import (
	"strconv"

	handlerCore "news-aggregator/internal/handlers/core"
	"news-aggregator/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Handler implements user-related operations independently.
type Handler struct {
	deps   *handlerCore.HandlerDependencies
	config handlerCore.HandlerConfig
	logger zerolog.Logger
}

// NewHandler creates a new independent user handler.
func NewHandler(deps *handlerCore.HandlerDependencies, config handlerCore.HandlerConfig) handlerCore.UserHandler {
	return &Handler{
		deps:   deps,
		config: config,
		logger: deps.Logger.With().Str("handler", "user").Logger(),
	}
}

// RegisterRoutes registers user routes.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	user := router.Group(h.GetBasePath())
	{
		// Profile endpoints
		user.GET("/profile", h.GetProfile)
		user.PUT("/profile", h.UpdateProfile)

		// Bookmark endpoints
		user.GET("/bookmarks", h.GetBookmarks)
		user.POST("/bookmarks", h.AddBookmark)
		user.DELETE("/bookmarks/:id", h.RemoveBookmark)

		// Preferences endpoint
		user.PUT("/preferences", h.UpdatePreferences)
	}
}

// GetBasePath returns the base path for user routes.
func (h *Handler) GetBasePath() string {
	return "/user"
}

// GetName returns a unique name for this handler.
func (h *Handler) GetName() string {
	return "user_handler"
}

// GetProfile retrieves user profile.
func (h *Handler) GetProfile(c *gin.Context) {
	userID, err := h.deps.ContextManager.GetUserID(c)
	if err != nil {
		h.deps.ResponseWriter.Unauthorized(c, "Unauthorized")
		return
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Str("user_id", userID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Get profile request")
	}

	user, err := h.deps.UserService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", userID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Failed to get user profile")

		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	h.deps.ResponseWriter.Success(c, user)

	if h.config.EnableLogging {
		h.logger.Info().
			Str("user_id", userID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("User profile retrieved successfully")
	}
}

// UpdateProfile updates user profile.
func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, err := h.deps.ContextManager.GetUserID(c)
	if err != nil {
		h.deps.ResponseWriter.Unauthorized(c, "Unauthorized")
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.deps.ResponseWriter.BadRequest(c, "Invalid request format")
		return
	}

	// Validate request if validation is enabled
	if h.config.EnableValidation {
		if err := h.deps.Validator.ValidateUpdateProfileRequest(&req); err != nil {
			h.deps.ResponseWriter.Error(c, err)
			return
		}
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Str("user_id", userID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Update profile request")
	}

	err = h.deps.UserService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", userID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Failed to update user profile")

		h.deps.ResponseWriter.Error(c, err)
		return
	}

	h.deps.ResponseWriter.Success(c, gin.H{
		"message": "Profile updated successfully",
	})

	if h.config.EnableLogging {
		h.logger.Info().
			Str("user_id", userID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("User profile updated successfully")
	}
}

// AddBookmark adds a news bookmark.
func (h *Handler) AddBookmark(c *gin.Context) {
	userID, err := h.deps.ContextManager.GetUserID(c)
	if err != nil {
		h.deps.ResponseWriter.Unauthorized(c, "Unauthorized")
		return
	}

	var req struct {
		ArticleID string `json:"article_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.deps.ResponseWriter.BadRequest(c, "Invalid request format")
		return
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Str("user_id", userID).
			Str("article_id", req.ArticleID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Add bookmark request")
	}

	err = h.deps.UserService.AddBookmark(c.Request.Context(), userID, req.ArticleID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", userID).
			Str("article_id", req.ArticleID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Failed to add bookmark")

		h.deps.ResponseWriter.Error(c, err)
		return
	}

	h.deps.ResponseWriter.Success(c, gin.H{
		"message": "Bookmark added successfully",
	})

	if h.config.EnableLogging {
		h.logger.Info().
			Str("user_id", userID).
			Str("article_id", req.ArticleID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Bookmark added successfully")
	}
}

// GetBookmarks retrieves user bookmarks.
func (h *Handler) GetBookmarks(c *gin.Context) {
	userID, err := h.deps.ContextManager.GetUserID(c)
	if err != nil {
		h.deps.ResponseWriter.Unauthorized(c, "Unauthorized")
		return
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Str("user_id", userID).
			Int("page", page).
			Int("limit", limit).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Get bookmarks request")
	}

	bookmarks, total, err := h.deps.UserService.GetBookmarks(c.Request.Context(), userID, page, limit)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", userID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Failed to get bookmarks")

		h.deps.ResponseWriter.InternalError(c, err)
		return
	}

	// Transform bookmarks to include full news data
	result := make([]gin.H, len(bookmarks))
	for i, bookmark := range bookmarks {
		result[i] = gin.H{
			"id":         bookmark.ID,
			"article_id": bookmark.NewsID,
			"news":       bookmark.News,
			"created_at": bookmark.CreatedAt,
		}
	}

	response := gin.H{
		"data":  result,
		"total": total,
		"page":  page,
		"limit": limit,
	}

	h.deps.ResponseWriter.Success(c, response)

	if h.config.EnableLogging {
		h.logger.Info().
			Str("user_id", userID).
			Int("count", len(bookmarks)).
			Int("total", total).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Bookmarks retrieved successfully")
	}
}

// RemoveBookmark removes a bookmark by article ID.
func (h *Handler) RemoveBookmark(c *gin.Context) {
	userID, err := h.deps.ContextManager.GetUserID(c)
	if err != nil {
		h.deps.ResponseWriter.Unauthorized(c, "Unauthorized")
		return
	}

	articleID := c.Param("id")
	if articleID == "" {
		h.deps.ResponseWriter.BadRequest(c, "Article ID is required")
		return
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Str("user_id", userID).
			Str("article_id", articleID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Remove bookmark request")
	}

	err = h.deps.UserService.RemoveBookmarkByArticle(c.Request.Context(), userID, articleID)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", userID).
			Str("article_id", articleID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Failed to remove bookmark")

		h.deps.ResponseWriter.Error(c, err)
		return
	}

	h.deps.ResponseWriter.Success(c, gin.H{
		"message": "Bookmark removed successfully",
	})

	if h.config.EnableLogging {
		h.logger.Info().
			Str("user_id", userID).
			Str("article_id", articleID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Bookmark removed successfully")
	}
}

// UpdatePreferences updates user preferences.
func (h *Handler) UpdatePreferences(c *gin.Context) {
	userID, err := h.deps.ContextManager.GetUserID(c)
	if err != nil {
		h.deps.ResponseWriter.Unauthorized(c, "Unauthorized")
		return
	}

	var req models.PreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.deps.ResponseWriter.BadRequest(c, "Invalid request format")
		return
	}

	// Validate request if validation is enabled
	if h.config.EnableValidation {
		if err := h.deps.Validator.ValidatePreferencesRequest(&req); err != nil {
			h.deps.ResponseWriter.Error(c, err)
			return
		}
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Str("user_id", userID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Update preferences request")
	}

	err = h.deps.UserService.UpdatePreferences(c.Request.Context(), userID, &req)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("user_id", userID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Failed to update user preferences")

		h.deps.ResponseWriter.Error(c, err)
		return
	}

	h.deps.ResponseWriter.Success(c, gin.H{
		"message": "Preferences updated successfully",
	})

	if h.config.EnableLogging {
		h.logger.Info().
			Str("user_id", userID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("User preferences updated successfully")
	}
}
