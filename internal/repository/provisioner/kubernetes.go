package provisioner

import (
	"context"
	"io"
	"sync"

	envModel "github.com/davidsugianto/idp-core/internal/model/environment"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// GetWorkloads returns all deployments in a namespace from the informer cache,
// falling back to direct API reads when informers are not running.
func (r *repository) GetWorkloads(namespace string) ([]*appsv1.Deployment, error) {
	if r.informerManager != nil && r.informerManager.IsStarted() {
		return r.informerManager.GetWorkloads(namespace)
	}

	deployments, err := r.k8sClient.Clientset.AppsV1().Deployments(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	items := make([]*appsv1.Deployment, 0, len(deployments.Items))
	for i := range deployments.Items {
		items = append(items, &deployments.Items[i])
	}
	return items, nil
}

// GetPods returns all pods in a namespace from the informer cache,
// falling back to direct API reads when informers are not running.
func (r *repository) GetPods(namespace string) ([]*corev1.Pod, error) {
	if r.informerManager != nil && r.informerManager.IsStarted() {
		return r.informerManager.GetPods(namespace)
	}

	pods, err := r.k8sClient.Clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	items := make([]*corev1.Pod, 0, len(pods.Items))
	for i := range pods.Items {
		items = append(items, &pods.Items[i])
	}
	return items, nil
}

func (r *repository) ResolvePodForWorkload(namespace, workloadName string) (*corev1.Pod, error) {
	if r.informerManager != nil && r.informerManager.IsStarted() {
		return r.informerManager.ResolvePodForWorkload(namespace, workloadName)
	}

	pods, err := r.GetPods(namespace)
	if err != nil {
		return nil, err
	}
	for _, pod := range pods {
		for _, owner := range pod.OwnerReferences {
			if owner.Kind == "ReplicaSet" {
				replicaSet, err := r.k8sClient.Clientset.AppsV1().ReplicaSets(namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					continue
				}
				for _, rsOwner := range replicaSet.OwnerReferences {
					if rsOwner.Kind == "Deployment" && rsOwner.Name == workloadName {
						return pod, nil
					}
				}
			}
			if owner.Kind == "StatefulSet" && owner.Name == workloadName {
				return pod, nil
			}
		}
	}
	return nil, nil
}

func (r *repository) StreamPodLogs(ctx context.Context, namespace, podName, containerName string, tailLines int64) (io.ReadCloser, error) {
	return r.k8sClient.Clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Follow:    true,
		Container: containerName,
		TailLines: func() *int64 {
			if tailLines <= 0 {
				return nil
			}
			return &tailLines
		}(),
	}).Stream(ctx)
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
