package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthHandler_Login(t *testing.T) {
	cfg := &config.AuthConfig{JWTSecret: "test-secret-key"}
	handler := NewAuthHandler(cfg)

	t.Run("successful login", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := LoginRequest{
			UserID: "user-123",
			TeamID: "team-456",
		}
		jsonBody, _ := json.Marshal(body)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data, ok := response["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.NotEmpty(t, data["token"])
		assert.Equal(t, "Bearer", data["type"])
	})

	t.Run("missing user_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := LoginRequest{
			TeamID: "team-456",
		}
		jsonBody, _ := json.Marshal(body)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing team_id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := LoginRequest{
			UserID: "user-123",
		}
		jsonBody, _ := json.Marshal(body)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("empty body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer([]byte{}))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestAuthHandler_Integration(t *testing.T) {
	cfg := &config.AuthConfig{JWTSecret: "test-secret-key"}
	handler := NewAuthHandler(cfg)

	t.Run("login returns valid token that can be used", func(t *testing.T) {
		// Step 1: Login
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body := LoginRequest{
			UserID: "user-123",
			TeamID: "team-456",
		}
		jsonBody, _ := json.Marshal(body)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		data, ok := response["data"].(map[string]interface{})
		assert.True(t, ok)
		token := data["token"].(string)

		// Step 2: Validate token using middleware
		w2 := httptest.NewRecorder()
		c2, _ := gin.CreateTestContext(w2)
		c2.Request = httptest.NewRequest(http.MethodGet, "/protected", nil)
		c2.Request.Header.Set("Authorization", "Bearer "+token)

		middleware.JWT(cfg)(c2)

		assert.False(t, c2.IsAborted())

		userID, exists := c2.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, "user-123", userID)

		teamID, exists := c2.Get("team_id")
		assert.True(t, exists)
		assert.Equal(t, "team-456", teamID)
	})
}
