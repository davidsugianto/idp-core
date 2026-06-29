package build_application

import buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"

const (
	notificationSeverityInfo    = "info"
	notificationSeveritySuccess = "success"
	notificationSeverityError   = "error"
)

func lifecycleNotificationSeverity(eventType string) string {
	switch eventType {
	case buildApplicationModel.EventTypeBuildFailed,
		buildApplicationModel.EventTypeDeployFailed,
		buildApplicationModel.EventTypeSecurityBlocked:
		return notificationSeverityError
	case buildApplicationModel.EventTypeBuildSucceeded,
		buildApplicationModel.EventTypeDeploySucceeded:
		return notificationSeveritySuccess
	default:
		return notificationSeverityInfo
	}
}

func lifecycleNotificationTitle(eventType string) string {
	switch eventType {
	case buildApplicationModel.EventTypeApplicationCreated:
		return "Build application created"
	case buildApplicationModel.EventTypeApplicationUpdated:
		return "Build application updated"
	case buildApplicationModel.EventTypeApplicationDeleted:
		return "Build application deleted"
	case buildApplicationModel.EventTypeBuildQueued:
		return "Build queued"
	case buildApplicationModel.EventTypeBuildRunning:
		return "Build running"
	case buildApplicationModel.EventTypeBuildCanceled:
		return "Build canceled"
	case buildApplicationModel.EventTypeBuildFailed:
		return "Build failed"
	case buildApplicationModel.EventTypeBuildSucceeded:
		return "Build completed"
	case buildApplicationModel.EventTypeSecuritySBOM:
		return "SBOM generated"
	case buildApplicationModel.EventTypeSecurityScan:
		return "Security scan completed"
	case buildApplicationModel.EventTypeSecuritySigning:
		return "Image signing completed"
	case buildApplicationModel.EventTypeSecurityBlocked:
		return "Security policy blocked deployment"
	case buildApplicationModel.EventTypeDeployStarted:
		return "Deployment update started"
	case buildApplicationModel.EventTypeDeploySucceeded:
		return "Deployment update succeeded"
	case buildApplicationModel.EventTypeDeployFailed:
		return "Deployment update failed"
	default:
		return "Build lifecycle updated"
	}
}
