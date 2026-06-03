package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/model/role"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestE2E_Phase2_UserJourney tests the complete user journey:
// Login → Create Environment → View Costs → Create Budget
func TestE2E_Phase2_UserJourney(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	gin.SetMode(gin.TestMode)

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key",
	}

	// Generate token for developer
	developerToken, err := middleware.GenerateToken(authConfig, "developer-1", "team-1", "", false)
	require.NoError(t, err)

	// Generate token for team admin
	adminToken, err := middleware.GenerateToken(authConfig, "team-admin-1", "team-1", "", false)
	require.NoError(t, err)

	// Create test router
	router := gin.New()
	router.Use(middleware.JWT(authConfig))

	// Mock handlers for user journey
	v1 := router.Group("/v1")
	{
		// Environment endpoints
		v1.GET("/environments", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"environments": []map[string]interface{}{
					{"id": "env-1", "name": "development", "team_id": "team-1"},
				},
				"total": 1,
			})
		})
		v1.POST("/environments", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID != "team-admin-1" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{
				"id":        "env-new",
				"team_id":   "team-1",
				"created_at": time.Now(),
			})
		})

		// Cost endpoints
		v1.GET("/costs", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"costs": []map[string]interface{}{
					{"namespace": "team-1-dev", "cost": 150.50, "period": "2026-05"},
				},
				"total": 150.50,
			})
		})

		// Budget endpoints
		v1.POST("/budgets", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID != "team-admin-1" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"id": "budget-1", "message": "budget created"})
		})
		v1.GET("/budgets", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"budgets": []map[string]interface{}{
					{"id": "budget-1", "name": "Dev Budget", "team_id": "team-1", "limit": 1000.00, "period": "monthly"},
				},
				"total": 1,
			})
		})
	}

	// Step 1: Login (token already generated)
	t.Run("step1_login", func(t *testing.T) {
		claims, err := middleware.ValidateToken(authConfig, developerToken)
		require.NoError(t, err)
		assert.Equal(t, "developer-1", claims.UserID)
		assert.Equal(t, "team-1", claims.TeamID)
		t.Log("✓ User logged in successfully")
	})

	// Step 2: View environments
	t.Run("step2_view_environments", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/environments", nil)
		req.Header.Set("Authorization", "Bearer "+developerToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		t.Log("✓ Retrieved environments list")
	})

	// Step 3: Admin creates environment
	t.Run("step3_create_environment", func(t *testing.T) {
		body := map[string]string{
			"name":         "staging",
			"git_repo_url": "https://github.com/org/repo.git",
			"manifest_path": "k8s/staging",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/environments", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		t.Log("✓ Admin created new environment")
	})

	// Step 4: View costs
	t.Run("step4_view_costs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/costs", nil)
		req.Header.Set("Authorization", "Bearer "+developerToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NotEmpty(t, resp["costs"])
		t.Log("✓ Viewed team costs")
	})

	// Step 5: Admin creates budget
	t.Run("step5_create_budget", func(t *testing.T) {
		body := map[string]interface{}{
			"name":    "Staging Budget",
			"team_id": "team-1",
			"limit":   2000.00,
			"period":  "monthly",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/budgets", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		t.Log("✓ Admin created budget")
	})

	t.Log("E2E User Journey completed successfully!")
}

