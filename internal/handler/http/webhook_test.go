package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestWebhookHandler_Validate(t *testing.T) {
	handler := NewWebhookHandler()

	t.Run("valid pod passes validation", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-pod",
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

		podJSON, _ := json.Marshal(pod)
		admissionReview := admissionv1.AdmissionReview{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AdmissionReview",
				APIVersion: "admission.k8s.io/v1",
			},
			Request: &admissionv1.AdmissionRequest{
				UID:  "test-uid-123",
				Kind: metav1.GroupVersionKind{Kind: "Pod"},
				Object: runtime.RawExtension{
					Raw: podJSON,
				},
			},
		}

		body, _ := json.Marshal(admissionReview)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/admission/validate", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Validate(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response admissionv1.AdmissionReview
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Response.Allowed)
		assert.Equal(t, "test-uid-123", string(response.Response.UID))
	})

	t.Run("invalid pod fails validation", func(t *testing.T) {
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:   "test-pod",
				Labels: map[string]string{}, // missing team-id label
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

		podJSON, _ := json.Marshal(pod)
		admissionReview := admissionv1.AdmissionReview{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AdmissionReview",
				APIVersion: "admission.k8s.io/v1",
			},
			Request: &admissionv1.AdmissionRequest{
				UID:  "test-uid-456",
				Kind: metav1.GroupVersionKind{Kind: "Pod"},
				Object: runtime.RawExtension{
					Raw: podJSON,
				},
			},
		}

		body, _ := json.Marshal(admissionReview)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/admission/validate", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Validate(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response admissionv1.AdmissionReview
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.False(t, response.Response.Allowed)
		assert.Contains(t, response.Response.Result.Message, "team-id")
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/admission/validate", bytes.NewBuffer([]byte("invalid")))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Validate(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("deployment validation", func(t *testing.T) {
		admissionReview := admissionv1.AdmissionReview{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AdmissionReview",
				APIVersion: "admission.k8s.io/v1",
			},
			Request: &admissionv1.AdmissionRequest{
				UID:  "test-uid-789",
				Kind: metav1.GroupVersionKind{Kind: "Deployment"},
				Object: runtime.RawExtension{
					Raw: []byte("{}"),
				},
			},
		}

		body, _ := json.Marshal(admissionReview)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/admission/validate", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Validate(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response admissionv1.AdmissionReview
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response.Response.Allowed)
	})
}

func TestNewWebhookHandler(t *testing.T) {
	handler := NewWebhookHandler()
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.validator)
}
