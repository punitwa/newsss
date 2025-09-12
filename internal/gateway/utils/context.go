// Package utils provides context management utilities.
package utils

import (
	"fmt"

	"news-aggregator/internal/gateway/core"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

// ContextManager implements context management functionality.
type ContextManager struct {
	logger zerolog.Logger
}

// NewContextManager creates a new context manager.
func NewContextManager(logger zerolog.Logger) core.ContextManager {
	return &ContextManager{
		logger: logger.With().Str("component", "context_manager").Logger(),
	}
}

// GetUserID extracts user ID from context.
func (cm *ContextManager) GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", core.ErrUnauthorized
	}
	
	id, ok := userID.(string)
	if !ok {
		cm.logger.Error().
			Interface("user_id", userID).
			Msg("Invalid user ID type in context")
		return "", core.ErrInternalError
	}
	
	if id == "" {
		return "", core.ErrUnauthorized
	}
	
	return id, nil
}

// SetUserID sets user ID in context.
func (cm *ContextManager) SetUserID(c *gin.Context, userID string) {
	c.Set("user_id", userID)
}

// GetUserRole extracts user role from context.
func (cm *ContextManager) GetUserRole(c *gin.Context) (string, error) {
	role, exists := c.Get("user_role")
	if !exists {
		return "", core.ErrUnauthorized
	}
	
	roleStr, ok := role.(string)
	if !ok {
		cm.logger.Error().
			Interface("user_role", role).
			Msg("Invalid user role type in context")
		return "", core.ErrInternalError
	}
	
	return roleStr, nil
}

// SetUserRole sets user role in context.
func (cm *ContextManager) SetUserRole(c *gin.Context, role string) {
	c.Set("user_role", role)
}

// IsAdmin checks if user is admin.
func (cm *ContextManager) IsAdmin(c *gin.Context) bool {
	role, err := cm.GetUserRole(c)
	if err != nil {
		return false
	}
	
	return role == "admin" || role == "administrator"
}

// GetRequestID gets or generates request ID.
func (cm *ContextManager) GetRequestID(c *gin.Context) string {
	// Try to get existing request ID
	requestID, exists := c.Get("request_id")
	if exists {
		if id, ok := requestID.(string); ok && id != "" {
			return id
		}
	}
	
	// Check for request ID in headers
	headerID := c.GetHeader("X-Request-ID")
	if headerID != "" {
		c.Set("request_id", headerID)
		return headerID
	}
	
	// Generate new request ID
	newID := uuid.New().String()
	c.Set("request_id", newID)
	c.Header("X-Request-ID", newID)
	
	return newID
}

// SetRequestID sets request ID in context.
func (cm *ContextManager) SetRequestID(c *gin.Context, requestID string) {
	c.Set("request_id", requestID)
	c.Header("X-Request-ID", requestID)
}

// GetAuthInfo extracts authentication information from context.
func (cm *ContextManager) GetAuthInfo(c *gin.Context) (*core.AuthInfo, error) {
	userID, err := cm.GetUserID(c)
	if err != nil {
		return nil, err
	}
	
	role, err := cm.GetUserRole(c)
	if err != nil {
		role = "user" // Default role
	}
	
	authInfo := &core.AuthInfo{
		UserID: userID,
		Role:   role,
	}
	
	// Extract additional auth info if available
	if tokenType, exists := c.Get("token_type"); exists {
		if tt, ok := tokenType.(string); ok {
			authInfo.TokenType = tt
		}
	}
	
	if scope, exists := c.Get("scope"); exists {
		if s, ok := scope.([]string); ok {
			authInfo.Scope = s
		}
	}
	
	return authInfo, nil
}

// SetAuthInfo sets authentication information in context.
func (cm *ContextManager) SetAuthInfo(c *gin.Context, authInfo *core.AuthInfo) {
	c.Set("user_id", authInfo.UserID)
	c.Set("user_role", authInfo.Role)
	
	if authInfo.TokenType != "" {
		c.Set("token_type", authInfo.TokenType)
	}
	
	if len(authInfo.Scope) > 0 {
		c.Set("scope", authInfo.Scope)
	}
}

