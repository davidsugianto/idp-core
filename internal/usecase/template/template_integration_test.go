package template

import (
	"reflect"
	"testing"
)

func TestUsecaseExposesPhase3TemplateOperations(t *testing.T) {
	usecaseType := reflect.TypeOf((*Usecase)(nil)).Elem()

	requiredMethods := []string{
		"ReplaceParameters",
		"ReplaceResources",
		"ValidateVersionInputs",
	}

	for _, methodName := range requiredMethods {
		t.Run(methodName, func(t *testing.T) {
			_, ok := usecaseType.MethodByName(methodName)
			if !ok {
				t.Fatalf("expected template.Usecase to define %s for Phase 3 template lifecycle support", methodName)
			}
		})
	}
}
