package provisioner

import (
	"context"

	"github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/davidsugianto/idp-core/internal/pkg/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// Repository defines the interface for Kubernetes provisioning operations
type Repository interface {
	// Namespace operations
	CreateNamespace(ctx context.Context, name string, labels map[string]string) error
	DeleteNamespace(ctx context.Context, name string) error
	GetNamespace(ctx context.Context, name string) (*corev1.Namespace, error)
	NamespaceExists(ctx context.Context, name string) (bool, error)

	// Resource Quota
	CreateResourceQuota(ctx context.Context, namespace, name string, cpu, memory string) error
	DeleteResourceQuota(ctx context.Context, namespace, name string) error

	// Network Policy
	CreateNetworkPolicy(ctx context.Context, namespace, name string, allowNamespaceLabels map[string]string) error
	DeleteNetworkPolicy(ctx context.Context, namespace, name string) error

	// Status from cache
	GetPodSummary(namespace string) (environment.PodSummary, bool)
	GetDeploymentSummary(namespace string) (environment.DeploymentSummary, bool)

	// Workloads and Pods from cache
	GetWorkloads(namespace string) ([]*appsv1.Deployment, error)
	GetPods(namespace string) ([]*corev1.Pod, error)

	// Informer management
	StartInformers(ctx context.Context) error
	StopInformers()
}

type repository struct {
	client          *kubernetes.Client
	statusStore     *statusStore
	informerManager *informerManager
}

type Dependencies struct {
	K8sClient *kubernetes.Client
}

func New(deps Dependencies) Repository {
	return &repository{
		client:          deps.K8sClient,
		statusStore:     globalStatusStore,
		informerManager: newInformerManager(deps.K8sClient, globalStatusStore),
	}
}
