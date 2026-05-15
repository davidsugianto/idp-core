package kubernetes

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetDeployment retrieves a deployment from Kubernetes
func (c *Client) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	deploy, err := c.Clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment %s/%s: %w", namespace, name, err)
	}
	return deploy, nil
}

// GetStatefulSet retrieves a statefulset from Kubernetes
func (c *Client) GetStatefulSet(ctx context.Context, namespace, name string) (*appsv1.StatefulSet, error) {
	sts, err := c.Clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get statefulset %s/%s: %w", namespace, name, err)
	}
	return sts, nil
}

// UpdateDeploymentResources updates the resource requests/limits for a container in a deployment
func (c *Client) UpdateDeploymentResources(ctx context.Context, namespace, name, containerName string, cpuRequest, cpuLimit, memoryRequest, memoryLimit string) error {
	deploy, err := c.GetDeployment(ctx, namespace, name)
	if err != nil {
		return err
	}

	found := false
	for i, container := range deploy.Spec.Template.Spec.Containers {
		if container.Name == containerName {
			found = true
			deploy.Spec.Template.Spec.Containers[i].Resources = buildResourceRequirements(cpuRequest, cpuLimit, memoryRequest, memoryLimit)
			break
		}
	}

	if !found {
		return fmt.Errorf("container %s not found in deployment %s/%s", containerName, namespace, name)
	}

	_, err = c.Clientset.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update deployment %s/%s: %w", namespace, name, err)
	}

	return nil
}

// UpdateStatefulSetResources updates the resource requests/limits for a container in a statefulset
func (c *Client) UpdateStatefulSetResources(ctx context.Context, namespace, name, containerName string, cpuRequest, cpuLimit, memoryRequest, memoryLimit string) error {
	sts, err := c.GetStatefulSet(ctx, namespace, name)
	if err != nil {
		return err
	}

	found := false
	for i, container := range sts.Spec.Template.Spec.Containers {
		if container.Name == containerName {
			found = true
			sts.Spec.Template.Spec.Containers[i].Resources = buildResourceRequirements(cpuRequest, cpuLimit, memoryRequest, memoryLimit)
			break
		}
	}

	if !found {
		return fmt.Errorf("container %s not found in statefulset %s/%s", containerName, namespace, name)
	}

	_, err = c.Clientset.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update statefulset %s/%s: %w", namespace, name, err)
	}

	return nil
}

// buildResourceRequirements constructs a ResourceRequirements from string values
func buildResourceRequirements(cpuRequest, cpuLimit, memoryRequest, memoryLimit string) corev1.ResourceRequirements {
	rr := corev1.ResourceRequirements{
		Requests: make(corev1.ResourceList),
		Limits:   make(corev1.ResourceList),
	}

	if cpuRequest != "" {
		if q, err := resource.ParseQuantity(cpuRequest); err == nil {
			rr.Requests[corev1.ResourceCPU] = q
		}
	}
	if cpuLimit != "" {
		if q, err := resource.ParseQuantity(cpuLimit); err == nil {
			rr.Limits[corev1.ResourceCPU] = q
		}
	}
	if memoryRequest != "" {
		if q, err := resource.ParseQuantity(memoryRequest); err == nil {
			rr.Requests[corev1.ResourceMemory] = q
		}
	}
	if memoryLimit != "" {
		if q, err := resource.ParseQuantity(memoryLimit); err == nil {
			rr.Limits[corev1.ResourceMemory] = q
		}
	}

	return rr
}
