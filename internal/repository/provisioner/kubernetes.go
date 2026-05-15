package provisioner

import (
	"context"
	"sync"

	envModel "github.com/davidsugianto/idp-core/internal/model/environment"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func (r *repository) CreateNamespace(ctx context.Context, name string, labels map[string]string) error {
	return r.k8sClient.CreateNamespace(ctx, name, labels)
}

func (r *repository) DeleteNamespace(ctx context.Context, name string) error {
	return r.k8sClient.DeleteNamespace(ctx, name)
}

func (r *repository) GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error) {
	return r.k8sClient.GetNamespace(ctx, name)
}

func (r *repository) NamespaceExists(ctx context.Context, name string) (bool, error) {
	return r.k8sClient.NamespaceExists(ctx, name)
}

// GetPodSummary returns the pod summary for a namespace from cache
func (r *repository) GetPodSummary(namespace string) (envModel.PodSummary, bool) {
	return r.statusStore.getPodSummary(namespace)
}

// GetDeploymentSummary returns the deployment summary for a namespace from cache
func (r *repository) GetDeploymentSummary(namespace string) (envModel.DeploymentSummary, bool) {
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

	PodSummaries        map[string]envModel.PodSummary
	DeploymentSummaries map[string]envModel.DeploymentSummary
}

func init() {
	globalStatusStore = &statusStore{
		PodSummaries:        make(map[string]envModel.PodSummary),
		DeploymentSummaries: make(map[string]envModel.DeploymentSummary),
	}
}

func (s *statusStore) getPodSummary(namespace string) (envModel.PodSummary, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	summary, ok := s.PodSummaries[namespace]
	return summary, ok
}

func (s *statusStore) getDeploymentSummary(namespace string) (envModel.DeploymentSummary, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	summary, ok := s.DeploymentSummaries[namespace]
	return summary, ok
}

func (s *statusStore) updatePodSummary(namespace string, summary envModel.PodSummary) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.PodSummaries[namespace] = summary
}

func (s *statusStore) updateDeploymentSummary(namespace string, summary envModel.DeploymentSummary) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.DeploymentSummaries[namespace] = summary
}
