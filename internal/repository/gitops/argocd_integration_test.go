package gitops

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/pkg/argocd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Skip if ArgoCD is not installed
func skipIfNoArgoCD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if kubeconfig exists
	home := os.Getenv("HOME")
	if home == "" {
		t.Skip("HOME environment variable not set")
	}

	kubeconfig := fmt.Sprintf("%s/.kube/config", home)
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		t.Skip("kubeconfig not found, skipping ArgoCD integration test")
	}

	// Try to connect and check if ArgoCD namespace exists
	client, err := argocd.NewClient(false, "")
	if err != nil {
		t.Skipf("Failed to create Kubernetes client: %v", err)
	}

	// Check if ArgoCD namespace exists
	ctx := context.Background()
	_, err = client.DynamicClient.Resource(schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "namespaces",
	}).Get(ctx, "argocd", metav1.GetOptions{})

	if err != nil {
		t.Skip("ArgoCD namespace not found. Run 'make dev-setup' to install ArgoCD.")
	}

	// Check if ArgoCD CRD is installed
	_, err = client.DynamicClient.Resource(schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}).Namespace("argocd").List(ctx, metav1.ListOptions{Limit: 1})

	if err != nil {
		t.Skip("ArgoCD Application CRD not found. ArgoCD may not be fully installed.")
	}
}

func setupArgoCDClient(t *testing.T) *argocd.Client {
	client, err := argocd.NewClient(false, "")
	require.NoError(t, err, "Failed to create ArgoCD client")
	return client
}

func setupGitopsRepo(t *testing.T) Repository {
	client := setupArgoCDClient(t)
	return New(Dependencies{
		ArgoCDClient:    client,
		ArgoCDNamespace: "argocd",
	})
}

func generateTestAppName() string {
	return fmt.Sprintf("test-app-%d", time.Now().UnixNano())
}

func cleanupApplication(t *testing.T, repo Repository, name string) {
	ctx := context.Background()
	err := repo.DeleteApplication(ctx, name)
	if err != nil && !errors.IsNotFound(err) {
		t.Logf("Warning: failed to cleanup application %s: %v", name, err)
	}
}

// TestArgoCDConnection tests basic connectivity to ArgoCD
func TestArgoCDConnection(t *testing.T) {
	skipIfNoArgoCD(t)

	client := setupArgoCDClient(t)

	// Try to list ArgoCD applications
	apps, err := client.DynamicClient.Resource(applicationGVR).Namespace("argocd").List(context.Background(), metav1.ListOptions{})
	require.NoError(t, err, "Failed to list ArgoCD applications")

	t.Logf("Connected to ArgoCD, found %d applications", len(apps.Items))
}

// TestApplicationCRUD tests ArgoCD Application creation and deletion
func TestApplicationCRUD(t *testing.T) {
	skipIfNoArgoCD(t)

	repo := setupGitopsRepo(t)
	ctx := context.Background()
	appName := generateTestAppName()
	namespace := fmt.Sprintf("test-ns-%d", time.Now().UnixNano())

	// Cleanup after test
	defer cleanupApplication(t, repo, appName)

	// Test CreateApplication
	spec := argocd.ApplicationSpec{
		Name:      appName,
		Namespace: namespace,
		RepoURL:   "https://github.com/argoproj/argocd-example-apps.git",
		Revision:  "HEAD",
		Path:      "guestbook",
		ServerURL: "https://kubernetes.default.svc",
	}

	err := repo.CreateApplication(ctx, spec)
	require.NoError(t, err, "Failed to create ArgoCD application")

	// Verify application was created
	client := setupArgoCDClient(t)
	app, err := client.DynamicClient.Resource(applicationGVR).Namespace("argocd").Get(ctx, appName, metav1.GetOptions{})
	require.NoError(t, err, "Failed to get ArgoCD application")
	assert.NotNil(t, app)

	// Verify labels
	labels := app.GetLabels()
	assert.Equal(t, "idp-core", labels["idp-core/managed-by"])

	// Test DeleteApplication
	err = repo.DeleteApplication(ctx, appName)
	require.NoError(t, err, "Failed to delete ArgoCD application")
}

// TestGetApplicationStatus tests fetching application status
func TestGetApplicationStatus(t *testing.T) {
	skipIfNoArgoCD(t)

	repo := setupGitopsRepo(t)
	ctx := context.Background()
	appName := generateTestAppName()
	namespace := fmt.Sprintf("test-ns-%d", time.Now().UnixNano())

	// Create application first
	spec := argocd.ApplicationSpec{
		Name:      appName,
		Namespace: namespace,
		RepoURL:   "https://github.com/argoproj/argocd-example-apps.git",
		Revision:  "HEAD",
		Path:      "guestbook",
		ServerURL: "https://kubernetes.default.svc",
	}

	err := repo.CreateApplication(ctx, spec)
	require.NoError(t, err, "Failed to create ArgoCD application")
	defer cleanupApplication(t, repo, appName)

	// Wait a bit for application to be processed
	time.Sleep(2 * time.Second)

	// Test GetApplicationStatus
	status, err := repo.GetApplicationStatus(ctx, appName)
	require.NoError(t, err, "Failed to get application status")
	assert.NotNil(t, status)

	t.Logf("Application status: Sync=%s, Health=%s", status.SyncStatus, status.HealthStatus)
}

// TestSyncApplication tests triggering application sync
func TestSyncApplication(t *testing.T) {
	skipIfNoArgoCD(t)

	repo := setupGitopsRepo(t)
	ctx := context.Background()
	appName := generateTestAppName()
	namespace := fmt.Sprintf("test-ns-%d", time.Now().UnixNano())

	// Create application first
	spec := argocd.ApplicationSpec{
		Name:      appName,
		Namespace: namespace,
		RepoURL:   "https://github.com/argoproj/argocd-example-apps.git",
		Revision:  "HEAD",
		Path:      "guestbook",
		ServerURL: "https://kubernetes.default.svc",
	}

	err := repo.CreateApplication(ctx, spec)
	require.NoError(t, err, "Failed to create ArgoCD application")
	defer cleanupApplication(t, repo, appName)

	// Wait a bit for application to be processed
	time.Sleep(2 * time.Second)

	// Test SyncApplication
	err = repo.SyncApplication(ctx, appName)
	require.NoError(t, err, "Failed to sync application")

	t.Logf("Successfully triggered sync for application %s", appName)
}

// TestGetApplicationStatusNotFound tests behavior when application doesn't exist
func TestGetApplicationStatusNotFound(t *testing.T) {
	skipIfNoArgoCD(t)

	repo := setupGitopsRepo(t)
	ctx := context.Background()

	// Test GetApplicationStatus for non-existent application
	_, err := repo.GetApplicationStatus(ctx, "non-existent-app")
	assert.Error(t, err, "Should return error for non-existent application")
}

// TestDeleteApplicationNotFound tests behavior when application doesn't exist
func TestDeleteApplicationNotFound(t *testing.T) {
	skipIfNoArgoCD(t)

	repo := setupGitopsRepo(t)
	ctx := context.Background()

	// Test DeleteApplication for non-existent application
	err := repo.DeleteApplication(ctx, "non-existent-app")
	assert.Error(t, err, "Should return error for non-existent application")
}
