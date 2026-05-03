package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}

		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, RequestIDKey, reqID)
		c.Request = c.Request.WithContext(ctx)

		c.Writer.Header().Set("X-Request-ID", reqID)

		c.Next()
	}
}
