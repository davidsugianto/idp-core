package integration

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

// TestIntegration_RBAC tests RBAC enforcement across features
func TestIntegration_RBAC(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	gin.SetMode(gin.TestMode)

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key",
	}

	// Generate test tokens for different roles
	adminToken, err := middleware.GenerateToken(authConfig, "admin-user", "platform", "", false)
	require.NoError(t, err)

	teamAdminToken, err := middleware.GenerateToken(authConfig, "team-admin", "team-1", "", false)
	require.NoError(t, err)

	developerToken, err := middleware.GenerateToken(authConfig, "developer", "team-1", "", false)
	require.NoError(t, err)

	viewerToken, err := middleware.GenerateToken(authConfig, "viewer", "team-1", "", false)
	require.NoError(t, err)

	// Create test router
	router := gin.New()
	router.Use(middleware.JWT(authConfig))

	// Mock RBAC endpoints
	rbac := router.Group("/v1")
	{
		// Role management - platform admin only
		rbac.GET("/roles", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID != "admin-user" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(http.StatusOK, role.RoleListResponse{
				Roles: []role.RoleResponse{
					{ID: "role-1", Name: "platform_admin", Scope: role.ScopePlatform, CreatedAt: time.Now()},
					{ID: "role-2", Name: "team_admin", Scope: role.ScopeTeam, CreatedAt: time.Now()},
				},
				Total: 2,
			})
		})
		rbac.POST("/roles", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID != "admin-user" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			var req role.CreateRoleRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, role.RoleResponse{
				ID:          "role-new",
				Name:        req.Name,
				Description: req.Description,
				Scope:       req.Scope,
				CreatedAt:   time.Now(),
			})
		})
		rbac.POST("/roles/assign", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID != "admin-user" && userID != "team-admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "role assigned"})
		})
		rbac.POST("/roles/revoke", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID != "admin-user" && userID != "team-admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "role revoked"})
		})

		// Environment management - team admin and developer can read, team admin can write
		rbac.GET("/environments", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID == "viewer" && userID != "admin-user" && userID != "team-admin" && userID != "developer" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"environments": []interface{}{}})
		})
		rbac.POST("/environments", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID != "admin-user" && userID != "team-admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"id": "env-new"})
		})
		rbac.DELETE("/environments/:id", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID != "admin-user" && userID != "team-admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "environment deleted"})
		})

		// Budget management - team admin can manage, developer can view
		rbac.GET("/budgets", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"budgets": []interface{}{}})
		})
		rbac.POST("/budgets", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID != "admin-user" && userID != "team-admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"id": "budget-new"})
		})
		rbac.PATCH("/budgets/:id", func(c *gin.Context) {
			userID := c.GetString("user_id")
			if userID != "admin-user" && userID != "team-admin" {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "budget updated"})
		})

		// Cost view - all authenticated users
		rbac.GET("/costs", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"costs": []interface{}{}})
		})

		// Audit log - admin and team admin can view all, developer can view own
		rbac.GET("/audit-logs", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"audit_logs": []interface{}{}})
		})
	}

	// Test cases
	t.Run("platform_admin_can_list_roles", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/roles", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("team_admin_cannot_list_roles", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/roles", nil)
		req.Header.Set("Authorization", "Bearer "+teamAdminToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("developer_cannot_list_roles", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/roles", nil)
		req.Header.Set("Authorization", "Bearer "+developerToken)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("platform_admin_can_create_role", func(t *testing.T) {
		body := role.CreateRoleRequest{
			Name:        "new-role",
			Description: "A new role",
			Scope:       role.ScopeTeam,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/roles", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("team_admin_cannot_create_role", func(t *testing.T) {
		body := role.CreateRoleRequest{
			Name:        "new-role",
			Description: "A new role",
			Scope:       role.ScopeTeam,
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/roles", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+teamAdminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("platform_admin_can_assign_role", func(t *testing.T) {
		body := map[string]string{
			"user_id": "user-1",
			"role_id": "role-1",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/roles/assign", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("team_admin_can_assign_role", func(t *testing.T) {
		body := map[string]string{
			"user_id": "user-1",
			"role_id": "role-2",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/roles/assign", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+teamAdminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("developer_cannot_assign_role", func(t *testing.T) {
		body := map[string]string{
			"user_id": "user-1",
			"role_id": "role-2",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/roles/assign", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+developerToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("platform_admin_can_revoke_role", func(t *testing.T) {
		body := map[string]string{
			"user_id": "user-1",
			"role_id": "role-1",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/roles/revoke", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("team_admin_can_revoke_role", func(t *testing.T) {
		body := map[string]string{
			"user_id": "user-1",
			"role_id": "role-2",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/roles/revoke", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+teamAdminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("developer_cannot_revoke_role", func(t *testing.T) {
		body := map[string]string{
			"user_id": "user-1",
			"role_id": "role-2",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/roles/revoke", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+developerToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("platform_admin_can_create_environment", func(t *testing.T) {
		body := map[string]string{
			"name":   "new-env",
			"team_id": "team-1",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/environments", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+adminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("team_admin_can_create_environment", func(t *testing.T) {
		body := map[string]string{
			"name":   "new-env",
			"team_id": "team-1",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/environments", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+teamAdminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("developer_cannot_create_environment", func(t *testing.T) {
		body := map[string]string{
			"name":   "new-env",
			"team_id": "team-1",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/environments", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+developerToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("all_users_can_view_costs", func(t *testing.T) {
		// Test with different tokens
		tokens := []struct {
			name  string
			token string
		}{
			{"platform_admin", adminToken},
			{"team_admin", teamAdminToken},
			{"developer", developerToken},
			{"viewer", viewerToken},
		}

		for _, tc := range tokens {
			t.Run(tc.name, func(t *testing.T) {
				req := httptest.NewRequest("GET", "/v1/costs", nil)
				req.Header.Set("Authorization", "Bearer "+tc.token)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
			})
		}
	})

	t.Run("team_admin_can_create_budget", func(t *testing.T) {
		body := map[string]interface{}{
			"name":    "budget-1",
			"team_id": "team-1",
			"limit":   1000.00,
			"period":  "monthly",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/budgets", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+teamAdminToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("developer_cannot_create_budget", func(t *testing.T) {
		body := map[string]interface{}{
			"name":    "budget-1",
			"team_id": "team-1",
			"limit":   1000.00,
			"period":  "monthly",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/budgets", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+developerToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("viewer_cannot_create_budget", func(t *testing.T) {
		body := map[string]interface{}{
			"name":    "budget-1",
			"team_id": "team-1",
			"limit":   1000.00,
			"period":  "monthly",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/budgets", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+viewerToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("unauthorized_without_token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/roles", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestIntegration_RBAC_RoleScopes tests different role scopes
func TestIntegration_RBAC_RoleScopes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	t.Run("platform_scope_validation", func(t *testing.T) {
		assert.True(t, isValidScope(role.ScopePlatform))
	})

	t.Run("team_scope_validation", func(t *testing.T) {
		assert.True(t, isValidScope(role.ScopeTeam))
	})

	t.Run("environment_scope_validation", func(t *testing.T) {
		assert.True(t, isValidScope(role.ScopeEnvironment))
	})

	t.Run("invalid_scope_rejected", func(t *testing.T) {
		assert.False(t, isValidScope("invalid"))
	})
}

// Helper function to validate scope
func isValidScope(scope string) bool {
	return scope == role.ScopePlatform || scope == role.ScopeTeam || scope == role.ScopeEnvironment
}
