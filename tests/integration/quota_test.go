package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/model/resourcequota"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_Quota tests the resource quota API
func TestIntegration_Quota(t *testing.T) {
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

	// Mock quota endpoints
	quotas := router.Group("/v1/quotas")
	{
		quotas.GET("", func(c *gin.Context) {
			teamID := c.Query("team_id")
			_ = teamID // Filter param
			c.JSON(http.StatusOK, resourcequota.ResourceQuotaListResponse{
				Quotas: []resourcequota.ResourceQuotaResponse{
					{
						ID:                 "quota-1",
						Namespace:          "default",
						TeamID:             "team-1",
						CPURequestLimit:    "4",
						MemoryRequestLimit: "8Gi",
						Enforce:            true,
						Status:             resourcequota.StatusActive,
						CreatedAt:          time.Now().Format(time.RFC3339),
					},
				},
				Total: 1,
			})
		})
		quotas.POST("", func(c *gin.Context) {
			var req resourcequota.CreateResourceQuotaRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, resourcequota.ResourceQuotaResponse{
				ID:                 "quota-new",
				Namespace:          req.Namespace,
				TeamID:             req.TeamID,
				EnvironmentID:      req.EnvironmentID,
				CPURequestLimit:    req.CPURequestLimit,
				MemoryRequestLimit: req.MemoryRequestLimit,
				Enforce:            req.Enforce,
				Status:             resourcequota.StatusActive,
				CreatedAt:          time.Now().Format(time.RFC3339),
			})
		})
		quotas.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			if id == "not-found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "quota not found"})
				return
			}
			c.JSON(http.StatusOK, resourcequota.ResourceQuotaResponse{
				ID:                 id,
				Namespace:          "default",
				TeamID:             "team-1",
				CPURequestLimit:    "4",
				MemoryRequestLimit: "8Gi",
				Enforce:            true,
				Status:             resourcequota.StatusActive,
				CreatedAt:          time.Now().Format(time.RFC3339),
			})
		})
		quotas.GET("/namespace/:namespace", func(c *gin.Context) {
			namespace := c.Param("namespace")
			if namespace == "not-found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "quota not found"})
				return
			}
			c.JSON(http.StatusOK, resourcequota.ResourceQuotaResponse{
				ID:                 "quota-1",
				Namespace:          namespace,
				TeamID:             "team-1",
				CPURequestLimit:    "4",
				MemoryRequestLimit: "8Gi",
				Enforce:            true,
				Status:             resourcequota.StatusActive,
				CreatedAt:          time.Now().Format(time.RFC3339),
			})
		})
		quotas.PATCH("/:id", func(c *gin.Context) {
			id := c.Param("id")
			var req resourcequota.UpdateResourceQuotaRequest
			_ = c.ShouldBindJSON(&req)

			resp := resourcequota.ResourceQuotaResponse{
				ID:                 id,
				Namespace:          "default",
				TeamID:             "team-1",
				CPURequestLimit:    "8",
				MemoryRequestLimit: "16Gi",
				Enforce:            true,
				Status:             resourcequota.StatusActive,
				CreatedAt:          time.Now().Format(time.RFC3339),
			}
			if req.CPURequestLimit != nil {
				resp.CPURequestLimit = *req.CPURequestLimit
			}
			if req.MemoryRequestLimit != nil {
				resp.MemoryRequestLimit = *req.MemoryRequestLimit
			}
			if req.Enforce != nil {
				resp.Enforce = *req.Enforce
			}
			c.JSON(http.StatusOK, resp)
		})
		quotas.DELETE("/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "quota deleted"})
		})
		quotas.GET("/namespace/:namespace/usage", func(c *gin.Context) {
			namespace := c.Param("namespace")
			c.JSON(http.StatusOK, resourcequota.UsageResponse{
				Namespace:     namespace,
				CPURequest:    "2",
				MemoryRequest: "4Gi",
				PodCount:      5,
				LastUpdated:   time.Now().Format(time.RFC3339),
			})
		})
		quotas.POST("/namespace/:namespace/usage/refresh", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "usage refreshed"})
		})
		quotas.POST("/check", func(c *gin.Context) {
			var req resourcequota.QuotaCheckRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Simple mock logic: allow if CPU request is small
			allowed := req.CPURequest == "" || req.CPURequest == "100m" || req.CPURequest == "500m"

			resp := resourcequota.QuotaCheckResponse{
				Allowed: allowed,
			}
			if !allowed {
				resp.Reasons = []resourcequota.QuotaExceededReason{
					{
						ResourceType: "cpu_request",
						Requested:    req.CPURequest,
						Limit:        "4",
						Current:      "3.5",
						Utilization:  87.5,
					},
				}
			}
			c.JSON(http.StatusOK, resp)
		})
	}

	// Test cases
	t.Run("list_quotas", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/quotas", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp resourcequota.ResourceQuotaListResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, int64(1), resp.Total)
		assert.Len(t, resp.Quotas, 1)
	})

	t.Run("create_quota", func(t *testing.T) {
		podLimit := 10
		body := resourcequota.CreateResourceQuotaRequest{
			Namespace:          "production",
			TeamID:             "team-1",
			CPURequestLimit:    "8",
			MemoryRequestLimit: "16Gi",
			PodCountLimit:      &podLimit,
			Enforce:            true,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/quotas", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp resourcequota.ResourceQuotaResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "production", resp.Namespace)
		assert.Equal(t, "8", resp.CPURequestLimit)
	})

	t.Run("get_quota", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/quotas/quota-1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp resourcequota.ResourceQuotaResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "quota-1", resp.ID)
		assert.Equal(t, "default", resp.Namespace)
	})

	t.Run("get_quota_not_found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/quotas/not-found", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("get_quota_by_namespace", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/quotas/namespace/default", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp resourcequota.ResourceQuotaResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "default", resp.Namespace)
	})

	t.Run("get_quota_by_namespace_not_found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/quotas/namespace/not-found", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("update_quota", func(t *testing.T) {
		newCPU := "8"
		newMemory := "16Gi"
		body := resourcequota.UpdateResourceQuotaRequest{
			CPURequestLimit:    &newCPU,
			MemoryRequestLimit: &newMemory,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/quotas/quota-1", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp resourcequota.ResourceQuotaResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "8", resp.CPURequestLimit)
		assert.Equal(t, "16Gi", resp.MemoryRequestLimit)
	})

	t.Run("update_quota_enforce_flag", func(t *testing.T) {
		enforce := false
		body := resourcequota.UpdateResourceQuotaRequest{
			Enforce: &enforce,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/quotas/quota-1", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp resourcequota.ResourceQuotaResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.False(t, resp.Enforce)
	})

	t.Run("delete_quota", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/v1/quotas/quota-1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("get_usage", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/quotas/namespace/default/usage", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp resourcequota.UsageResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "default", resp.Namespace)
		assert.Equal(t, 5, resp.PodCount)
	})

	t.Run("refresh_usage", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/quotas/namespace/default/usage/refresh", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "usage refreshed", resp["message"])
	})

	t.Run("check_quota_allowed", func(t *testing.T) {
		body := resourcequota.QuotaCheckRequest{
			Namespace:  "default",
			CPURequest: "500m",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/quotas/check", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp resourcequota.QuotaCheckResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.Allowed)
	})

	t.Run("check_quota_exceeded", func(t *testing.T) {
		body := resourcequota.QuotaCheckRequest{
			Namespace:  "default",
			CPURequest: "2000m",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/quotas/check", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp resourcequota.QuotaCheckResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.False(t, resp.Allowed)
		assert.Len(t, resp.Reasons, 1)
		assert.Equal(t, "cpu_request", resp.Reasons[0].ResourceType)
	})

	t.Run("check_quota_empty_request", func(t *testing.T) {
		body := resourcequota.QuotaCheckRequest{
			Namespace: "default",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/quotas/check", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp resourcequota.QuotaCheckResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.Allowed)
	})

	t.Run("unauthorized_without_token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/quotas", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestIntegration_Quota_Status tests quota status values
func TestIntegration_Quota_Status(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	statuses := []string{resourcequota.StatusActive, resourcequota.StatusInactive, resourcequota.StatusExceeded}

	for _, s := range statuses {
		t.Run(s, func(t *testing.T) {
			assert.True(t, resourcequota.ValidStatus(s), "Status %s should be valid", s)
		})
	}

	t.Run("invalid_status", func(t *testing.T) {
		assert.False(t, resourcequota.ValidStatus("invalid"), "Invalid status should be rejected")
	})
}

// TestIntegration_Quota_ResourceLimits tests quota with different resource limits
func TestIntegration_Quota_ResourceLimits(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	gin.SetMode(gin.TestMode)

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key",
	}

	token, err := middleware.GenerateToken(authConfig, "test-user", "test-team")
	require.NoError(t, err)

	router := gin.New()
	router.Use(middleware.JWT(authConfig))

	quotas := router.Group("/v1/quotas")
	quotas.POST("", func(c *gin.Context) {
		var req resourcequota.CreateResourceQuotaRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, resourcequota.ResourceQuotaResponse{
			ID:                    "quota-new",
			Namespace:             req.Namespace,
			TeamID:                req.TeamID,
			CPURequestLimit:       req.CPURequestLimit,
			MemoryRequestLimit:    req.MemoryRequestLimit,
			StorageRequestLimit:   req.StorageRequestLimit,
			PodCountLimit:         req.PodCountLimit,
			ConfigMapCountLimit:   req.ConfigMapCountLimit,
			SecretCountLimit:      req.SecretCountLimit,
			PVCCountLimit:         req.PVCCountLimit,
			Enforce:               req.Enforce,
			Status:                resourcequota.StatusActive,
			CreatedAt:             time.Now().Format(time.RFC3339),
		})
	})

	t.Run("create_with_all_limits", func(t *testing.T) {
		podLimit := 100
		configMapLimit := 50
		secretLimit := 50
		pvcLimit := 10

		body := resourcequota.CreateResourceQuotaRequest{
			Namespace:            "production",
			TeamID:               "team-1",
			CPURequestLimit:      "16",
			MemoryRequestLimit:   "32Gi",
			StorageRequestLimit:  "100Gi",
			PodCountLimit:        &podLimit,
			ConfigMapCountLimit:  &configMapLimit,
			SecretCountLimit:     &secretLimit,
			PVCCountLimit:        &pvcLimit,
			Enforce:              true,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/quotas", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp resourcequota.ResourceQuotaResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "16", resp.CPURequestLimit)
		assert.Equal(t, "32Gi", resp.MemoryRequestLimit)
		assert.Equal(t, "100Gi", resp.StorageRequestLimit)
		assert.Equal(t, 100, *resp.PodCountLimit)
	})

	t.Run("create_minimal", func(t *testing.T) {
		body := resourcequota.CreateResourceQuotaRequest{
			Namespace: "minimal",
			TeamID:    "team-1",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/quotas", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestIntegration_Quota_PodDelta tests quota check with pod delta
func TestIntegration_Quota_PodDelta(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	gin.SetMode(gin.TestMode)

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key",
	}

	token, err := middleware.GenerateToken(authConfig, "test-user", "test-team")
	require.NoError(t, err)

	router := gin.New()
	router.Use(middleware.JWT(authConfig))

	quotas := router.Group("/v1/quotas")
	quotas.POST("/check", func(c *gin.Context) {
		var req resourcequota.QuotaCheckRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Mock: allow if pod delta <= 2
		allowed := req.PodDelta <= 2

		c.JSON(http.StatusOK, resourcequota.QuotaCheckResponse{
			Allowed: allowed,
		})
	})

	t.Run("pod_delta_create", func(t *testing.T) {
		body := resourcequota.QuotaCheckRequest{
			Namespace: "default",
			PodDelta:  1, // Creating one pod
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/quotas/check", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp resourcequota.QuotaCheckResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.Allowed)
	})

	t.Run("pod_delta_delete", func(t *testing.T) {
		body := resourcequota.QuotaCheckRequest{
			Namespace: "default",
			PodDelta:  -1, // Deleting one pod
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/quotas/check", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp resourcequota.QuotaCheckResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.Allowed) // Negative delta should always be allowed
	})

	t.Run("pod_delta_scale_up", func(t *testing.T) {
		body := resourcequota.QuotaCheckRequest{
			Namespace: "default",
			PodDelta:  5, // Scaling up by 5 pods
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/quotas/check", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp resourcequota.QuotaCheckResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.False(t, resp.Allowed) // Exceeds mock limit
	})
}
