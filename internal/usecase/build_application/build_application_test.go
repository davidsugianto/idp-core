package build_application

import (
	"errors"
	"testing"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
	"github.com/stretchr/testify/assert"
)

func TestValidateRegistryTargetForApplication(t *testing.T) {
	u := &usecase{}

	t.Run("empty registry target is allowed", func(t *testing.T) {
		err := u.validateRegistryTargetForApplication(t.Context(), "")
		assert.NoError(t, err)
	})

	t.Run("non-empty registry target without resolved type is rejected", func(t *testing.T) {
		err := u.validateRegistryTargetForApplication(t.Context(), "target-1")
		assert.True(t, errors.Is(err, ErrInvalidRegistryType))
	})
}

func TestLifecycleNotificationMetadata(t *testing.T) {
	assert.Equal(t, notificationSeverityError, lifecycleNotificationSeverity(buildApplicationModel.EventTypeBuildFailed))
	assert.Equal(t, notificationSeveritySuccess, lifecycleNotificationSeverity(buildApplicationModel.EventTypeBuildSucceeded))
	assert.Equal(t, notificationSeverityInfo, lifecycleNotificationSeverity(buildApplicationModel.EventTypeBuildQueued))

	assert.Equal(t, "Build failed", lifecycleNotificationTitle(buildApplicationModel.EventTypeBuildFailed))
	assert.Equal(t, "Deployment update succeeded", lifecycleNotificationTitle(buildApplicationModel.EventTypeDeploySucceeded))
	assert.Equal(t, "Build lifecycle updated", lifecycleNotificationTitle("unknown.event"))
}
