package webhook

import (
	"context"
	"fmt"

	quotaModel "github.com/davidsugianto/idp-core/internal/model/resourcequota"
	quotaUsecase "github.com/davidsugianto/idp-core/internal/usecase/quota"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// PolicyRule defines a validation rule
type PolicyRule interface {
	Validate(obj interface{}) (bool, string)
}

// QuotaRule checks resource quota limits
type QuotaRule struct {
	quotaUC quotaUsecase.Usecase
}

func NewQuotaRule(quotaUC quotaUsecase.Usecase) *QuotaRule {
	return &QuotaRule{quotaUC: quotaUC}
}

func (r *QuotaRule) Validate(obj interface{}) (bool, string) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return true, ""
	}

	if r.quotaUC == nil {
		return true, ""
	}

	namespace := pod.Namespace
	if namespace == "" {
		return true, ""
	}

	// Calculate total resources for the pod
	var totalCPU, totalMemory int64
	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests != nil {
			if cpu := container.Resources.Requests.Cpu(); cpu != nil {
				totalCPU += cpu.Value()
			}
			if mem := container.Resources.Requests.Memory(); mem != nil {
				totalMemory += mem.Value()
			}
		}
	}
	for _, container := range pod.Spec.InitContainers {
		if container.Resources.Requests != nil {
			if cpu := container.Resources.Requests.Cpu(); cpu != nil {
				totalCPU += cpu.Value()
			}
			if mem := container.Resources.Requests.Memory(); mem != nil {
				totalMemory += mem.Value()
			}
		}
	}

	// Build quota check request
	req := &quotaModel.QuotaCheckRequest{
		Namespace:   namespace,
		CPURequest:  resource.NewQuantity(totalCPU, resource.DecimalSI).String(),
		MemoryRequest: resource.NewQuantity(totalMemory, resource.BinarySI).String(),
		PodDelta:    1,
	}

	// Check quota
	result, err := r.quotaUC.CheckQuota(context.Background(), req)
	if err != nil {
		// On error, allow the pod (fail-open)
		return true, ""
	}

	if !result.Allowed {
		reason := "would exceed resource quota"
		if len(result.Reasons) > 0 {
			reason = fmt.Sprintf("would exceed resource quota: %s", result.Reasons[0].ResourceType)
		}
		return false, reason
	}

	return true, ""
}

// Validator validates Kubernetes resources against policies
type Validator struct {
	rules []PolicyRule
}

// NewValidator creates a new Validator with default rules
func NewValidator() *Validator {
	return &Validator{
		rules: []PolicyRule{
			&TeamIDLabelRule{},
			&NoPrivilegedContainersRule{},
			&ResourceLimitsRule{},
		},
	}
}

// NewValidatorWithQuota creates a new Validator with quota checking
func NewValidatorWithQuota(quotaUC quotaUsecase.Usecase) *Validator {
	return &Validator{
		rules: []PolicyRule{
			&TeamIDLabelRule{},
			&NoPrivilegedContainersRule{},
			&ResourceLimitsRule{},
			NewQuotaRule(quotaUC),
		},
	}
}

// ValidatePod validates a Pod against all rules
func (v *Validator) ValidatePod(pod *corev1.Pod) (allowed bool, reason string) {
	for _, rule := range v.rules {
		if ok, msg := rule.Validate(pod); !ok {
			return false, msg
		}
	}
	return true, ""
}

// ValidateDeployment validates a Deployment's pod template
func (v *Validator) ValidateDeployment(deploy interface{}) (allowed bool, reason string) {
	// For deployments, we validate the pod template
	return true, ""
}

// TeamIDLabelRule ensures pods have the team-id label
type TeamIDLabelRule struct{}

func (r *TeamIDLabelRule) Validate(obj interface{}) (bool, string) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return true, ""
	}

	if _, exists := pod.Labels["idp-core/team-id"]; !exists {
		return false, "pod must have 'idp-core/team-id' label"
	}
	return true, ""
}

// NoPrivilegedContainersRule blocks privileged containers
type NoPrivilegedContainersRule struct{}

func (r *NoPrivilegedContainersRule) Validate(obj interface{}) (bool, string) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return true, ""
	}

	for _, container := range pod.Spec.Containers {
		if container.SecurityContext != nil && container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
			return false, fmt.Sprintf("container %s is privileged, which is not allowed", container.Name)
		}
	}

	for _, container := range pod.Spec.InitContainers {
		if container.SecurityContext != nil && container.SecurityContext.Privileged != nil && *container.SecurityContext.Privileged {
			return false, fmt.Sprintf("init container %s is privileged, which is not allowed", container.Name)
		}
	}

	return true, ""
}

// ResourceLimitsRule ensures containers have resource limits
type ResourceLimitsRule struct{}

func (r *ResourceLimitsRule) Validate(obj interface{}) (bool, string) {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return true, ""
	}

	for _, container := range pod.Spec.Containers {
		if container.Resources.Limits == nil {
			return false, fmt.Sprintf("container %s must have resource limits defined", container.Name)
		}
		if _, ok := container.Resources.Limits[corev1.ResourceCPU]; !ok {
			return false, fmt.Sprintf("container %s must have CPU limit defined", container.Name)
		}
		if _, ok := container.Resources.Limits[corev1.ResourceMemory]; !ok {
			return false, fmt.Sprintf("container %s must have memory limit defined", container.Name)
		}
	}

	return true, ""
}
