package http

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeliveryTargetHandlerExposesUS2Endpoints(t *testing.T) {
	handlerType := reflect.TypeOf(&Handler{})

	requiredMethods := []string{
		"ListDeliveryTargets",
		"CreateDeliveryTarget",
		"GetDeliveryTarget",
		"UpdateDeliveryTarget",
		"DeleteDeliveryTarget",
	}

	for _, methodName := range requiredMethods {
		t.Run(methodName, func(t *testing.T) {
			_, ok := handlerType.MethodByName(methodName)
			assert.True(t, ok, "expected Handler to expose %s for Phase 3 delivery target contracts", methodName)
		})
	}
}
