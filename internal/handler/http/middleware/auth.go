package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/davidsugianto/idp-core/internal/pkg/config"

	"github.com/davidsugianto/go-pkgs/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID  string `json:"user_id"`
	TeamID  string `json:"team_id,omitempty"`
	Email   string `json:"email,omitempty"`
	IsAdmin bool   `json:"is_admin,omitempty"`
	jwt.RegisteredClaims
}

func JWT(cfg *config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := extractToken(c)
		if tokenString == "" {
			response.GinUnauthorized(c, fmt.Errorf("authorization header required"))
			c.Abort()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			response.GinUnauthorized(c, err)
			c.Abort()
			return
		}

		if !token.Valid {
			response.GinUnauthorized(c, fmt.Errorf("invalid token"))
			c.Abort()
			return
		}

		// Add user info to context
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "team_id", claims.TeamID)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "is_admin", claims.IsAdmin)
		c.Request = c.Request.WithContext(ctx)

		c.Set("user_id", claims.UserID)
		c.Set("team_id", claims.TeamID)
		c.Set("email", claims.Email)
		c.Set("is_admin", claims.IsAdmin)

		c.Next()
	}
}

// extractToken extracts the JWT from the Authorization header or auth_token cookie.
func extractToken(c *gin.Context) string {
	// Authorization header takes precedence
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// Fallback to cookie
	if cookie, err := c.Cookie("auth_token"); err == nil && cookie != "" {
		return cookie
	}

	return ""
}

func GetTeamID(c *gin.Context) string {
	teamID, exists := c.Get("team_id")
	if !exists {
		return ""
	}
	if str, ok := teamID.(string); ok {
		return str
	}
	return ""
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

func JWTExpiryDuration(cfg *config.AuthConfig) time.Duration {
	if cfg != nil && cfg.JWTExpiry != "" {
		if d, err := time.ParseDuration(cfg.JWTExpiry); err == nil {
			return d
		}
	}

	return 24 * time.Hour
}

// GenerateToken is a helper for testing/development
func GenerateToken(cfg *config.AuthConfig, userID, teamID, email string, isAdmin bool) (string, error) {
	expiresAt := time.Now().Add(JWTExpiryDuration(cfg))

	claims := &Claims{
		UserID:  userID,
		TeamID:  teamID,
		Email:   email,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "idp-core",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(cfg *config.AuthConfig, tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}