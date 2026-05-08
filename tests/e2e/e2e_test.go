package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/handler/http/middleware"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/davidsugianto/idp-core/internal/pkg/kubernetes"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// skipIfNoK8s skips the test if Kubernetes is not available
func skipIfNoK8s(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	_, err := kubernetes.NewClient(false, "")
	if err != nil {
		t.Skipf("Kubernetes not available: %v", err)
	}
}

// skipIfNoArgoCD skips the test if ArgoCD is not available
func skipIfNoArgoCD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Check if ArgoCD namespace exists
	client, err := kubernetes.NewClient(false, "")
	if err != nil {
		t.Skipf("Kubernetes not available: %v", err)
	}

	ctx := context.Background()
	_, err = client.Clientset.CoreV1().Namespaces().Get(ctx, "argocd", metav1.GetOptions{})
	if err != nil {
		t.Skip("ArgoCD not installed. Run 'make dev-k8s-setup' first.")
	}
}

// TestE2E_HealthCheck tests the health endpoint
func TestE2E_HealthCheck(t *testing.T) {
	skipIfNoK8s(t)

	// Create a test server
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

// TestE2E_FullEnvironmentFlow tests the complete environment lifecycle
// This test requires both Kubernetes and ArgoCD to be available
func TestE2E_FullEnvironmentFlow(t *testing.T) {
	skipIfNoArgoCD(t)

	ctx := context.Background()
	k8sClient, err := kubernetes.NewClient(false, "")
	require.NoError(t, err, "Failed to create Kubernetes client")

	// Generate unique test namespace
	namespace := fmt.Sprintf("idp-e2e-test-%d", time.Now().UnixNano())

	// Cleanup function
	cleanup := func() {
		// Delete namespace
		_ = k8sClient.Clientset.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	}
	defer cleanup()

	// Step 1: Create namespace (simulating environment creation)
	t.Log("Step 1: Creating namespace...")
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"idp-core/managed-by":  "idp-core",
				"idp-core/test":        "true",
				"idp-core/environment": "e2e-test",
			},
		},
	}
	_, err = k8sClient.Clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	require.NoError(t, err, "Failed to create namespace")

	// Step 2: Verify namespace exists
	t.Log("Step 2: Verifying namespace exists...")
	_, err = k8sClient.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	require.NoError(t, err, "Failed to get namespace")

	// Step 3: Create a test pod in the namespace
	t.Log("Step 3: Creating test pod...")
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: namespace,
			Labels: map[string]string{
				"app": "test",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "test",
					Image: "nginx:alpine",
				},
			},
		},
	}
	_, err = k8sClient.Clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	require.NoError(t, err, "Failed to create pod")

	// Step 4: Verify pod exists
	t.Log("Step 4: Verifying pod exists...")
	_, err = k8sClient.Clientset.CoreV1().Pods(namespace).Get(ctx, "test-pod", metav1.GetOptions{})
	require.NoError(t, err, "Failed to get pod")

	// Step 5: Delete namespace (cleanup)
	t.Log("Step 5: Deleting namespace...")
	err = k8sClient.Clientset.CoreV1().Namespaces().Delete(ctx, namespace, metav1.DeleteOptions{})
	require.NoError(t, err, "Failed to delete namespace")

	// Wait for namespace deletion
	t.Log("Waiting for namespace deletion...")
	for i := 0; i < 30; i++ {
		_, err = k8sClient.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
		if err != nil {
			t.Log("Namespace deleted successfully")
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	t.Log("E2E test completed successfully!")
}

// TestE2E_AuthFlow tests the authentication flow
func TestE2E_AuthFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Test JWT token generation and validation
	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key",
	}

	// Generate token
	token, err := middleware.GenerateToken(authConfig, "test-user", "test-team")
	require.NoError(t, err, "Failed to generate token")
	assert.NotEmpty(t, token, "Token should not be empty")

	// Validate token
	claims, err := middleware.ValidateToken(authConfig, token)
	require.NoError(t, err, "Failed to validate token")
	assert.Equal(t, "test-user", claims.UserID)
	assert.Equal(t, "test-team", claims.TeamID)
}

// TestE2E_APIAuthentication tests API authentication middleware
func TestE2E_APIAuthentication(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	authConfig := &config.AuthConfig{
		JWTSecret: "test-secret-key",
	}

	// Create test router with auth middleware
	router := gin.New()
	router.Use(middleware.JWT(authConfig))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test without token
	t.Run("missing_token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test with invalid token
	t.Run("invalid_token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	// Test with valid token
	t.Run("valid_token", func(t *testing.T) {
		token, _ := middleware.GenerateToken(authConfig, "test-user", "test-team")
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestE2E_NamespaceNaming tests namespace naming conventions
func TestE2E_NamespaceNaming(t *testing.T) {
	tests := []struct {
		name      string
		teamSlug  string
		envSlug   string
		maxLength int
	}{
		{
			name:      "normal names",
			teamSlug:  "team-a",
			envSlug:   "dev",
			maxLength: 63,
		},
		{
			name:      "long names",
			teamSlug:  "very-long-team-name-that-exceeds-limit",
			envSlug:   "development-environment",
			maxLength: 63,
		},
		{
			name:      "special characters",
			teamSlug:  "Team_A",
			envSlug:   "Dev-Env",
			maxLength: 63,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			namespace := fmt.Sprintf("idp-%s-%s", tt.teamSlug, tt.envSlug)

			// Convert to lowercase
			namespace = strings.ToLower(namespace)

			// Replace underscores with hyphens
			namespace = strings.ReplaceAll(namespace, "_", "-")

			// Remove any characters that aren't lowercase alphanumeric or hyphens
			reg := regexp.MustCompile("[^a-z0-9-]")
			namespace = reg.ReplaceAllString(namespace, "")

			// Truncate if too long
			if len(namespace) > tt.maxLength {
				namespace = namespace[:tt.maxLength]
			}

			// Ensure it doesn't end with a hyphen
			namespace = strings.TrimRight(namespace, "-")

			assert.LessOrEqual(t, len(namespace), tt.maxLength)
			assert.Regexp(t, "^[a-z0-9]([a-z0-9-]*[a-z0-9])?$", namespace)
		})
	}
}
