package live_update

import (
	"reflect"
	"testing"
)

func TestUsecaseExposesUS3LiveUpdateOperations(t *testing.T) {
	usecaseType := reflect.TypeOf((*Usecase)(nil)).Elem()

	requiredMethods := []string{
		"Subscribe",
		"Unsubscribe",
		"ListNotifications",
		"StreamEvents",
		"StreamLogs",
	}

	for _, methodName := range requiredMethods {
		t.Run(methodName, func(t *testing.T) {
			_, ok := usecaseType.MethodByName(methodName)
			if !ok {
				t.Fatalf("expected live_update.Usecase to define %s for Phase 3 event streaming, log streaming, notification history, stream expiry, and access-loss termination", methodName)
			}
		})
	}
}
