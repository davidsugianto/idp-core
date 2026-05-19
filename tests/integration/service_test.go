package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/model/service"
	"github.com/davidsugianto/idp-core/internal/model/service_dependency"
	"github.com/davidsugianto/idp-core/internal/model/service_endpoint"
	"github.com/davidsugianto/idp-core/internal/model/service_environment"
	"github.com/davidsugianto/idp-core/internal/model/service_version"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegration_ServiceCatalog tests the service catalog API
func TestIntegration_ServiceCatalog(t *testing.T) {
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

	// Mock service endpoints
	services := router.Group("/v1/services")
	{
		services.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, service.ServiceListResponse{
				Services: []service.ServiceResponse{
					{ID: "svc-1", Name: "api-gateway", TeamID: "team-1", Visibility: "team", Status: "active"},
				},
				Total: 1,
			})
		})
		services.POST("", func(c *gin.Context) {
			var req service.CreateServiceRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, service.ServiceResponse{
				ID:          "svc-new",
				Name:        req.Name,
				Description: req.Description,
				TeamID:      req.TeamID,
				Visibility:  "team",
				Status:      "active",
				CreatedAt:   time.Now().Format(time.RFC3339),
			})
		})
		services.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			if id == "not-found" {
				c.JSON(http.StatusNotFound, gin.H{"error": "service not found"})
				return
			}
			c.JSON(http.StatusOK, service.ServiceResponse{
				ID:        id,
				Name:      "test-service",
				TeamID:    "team-1",
				Visibility: "team",
				Status:    "active",
			})
		})
		services.PATCH("/:id", func(c *gin.Context) {
			id := c.Param("id")
			c.JSON(http.StatusOK, service.ServiceResponse{
				ID:        id,
				Name:      "updated-service",
				TeamID:    "team-1",
				Visibility: "public",
				Status:    "active",
			})
		})
		services.DELETE("/:id", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "service deregistered"})
		})

		// Version routes
		services.GET("/:id/versions", func(c *gin.Context) {
			c.JSON(http.StatusOK, service_version.ServiceVersionListResponse{
				Versions: []service_version.ServiceVersionResponse{
					{ID: "ver-1", ServiceID: c.Param("id"), Version: "1.0.0", Status: "active"},
				},
				Total: 1,
			})
		})
		services.POST("/:id/versions", func(c *gin.Context) {
			var req service_version.CreateServiceVersionRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, service_version.ServiceVersionResponse{
				ID:        "ver-new",
				ServiceID: c.Param("id"),
				Version:   req.Version,
				GitRef:    req.GitRef,
				Status:    "active",
			})
		})

		// Endpoint routes
		services.GET("/:id/versions/:versionId/endpoints", func(c *gin.Context) {
			c.JSON(http.StatusOK, service_endpoint.ServiceEndpointListResponse{
				Endpoints: []service_endpoint.ServiceEndpointResponse{
					{ID: "ep-1", ServiceVersionID: c.Param("versionId"), URL: "https://api.example.com", Type: "http", Status: "active"},
				},
			})
		})
		services.POST("/:id/versions/:versionId/endpoints", func(c *gin.Context) {
			var req service_endpoint.CreateServiceEndpointRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, service_endpoint.ServiceEndpointResponse{
				ID:               "ep-new",
				ServiceVersionID: c.Param("versionId"),
				URL:              req.URL,
				Type:             "http",
				Status:           "active",
			})
		})

		// Dependency routes
		services.GET("/:id/dependencies", func(c *gin.Context) {
			c.JSON(http.StatusOK, service_dependency.DependencyListResponse{
				Dependencies: []service_dependency.DependencyResponse{
					{ID: "dep-1", ServiceID: c.Param("id"), DependsOnServiceID: "svc-2", DependencyType: "runtime"},
				},
				Total: 1,
			})
		})
		services.POST("/:id/dependencies", func(c *gin.Context) {
			var req service_dependency.CreateDependencyRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, service_dependency.DependencyResponse{
				ID:                 "dep-new",
				ServiceID:          c.Param("id"),
				DependsOnServiceID: req.DependsOnServiceID,
				DependencyType:     req.DependencyType,
			})
		})
		services.GET("/:id/dependencies/graph", func(c *gin.Context) {
			c.JSON(http.StatusOK, service_dependency.DependencyGraphResponse{
				ServiceID:   c.Param("id"),
				ServiceName: "test-service",
				Nodes: []service_dependency.GraphNode{
					{ID: c.Param("id"), Name: "test-service", Type: "root"},
					{ID: "svc-2", Name: "database", Type: "dependency"},
				},
				Edges: []service_dependency.GraphEdge{
					{From: c.Param("id"), To: "svc-2", Type: "runtime"},
				},
			})
		})
		services.GET("/:id/dependents", func(c *gin.Context) {
			c.JSON(http.StatusOK, service_dependency.DependencyListResponse{
				Dependencies: []service_dependency.DependencyResponse{},
				Total:        0,
			})
		})

		// Deployment routes
		services.GET("/:id/environments", func(c *gin.Context) {
			c.JSON(http.StatusOK, service_environment.ServiceEnvironmentListResponse{
				Deployments: []service_environment.ServiceEnvironmentResponse{
					{ID: "deploy-1", ServiceVersionID: "ver-1", EnvironmentID: "env-1", Status: "deployed"},
				},
				Total: 1,
			})
		})
		services.POST("/:id/versions/:versionId/deploy", func(c *gin.Context) {
			var req service_environment.DeployRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, service_environment.ServiceEnvironmentResponse{
				ID:               "deploy-new",
				ServiceVersionID: c.Param("versionId"),
				EnvironmentID:    req.EnvironmentID,
				Status:           "deployed",
				DeployedBy:       "test-user",
			})
		})
	}

	// Test cases
	t.Run("list_services", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/services", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp service.ServiceListResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, int64(1), resp.Total)
		assert.Len(t, resp.Services, 1)
	})

	t.Run("create_service", func(t *testing.T) {
		body := service.CreateServiceRequest{
			Name:        "new-service",
			Description: "A new service",
			TeamID:      "team-1",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/services", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp service.ServiceResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "new-service", resp.Name)
	})

	t.Run("get_service", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/services/svc-1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("update_service", func(t *testing.T) {
		body := service.UpdateServiceRequest{
			Name:        ptr("updated-service"),
			Description: ptr("Updated description"),
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("PATCH", "/v1/services/svc-1", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("delete_service", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/v1/services/svc-1", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("list_versions", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/services/svc-1/versions", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("create_version", func(t *testing.T) {
		body := service_version.CreateServiceVersionRequest{
			Version: "2.0.0",
			GitRef:  "main",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/services/svc-1/versions", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("list_endpoints", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/services/svc-1/versions/ver-1/endpoints", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("create_endpoint", func(t *testing.T) {
		body := service_endpoint.CreateServiceEndpointRequest{
			URL:  "https://api.example.com/v2",
			Type: "http",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/services/svc-1/versions/ver-1/endpoints", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("list_dependencies", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/services/svc-1/dependencies", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("create_dependency", func(t *testing.T) {
		body := service_dependency.CreateDependencyRequest{
			DependsOnServiceID: "svc-2",
			DependencyType:     "runtime",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/services/svc-1/dependencies", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("get_dependency_graph", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/services/svc-1/dependencies/graph", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp service_dependency.DependencyGraphResponse
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Len(t, resp.Nodes, 2)
		assert.Len(t, resp.Edges, 1)
	})

	t.Run("list_dependents", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/services/svc-1/dependents", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("list_deployments", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/services/svc-1/environments", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("deploy_version", func(t *testing.T) {
		body := service_environment.DeployRequest{
			EnvironmentID: "env-1",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/v1/services/svc-1/versions/ver-1/deploy", bytes.NewReader(bodyBytes))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized_without_token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/services", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestIntegration_ServiceCatalog_DependencyTypes tests different dependency types
func TestIntegration_ServiceCatalog_DependencyTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	types := []string{"runtime", "build", "data", "api"}

	for _, depType := range types {
		t.Run(depType, func(t *testing.T) {
			assert.True(t, service_dependency.ValidDependencyType(depType), "Dependency type %s should be valid", depType)
		})
	}

	t.Run("invalid_type", func(t *testing.T) {
		assert.False(t, service_dependency.ValidDependencyType("invalid"), "Invalid type should be rejected")
	})
}

// TestIntegration_ServiceCatalog_Visibility tests service visibility levels
func TestIntegration_ServiceCatalog_Visibility(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	visibilities := []string{"public", "team", "private"}

	for _, vis := range visibilities {
		t.Run(vis, func(t *testing.T) {
			assert.True(t, service.ValidVisibility(vis), "Visibility %s should be valid", vis)
		})
	}

	t.Run("invalid_visibility", func(t *testing.T) {
		assert.False(t, service.ValidVisibility("invalid"), "Invalid visibility should be rejected")
	})
}

// TestIntegration_ServiceCatalog_DeploymentStatus tests deployment status values
func TestIntegration_ServiceCatalog_DeploymentStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	statuses := []string{"deployed", "deploying", "failed", "rolled_back"}

	for _, status := range statuses {
		t.Run(status, func(t *testing.T) {
			assert.True(t, service_environment.ValidStatus(status), "Status %s should be valid", status)
		})
	}

	t.Run("invalid_status", func(t *testing.T) {
		assert.False(t, service_environment.ValidStatus("invalid"), "Invalid status should be rejected")
	})
}

// Helper function
func ptr(s string) *string {
	return &s
}
