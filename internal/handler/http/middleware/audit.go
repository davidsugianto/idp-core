package middleware

import (
	"bytes"
	"time"

	auditlogModel "github.com/davidsugianto/idp-core/internal/model/auditlog"
	auditlogUsecase "github.com/davidsugianto/idp-core/internal/usecase/auditlog"
	"github.com/gin-gonic/gin"
)

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// AuditLog creates middleware that automatically logs API requests to the audit log
func AuditLog(uc auditlogUsecase.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Capture response body
		blw := &responseBodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		duration := time.Since(start)

		// Determine actor info from context
		userID, _ := c.Get("user_id")
		userEmail, _ := c.Get("user_email")
		actorType, _ := c.Get("actor_type")
		teamID, _ := c.Get("team_id")
		apiKeyTeam, _ := c.Get("api_key_team")

		uid := ""
		if str, ok := userID.(string); ok {
			uid = str
		}

		email := ""
		if str, ok := userEmail.(string); ok {
			email = str
		}

		aType := auditlogModel.ActorTypeUser
		if str, ok := actorType.(string); ok && str != "" {
			aType = str
		}

		tID := ""
		if str, ok := teamID.(string); ok {
			tID = str
		}
		// API key team takes precedence if set
		if str, ok := apiKeyTeam.(string); ok && str != "" {
			tID = str
		}

		status := auditlogModel.StatusSuccess
		if c.Writer.Status() >= 400 {
			status = auditlogModel.StatusFailure
		}

		requestID, _ := c.Get("request_id")
		reqID := ""
		if str, ok := requestID.(string); ok {
			reqID = str
		}

		req := auditlogModel.CreateAuditLogRequest{
			UserID:        uid,
			UserEmail:     email,
			ActorType:     aType,
			Action:        c.Request.Method,
			ResourceType:  c.FullPath(),
			ResourceID:    c.Param("id"),
			TeamID:        tID,
			IPAddress:     c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
			RequestMethod: c.Request.Method,
			RequestPath:   c.Request.URL.Path,
			RequestID:     reqID,
			NewValues: auditlogModel.Map{
				"status_code": c.Writer.Status(),
				"duration_ms": duration.Milliseconds(),
				"path":        c.Request.URL.Path,
				"method":      c.Request.Method,
			},
			Status: status,
		}

		if status == auditlogModel.StatusFailure {
			req.ErrorMessage = blw.body.String()
		}

		// Fire and forget — don't block the response on audit logging
		go func() {
			_, _ = uc.Create(c.Copy(), req)
		}()
	}
}