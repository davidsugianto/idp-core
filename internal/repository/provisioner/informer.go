package provisioner

import (
	"context"
	"sync"
	"time"

	envModel "github.com/davidsugianto/idp-core/internal/model/environment"
	k8sPkg "github.com/davidsugianto/idp-core/internal/pkg/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

// informerManager manages Kubernetes informers for status tracking
type informerManager struct {
	client      *k8sPkg.Client
	statusStore *statusStore
	factory     informers.SharedInformerFactory
	stopCh      chan struct{}
	started     bool
	mu          sync.RWMutex
}

// newInformerManager creates a new informer manager
func newInformerManager(client *k8sPkg.Client, store *statusStore) *informerManager {
	return &informerManager{
		client:      client,
		statusStore: store,
		stopCh:      make(chan struct{}),
	}
}

// StartInformers starts the shared informers for Pods and Deployments
func (r *repository) StartInformers(ctx context.Context) error {
	r.informerManager.mu.Lock()
	defer r.informerManager.mu.Unlock()

	if r.informerManager.started {
		return nil
	}

	// Create shared informer factory watching all namespaces
	r.informerManager.factory = informers.NewSharedInformerFactory(r.k8sClient.Clientset, 30*time.Second)

	// Set up pod informer
	podInformer := r.informerManager.factory.Core().V1().Pods().Informer()
	podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { r.informerManager.updatePodStatus(obj) },
		UpdateFunc: func(oldObj, newObj interface{}) { r.informerManager.updatePodStatus(newObj) },
		DeleteFunc: func(obj interface{}) { r.informerManager.updatePodStatus(obj) },
	})

	// Set up deployment informer
	deployInformer := r.informerManager.factory.Apps().V1().Deployments().Informer()
	deployInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { r.informerManager.updateDeploymentStatus(obj) },
		UpdateFunc: func(oldObj, newObj interface{}) { r.informerManager.updateDeploymentStatus(newObj) },
		DeleteFunc: func(obj interface{}) { r.informerManager.updateDeploymentStatus(obj) },
	})

	// Start informers
	r.informerManager.factory.Start(r.informerManager.stopCh)

	// Wait for cache sync with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if !cache.WaitForCacheSync(ctx.Done(),
		podInformer.HasSynced,
		deployInformer.HasSynced,
	) {
		// Continue even if sync times out
	}

	r.informerManager.started = true
	return nil
}

// StopInformers stops the informers
func (r *repository) StopInformers() {
	r.informerManager.mu.Lock()
	defer r.informerManager.mu.Unlock()

	if r.informerManager.started {
		close(r.informerManager.stopCh)
		r.informerManager.started = false
	}
}

func (m *informerManager) updatePodStatus(obj interface{}) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return
	}

	namespace := pod.Namespace

	// Get all pods in namespace from cache
	pods, err := m.factory.Core().V1().Pods().Lister().Pods(namespace).List(labels.Everything())
	if err != nil {
		return
	}

	summary := envModel.PodSummary{}
	for _, p := range pods {
		summary.Total++
		switch p.Status.Phase {
		case corev1.PodRunning:
			summary.Running++
		case corev1.PodPending:
			summary.Pending++
		case corev1.PodFailed:
			summary.Failed++
		}
	}

	m.statusStore.updatePodSummary(namespace, summary)
}

func (m *informerManager) updateDeploymentStatus(obj interface{}) {
	deploy, ok := obj.(*appsv1.Deployment)
	if !ok {
		return
	}

	namespace := deploy.Namespace

	// Get all deployments in namespace from cache
	deploys, err := m.factory.Apps().V1().Deployments().Lister().Deployments(namespace).List(labels.Everything())
	if err != nil {
		return
	}

	summary := envModel.DeploymentSummary{}
	for _, d := range deploys {
		if d.Spec.Replicas != nil {
			summary.Desired += int(*d.Spec.Replicas)
		}
		summary.Ready += int(d.Status.ReadyReplicas)
		summary.Updated += int(d.Status.UpdatedReplicas)
		summary.Available += int(d.Status.AvailableReplicas)
	}

	m.statusStore.updateDeploymentSummary(namespace, summary)
}

// GetWorkloads returns all deployments in a namespace
func (m *informerManager) GetWorkloads(namespace string) ([]*appsv1.Deployment, error) {
	if !m.started {
		return nil, nil
	}
	return m.factory.Apps().V1().Deployments().Lister().Deployments(namespace).List(labels.Everything())
}

// GetPods returns all pods in a namespace
func (m *informerManager) GetPods(namespace string) ([]*corev1.Pod, error) {
	if !m.started {
		return nil, nil
	}
	return m.factory.Core().V1().Pods().Lister().Pods(namespace).List(labels.Everything())
}
