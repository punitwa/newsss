// Package auth provides authentication-related HTTP handlers that are independent of any gateway.
package auth

import (
	"net/http"

	handlerCore "news-aggregator/internal/handlers/core"
	"news-aggregator/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Handler implements authentication-related operations independently.
type Handler struct {
	deps   *handlerCore.HandlerDependencies
	config handlerCore.HandlerConfig
	logger zerolog.Logger
}

// NewHandler creates a new independent authentication handler.
func NewHandler(deps *handlerCore.HandlerDependencies, config handlerCore.HandlerConfig) handlerCore.AuthHandler {
	return &Handler{
		deps:   deps,
		config: config,
		logger: deps.Logger.With().Str("handler", "auth").Logger(),
	}
}

// RegisterRoutes registers authentication routes.
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	auth := router.Group(h.GetBasePath())
	{
		auth.POST("/login", h.Login)
		auth.POST("/register", h.Register)
		auth.POST("/refresh", h.RefreshToken)
		auth.POST("/logout", h.Logout)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/reset-password", h.ResetPassword)
		auth.GET("/verify-email/:token", h.VerifyEmail)
		auth.POST("/resend-verification", h.ResendVerification)
		auth.GET("/status", h.GetAuthStatus)
	}
}

// GetBasePath returns the base path for authentication routes.
func (h *Handler) GetBasePath() string {
	return "/auth"
}

// GetName returns a unique name for this handler.
func (h *Handler) GetName() string {
	return "auth_handler"
}

// Login handles user login.
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.deps.ResponseWriter.BadRequest(c, "Invalid request format")
		return
	}

	// Validate request if validation is enabled
	if h.config.EnableValidation {
		if err := h.deps.Validator.ValidateLogin(&req); err != nil {
			h.deps.ResponseWriter.Error(c, err)
			return
		}
	}

	// Log request if logging is enabled
	if h.config.EnableLogging {
		h.logger.Info().
			Str("email", req.Email).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Login attempt")
	}

	// Authenticate user
	token, user, err := h.deps.UserService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("email", req.Email).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Login failed")

		h.deps.ResponseWriter.Unauthorized(c, "Invalid credentials")
		return
	}

	// Prepare response
	response := gin.H{
		"token":      token,
		"user":       user,
		"expires_in": h.deps.Config.JWT.ExpirationTime.Seconds(),
		"token_type": "Bearer",
	}

	h.deps.ResponseWriter.Success(c, response)

	if h.config.EnableLogging {
		h.logger.Info().
			Str("email", req.Email).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("User logged in successfully")
	}
}

// Register handles user registration.
func (h *Handler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.deps.ResponseWriter.BadRequest(c, "Invalid request format")
		return
	}

	// Validate request if validation is enabled
	if h.config.EnableValidation {
		if err := h.deps.Validator.ValidateRegistration(&req); err != nil {
			h.deps.ResponseWriter.Error(c, err)
			return
		}
	}

	// Log request if logging is enabled
	if h.config.EnableLogging {
		h.logger.Info().
			Str("email", req.Email).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Registration attempt")
	}

	// Create user
	user, err := h.deps.UserService.Register(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error().
			Err(err).
			Str("email", req.Email).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Registration failed")

		h.deps.ResponseWriter.Error(c, err)
		return
	}

	// Prepare response
	response := gin.H{
		"user":    user,
		"message": "User registered successfully",
	}

	c.JSON(http.StatusCreated, response)

	if h.config.EnableLogging {
		h.logger.Info().
			Str("email", req.Email).
			Str("user_id", user.ID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("User registered successfully")
	}
}

// RefreshToken handles token refresh.
func (h *Handler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.deps.ResponseWriter.BadRequest(c, "Invalid request format")
		return
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Token refresh attempt")
	}

	// TODO: Implement token refresh functionality
	// For now, return an error indicating it's not implemented
	h.logger.Warn().
		Str("request_id", h.deps.ContextManager.GetRequestID(c)).
		Msg("Token refresh not implemented")

	h.deps.ResponseWriter.ErrorWithCode(c, http.StatusNotImplemented, "Token refresh not implemented")
}

// Logout handles user logout.
func (h *Handler) Logout(c *gin.Context) {
	// Extract user ID from context
	userID, err := h.deps.ContextManager.GetUserID(c)
	if err != nil {
		h.deps.ResponseWriter.Unauthorized(c, "Unauthorized")
		return
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Str("user_id", userID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Logout attempt")
	}

	// TODO: Implement token invalidation/blacklist functionality
	// For now, just return success (stateless JWT tokens)

	h.deps.ResponseWriter.Success(c, gin.H{
		"message": "Logged out successfully",
	})

	if h.config.EnableLogging {
		h.logger.Info().
			Str("user_id", userID).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("User logged out successfully")
	}
}

// ForgotPassword handles forgot password requests.
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.deps.ResponseWriter.BadRequest(c, "Invalid request format")
		return
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Str("email", req.Email).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Forgot password request")
	}

	// TODO: Implement forgot password functionality
	// For now, return a generic message for security
	h.deps.ResponseWriter.Success(c, gin.H{
		"message": "If the email exists, a password reset link has been sent",
	})
}

// ResetPassword handles password reset.
func (h *Handler) ResetPassword(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.deps.ResponseWriter.BadRequest(c, "Invalid request format")
		return
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Password reset attempt")
	}

	// TODO: Implement password reset functionality
	h.deps.ResponseWriter.ErrorWithCode(c, http.StatusNotImplemented, "Password reset not implemented")
}

// VerifyEmail handles email verification.
func (h *Handler) VerifyEmail(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		h.deps.ResponseWriter.BadRequest(c, "Verification token is required")
		return
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Str("token", token).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Email verification attempt")
	}

	// TODO: Implement email verification functionality
	h.deps.ResponseWriter.ErrorWithCode(c, http.StatusNotImplemented, "Email verification not implemented")
}

// ResendVerification handles resending verification email.
func (h *Handler) ResendVerification(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.deps.ResponseWriter.BadRequest(c, "Invalid request format")
		return
	}

	if h.config.EnableLogging {
		h.logger.Info().
			Str("email", req.Email).
			Str("request_id", h.deps.ContextManager.GetRequestID(c)).
			Msg("Resend verification attempt")
	}

	// TODO: Implement resend verification functionality
	// For now, return a generic message for security
	h.deps.ResponseWriter.Success(c, gin.H{
		"message": "If the email exists and is not verified, a verification link has been sent",
	})
}

// GetAuthStatus returns the current authentication status.
func (h *Handler) GetAuthStatus(c *gin.Context) {
	userID, err := h.deps.ContextManager.GetUserID(c)
	if err != nil {
		h.deps.ResponseWriter.Unauthorized(c, "Not authenticated")
		return
	}

	role, _ := h.deps.ContextManager.GetUserRole(c)

	h.deps.ResponseWriter.Success(c, gin.H{
		"authenticated": true,
		"user_id":       userID,
		"role":          role,
	})
}