// GetClientIP gets the real client IP address.
func (cm *ContextManager) GetClientIP(c *gin.Context) string {
	// Check for forwarded IP headers
	if ip := c.GetHeader("X-Forwarded-For"); ip != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := len(ip); idx > 0 {
			if commaIdx := 0; commaIdx < idx {
				for i, char := range ip {
					if char == ',' {
						commaIdx = i
						break
					}
				}
				if commaIdx > 0 {
					return ip[:commaIdx]
				}
			}
			return ip
		}
	}
	
	if ip := c.GetHeader("X-Real-IP"); ip != "" {
		return ip
	}
	
	if ip := c.GetHeader("X-Forwarded-Host"); ip != "" {
		return ip
	}
	
	// Fall back to Gin's ClientIP method
	return c.ClientIP()
}

// GetUserAgent gets the user agent string.
func (cm *ContextManager) GetUserAgent(c *gin.Context) string {
	return c.GetHeader("User-Agent")
}

// CreateRequestContext creates a request context with all relevant information.
func (cm *ContextManager) CreateRequestContext(c *gin.Context) *core.RequestContext {
	userID, _ := cm.GetUserID(c)
	userRole, _ := cm.GetUserRole(c)
	
	return &core.RequestContext{
		RequestID: cm.GetRequestID(c),
		UserID:    userID,
		UserRole:  userRole,
		Method:    c.Request.Method,
		Path:      c.Request.URL.Path,
		IP:        cm.GetClientIP(c),
		UserAgent: cm.GetUserAgent(c),
	}
}

// RequireAuth ensures user is authenticated.
func (cm *ContextManager) RequireAuth(c *gin.Context) error {
	_, err := cm.GetUserID(c)
	return err
}

// RequireAdmin ensures user is admin.
func (cm *ContextManager) RequireAdmin(c *gin.Context) error {
	if err := cm.RequireAuth(c); err != nil {
		return err
	}
	
	if !cm.IsAdmin(c) {
		return core.ErrForbidden
	}
	
	return nil
}

// RequireRole ensures user has the specified role.
func (cm *ContextManager) RequireRole(c *gin.Context, requiredRole string) error {
	if err := cm.RequireAuth(c); err != nil {
		return err
	}
	
	role, err := cm.GetUserRole(c)
	if err != nil {
		return err
	}
	
	if role != requiredRole && !cm.IsAdmin(c) {
		return core.NewAuthorizationError(fmt.Sprintf("required role: %s", requiredRole))
	}
	
	return nil
}

// RequireScope ensures user has the specified scope.
func (cm *ContextManager) RequireScope(c *gin.Context, requiredScope string) error {
	if err := cm.RequireAuth(c); err != nil {
		return err
	}
	
	scope, exists := c.Get("scope")
	if !exists {
		return core.ErrForbidden
	}
	
	scopes, ok := scope.([]string)
	if !ok {
		return core.ErrInternalError
	}
	
	for _, s := range scopes {
		if s == requiredScope {
			return nil
		}
	}
	
	return core.NewAuthorizationError(fmt.Sprintf("required scope: %s", requiredScope))
}

// SetCacheHeaders sets appropriate cache headers.
func (cm *ContextManager) SetCacheHeaders(c *gin.Context, maxAge int) {
	if maxAge > 0 {
		c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
	} else {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
	}
}

// SetSecurityHeaders sets security headers.
func (cm *ContextManager) SetSecurityHeaders(c *gin.Context) {
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "DENY")
	c.Header("X-XSS-Protection", "1; mode=block")
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	c.Header("Content-Security-Policy", "default-src 'self'")
}

// SetCORSHeaders sets CORS headers.
func (cm *ContextManager) SetCORSHeaders(c *gin.Context, allowedOrigins []string) {
	origin := c.GetHeader("Origin")
	
	// Check if origin is allowed
	allowed := false
	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			allowed = true
			break
		}
	}
	
	if allowed {
		c.Header("Access-Control-Allow-Origin", origin)
	}
	
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("Access-Control-Max-Age", "3600")
}

// LogRequest logs request information.
func (cm *ContextManager) LogRequest(c *gin.Context) {
	requestCtx := cm.CreateRequestContext(c)
	
	cm.logger.Info().
		Str("request_id", requestCtx.RequestID).
		Str("method", requestCtx.Method).
		Str("path", requestCtx.Path).
		Str("ip", requestCtx.IP).
		Str("user_agent", requestCtx.UserAgent).
		Str("user_id", requestCtx.UserID).
		Msg("Request received")
}
