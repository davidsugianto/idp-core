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
	UserID string `json:"user_id"`
	TeamID string `json:"team_id"`
	jwt.RegisteredClaims
}

func JWT(cfg *config.AuthConfig) gin.HandlerFunc {
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
		c.Request = c.Request.WithContext(ctx)

		c.Set("user_id", claims.UserID)
		c.Set("team_id", claims.TeamID)

		c.Next()
	}
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

// GenerateToken is a helper for testing/development
func GenerateToken(cfg *config.AuthConfig, userID, teamID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		TeamID: teamID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "sentinel-incident",
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
