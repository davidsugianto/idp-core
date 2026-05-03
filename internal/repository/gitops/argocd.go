package gitops

import (
	"context"

	"github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/davidsugianto/idp-core/internal/pkg/argocd"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var applicationGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "applications",
}

// CreateApplication creates an ArgoCD Application CRD
func (r *repository) CreateApplication(ctx context.Context, spec argocd.ApplicationSpec) error {
	// Default project to "default" if not specified
	project := spec.Project
	if project == "" {
		project = "default"
	}

	app := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "argoproj.io/v1alpha1",
			"kind":       "Application",
			"metadata": map[string]interface{}{
				"name":      spec.Name,
				"namespace": r.argocdNS,
				"labels": map[string]interface{}{
					"idp-core/managed-by": "idp-core",
				},
			},
			"spec": map[string]interface{}{
				"project": project,
				"destination": map[string]interface{}{
					"namespace": spec.Namespace,
					"server":    spec.ServerURL,
				},
				"source": map[string]interface{}{
					"repoURL":        spec.RepoURL,
					"targetRevision": spec.Revision,
					"path":           spec.Path,
				},
				"syncPolicy": map[string]interface{}{
					"automated": map[string]interface{}{
						"prune":    true,
						"selfHeal": true,
					},
				},
			},
		},
	}

	_, err := r.client.DynamicClient.Resource(applicationGVR).Namespace(r.argocdNS).Create(ctx, app, metav1.CreateOptions{})
	return err
}

// GetApplicationStatus retrieves the status of an ArgoCD Application
func (r *repository) GetApplicationStatus(ctx context.Context, name string) (*environment.ArgoStatus, error) {
	app, err := r.client.DynamicClient.Resource(applicationGVR).Namespace(r.argocdNS).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	status := &environment.ArgoStatus{}

	// Extract sync status
	if syncStatus, ok, _ := unstructured.NestedString(app.Object, "status", "sync", "status"); ok {
		status.SyncStatus = syncStatus
	}

	// Extract health status
	if healthStatus, ok, _ := unstructured.NestedString(app.Object, "status", "health", "status"); ok {
		status.HealthStatus = healthStatus
	}

	// Extract revision
	if revision, ok, _ := unstructured.NestedString(app.Object, "status", "sync", "revision"); ok {
		status.Revision = revision
	}

	// Extract message
	if message, ok, _ := unstructured.NestedString(app.Object, "status", "operationState", "message"); ok {
		status.Message = message
	}

	return status, nil
}

// SyncApplication triggers a manual sync of an ArgoCD Application
func (r *repository) SyncApplication(ctx context.Context, name string) error {
	// First, get the current application to retrieve the resourceVersion
	app, err := r.client.DynamicClient.Resource(applicationGVR).Namespace(r.argocdNS).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Create an operation to sync the application
	operation := map[string]interface{}{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Application",
		"metadata": map[string]interface{}{
			"name":              name,
			"namespace":         r.argocdNS,
			"resourceVersion":   app.GetResourceVersion(),
		},
		"operation": map[string]interface{}{
			"sync": map[string]interface{}{
				"syncStrategy": map[string]interface{}{
					"apply": map[string]interface{}{},
				},
			},
		},
	}

	patch := &unstructured.Unstructured{Object: operation}

	// Use JSON patch to add the operation
	_, err = r.client.DynamicClient.Resource(applicationGVR).Namespace(r.argocdNS).Update(ctx, patch, metav1.UpdateOptions{})
	return err
}

// DeleteApplication deletes an ArgoCD Application
func (r *repository) DeleteApplication(ctx context.Context, name string) error {
	return r.client.DynamicClient.Resource(applicationGVR).Namespace(r.argocdNS).Delete(ctx, name, metav1.DeleteOptions{})
}
