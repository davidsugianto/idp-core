package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/model/budget"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_Budget tests the budget API
func TestIntegration_Budget(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	gin.SetMode(gin.TestMode)

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key",
	}

	// Generate test token
	token, err := middleware.GenerateToken(authConfig, "test-user", "test-team")
	require.NoError(t, err)

	// Create test router
	router := gin.New()
	router.Use(middleware.JWT(authConfig))

	// Mock budget endpoints
	budgets := router.Group("/v1/budgets")
	{
		budgets.GET("", func(c *gin.Context) {
			teamID := c.Query("team_id")
			if teamID == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "team_id required"})
				return
			}
			c.JSON(http.StatusOK, budget.BudgetListResponse{
				Budgets: []budget.BudgetResponse{
					{
						ID:              "budget-1",
						TeamID:          teamID,
						Name:            "Monthly Budget",
						Limit:           10000.00,
						Period:          budget.PeriodMonthly,
						AlertThresholds: []int{80, 90, 100},
						AlertChannels:   []string{"slack"},
						Status:          budget.StatusActive,
						CreatedAt:       time.Now().Format(time.RFC3339),
					},
				},
				Total: 1,
			})
		})
		budgets.POST("", func(c *gin.Context) {
			var req budget.CreateBudgetRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, budget.BudgetResponse{
				ID:              "budget-new",
				TeamID:          req.TeamID,
				EnvironmentID:   req.EnvironmentID,
				Name:            req.Name,
				Limit:           req.Limit,
				Period:          req.Period,
				AlertThresholds: req.AlertThresholds,
				AlertChannels:   req.AlertChannels,
				Status:          budget.StatusActive,
				CreatedAt:       time.Now().Format(time.RFC3339),
			})
		})
		budgets.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			if id == "not-found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "budget not found"})
				return
			}
			c.JSON(http.StatusOK, budget.BudgetResponse{
				ID:              id,
				TeamID:          "team-1",
				Name:            "Test Budget",
				Limit:           5000.00,
				Period:          budget.PeriodMonthly,
				AlertThresholds: []int{80, 90, 100},
				AlertChannels:   []string{"slack"},
				Status:          budget.StatusActive,
				CreatedAt:       time.Now().Format(time.RFC3339),
			})
		})
		budgets.PATCH("/:id", func(c *gin.Context) {
			id := c.Param("id")
			var req budget.UpdateBudgetRequest
			_ = c.ShouldBindJSON(&req)

			resp := budget.BudgetResponse{
				ID:              id,
				TeamID:          "team-1",
				Name:            "Updated Budget",
				Limit:           7500.00,
				Period:          budget.PeriodMonthly,
				AlertThresholds: []int{70, 85, 100},
				AlertChannels:   []string{"slack", "email"},
				Status:          budget.StatusActive,
				CreatedAt:       time.Now().Format(time.RFC3339),
			}
			if req.Name != nil {
				resp.Name = *req.Name
			}
			if req.Limit != nil {
				resp.Limit = *req.Limit
			}
			if req.Status != nil {
				resp.Status = *req.Status
			}
			c.JSON(http.StatusOK, resp)
		})
		budgets.DELETE("/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "budget deleted"})
		})
		budgets.GET("/:id/alerts", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, []budget.BudgetAlertResponse{
				{
					ID:           "alert-1",
					BudgetID:     id,
					Threshold:    80,
					CurrentSpend: 8000.00,
					Limit:        10000.00,
					Percentage:   80.0,
					Status:       budget.AlertStatusSent,
					Timestamp:    time.Now().Format(time.RFC3339),
					CreatedAt:    time.Now().Format(time.RFC3339),
				},
			})
		})
	}

	// Test cases
	t.Run("list_budgets", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/budgets?team_id=team-1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp budget.BudgetListResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, int64(1), resp.Total)
		assert.Len(t, resp.Budgets, 1)
		assert.Equal(t, "Monthly Budget", resp.Budgets[0].Name)
	})

	t.Run("list_budgets_missing_team_id", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/budgets", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("create_budget", func(t *testing.T) {
		body := budget.CreateBudgetRequest{
			TeamID:          "team-1",
			Name:            "New Budget",
			Limit:           15000.00,
			Period:          budget.PeriodMonthly,
			AlertThresholds: []int{80, 90, 100},
			AlertChannels:   []string{"slack"},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/budgets", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp budget.BudgetResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "New Budget", resp.Name)
		assert.Equal(t, 15000.00, resp.Limit)
	})

	t.Run("get_budget", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/budgets/budget-1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp budget.BudgetResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "budget-1", resp.ID)
		assert.Equal(t, 5000.00, resp.Limit)
	})

	t.Run("get_budget_not_found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/budgets/not-found", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("update_budget", func(t *testing.T) {
		newLimit := 7500.00
		newName := "Updated Budget Name"
		body := budget.UpdateBudgetRequest{
			Name:  &newName,
			Limit: &newLimit,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/budgets/budget-1", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp budget.BudgetResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, newName, resp.Name)
	})

	t.Run("update_budget_status", func(t *testing.T) {
		newStatus := budget.StatusPaused
		body := budget.UpdateBudgetRequest{
			Status: &newStatus,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/budgets/budget-1", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("delete_budget", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/v1/budgets/budget-1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("list_budget_alerts", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/budgets/budget-1/alerts", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp []budget.BudgetAlertResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, 80, resp[0].Threshold)
		assert.Equal(t, 80.0, resp[0].Percentage)
	})

	t.Run("unauthorized_without_token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/budgets?team_id=team-1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestIntegration_Budget_Periods tests different budget periods
func TestIntegration_Budget_Periods(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	periods := []string{budget.PeriodDaily, budget.PeriodWeekly, budget.PeriodMonthly}

	for _, period := range periods {
		t.Run(period, func(t *testing.T) {
			assert.True(t, budget.ValidPeriod(period), "Period %s should be valid", period)
		})
	}

	t.Run("invalid_period", func(t *testing.T) {
		assert.False(t, budget.ValidPeriod("yearly"), "Invalid period should be rejected")
	})
}

// TestIntegration_Budget_AlertThresholds tests alert threshold validation
func TestIntegration_Budget_AlertThresholds(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("valid_thresholds", func(t *testing.T) {
		thresholds := []int{80, 90, 100}
		formatted := budget.FormatThresholds(thresholds)
		parsed := budget.ParseThresholds(formatted)

		assert.Equal(t, thresholds, parsed)
	})

	t.Run("parse_empty", func(t *testing.T) {
		parsed := budget.ParseThresholds("")
		assert.Nil(t, parsed)
	})

	t.Run("format_single", func(t *testing.T) {
		thresholds := []int{80}
		formatted := budget.FormatThresholds(thresholds)
		assert.Equal(t, "80", formatted)
	})
}

// TestIntegration_Budget_AlertChannels tests alert channel handling
func TestIntegration_Budget_AlertChannels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("valid_channels", func(t *testing.T) {
		channels := []string{"slack", "email", "pagerduty"}
		formatted := budget.FormatChannels(channels)
		parsed := budget.ParseChannels(formatted)

		assert.Equal(t, channels, parsed)
	})

	t.Run("parse_empty", func(t *testing.T) {
		parsed := budget.ParseChannels("")
		assert.Nil(t, parsed)
	})

	t.Run("single_channel", func(t *testing.T) {
		channels := []string{"slack"}
		formatted := budget.FormatChannels(channels)
		assert.Equal(t, `["slack"]`, formatted)
	})
}
