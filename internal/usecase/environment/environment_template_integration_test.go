package environment

import (
	"reflect"
	"testing"

	environmentModel "github.com/davidsugianto/idp-core/internal/model/environment"
)

func TestCreateEnvironmentRequestIncludesTemplateFields(t *testing.T) {
	requestType := reflect.TypeOf(environmentModel.CreateEnvironmentRequest{})

	requiredFields := []string{
		"TemplateVersionID",
		"TemplateInputs",
		"DeliveryTargetID",
	}

	for _, fieldName := range requiredFields {
		t.Run(fieldName, func(t *testing.T) {
			_, ok := requestType.FieldByName(fieldName)
			if !ok {
				t.Fatalf("expected CreateEnvironmentRequest to include %s for template-backed provisioning", fieldName)
			}
		})
	}
}

func TestEnvironmentResponseIncludesTemplateHistory(t *testing.T) {
	responseType := reflect.TypeOf(environmentModel.EnvironmentResponse{})

	requiredFields := []string{
		"TemplateInstanceID",
		"DeliveryTargetID",
	}

	for _, fieldName := range requiredFields {
		t.Run(fieldName, func(t *testing.T) {
			_, ok := responseType.FieldByName(fieldName)
			if !ok {
				t.Fatalf("expected EnvironmentResponse to include %s for template and placement history", fieldName)
			}
		})
	}
}
