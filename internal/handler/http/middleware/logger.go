package middleware

import (
	"time"

	"github.com/davidsugianto/go-pkgs/logger"
	"github.com/gin-gonic/gin"
)

func Logger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", status).
			Str("latency", latency.String()).
			Str("clientIP", c.ClientIP()).
			Msg("incoming request")
	}
}
