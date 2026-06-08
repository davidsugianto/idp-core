package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/mocks"
	teamModel "github.com/davidsugianto/idp-core/internal/model/team"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/webhook"
	teamUsecase "github.com/davidsugianto/idp-core/internal/usecase/team"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandler_Login(t *testing.T) {
	cfg := &config.AuthConfig{JWTSecret: "test-secret-key"}
	handler := New(Dependencies{
		AuthConfig:       cfg,
		WebhookValidator: webhook.NewValidator(),
	})

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

func TestHandler_ResolveOIDCRedirectURI(t *testing.T) {
	handler := New(Dependencies{
		WebhookValidator: webhook.NewValidator(),
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
	})

	t.Run("accepts allowed origin with path", func(t *testing.T) {
		redirectURI := handler.resolveOIDCRedirectURI("http://localhost:3000/dashboard")
		assert.Equal(t, "http://localhost:3000/dashboard", redirectURI)
	})

	t.Run("falls back to first allowed origin for invalid redirect uri", func(t *testing.T) {
		redirectURI := handler.resolveOIDCRedirectURI("http://evil.example.com")
		assert.Equal(t, "http://localhost:3000", redirectURI)
	})

	t.Run("falls back to first allowed origin when redirect uri is empty", func(t *testing.T) {
		redirectURI := handler.resolveOIDCRedirectURI("")
		assert.Equal(t, "http://localhost:3000", redirectURI)
	})
}

func TestHandler_Login_Integration(t *testing.T) {
	cfg := &config.AuthConfig{JWTSecret: "test-secret-key"}
	handler := New(Dependencies{
		AuthConfig:       cfg,
		WebhookValidator: webhook.NewValidator(),
	})

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

func TestHandler_ResolveOIDCTeamID(t *testing.T) {
	t.Run("returns empty team id when handler has no team usecase", func(t *testing.T) {
		handler := New(Dependencies{WebhookValidator: webhook.NewValidator()})

		teamID, err := handler.resolveOIDCTeamID(context.Background(), "user-123")
		assert.NoError(t, err)
		assert.Empty(t, teamID)
	})

	t.Run("returns empty team id when user has no team memberships", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTeamRepo := mocks.NewMockTeamRepository(ctrl)
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		mockTeamRepo.EXPECT().ListTeamsByUser(gomock.Any(), "user-123").Return([]teamModel.TeamMember{}, nil)

		handler := New(Dependencies{
			TeamUseCase: teamUsecase.New(teamUsecase.Dependencies{
				TeamRepo: mockTeamRepo,
				UserRepo: mockUserRepo,
			}),
			WebhookValidator: webhook.NewValidator(),
		})

		teamID, err := handler.resolveOIDCTeamID(context.Background(), "user-123")
		assert.NoError(t, err)
		assert.Empty(t, teamID)
	})

	t.Run("returns first team id from user memberships", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTeamRepo := mocks.NewMockTeamRepository(ctrl)
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		mockTeamRepo.EXPECT().ListTeamsByUser(gomock.Any(), "user-123").Return([]teamModel.TeamMember{{TeamID: "team-1", UserID: "user-123"}, {TeamID: "team-2", UserID: "user-123"}}, nil)
		mockTeamRepo.EXPECT().GetByID(gomock.Any(), "team-1").Return(&teamModel.Team{ID: "team-1", Name: "Team One"}, nil)
		mockTeamRepo.EXPECT().GetByID(gomock.Any(), "team-2").Return(&teamModel.Team{ID: "team-2", Name: "Team Two"}, nil)

		handler := New(Dependencies{
			TeamUseCase: teamUsecase.New(teamUsecase.Dependencies{
				TeamRepo: mockTeamRepo,
				UserRepo: mockUserRepo,
			}),
			WebhookValidator: webhook.NewValidator(),
		})

		teamID, err := handler.resolveOIDCTeamID(context.Background(), "user-123")
		assert.NoError(t, err)
		assert.Equal(t, "team-1", teamID)
	})

	t.Run("returns repository error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTeamRepo := mocks.NewMockTeamRepository(ctrl)
		mockUserRepo := mocks.NewMockUserRepository(ctrl)
		mockTeamRepo.EXPECT().ListTeamsByUser(gomock.Any(), "user-123").Return(nil, assert.AnError)

		handler := New(Dependencies{
			TeamUseCase: teamUsecase.New(teamUsecase.Dependencies{
				TeamRepo: mockTeamRepo,
				UserRepo: mockUserRepo,
			}),
			WebhookValidator: webhook.NewValidator(),
		})

		teamID, err := handler.resolveOIDCTeamID(context.Background(), "user-123")
		assert.ErrorIs(t, err, assert.AnError)
		assert.Empty(t, teamID)
	})
}
