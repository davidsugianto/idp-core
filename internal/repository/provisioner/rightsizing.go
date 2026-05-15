package provisioner

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
)

func (r *repository) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return r.k8sClient.GetDeployment(ctx, namespace, name)
}

func (r *repository) GetStatefulSet(ctx context.Context, namespace, name string) (*appsv1.StatefulSet, error) {
	return r.k8sClient.GetStatefulSet(ctx, namespace, name)
}

func (r *repository) UpdateDeploymentResources(ctx context.Context, namespace, name, containerName string, cpuRequest, cpuLimit, memoryRequest, memoryLimit string) error {
	return r.k8sClient.UpdateDeploymentResources(ctx, namespace, name, containerName, cpuRequest, cpuLimit, memoryRequest, memoryLimit)
}

func (r *repository) UpdateStatefulSetResources(ctx context.Context, namespace, name, containerName string, cpuRequest, cpuLimit, memoryRequest, memoryLimit string) error {
	return r.k8sClient.UpdateStatefulSetResources(ctx, namespace, name, containerName, cpuRequest, cpuLimit, memoryRequest, memoryLimit)
}
