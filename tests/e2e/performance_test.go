package e2e

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPerformance_TokenGeneration benchmarks JWT token generation
func TestPerformance_TokenGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key-for-performance-testing",
	}

	iterations := 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		_, err := middleware.GenerateToken(authConfig, fmt.Sprintf("user-%d", i), fmt.Sprintf("team-%d", i%10), "", false)
		require.NoError(t, err)
	}

	elapsed := time.Since(start)
	t.Logf("Generated %d tokens in %v (%.2f tokens/sec)", iterations, elapsed, float64(iterations)/elapsed.Seconds())

	// Assert minimum performance: at least 1000 tokens/sec
	assert.GreaterOrEqual(t, float64(iterations)/elapsed.Seconds(), 1000.0, "Token generation should be at least 1000 tokens/sec")
}

// TestPerformance_TokenValidation benchmarks JWT token validation
func TestPerformance_TokenValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key-for-performance-testing",
	}

	// Generate a single token to validate repeatedly
	token, err := middleware.GenerateToken(authConfig, "test-user", "test-team", "", false)
	require.NoError(t, err)

	iterations := 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		_, err := middleware.ValidateToken(authConfig, token)
		require.NoError(t, err)
	}

	elapsed := time.Since(start)
	t.Logf("Validated %d tokens in %v (%.2f validations/sec)", iterations, elapsed, float64(iterations)/elapsed.Seconds())

	// Assert minimum performance: at least 2000 validations/sec
	assert.GreaterOrEqual(t, float64(iterations)/elapsed.Seconds(), 2000.0, "Token validation should be at least 2000 validations/sec")
}

// TestPerformance_AuthMiddleware benchmarks auth middleware overhead
func TestPerformance_AuthMiddleware(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	gin.SetMode(gin.TestMode)

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key-for-performance-testing",
	}

	// Create test router with auth middleware
	router := gin.New()
	router.Use(middleware.JWT(authConfig))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Generate token
	token, err := middleware.GenerateToken(authConfig, "test-user", "test-team", "", false)
	require.NoError(t, err)

	iterations := 1000
	start := time.Now()

	for i := 0; i < iterations; i++ {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	elapsed := time.Since(start)
	t.Logf("Processed %d authenticated requests in %v (%.2f req/sec)", iterations, elapsed, float64(iterations)/elapsed.Seconds())

	// Assert minimum performance: at least 5000 req/sec
	assert.GreaterOrEqual(t, float64(iterations)/elapsed.Seconds(), 5000.0, "Auth middleware should handle at least 5000 req/sec")
}

// TestPerformance_CostQuery benchmarks cost query endpoint
func TestPerformance_CostQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	gin.SetMode(gin.TestMode)

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key-for-performance-testing",
	}

	// Create test router with mock cost data
	router := gin.New()
	router.Use(middleware.JWT(authConfig))

	// Mock cost data (100 records)
	costData := make([]map[string]interface{}, 100)
	for i := 0; i < 100; i++ {
		costData[i] = map[string]interface{}{
			"namespace":    fmt.Sprintf("team-%d-prod", i%10),
			"cost":         float64(i) * 10.5,
			"period_start": "2026-05-01",
			"period_end":   "2026-05-31",
		}
	}

	router.GET("/v1/costs", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"costs": costData,
			"total": len(costData),
		})
	})

	// Generate token
	token, err := middleware.GenerateToken(authConfig, "test-user", "test-team", "", false)
	require.NoError(t, err)

	iterations := 500
	start := time.Now()

	for i := 0; i < iterations; i++ {
		req := httptest.NewRequest("GET", "/v1/costs", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	elapsed := time.Since(start)
	t.Logf("Processed %d cost queries in %v (%.2f req/sec)", iterations, elapsed, float64(iterations)/elapsed.Seconds())

	// Assert minimum performance: at least 3000 req/sec
	assert.GreaterOrEqual(t, float64(iterations)/elapsed.Seconds(), 3000.0, "Cost query should handle at least 3000 req/sec")
}

// TestPerformance_ConcurrentRequests benchmarks concurrent request handling
func TestPerformance_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	gin.SetMode(gin.TestMode)

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key-for-performance-testing",
	}

	// Create test router
	router := gin.New()
	router.Use(middleware.JWT(authConfig))
	router.GET("/v1/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Generate token
	token, err := middleware.GenerateToken(authConfig, "test-user", "test-team", "", false)
	require.NoError(t, err)

	concurrency := 10
	iterationsPerWorker := 100
	totalIterations := concurrency * iterationsPerWorker

	var wg sync.WaitGroup
	wg.Add(concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterationsPerWorker; j++ {
				req := httptest.NewRequest("GET", "/v1/health", nil)
				req.Header.Set("Authorization", "Bearer "+token)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
			}
		}()
	}

	wg.Wait()
	elapsed := time.Since(start)

	t.Logf("Processed %d concurrent requests in %v (%.2f req/sec)", totalIterations, elapsed, float64(totalIterations)/elapsed.Seconds())

	// Assert minimum performance: at least 10000 req/sec under concurrent load
	assert.GreaterOrEqual(t, float64(totalIterations)/elapsed.Seconds(), 10000.0, "Concurrent handling should achieve at least 10000 req/sec")
}

