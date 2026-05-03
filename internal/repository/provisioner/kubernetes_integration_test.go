package provisioner

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/pkg/kubernetes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Skip if not running integration tests
func skipIfNoK8s(t *testing.T) {
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
		t.Skip("kubeconfig not found, skipping K8s integration test")
	}

	// Verify we can connect to the cluster
	client, err := kubernetes.NewClient(false, "")
	if err != nil {
		t.Skipf("Failed to connect to Kubernetes: %v", err)
	}

	// Check if cluster is accessible
	ctx := context.Background()
	_, err = client.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		t.Skipf("Kubernetes cluster not accessible: %v", err)
	}
}

func setupK8sClient(t *testing.T) *kubernetes.Client {
	client, err := kubernetes.NewClient(false, "")
	require.NoError(t, err, "Failed to create Kubernetes client")
	return client
}

func setupProvisionerRepo(t *testing.T) Repository {
	client := setupK8sClient(t)
	return New(Dependencies{K8sClient: client})
}

func generateTestNamespace() string {
	return fmt.Sprintf("idp-test-%d", time.Now().UnixNano())
}

func cleanupNamespace(t *testing.T, repo Repository, namespace string) {
	ctx := context.Background()
	err := repo.DeleteNamespace(ctx, namespace)
	if err != nil && !errors.IsNotFound(err) {
		t.Logf("Warning: failed to cleanup namespace %s: %v", namespace, err)
	}
}

// TestK8sConnection tests basic connectivity to the Kubernetes cluster
func TestK8sConnection(t *testing.T) {
	skipIfNoK8s(t)

	client := setupK8sClient(t)

	ctx := context.Background()
	nodes, err := client.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	require.NoError(t, err, "Failed to list nodes")

	assert.NotEmpty(t, nodes.Items, "Cluster should have at least one node")
	t.Logf("Connected to cluster with %d nodes", len(nodes.Items))
}

