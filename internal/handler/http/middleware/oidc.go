package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/oidc"
	authUsecase "github.com/davidsugianto/idp-core/internal/usecase/auth"
	"github.com/davidsugianto/go-pkgs/response"
	"github.com/gin-gonic/gin"
)

// OIDCConfig holds OIDC middleware configuration
type OIDCConfig struct {
	OIDCCfg     *config.OIDCConfig
	OIDCClient  *oidc.Client
	OIDCVerifier *oidc.Verifier
	RBACEngine  *authUsecase.RBACEngine
}

// OIDCAuth creates an OIDC authentication middleware
func OIDCAuth(cfg *OIDCConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.GinUnauthorized(c, fmt.Errorf("authorization header required"))
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.GinUnauthorized(c, fmt.Errorf("invalid authorization header format"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Verify token with OIDC provider
		userInfo, err := cfg.OIDCVerifier.VerifyAndExtract(c.Request.Context(), tokenString)
		if err != nil {
			response.GinUnauthorized(c, fmt.Errorf("invalid token: %w", err))
			c.Abort()
			return
		}

		// Add user info to context
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "user_id", userInfo.Subject)
		ctx = context.WithValue(ctx, "email", userInfo.Email)
		ctx = context.WithValue(ctx, "groups", userInfo.Groups)
		c.Request = c.Request.WithContext(ctx)

		c.Set("user_id", userInfo.Subject)
		c.Set("email", userInfo.Email)
		c.Set("groups", userInfo.Groups)
		c.Set("is_admin", cfg.OIDCVerifier.IsAdmin(userInfo))

		c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}
	if str, ok := userID.(string); ok {
		return str
	}
	return ""
}

// GetUserEmail extracts user email from context
func GetUserEmail(c *gin.Context) string {
	email, exists := c.Get("email")
	if !exists {
		return ""
	}
	if str, ok := email.(string); ok {
		return str
	}
	return ""
}

// GetUserGroups extracts user groups from context
func GetUserGroups(c *gin.Context) []string {
	groups, exists := c.Get("groups")
	if !exists {
		return nil
	}
	if arr, ok := groups.([]string); ok {
		return arr
	}
	return nil
}

// IsAdmin checks if user is platform admin
func IsAdmin(c *gin.Context) bool {
	isAdmin, exists := c.Get("is_admin")
	if !exists {
		return false
	}
	if b, ok := isAdmin.(bool); ok {
		return b
	}
	return false
}

// RequirePermission creates a middleware that checks for specific permission
func RequirePermission(rbac *authUsecase.RBACEngine, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			response.GinUnauthorized(c, fmt.Errorf("user not authenticated"))
			c.Abort()
			return
		}

		// Platform admins bypass permission checks
		if IsAdmin(c) {
			c.Next()
			return
		}

		// Check permission
		teamID := GetTeamID(c)
		var hasPermission bool
		var err error

		if teamID != "" {
			hasPermission, err = rbac.CheckTeamPermission(c.Request.Context(), userID, teamID, resource, action)
		} else {
			hasPermission, err = rbac.CheckPermission(c.Request.Context(), userID, resource, action)
		}

		if err != nil {
			response.GinInternalServerError(c, err)
			c.Abort()
			return
		}

		if !hasPermission {
			response.GinUnauthorized(c, fmt.Errorf("permission denied: %s:%s", resource, action))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireTeamPermission creates a middleware that checks for team-specific permission
func RequireTeamPermission(rbac *authUsecase.RBACEngine, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		teamID := GetTeamID(c)

		if userID == "" {
			response.GinUnauthorized(c, fmt.Errorf("user not authenticated"))
			c.Abort()
			return
		}

		if teamID == "" {
			response.GinBadRequest(c, fmt.Errorf("team_id required"))
			c.Abort()
			return
		}

		// Platform admins bypass permission checks
		if IsAdmin(c) {
			c.Next()
			return
		}

		hasPermission, err := rbac.CheckTeamPermission(c.Request.Context(), userID, teamID, resource, action)
		if err != nil {
			response.GinInternalServerError(c, err)
			c.Abort()
			return
		}

		if !hasPermission {
			response.GinUnauthorized(c, fmt.Errorf("permission denied: %s:%s", resource, action))
			c.Abort()
			return
		}

		c.Next()
	}
}
