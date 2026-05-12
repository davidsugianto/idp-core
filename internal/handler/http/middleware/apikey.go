package middleware

import (
	"context"
	"fmt"
	"strings"

	apikeyUsecase "github.com/davidsugianto/idp-core/internal/usecase/apikey"

	"github.com/davidsugianto/go-pkgs/response"
	"github.com/gin-gonic/gin"
)

const (
	ActorTypeAPIKey string = "api_key"
)

// APIKeyAuth creates middleware that validates API key authentication
func APIKeyAuth(uc apikeyUsecase.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := extractAPIKey(c)
		if apiKey == "" {
			response.GinUnauthorized(c, fmt.Errorf("api key required"))
			c.Abort()
			return
		}

		key, err := uc.Validate(c.Request.Context(), apiKey)
		if err != nil {
			response.GinUnauthorized(c, err)
			c.Abort()
			return
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, contextKey("actor_type"), ActorTypeAPIKey)
		ctx = context.WithValue(ctx, contextKey("api_key_id"), key.ID)
		ctx = context.WithValue(ctx, contextKey("api_key_team"), key.TeamID)
		c.Request = c.Request.WithContext(ctx)

		c.Set("actor_type", ActorTypeAPIKey)
		c.Set("api_key_id", key.ID)
		c.Set("api_key_team", key.TeamID)

		c.Next()
	}
}

func extractAPIKey(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		authHeader = c.GetHeader("X-API-Key")
		if authHeader != "" {
			return authHeader
		}
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) == 2 && parts[0] == "Bearer" {
		return parts[1]
	}

	return ""
}

// GetAPIKeyID returns the API key ID from context
func GetAPIKeyID(c *gin.Context) string {
	id, _ := c.Get("api_key_id")
	if str, ok := id.(string); ok {
		return str
	}
	return ""
}

// GetAPIKeyTeam returns the API key's team from context
func GetAPIKeyTeam(c *gin.Context) string {
	team, _ := c.Get("api_key_team")
	if str, ok := team.(string); ok {
		return str
	}
	return ""
}