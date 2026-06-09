package delivery_target

import (
	"reflect"
	"testing"
)

func TestUsecaseExposesUS2DeliveryTargetOperations(t *testing.T) {
	usecaseType := reflect.TypeOf((*Usecase)(nil)).Elem()

	requiredMethods := []string{
		"Create",
		"Get",
		"List",
		"Update",
		"Delete",
	}

	for _, methodName := range requiredMethods {
		t.Run(methodName, func(t *testing.T) {
			_, ok := usecaseType.MethodByName(methodName)
			if !ok {
				t.Fatalf("expected delivery_target.Usecase to define %s for Phase 3 target management", methodName)
			}
		})
	}
}