// TestE2E_Phase2_AdminJourney tests the admin journey:
// Manage Roles → Set Budgets → View Audit Logs
func TestE2E_Phase2_AdminJourney(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	gin.SetMode(gin.TestMode)

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key",
	}

	// Generate tokens for different roles
	platformAdminToken, err := middleware.GenerateToken(authConfig, "platform-admin", "platform", "", false)
	require.NoError(t, err)

	teamAdminToken, err := middleware.GenerateToken(authConfig, "team-admin", "team-1", "", false)
	require.NoError(t, err)

	// Create test router
	router := gin.New()
	router.Use(middleware.JWT(authConfig))

	v1 := router.Group("/v1")
	{
		// Role management
		v1.GET("/roles", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID != "platform-admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(http.StatusOK, role.RoleListResponse{
				Roles: []role.RoleResponse{
					{ID: "role-1", Name: "platform_admin", Scope: role.ScopePlatform, CreatedAt: time.Now()},
					{ID: "role-2", Name: "team_admin", Scope: role.ScopeTeam, CreatedAt: time.Now()},
					{ID: "role-3", Name: "developer", Scope: role.ScopeTeam, CreatedAt: time.Now()},
				},
				Total: 3,
			})
		})
		v1.POST("/roles", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID != "platform-admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"id": "role-new", "message": "role created"})
		})
		v1.POST("/roles/assign", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID != "platform-admin" && userID != "team-admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "role assigned"})
		})

		// Budget management
		v1.POST("/budgets", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID != "platform-admin" && userID != "team-admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"id": "budget-new"})
		})

		// Audit logs
		v1.GET("/audit-logs", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"audit_logs": []map[string]interface{}{
					{"action": "role.assign", "user_id": "platform-admin", "resource_type": "role"},
					{"action": "budget.create", "user_id": "team-admin", "resource_type": "budget"},
				},
				"total": 2,
			})
		})
	}

	// Step 1: Platform admin lists roles
	t.Run("step1_list_roles", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/roles", nil)
		req.Header.Set("Authorization", "Bearer "+platformAdminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp role.RoleListResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, int64(3), resp.Total)
		t.Log("✓ Platform admin listed roles")
	})

	// Step 2: Platform admin creates role
	t.Run("step2_create_role", func(t *testing.T) {
		body := role.CreateRoleRequest{
			Name:        "viewer",
			Description: "Read-only access",
			Scope:       role.ScopeTeam,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/roles", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+platformAdminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		t.Log("✓ Platform admin created role")
	})

	// Step 3: Team admin assigns role
	t.Run("step3_assign_role", func(t *testing.T) {
		body := map[string]string{
			"user_id": "user-123",
			"role_id": "role-3",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/roles/assign", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+teamAdminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		t.Log("✓ Team admin assigned role")
	})

	// Step 4: Team admin creates budget
	t.Run("step4_create_budget", func(t *testing.T) {
		body := map[string]interface{}{
			"name":    "Production Budget",
			"team_id": "team-1",
			"limit":   5000.00,
			"period":  "monthly",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/budgets", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+teamAdminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		t.Log("✓ Team admin created budget")
	})

	// Step 5: View audit logs
	t.Run("step5_view_audit_logs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/audit-logs", nil)
		req.Header.Set("Authorization", "Bearer "+platformAdminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.NotEmpty(t, resp["audit_logs"])
		t.Log("✓ Viewed audit logs")
	})

	t.Log("E2E Admin Journey completed successfully!")
}

// TestE2E_Phase2_ServiceCatalogJourney tests the service catalog workflow:
// Register Service → Add Version → Add Dependencies → Deploy
func TestE2E_Phase2_ServiceCatalogJourney(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	gin.SetMode(gin.TestMode)

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key",
	}

	token, err := middleware.GenerateToken(authConfig, "developer-1", "team-1", "", false)
	require.NoError(t, err)

	// Create test router
	router := gin.New()
	router.Use(middleware.JWT(authConfig))

	v1 := router.Group("/v1")
	{
		// Service endpoints
		v1.GET("/services", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"services": []map[string]interface{}{
					{"id": "svc-1", "name": "user-service", "team_id": "team-1", "visibility": "team"},
				},
				"total": 1,
			})
		})
		v1.POST("/services", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"id": "svc-new"})
		})
		v1.GET("/services/:id/versions", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"versions": []map[string]interface{}{
					{"id": "ver-1", "version": "1.0.0"},
					{"id": "ver-2", "version": "1.1.0"},
				},
			})
		})
		v1.POST("/services/:id/versions", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"id": "ver-new"})
		})
		v1.GET("/services/:id/dependencies", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"dependencies": []map[string]interface{}{
					{"depends_on_service_id": "svc-db", "dependency_type": "runtime"},
				},
			})
		})
		v1.POST("/services/:id/dependencies", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"message": "dependency added"})
		})
		v1.POST("/services/:id/versions/:versionId/deploy", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"deployment_id": "deploy-1", "status": "deploying"})
		})
	}

	// Step 1: Register service
	t.Run("step1_register_service", func(t *testing.T) {
		body := map[string]interface{}{
			"name":        "payment-service",
			"description": "Payment processing service",
			"team_id":     "team-1",
			"visibility":  "team",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/services", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		t.Log("✓ Registered new service")
	})

	// Step 2: Add version
	t.Run("step2_add_version", func(t *testing.T) {
		body := map[string]string{
			"version":   "1.0.0",
			"git_ref":   "refs/tags/v1.0.0",
			"changelog": "Initial release",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/services/svc-new/versions", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		t.Log("✓ Added service version")
	})

	// Step 3: Add dependency
	t.Run("step3_add_dependency", func(t *testing.T) {
		body := map[string]string{
			"depends_on_service_id": "svc-db",
			"dependency_type":       "runtime",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/services/svc-new/dependencies", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		t.Log("✓ Added service dependency")
	})

	// Step 4: Deploy version
	t.Run("step4_deploy_version", func(t *testing.T) {
		body := map[string]string{
			"environment_id": "env-production",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/services/svc-new/versions/ver-new/deploy", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		t.Log("✓ Deployed version to environment")
	})

	t.Log("E2E Service Catalog Journey completed successfully!")
}

// TestE2E_Phase2_RightsizingJourney tests the rightsizing workflow:
// List Recommendations → Apply → Verify
func TestE2E_Phase2_RightsizingJourney(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	gin.SetMode(gin.TestMode)

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key",
	}

	token, err := middleware.GenerateToken(authConfig, "team-admin-1", "team-1", "", false)
	require.NoError(t, err)

	// Create test router
	router := gin.New()
	router.Use(middleware.JWT(authConfig))

	v1 := router.Group("/v1")
	{
		v1.GET("/rightsizing/recommendations", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"recommendations": []map[string]interface{}{
					{
						"id":              "rec-1",
						"workload_name":   "api-server",
						"workload_type":   "Deployment",
						"namespace":       "team-1-prod",
						"recommendation":  "scale_down",
						"status":          "pending",
						"confidence":      85,
					},
				},
			})
		})
		v1.GET("/rightsizing/recommendations/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"id":             c.Param("id"),
				"workload_name":  "api-server",
				"status":         "pending",
				"confidence":     85,
			})
		})
		v1.POST("/rightsizing/recommendations/:id/apply", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"id":      c.Param("id"),
				"status":  "applied",
				"message": "Recommendation applied successfully",
			})
		})
	}

	// Step 1: List recommendations
	t.Run("step1_list_recommendations", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/rightsizing/recommendations?status=pending", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		recs := resp["recommendations"].([]interface{})
		assert.NotEmpty(t, recs)
		t.Log("✓ Listed pending recommendations")
	})

	// Step 2: Get recommendation details
	t.Run("step2_get_recommendation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/rightsizing/recommendations/rec-1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		t.Log("✓ Retrieved recommendation details")
	})

	// Step 3: Apply recommendation
	t.Run("step3_apply_recommendation", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/v1/rightsizing/recommendations/rec-1/apply", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "applied", resp["status"])
		t.Log("✓ Applied recommendation")
	})

	t.Log("E2E Rightsizing Journey completed successfully!")
}

