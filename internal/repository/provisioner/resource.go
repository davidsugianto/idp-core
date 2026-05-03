package provisioner

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateResourceQuota creates a ResourceQuota for a namespace
func (r *repository) CreateResourceQuota(ctx context.Context, namespace, name string, cpu, memory string) error {
	quota := &corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"idp-core/managed-by": "idp-core",
			},
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{},
		},
	}

	if cpu != "" {
		quota.Spec.Hard[corev1.ResourceRequestsCPU] = resource.MustParse(cpu)
		quota.Spec.Hard[corev1.ResourceLimitsCPU] = resource.MustParse(cpu)
	}

	if memory != "" {
		quota.Spec.Hard[corev1.ResourceRequestsMemory] = resource.MustParse(memory)
		quota.Spec.Hard[corev1.ResourceLimitsMemory] = resource.MustParse(memory)
	}

	_, err := r.client.Clientset.CoreV1().ResourceQuotas(namespace).Create(ctx, quota, metav1.CreateOptions{})
	return err
}

// DeleteResourceQuota deletes a ResourceQuota
func (r *repository) DeleteResourceQuota(ctx context.Context, namespace, name string) error {
	return r.client.Clientset.CoreV1().ResourceQuotas(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// CreateNetworkPolicy creates a NetworkPolicy to isolate the namespace
func (r *repository) CreateNetworkPolicy(ctx context.Context, namespace, name string, allowNamespaceLabels map[string]string) error {
	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels: map[string]string{
				"idp-core/managed-by": "idp-core",
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
				networkingv1.PolicyTypeEgress,
			},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{
					// Allow ingress from pods in the same namespace
					From: []networkingv1.NetworkPolicyPeer{
						{
							PodSelector: &metav1.LabelSelector{},
							NamespaceSelector: &metav1.LabelSelector{
								MatchLabels: allowNamespaceLabels,
							},
						},
					},
				},
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				{
					// Allow all egress by default
					To: []networkingv1.NetworkPolicyPeer{
						{
							IPBlock: &networkingv1.IPBlock{
								CIDR: "0.0.0.0/0",
								Except: []string{
									"10.0.0.0/8",
									"172.16.0.0/12",
									"192.168.0.0/16",
								},
							},
						},
					},
				},
			},
		},
	}

	_, err := r.client.Clientset.NetworkingV1().NetworkPolicies(namespace).Create(ctx, policy, metav1.CreateOptions{})
	return err
}

// DeleteNetworkPolicy deletes a NetworkPolicy
func (r *repository) DeleteNetworkPolicy(ctx context.Context, namespace, name string) error {
	return r.client.Clientset.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