// TestNamespaceCRUD tests namespace creation and deletion
func TestNamespaceCRUD(t *testing.T) {
	skipIfNoK8s(t)

	repo := setupProvisionerRepo(t)
	ctx := context.Background()
	namespace := generateTestNamespace()

	// Test CreateNamespace
	labels := map[string]string{
		"idp-core/test":       "true",
		"idp-core/managed-by": "idp-core",
	}

	err := repo.CreateNamespace(ctx, namespace, labels)
	require.NoError(t, err, "Failed to create namespace")

	// Test NamespaceExists
	exists, err := repo.NamespaceExists(ctx, namespace)
	require.NoError(t, err, "Failed to check namespace exists")
	assert.True(t, exists, "Namespace should exist")

	// Test GetNamespace
	ns, err := repo.GetNamespace(ctx, namespace)
	require.NoError(t, err, "Failed to get namespace")
	assert.Equal(t, namespace, ns.Name)
	assert.Equal(t, "idp-core", ns.Labels["idp-core/managed-by"])

	// Test DeleteNamespace
	err = repo.DeleteNamespace(ctx, namespace)
	require.NoError(t, err, "Failed to delete namespace")

	// Wait for namespace to actually be deleted (async in Kubernetes)
	for i := 0; i < 30; i++ {
		exists, err = repo.NamespaceExists(ctx, namespace)
		if err == nil && !exists {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	// Verify deletion
	exists, err = repo.NamespaceExists(ctx, namespace)
	require.NoError(t, err, "Failed to check namespace exists")
	assert.False(t, exists, "Namespace should not exist after deletion")
}

// TestResourceQuota tests ResourceQuota creation and deletion
func TestResourceQuota(t *testing.T) {
	skipIfNoK8s(t)

	repo := setupProvisionerRepo(t)
	ctx := context.Background()
	namespace := generateTestNamespace()

	// Create namespace first
	err := repo.CreateNamespace(ctx, namespace, map[string]string{"idp-core/test": "true"})
	require.NoError(t, err, "Failed to create namespace")
	defer cleanupNamespace(t, repo, namespace)

	// Test CreateResourceQuota
	err = repo.CreateResourceQuota(ctx, namespace, "test-quota", "2", "4Gi")
	require.NoError(t, err, "Failed to create resource quota")

	// Verify quota was created
	client := setupK8sClient(t)
	quota, err := client.Clientset.CoreV1().ResourceQuotas(namespace).Get(ctx, "test-quota", metav1.GetOptions{})
	require.NoError(t, err, "Failed to get resource quota")
	assert.NotNil(t, quota)
	if cpu, ok := quota.Spec.Hard[corev1.ResourceRequestsCPU]; ok {
		assert.Equal(t, "2", cpu.String())
	}

	// Test DeleteResourceQuota
	err = repo.DeleteResourceQuota(ctx, namespace, "test-quota")
	require.NoError(t, err, "Failed to delete resource quota")
}

// TestNetworkPolicy tests NetworkPolicy creation and deletion
func TestNetworkPolicy(t *testing.T) {
	skipIfNoK8s(t)

	repo := setupProvisionerRepo(t)
	ctx := context.Background()
	namespace := generateTestNamespace()

	// Create namespace first
	err := repo.CreateNamespace(ctx, namespace, map[string]string{"idp-core/test": "true"})
	require.NoError(t, err, "Failed to create namespace")
	defer cleanupNamespace(t, repo, namespace)

	// Test CreateNetworkPolicy
	allowLabels := map[string]string{
		"idp-core/test": "true",
	}
	err = repo.CreateNetworkPolicy(ctx, namespace, "test-policy", allowLabels)
	require.NoError(t, err, "Failed to create network policy")

	// Verify policy was created
	client := setupK8sClient(t)
	policy, err := client.Clientset.NetworkingV1().NetworkPolicies(namespace).Get(ctx, "test-policy", metav1.GetOptions{})
	require.NoError(t, err, "Failed to get network policy")
	assert.NotNil(t, policy)
	assert.Equal(t, "test-policy", policy.Name)

	// Test DeleteNetworkPolicy
	err = repo.DeleteNetworkPolicy(ctx, namespace, "test-policy")
	require.NoError(t, err, "Failed to delete network policy")
}

// TestPodSummary tests pod status caching
func TestPodSummary(t *testing.T) {
	skipIfNoK8s(t)

	repo := setupProvisionerRepo(t)
	ctx := context.Background()
	namespace := generateTestNamespace()

	// Create namespace
	err := repo.CreateNamespace(ctx, namespace, map[string]string{"idp-core/test": "true"})
	require.NoError(t, err, "Failed to create namespace")
	defer cleanupNamespace(t, repo, namespace)

	// Start informers
	err = repo.StartInformers(ctx)
	require.NoError(t, err, "Failed to start informers")
	defer repo.StopInformers()

	// Wait for cache sync
	time.Sleep(2 * time.Second)

	// Test GetPodSummary for empty namespace
	summary, ok := repo.GetPodSummary(namespace)
	// May not be cached yet, or may be empty
	if ok {
		assert.Equal(t, 0, summary.Total, "Empty namespace should have 0 pods")
	}

	// Create a test pod
	client := setupK8sClient(t)
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

	_, err = client.Clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	require.NoError(t, err, "Failed to create test pod")

	// Cleanup pod
	defer client.Clientset.CoreV1().Pods(namespace).Delete(ctx, "test-pod", metav1.DeleteOptions{})

	// Wait for pod to be processed
	time.Sleep(2 * time.Second)

	// Test GetPodSummary
	summary, ok = repo.GetPodSummary(namespace)
	if ok {
		assert.GreaterOrEqual(t, summary.Total, 1, "Should have at least 1 pod")
	}
}

// TestDeploymentSummary tests deployment status caching
func TestDeploymentSummary(t *testing.T) {
	skipIfNoK8s(t)

	repo := setupProvisionerRepo(t)
	ctx := context.Background()
	namespace := generateTestNamespace()

	// Create namespace
	err := repo.CreateNamespace(ctx, namespace, map[string]string{"idp-core/test": "true"})
	require.NoError(t, err, "Failed to create namespace")
	defer cleanupNamespace(t, repo, namespace)

	// Start informers
	err = repo.StartInformers(ctx)
	require.NoError(t, err, "Failed to start informers")
	defer repo.StopInformers()

	// Wait for cache sync
	time.Sleep(2 * time.Second)

	// Test GetDeploymentSummary for empty namespace
	summary, ok := repo.GetDeploymentSummary(namespace)
	if ok {
		assert.Equal(t, 0, summary.Desired, "Empty namespace should have 0 desired replicas")
	}
}

// TestGetWorkloads tests fetching deployments from cache
func TestGetWorkloads(t *testing.T) {
	skipIfNoK8s(t)

	repo := setupProvisionerRepo(t)
	ctx := context.Background()
	namespace := generateTestNamespace()

	// Create namespace
	err := repo.CreateNamespace(ctx, namespace, map[string]string{"idp-core/test": "true"})
	require.NoError(t, err, "Failed to create namespace")
	defer cleanupNamespace(t, repo, namespace)

	// Start informers
	err = repo.StartInformers(ctx)
	require.NoError(t, err, "Failed to start informers")
	defer repo.StopInformers()

	// Wait for cache sync
	time.Sleep(2 * time.Second)

	// Test GetWorkloads - should return empty slice for empty namespace
	workloads, err := repo.GetWorkloads(namespace)
	require.NoError(t, err, "Failed to get workloads")
	// For empty namespace, workloads can be nil or empty slice - both are valid
	if workloads != nil {
		assert.Empty(t, workloads, "Empty namespace should have no workloads")
	}
}

// TestGetPods tests fetching pods from cache
func TestGetPods(t *testing.T) {
	skipIfNoK8s(t)

	repo := setupProvisionerRepo(t)
	ctx := context.Background()
	namespace := generateTestNamespace()

	// Create namespace
	err := repo.CreateNamespace(ctx, namespace, map[string]string{"idp-core/test": "true"})
	require.NoError(t, err, "Failed to create namespace")
	defer cleanupNamespace(t, repo, namespace)

	// Start informers
	err = repo.StartInformers(ctx)
	require.NoError(t, err, "Failed to start informers")
	defer repo.StopInformers()

	// Wait for cache sync
	time.Sleep(2 * time.Second)

	// Test GetPods - should return empty slice for empty namespace
	pods, err := repo.GetPods(namespace)
	require.NoError(t, err, "Failed to get pods")
	// For empty namespace, pods can be nil or empty slice - both are valid
	if pods != nil {
		assert.Empty(t, pods, "Empty namespace should have no pods")
	}
}
