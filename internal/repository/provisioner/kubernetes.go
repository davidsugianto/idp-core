package provisioner

import (
	"context"
	"sync"

	"github.com/davidsugianto/idp-core/internal/model/environment"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateNamespace creates a new namespace with the given name and labels
func (r *repository) CreateNamespace(ctx context.Context, name string, labels map[string]string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
	}

	_, err := r.client.Clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	return err
}

// DeleteNamespace deletes the namespace with the given name
func (r *repository) DeleteNamespace(ctx context.Context, name string) error {
	return r.client.Clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
}

// GetNamespace retrieves a namespace by name
func (r *repository) GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	return r.client.Clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
}

// NamespaceExists checks if a namespace exists
func (r *repository) NamespaceExists(ctx context.Context, name string) (bool, error) {
	_, err := r.client.Clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetPodSummary returns the pod summary for a namespace from cache
func (r *repository) GetPodSummary(namespace string) (environment.PodSummary, bool) {
	return r.statusStore.getPodSummary(namespace)
}

// GetDeploymentSummary returns the deployment summary for a namespace from cache
func (r *repository) GetDeploymentSummary(namespace string) (environment.DeploymentSummary, bool) {
	return r.statusStore.getDeploymentSummary(namespace)
}

// GetWorkloads returns all deployments in a namespace from the informer cache
func (r *repository) GetWorkloads(namespace string) ([]*appsv1.Deployment, error) {
	return r.informerManager.GetWorkloads(namespace)
}

// GetPods returns all pods in a namespace from the informer cache
func (r *repository) GetPods(namespace string) ([]*corev1.Pod, error) {
	return r.informerManager.GetPods(namespace)
}

// Global status store for caching pod/deployment status
var (
	globalStatusStore *statusStore
	globalStoreMu     sync.Once
)

type statusStore struct {
	mu sync.RWMutex

	podSummaries        map[string]environment.PodSummary
	deploymentSummaries map[string]environment.DeploymentSummary
}

func init() {
	globalStatusStore = &statusStore{
		podSummaries:        make(map[string]environment.PodSummary),
		deploymentSummaries: make(map[string]environment.DeploymentSummary),
	}
}

func (s *statusStore) getPodSummary(namespace string) (environment.PodSummary, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	summary, ok := s.podSummaries[namespace]
	return summary, ok
}

func (s *statusStore) getDeploymentSummary(namespace string) (environment.DeploymentSummary, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	summary, ok := s.deploymentSummaries[namespace]
	return summary, ok
}

func (s *statusStore) updatePodSummary(namespace string, summary environment.PodSummary) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.podSummaries[namespace] = summary
}

func (s *statusStore) updateDeploymentSummary(namespace string, summary environment.DeploymentSummary) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.deploymentSummaries[namespace] = summary
}
