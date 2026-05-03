package webhook

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestTeamIDLabelRule(t *testing.T) {
	rule := &TeamIDLabelRule{}

	t.Run("pod with team-id label passes", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"idp-core/team-id": "team-123",
				},
			},
		}
		allowed, reason := rule.Validate(pod)
		assert.True(t, allowed)
		assert.Empty(t, reason)
	})

	t.Run("pod without team-id label fails", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "myapp",
				},
			},
		}
		allowed, reason := rule.Validate(pod)
		assert.False(t, allowed)
		assert.Contains(t, reason, "idp-core/team-id")
	})

	t.Run("pod with no labels fails", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{},
		}
		allowed, reason := rule.Validate(pod)
		assert.False(t, allowed)
		assert.Contains(t, reason, "idp-core/team-id")
	})

	t.Run("non-pod object passes", func(t *testing.T) {
		allowed, reason := rule.Validate("not a pod")
		assert.True(t, allowed)
		assert.Empty(t, reason)
	})
}

func TestNoPrivilegedContainersRule(t *testing.T) {
	rule := &NoPrivilegedContainersRule{}
	privileged := true

	t.Run("pod without privileged containers passes", func(t *testing.T) {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "app"},
				},
			},
		}
		allowed, reason := rule.Validate(pod)
		assert.True(t, allowed)
		assert.Empty(t, reason)
	})

	t.Run("pod with privileged container fails", func(t *testing.T) {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "app",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
						},
					},
				},
			},
		}
		allowed, reason := rule.Validate(pod)
		assert.False(t, allowed)
		assert.Contains(t, reason, "privileged")
		assert.Contains(t, reason, "app")
	})

	t.Run("pod with privileged init container fails", func(t *testing.T) {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				InitContainers: []corev1.Container{
					{
						Name: "init",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &privileged,
						},
					},
				},
				Containers: []corev1.Container{
					{Name: "app"},
				},
			},
		}
		allowed, reason := rule.Validate(pod)
		assert.False(t, allowed)
		assert.Contains(t, reason, "init container")
		assert.Contains(t, reason, "privileged")
	})

	t.Run("pod with non-privileged security context passes", func(t *testing.T) {
		notPrivileged := false
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "app",
						SecurityContext: &corev1.SecurityContext{
							Privileged: &notPrivileged,
						},
					},
				},
			},
		}
		allowed, reason := rule.Validate(pod)
		assert.True(t, allowed)
		assert.Empty(t, reason)
	})

	t.Run("non-pod object passes", func(t *testing.T) {
		allowed, reason := rule.Validate("not a pod")
		assert.True(t, allowed)
		assert.Empty(t, reason)
	})
}

func TestResourceLimitsRule(t *testing.T) {
	rule := &ResourceLimitsRule{}

	t.Run("pod with resource limits passes", func(t *testing.T) {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "app",
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
						},
					},
				},
			},
		}
		allowed, reason := rule.Validate(pod)
		assert.True(t, allowed)
		assert.Empty(t, reason)
	})

	t.Run("pod without resource limits fails", func(t *testing.T) {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{Name: "app"},
				},
			},
		}
		allowed, reason := rule.Validate(pod)
		assert.False(t, allowed)
		assert.Contains(t, reason, "resource limits")
	})

	t.Run("pod with limits but missing CPU fails", func(t *testing.T) {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "app",
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
						},
					},
				},
			},
		}
		allowed, reason := rule.Validate(pod)
		assert.False(t, allowed)
		assert.Contains(t, reason, "CPU limit")
	})

	t.Run("pod with limits but missing memory fails", func(t *testing.T) {
		pod := &corev1.Pod{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "app",
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU: resource.MustParse("100m"),
							},
						},
					},
				},
			},
		}
		allowed, reason := rule.Validate(pod)
		assert.False(t, allowed)
		assert.Contains(t, reason, "memory limit")
	})

	t.Run("non-pod object passes", func(t *testing.T) {
		allowed, reason := rule.Validate("not a pod")
		assert.True(t, allowed)
		assert.Empty(t, reason)
	})
}

func TestValidator(t *testing.T) {
	validator := NewValidator()

	t.Run("valid pod passes all rules", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"idp-core/team-id": "team-123",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "app",
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
						},
					},
				},
			},
		}
		allowed, reason := validator.ValidatePod(pod)
		assert.True(t, allowed)
		assert.Empty(t, reason)
	})

	t.Run("invalid pod fails first rule", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{
					"app": "myapp",
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "app",
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("128Mi"),
							},
						},
					},
				},
			},
		}
		allowed, reason := validator.ValidatePod(pod)
		assert.False(t, allowed)
		assert.Contains(t, reason, "team-id")
	})

	t.Run("ValidateDeployment returns allowed", func(t *testing.T) {
		allowed, reason := validator.ValidateDeployment(nil)
		assert.True(t, allowed)
		assert.Empty(t, reason)
	})
}
