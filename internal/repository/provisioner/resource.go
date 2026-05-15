package provisioner

import (
	"context"
)

func (r *repository) CreateResourceQuota(ctx context.Context, namespace, name string, cpu, memory string) error {
	return r.k8sClient.CreateResourceQuota(ctx, namespace, name, cpu, memory)
}

func (r *repository) DeleteResourceQuota(ctx context.Context, namespace, name string) error {
	return r.k8sClient.DeleteResourceQuota(ctx, namespace, name)
}

func (r *repository) CreateNetworkPolicy(ctx context.Context, namespace, name string, allowNamespaceLabels map[string]string) error {
	return r.k8sClient.CreateNetworkPolicy(ctx, namespace, name, allowNamespaceLabels)
}

func (r *repository) DeleteNetworkPolicy(ctx context.Context, namespace, name string) error {
	return r.k8sClient.DeleteNetworkPolicy(ctx, namespace, name)
}
