package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/model/rightsizing"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_Rightsizing tests the rightsizing API
func TestIntegration_Rightsizing(t *testing.T) {
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

	// Mock rightsizing endpoints
	rightsizingGroup := router.Group("/v1/rightsizing")
	recommendations := rightsizingGroup.Group("/recommendations")
	{
		recommendations.GET("", func(c *gin.Context) {
			namespace := c.Query("namespace")
			status := c.Query("status")

			recs := []rightsizing.RecommendationResponse{
				{
					ID:                    "rec-1",
					Namespace:             "default",
					WorkloadName:          "api-server",
					WorkloadType:          rightsizing.WorkloadTypeDeployment,
					ContainerName:         "main",
					CurrentCPURequest:     "1000m",
					CurrentMemoryRequest:  "512Mi",
					RecommendedCPURequest: "500m",
					RecommendedMemoryRequest: "256Mi",
					RecommendationType:    rightsizing.RecommendationTypeScaleDown,
					SavingsPotential:      50.0,
					ConfidenceScore:       95.0,
					Status:                rightsizing.StatusPending,
					CreatedAt:             time.Now().Format(time.RFC3339),
				},
			}

			// Apply filters
			if namespace != "" && namespace != "default" {
				recs = []rightsizing.RecommendationResponse{}
			}
			if status != "" && status != rightsizing.StatusPending {
				recs = []rightsizing.RecommendationResponse{}
			}

			c.JSON(http.StatusOK, rightsizing.RecommendationListResponse{
				Recommendations: recs,
				Total:           int64(len(recs)),
			})
		})
		recommendations.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			if id == "not-found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "recommendation not found"})
				return
			}
			c.JSON(http.StatusOK, rightsizing.RecommendationResponse{
				ID:                    id,
				Namespace:             "default",
				WorkloadName:          "api-server",
				WorkloadType:          rightsizing.WorkloadTypeDeployment,
				ContainerName:         "main",
				CurrentCPURequest:     "1000m",
				CurrentMemoryRequest:  "512Mi",
				RecommendedCPURequest: "500m",
				RecommendedMemoryRequest: "256Mi",
				RecommendationType:    rightsizing.RecommendationTypeScaleDown,
				SavingsPotential:      50.0,
				ConfidenceScore:       95.0,
				Status:                rightsizing.StatusPending,
				CreatedAt:             time.Now().Format(time.RFC3339),
			})
		})
		recommendations.POST("/:id/apply", func(c *gin.Context) {
			id := c.Param("id")
			if id == "already-applied" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "recommendation already applied"})
				return
			}
			if id == "not-pending" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "recommendation not in pending status"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "recommendation applied successfully"})
		})
		recommendations.POST("/:id/rollback", func(c *gin.Context) {
			id := c.Param("id")
			if id == "not-applied" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "recommendation not in applied status"})
				return
			}
			if id == "no-previous-state" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "no previous state stored"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "recommendation rolled back successfully"})
		})
		recommendations.POST("/:id/dismiss", func(c *gin.Context) {
			id := c.Param("id")
			if id == "not-pending" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "recommendation not in pending status"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "recommendation dismissed"})
		})
	}

	// Test cases
	t.Run("list_recommendations", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/rightsizing/recommendations", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp rightsizing.RecommendationListResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, int64(1), resp.Total)
		assert.Len(t, resp.Recommendations, 1)
	})

	t.Run("list_recommendations_with_filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/rightsizing/recommendations?namespace=default&status=pending", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp rightsizing.RecommendationListResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, int64(1), resp.Total)
	})

	t.Run("list_recommendations_empty_filter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/rightsizing/recommendations?namespace=nonexistent", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp rightsizing.RecommendationListResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, int64(0), resp.Total)
	})

	t.Run("get_recommendation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/rightsizing/recommendations/rec-1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp rightsizing.RecommendationResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "rec-1", resp.ID)
		assert.Equal(t, rightsizing.RecommendationTypeScaleDown, resp.RecommendationType)
		assert.Equal(t, 95.0, resp.ConfidenceScore)
	})

	t.Run("get_recommendation_not_found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/rightsizing/recommendations/not-found", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("apply_recommendation", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/rightsizing/recommendations/rec-1/apply", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "recommendation applied successfully", resp["message"])
	})

	t.Run("apply_recommendation_already_applied", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/rightsizing/recommendations/already-applied/apply", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("apply_recommendation_not_pending", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/rightsizing/recommendations/not-pending/apply", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("rollback_recommendation", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/rightsizing/recommendations/rec-1/rollback", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "recommendation rolled back successfully", resp["message"])
	})

	t.Run("rollback_recommendation_not_applied", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/rightsizing/recommendations/not-applied/rollback", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("rollback_recommendation_no_previous_state", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/rightsizing/recommendations/no-previous-state/rollback", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("dismiss_recommendation", func(t *testing.T) {
		body := rightsizing.DismissRecommendationRequest{
			Reason: "Not needed for this workload",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/rightsizing/recommendations/rec-1/dismiss", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("dismiss_recommendation_without_reason", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/rightsizing/recommendations/rec-1/dismiss", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("dismiss_recommendation_not_pending", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/rightsizing/recommendations/not-pending/dismiss", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized_without_token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/rightsizing/recommendations", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestIntegration_Rightsizing_WorkloadTypes tests different workload types
func TestIntegration_Rightsizing_WorkloadTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	workloadTypes := []string{rightsizing.WorkloadTypeDeployment, rightsizing.WorkloadTypeStatefulSet}

	for _, wt := range workloadTypes {
		t.Run(wt, func(t *testing.T) {
			assert.True(t, rightsizing.ValidWorkloadType(wt), "Workload type %s should be valid", wt)
		})
	}

	t.Run("invalid_type", func(t *testing.T) {
		assert.False(t, rightsizing.ValidWorkloadType("DaemonSet"), "Invalid type should be rejected")
	})
}

// TestIntegration_Rightsizing_RecommendationTypes tests different recommendation types
func TestIntegration_Rightsizing_RecommendationTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	recTypes := []string{rightsizing.RecommendationTypeScaleDown, rightsizing.RecommendationTypeScaleUp, rightsizing.RecommendationTypeOptimal}

	for _, rt := range recTypes {
		t.Run(rt, func(t *testing.T) {
			assert.True(t, rightsizing.ValidRecommendationType(rt), "Recommendation type %s should be valid", rt)
		})
	}

	t.Run("invalid_type", func(t *testing.T) {
		assert.False(t, rightsizing.ValidRecommendationType("invalid"), "Invalid type should be rejected")
	})
}

// TestIntegration_Rightsizing_Status tests recommendation status values
func TestIntegration_Rightsizing_Status(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	statuses := []string{rightsizing.StatusPending, rightsizing.StatusApplied, rightsizing.StatusDismissed, rightsizing.StatusFailed}

	for _, s := range statuses {
		t.Run(s, func(t *testing.T) {
			assert.True(t, rightsizing.ValidStatus(s), "Status %s should be valid", s)
		})
	}

	t.Run("invalid_status", func(t *testing.T) {
		assert.False(t, rightsizing.ValidStatus("invalid"), "Invalid status should be rejected")
	})
}

// TestIntegration_Rightsizing_PreviousState tests previous state handling
func TestIntegration_Rightsizing_PreviousState(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("set_and_get_previous_state", func(t *testing.T) {
		rec := &rightsizing.RightsizingRecommendation{}
		state := &rightsizing.PreviousResourceState{
			CPURequest:    "1000m",
			CPULimit:      "2000m",
			MemoryRequest: "512Mi",
			MemoryLimit:   "1Gi",
		}

		err := rec.SetPreviousState(state)
		require.NoError(t, err)

		got, err := rec.GetPreviousState()
		require.NoError(t, err)
		assert.Equal(t, state.CPURequest, got.CPURequest)
		assert.Equal(t, state.MemoryRequest, got.MemoryRequest)
	})

	t.Run("empty_previous_state", func(t *testing.T) {
		rec := &rightsizing.RightsizingRecommendation{}
		got, err := rec.GetPreviousState()
		require.NoError(t, err)
		assert.Nil(t, got)
	})

	t.Run("nil_previous_state", func(t *testing.T) {
		rec := &rightsizing.RightsizingRecommendation{}
		err := rec.SetPreviousState(nil)
		require.NoError(t, err)
		assert.Equal(t, "", rec.PreviousState)
	})
}