// TestPerformance_BudgetQuery benchmarks budget query endpoint
func TestPerformance_BudgetQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	gin.SetMode(gin.TestMode)

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key-for-performance-testing",
	}

	// Create test router with mock budget data
	router := gin.New()
	router.Use(middleware.JWT(authConfig))

	// Mock budget data (50 budgets)
	budgetData := make([]map[string]interface{}, 50)
	for i := 0; i < 50; i++ {
		budgetData[i] = map[string]interface{}{
			"id":       fmt.Sprintf("budget-%d", i),
			"name":     fmt.Sprintf("Budget %d", i),
			"team_id":  fmt.Sprintf("team-%d", i%10),
			"limit":    float64(i) * 100,
			"period":   "monthly",
			"status":   "active",
		}
	}

	router.GET("/v1/budgets", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"budgets": budgetData,
			"total":   len(budgetData),
		})
	})

	// Generate token
	token, err := middleware.GenerateToken(authConfig, "test-user", "test-team", "", false)
	require.NoError(t, err)

	iterations := 500
	start := time.Now()

	for i := 0; i < iterations; i++ {
		req := httptest.NewRequest("GET", "/v1/budgets", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	elapsed := time.Since(start)
	t.Logf("Processed %d budget queries in %v (%.2f req/sec)", iterations, elapsed, float64(iterations)/elapsed.Seconds())

	// Assert minimum performance: at least 3000 req/sec
	assert.GreaterOrEqual(t, float64(iterations)/elapsed.Seconds(), 3000.0, "Budget query should handle at least 3000 req/sec")
}

// TestPerformance_RightsizingQuery benchmarks rightsizing recommendation query
func TestPerformance_RightsizingQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	gin.SetMode(gin.TestMode)

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key-for-performance-testing",
	}

	// Create test router with mock rightsizing data
	router := gin.New()
	router.Use(middleware.JWT(authConfig))

	// Mock recommendations (200 recommendations)
	recommendations := make([]map[string]interface{}, 200)
	for i := 0; i < 200; i++ {
		recommendations[i] = map[string]interface{}{
			"id":             fmt.Sprintf("rec-%d", i),
			"workload_name":  fmt.Sprintf("workload-%d", i),
			"workload_type":  "Deployment",
			"namespace":      fmt.Sprintf("team-%d-prod", i%10),
			"recommendation": "scale_down",
			"status":         "pending",
			"confidence":     85,
		}
	}

	router.GET("/v1/rightsizing/recommendations", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"recommendations": recommendations,
			"total":           len(recommendations),
		})
	})

	// Generate token
	token, err := middleware.GenerateToken(authConfig, "test-user", "test-team", "", false)
	require.NoError(t, err)

	iterations := 500
	start := time.Now()

	for i := 0; i < iterations; i++ {
		req := httptest.NewRequest("GET", "/v1/rightsizing/recommendations", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	elapsed := time.Since(start)
	t.Logf("Processed %d rightsizing queries in %v (%.2f req/sec)", iterations, elapsed, float64(iterations)/elapsed.Seconds())

	// Assert minimum performance: at least 500 req/sec (larger JSON payload)
	assert.GreaterOrEqual(t, float64(iterations)/elapsed.Seconds(), 500.0, "Rightsizing query should handle at least 500 req/sec")
}
