package http

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Validate handles admission review requests
func (h *Handler) Validate(c *gin.Context) {
	var admissionReview admissionv1.AdmissionReview
	if err := json.NewDecoder(c.Request.Body).Decode(&admissionReview); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid admission review"})
		return
	}

	response := admissionv1.AdmissionReview{
		TypeMeta: admissionReview.TypeMeta,
		Response: &admissionv1.AdmissionResponse{
			UID: admissionReview.Request.UID,
		},
	}

	// Parse the object based on kind
	kind := admissionReview.Request.Kind.Kind
	raw := admissionReview.Request.Object.Raw

	allowed := true
	reason := ""

	switch kind {
	case "Pod":
		var pod corev1.Pod
		if err := json.Unmarshal(raw, &pod); err == nil {
			allowed, reason = h.webhookValidator.ValidatePod(&pod)
		}
	case "Deployment":
		allowed, reason = h.webhookValidator.ValidateDeployment(raw)
	}

	response.Response.Allowed = allowed
	if !allowed {
		response.Response.Result = &metav1.Status{
			Message: reason,
		}
	}

	c.JSON(http.StatusOK, response)
}