// TestE2E_Phase2_QuotaEnforcementJourney tests resource quota enforcement:
// Create Quota → Deploy Pod (allowed) → Deploy Pod (exceeded)
func TestE2E_Phase2_QuotaEnforcementJourney(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	gin.SetMode(gin.TestMode)

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key",
	}

	token, err := middleware.GenerateToken(authConfig, "team-admin-1", "team-1", "", false)
	require.NoError(t, err)

	// Create test router
	router := gin.New()
	router.Use(middleware.JWT(authConfig))

	v1 := router.Group("/v1")
	{
		v1.POST("/quotas", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"id": "quota-1"})
		})
		v1.GET("/quotas/namespace/:namespace", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"id":                "quota-1",
				"namespace":         c.Param("namespace"),
				"cpu_request_limit": "4",
				"memory_request_limit": "8Gi",
				"pod_count_limit":   50,
			})
		})
		v1.GET("/quotas/namespace/:namespace/usage", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"cpu_request":     "3.5",
				"memory_request":  "7Gi",
				"pod_count":       45,
			})
		})
		v1.POST("/quotas/check", func(c *gin.Context) {
			var req map[string]interface{}
			c.ShouldBindJSON(&req)

			// Simulate quota check
			cpuRequest := req["cpu_request"].(float64)
			if cpuRequest > 0.5 {
				c.JSON(http.StatusOK, gin.H{
					"allowed": false,
					"reason":  "cpu_request would exceed limit",
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"allowed": true,
				"message": "request within limits",
			})
		})
	}

	// Step 1: Create quota
	t.Run("step1_create_quota", func(t *testing.T) {
		body := map[string]interface{}{
			"namespace":          "team-1-prod",
			"team_id":            "team-1",
			"cpu_request_limit":  "4",
			"memory_request_limit": "8Gi",
			"pod_count_limit":    50,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/quotas", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		t.Log("✓ Created resource quota")
	})

	// Step 2: View usage
	t.Run("step2_view_usage", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/quotas/namespace/team-1-prod/usage", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		t.Log("✓ Viewed current usage")
	})

	// Step 3: Check quota (allowed)
	t.Run("step3_check_allowed", func(t *testing.T) {
		body := map[string]interface{}{
			"namespace":    "team-1-prod",
			"cpu_request":  0.3,
			"memory_request": "512Mi",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/quotas/check", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, true, resp["allowed"])
		t.Log("✓ Quota check passed (within limits)")
	})

	// Step 4: Check quota (exceeded)
	t.Run("step4_check_exceeded", func(t *testing.T) {
		body := map[string]interface{}{
			"namespace":    "team-1-prod",
			"cpu_request":  0.6,
			"memory_request": "1Gi",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/quotas/check", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["allowed"])
		t.Log("✓ Quota check rejected (would exceed)")
	})

	t.Log("E2E Quota Enforcement Journey completed successfully!")
}
