package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsugianto/go-pkgs/logger"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRequestID(t *testing.T) {
	t.Run("generates request ID if not provided", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

		RequestID()(c)

		reqID := c.Writer.Header().Get("X-Request-ID")
		assert.NotEmpty(t, reqID)

		ctxReqID := c.Request.Context().Value(RequestIDKey)
		assert.Equal(t, reqID, ctxReqID)
	})

	t.Run("uses provided request ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("X-Request-ID", "test-request-id-123")

		RequestID()(c)

		reqID := c.Writer.Header().Get("X-Request-ID")
		assert.Equal(t, "test-request-id-123", reqID)

		ctxReqID := c.Request.Context().Value(RequestIDKey)
		assert.Equal(t, "test-request-id-123", ctxReqID)
	})
}

func TestJWT(t *testing.T) {
	cfg := &config.AuthConfig{JWTSecret: "test-secret-key"}

	t.Run("missing authorization header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

		JWT(cfg)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid authorization header format", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("Authorization", "InvalidFormat")

		JWT(cfg)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("Authorization", "Bearer invalid-token")

		JWT(cfg)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid token", func(t *testing.T) {
		// Generate a valid token
		token, err := GenerateToken(cfg, "user-123", "team-456")
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("Authorization", "Bearer "+token)

		JWT(cfg)(c)

		assert.False(t, c.IsAborted())

		// Check context values
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, "user-123", userID)

		teamID, exists := c.Get("team_id")
		assert.True(t, exists)
		assert.Equal(t, "team-456", teamID)

		// Check request context
		ctx := c.Request.Context()
		assert.Equal(t, "user-123", ctx.Value("user_id"))
		assert.Equal(t, "team-456", ctx.Value("team_id"))
	})

	t.Run("wrong secret key", func(t *testing.T) {
		// Generate token with different secret
		differentCfg := &config.AuthConfig{JWTSecret: "different-secret"}
		token, err := GenerateToken(differentCfg, "user-123", "team-456")
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("Authorization", "Bearer "+token)

		JWT(cfg)(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestGetTeamID(t *testing.T) {
	t.Run("returns team ID when set", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("team_id", "team-123")

		teamID := GetTeamID(c)
		assert.Equal(t, "team-123", teamID)
	})

	t.Run("returns empty string when not set", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		teamID := GetTeamID(c)
		assert.Empty(t, teamID)
	})

	t.Run("returns empty string when type is wrong", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("team_id", 123) // wrong type

		teamID := GetTeamID(c)
		assert.Empty(t, teamID)
	})
}

func TestGenerateToken(t *testing.T) {
	cfg := &config.AuthConfig{JWTSecret: "test-secret-key"}

	t.Run("generates valid token", func(t *testing.T) {
		token, err := GenerateToken(cfg, "user-123", "team-456")
		assert.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("generated token can be validated", func(t *testing.T) {
		token, err := GenerateToken(cfg, "user-123", "team-456")
		assert.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Request.Header.Set("Authorization", "Bearer "+token)

		JWT(cfg)(c)

		assert.False(t, c.IsAborted())
	})
}

func TestLogger(t *testing.T) {
	log := logger.New()

	t.Run("logs request", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)
		c.Request.Header.Set("X-Real-IP", "192.168.1.1")

		Logger(log)(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("logs request with status code from handler", func(t *testing.T) {
		router := gin.New()
		router.Use(Logger(log))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}
