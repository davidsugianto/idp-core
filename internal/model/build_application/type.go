package build_application

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	ApplicationStatusActive    = "active"
	ApplicationStatusSuspended = "suspended"
	ApplicationStatusDeleting  = "deleting"
	ApplicationStatusDeleted   = "deleted"
)

const (
	RuntimeFamilyGo     = "go"
	RuntimeFamilyJava   = "java"
	RuntimeFamilyNode   = "node"
	RuntimeFamilyPython = "python"
	RuntimeFamilyDotNet = "dotnet"
)

const (
	RuntimeDetectionModeAuto   = "auto"
	RuntimeDetectionModeManual = "manual"
)

const (
	BuildStatusPending         = "pending"
	BuildStatusQueued          = "queued"
	BuildStatusRunning         = "running"
	BuildStatusCanceling       = "canceling"
	BuildStatusCanceled        = "canceled"
	BuildStatusFailed          = "failed"
	BuildStatusSucceeded       = "succeeded"
	BuildStatusBlocked         = "blocked"
	BuildStatusDeploymentReady = "deployment_ready"
)

const (
	BuildTriggerTypeManual = "manual"
	BuildTriggerTypeRetry  = "retry"
	BuildTriggerTypeSystem = "system"
)

const (
	SecurityStatusPending = "pending"
	SecurityStatusPassed  = "passed"
	SecurityStatusFailed  = "failed"
	SecurityStatusWaived  = "waived"
)

const (
	DeploymentUpdateStatusPending    = "pending"
	DeploymentUpdateStatusInProgress = "in_progress"
	DeploymentUpdateStatusSucceeded  = "succeeded"
	DeploymentUpdateStatusFailed     = "failed"
)

const (
	EventTypeApplicationCreated = "application.created"
	EventTypeApplicationUpdated = "application.updated"
	EventTypeApplicationDeleted = "application.deleted"
	EventTypeBuildQueued        = "build.queued"
	EventTypeBuildRunning       = "build.running"
	EventTypeBuildCanceled      = "build.canceled"
	EventTypeBuildFailed        = "build.failed"
	EventTypeBuildSucceeded     = "build.succeeded"
	EventTypeSecuritySBOM       = "security.sbom_generated"
	EventTypeSecurityScan       = "security.scan_completed"
	EventTypeSecuritySigning    = "security.signing_completed"
	EventTypeSecurityBlocked    = "security.policy_blocked"
	EventTypeDeployStarted      = "deployment.update_started"
	EventTypeDeploySucceeded    = "deployment.update_succeeded"
	EventTypeDeployFailed       = "deployment.update_failed"
)

const (
	RegistryTypeHarbor = "harbor"
	RegistryTypeGHCR   = "ghcr"
	RegistryTypeECR    = "ecr"
	RegistryTypeGCR    = "gcr"
)

type BuildApplication struct {
	ID                          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TeamID                      string         `gorm:"type:varchar(36);not null;index" json:"team_id"`
	Name                        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description                 string         `gorm:"type:text" json:"description,omitempty"`
	Status                      string         `gorm:"type:varchar(32);not null" json:"status"`
	RepositoryURL               string         `gorm:"type:varchar(512);not null" json:"repository_url"`
	RepositoryProvider          string         `gorm:"type:varchar(64)" json:"repository_provider,omitempty"`
	DefaultBranch               string         `gorm:"type:varchar(128)" json:"default_branch,omitempty"`
	ApplicationDescriptorPath   string         `gorm:"type:varchar(512);not null" json:"application_descriptor_path"`
	RuntimeFamily               string         `gorm:"type:varchar(32)" json:"runtime_family,omitempty"`
	RuntimeDetectionMode        string         `gorm:"type:varchar(32)" json:"runtime_detection_mode,omitempty"`
	BuilderProfileID            string         `gorm:"type:varchar(36)" json:"builder_profile_id,omitempty"`
	RegistryTargetID            string         `gorm:"type:varchar(36)" json:"registry_target_id,omitempty"`
	DeploymentAutomationEnabled bool           `gorm:"not null;default:false" json:"deployment_automation_enabled"`
	GitOpsTargetID              string         `gorm:"type:varchar(36)" json:"gitops_target_id,omitempty"`
	CreatedBy                   string         `gorm:"type:varchar(36)" json:"created_by,omitempty"`
	UpdatedBy                   string         `gorm:"type:varchar(36)" json:"updated_by,omitempty"`
	CreatedAt                   time.Time      `json:"created_at"`
	UpdatedAt                   time.Time      `json:"updated_at"`
	DeletedAt                   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (BuildApplication) TableName() string {
	return "build_applications"
}

type Build struct {
	ID                      string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	ApplicationID           string     `gorm:"type:varchar(36);not null;index" json:"application_id"`
	TeamID                  string     `gorm:"type:varchar(36);not null;index" json:"team_id"`
	SequenceNumber          int64      `gorm:"not null" json:"sequence_number"`
	Status                  string     `gorm:"type:varchar(32);not null;index" json:"status"`
	TriggerType             string     `gorm:"type:varchar(32);not null" json:"trigger_type"`
	TriggeredBy             string     `gorm:"type:varchar(36)" json:"triggered_by,omitempty"`
	IdempotencyKey          string     `gorm:"type:varchar(255)" json:"idempotency_key,omitempty"`
	SourceRevisionRequested string     `gorm:"type:varchar(255)" json:"source_revision_requested,omitempty"`
	SourceRevisionResolved  string     `gorm:"type:varchar(255)" json:"source_revision_resolved,omitempty"`
	RetryOfBuildID          string     `gorm:"type:varchar(36)" json:"retry_of_build_id,omitempty"`
	CancelRequestedBy       string     `gorm:"type:varchar(36)" json:"cancel_requested_by,omitempty"`
	FailureReason           string     `gorm:"type:text" json:"failure_reason,omitempty"`
	ExecutionWorkerID       string     `gorm:"type:varchar(128);index" json:"-"`
	ExecutionClaimedAt      *time.Time `json:"-"`
	ExecutionLeaseExpiresAt *time.Time `json:"-"`
	ExecutionAttempts       int        `gorm:"not null;default:0" json:"-"`
	QueuedAt                *time.Time `json:"queued_at,omitempty"`
	StartedAt               *time.Time `json:"started_at,omitempty"`
	FinishedAt              *time.Time `json:"finished_at,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

func (Build) TableName() string {
	return "builds"
}

type BuildArtifact struct {
	ID                     string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	BuildID                string     `gorm:"type:varchar(36);not null;index" json:"build_id"`
	ApplicationID          string     `gorm:"type:varchar(36);not null;index" json:"application_id"`
	RegistryTargetID       string     `gorm:"type:varchar(36)" json:"registry_target_id,omitempty"`
	ImageRepository        string     `gorm:"type:varchar(512)" json:"image_repository,omitempty"`
	ImageTag               string     `gorm:"type:varchar(255)" json:"image_tag,omitempty"`
	ImageDigest            string     `gorm:"type:varchar(255)" json:"image_digest,omitempty"`
	PublishedImageReference string    `gorm:"type:varchar(1024)" json:"published_image_reference,omitempty"`
	PublishedAt            *time.Time `json:"published_at,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

func (BuildArtifact) TableName() string {
	return "build_artifacts"
}

type SecurityVerification struct {
	ID                 string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	BuildID            string     `gorm:"type:varchar(36);not null;index" json:"build_id"`
	ArtifactID         string     `gorm:"type:varchar(36)" json:"artifact_id,omitempty"`
	Status             string     `gorm:"type:varchar(32);not null" json:"status"`
	SBOMStatus         string     `gorm:"type:varchar(32)" json:"sbom_status,omitempty"`
	SBOMReference      string     `gorm:"type:varchar(1024)" json:"sbom_reference,omitempty"`
	ScanStatus         string     `gorm:"type:varchar(32)" json:"scan_status,omitempty"`
	ScanReference      string     `gorm:"type:varchar(1024)" json:"scan_reference,omitempty"`
	ScanSummary        string     `gorm:"type:text" json:"scan_summary,omitempty"`
	SigningStatus      string     `gorm:"type:varchar(32)" json:"signing_status,omitempty"`
	SignatureReference string     `gorm:"type:varchar(1024)" json:"signature_reference,omitempty"`
	PolicyGateStatus   string     `gorm:"type:varchar(32)" json:"policy_gate_status,omitempty"`
	PolicyGateReason   string     `gorm:"type:text" json:"policy_gate_reason,omitempty"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (SecurityVerification) TableName() string {
	return "security_verifications"
}

type DeploymentUpdate struct {
	ID                    string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	BuildID               string     `gorm:"type:varchar(36);not null;index" json:"build_id"`
	ApplicationID         string     `gorm:"type:varchar(36);not null;index" json:"application_id"`
	Status                string     `gorm:"type:varchar(32);not null" json:"status"`
	GitOpsTargetID        string     `gorm:"type:varchar(36)" json:"gitops_target_id,omitempty"`
	RequestedImageReference string   `gorm:"type:varchar(1024)" json:"requested_image_reference,omitempty"`
	RequestedManifestPath string     `gorm:"type:varchar(1024)" json:"requested_manifest_path,omitempty"`
	ResultingRevision     string     `gorm:"type:varchar(255)" json:"resulting_revision,omitempty"`
	FailureReason         string     `gorm:"type:text" json:"failure_reason,omitempty"`
	StartedAt             *time.Time `json:"started_at,omitempty"`
	FinishedAt            *time.Time `json:"finished_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

func (DeploymentUpdate) TableName() string {
	return "deployment_updates"
}

type LifecycleEvent struct {
	ID             string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TeamID         string    `gorm:"type:varchar(36);not null;index" json:"team_id"`
	ApplicationID  string    `gorm:"type:varchar(36);index" json:"application_id,omitempty"`
	BuildID        string    `gorm:"type:varchar(36);index" json:"build_id,omitempty"`
	EventType      string    `gorm:"type:varchar(64);not null" json:"event_type"`
	EventSource    string    `gorm:"type:varchar(64)" json:"event_source,omitempty"`
	PayloadSummary string    `gorm:"type:text" json:"payload_summary,omitempty"`
	OccurredAt     time.Time `json:"occurred_at"`
	CreatedAt      time.Time `json:"created_at"`
}

func (LifecycleEvent) TableName() string {
	return "lifecycle_events"
}

type BuildLog struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	BuildID   string    `gorm:"type:varchar(36);not null;index" json:"build_id"`
	Sequence  int64     `gorm:"not null" json:"sequence"`
	Line      string    `gorm:"type:text;not null" json:"line"`
	CreatedAt time.Time `json:"created_at"`
}

func (BuildLog) TableName() string {
	return "build_logs"
}

type BuildApplicationBuilderProfile struct {
	ID                       string   `json:"id"`
	Name                     string   `json:"name"`
	SupportedRuntimeFamilies []string `json:"supported_runtime_families,omitempty"`
}

type BuildApplicationRegistryTarget struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	RegistryType   string `json:"registry_type"`
	Host           string `json:"host,omitempty"`
	RepositoryPath string `json:"repository_path,omitempty"`
}

type BuildApplicationGitOpsTarget struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateBuildApplicationRequest struct {
	Name                        string `json:"name" binding:"required"`
	Description                 string `json:"description"`
	RepositoryURL               string `json:"repository_url" binding:"required"`
	ApplicationDescriptorPath   string `json:"application_descriptor_path" binding:"required"`
	RuntimeFamily               string `json:"runtime_family"`
	RuntimeDetectionMode        string `json:"runtime_detection_mode"`
	BuilderProfileID            string `json:"builder_profile_id"`
	RegistryTargetID            string `json:"registry_target_id"`
	DeploymentAutomationEnabled bool   `json:"deployment_automation_enabled"`
	GitOpsTargetID              string `json:"gitops_target_id"`
}

type UpdateBuildApplicationRequest struct {
	Description                 *string `json:"description"`
	RuntimeFamily               *string `json:"runtime_family"`
	RuntimeDetectionMode        *string `json:"runtime_detection_mode"`
	BuilderProfileID            *string `json:"builder_profile_id"`
	RegistryTargetID            *string `json:"registry_target_id"`
	DeploymentAutomationEnabled *bool   `json:"deployment_automation_enabled"`
	GitOpsTargetID              *string `json:"gitops_target_id"`
	Status                      *string `json:"status"`
}

type ListBuildApplicationsRequest struct {
	Status                     string `form:"status"`
	RuntimeFamily              string `form:"runtime_family"`
	RegistryTargetID           string `form:"registry_target_id"`
	DeploymentAutomationEnabled *bool `form:"deployment_automation_enabled"`
	Limit                      int    `form:"limit"`
	Offset                     int    `form:"offset"`
}

type BuildApplicationResponse struct {
	ID                          string                                `json:"id"`
	TeamID                      string                                `json:"team_id"`
	Name                        string                                `json:"name"`
	Description                 string                                `json:"description,omitempty"`
	Status                      string                                `json:"status"`
	RepositoryURL               string                                `json:"repository_url"`
	ApplicationDescriptorPath   string                                `json:"application_descriptor_path"`
	RuntimeFamily               string                                `json:"runtime_family,omitempty"`
	RuntimeDetectionMode        string                                `json:"runtime_detection_mode,omitempty"`
	BuilderProfile              *BuildApplicationBuilderProfile       `json:"builder_profile,omitempty"`
	RegistryTarget              *BuildApplicationRegistryTarget       `json:"registry_target,omitempty"`
	DeploymentAutomationEnabled bool                                  `json:"deployment_automation_enabled"`
	GitOpsTarget                *BuildApplicationGitOpsTarget         `json:"gitops_target,omitempty"`
	CreatedAt                   string                                `json:"created_at"`
	UpdatedAt                   string                                `json:"updated_at"`
}

type BuildApplicationListResponse struct {
	Applications []BuildApplicationResponse `json:"applications"`
	Total        int64                      `json:"total"`
}

type TriggerBuildRequest struct {
	SourceRevision  string `json:"source_revision"`
	IdempotencyKey  string `json:"idempotency_key"`
}

type BuildActionResponse struct {
	Build BuildResponse `json:"build"`
}

type BuildResponse struct {
	ID                     string                        `json:"id"`
	ApplicationID          string                        `json:"application_id"`
	TeamID                 string                        `json:"team_id"`
	Status                 string                        `json:"status"`
	TriggerType            string                        `json:"trigger_type"`
	SourceRevisionRequested string                       `json:"source_revision_requested,omitempty"`
	SourceRevisionResolved string                        `json:"source_revision_resolved,omitempty"`
	RetryOfBuildID         string                        `json:"retry_of_build_id,omitempty"`
	QueuedAt               *string                       `json:"queued_at,omitempty"`
	StartedAt              *string                       `json:"started_at,omitempty"`
	FinishedAt             *string                       `json:"finished_at,omitempty"`
	FailureReason          string                        `json:"failure_reason,omitempty"`
	Artifact               *BuildArtifactResponse        `json:"artifact,omitempty"`
	SecurityVerification   *SecurityVerificationResponse `json:"security_verification,omitempty"`
	DeploymentUpdate       *DeploymentUpdateResponse     `json:"deployment_update,omitempty"`
}

type BuildHistoryResponse struct {
	Builds []BuildResponse `json:"builds"`
	Total  int64           `json:"total"`
}

type BuildArtifactResponse struct {
	PublishedImageReference string `json:"published_image_reference,omitempty"`
	ImageDigest             string `json:"image_digest,omitempty"`
}

type SecurityVerificationResponse struct {
	Status           string `json:"status"`
	SBOMStatus       string `json:"sbom_status,omitempty"`
	ScanStatus       string `json:"scan_status,omitempty"`
	SigningStatus    string `json:"signing_status,omitempty"`
	PolicyGateStatus string `json:"policy_gate_status,omitempty"`
	PolicyGateReason string `json:"policy_gate_reason,omitempty"`
}

type DeploymentUpdateResponse struct {
	Status            string `json:"status"`
	ResultingRevision string `json:"resulting_revision,omitempty"`
	FailureReason     string `json:"failure_reason,omitempty"`
}

type LifecycleEventResponse struct {
	EventID     string `json:"event_id"`
	EventType   string `json:"event_type"`
	TeamID      string `json:"team_id"`
	ApplicationID string `json:"application_id,omitempty"`
	BuildID     string `json:"build_id,omitempty"`
	OccurredAt  string `json:"occurred_at"`
	Summary     string `json:"summary"`
}

type BuildLogStreamResponse struct {
	BuildID         string  `json:"build_id"`
	StreamState     string  `json:"stream_state"`
	LastSequence    int64   `json:"last_sequence"`
	TerminalSummary string  `json:"terminal_summary,omitempty"`
	Lines           []string `json:"lines,omitempty"`
}

func ValidateRuntimeFamily(runtimeFamily string) error {
	normalized := strings.ToLower(strings.TrimSpace(runtimeFamily))
	if normalized == "" {
		return nil
	}
	switch normalized {
	case RuntimeFamilyGo, RuntimeFamilyJava, RuntimeFamilyNode, RuntimeFamilyPython, RuntimeFamilyDotNet:
		return nil
	default:
		return errors.New("unsupported runtime family")
	}
}

func ValidateRegistryType(registryType string) error {
	normalized := strings.ToLower(strings.TrimSpace(registryType))
	if normalized == "" {
		return nil
	}
	switch normalized {
	case RegistryTypeHarbor, RegistryTypeGHCR, RegistryTypeECR, RegistryTypeGCR:
		return nil
	default:
		return errors.New("unsupported registry type")
	}
}

func ValidApplicationStatus(status string) bool {
	switch status {
	case ApplicationStatusActive, ApplicationStatusSuspended, ApplicationStatusDeleting, ApplicationStatusDeleted:
		return true
	default:
		return false
	}
}

func ValidBuildStatus(status string) bool {
	switch status {
	case BuildStatusPending, BuildStatusQueued, BuildStatusRunning, BuildStatusCanceling, BuildStatusCanceled, BuildStatusFailed, BuildStatusSucceeded, BuildStatusBlocked, BuildStatusDeploymentReady:
		return true
	default:
		return false
	}
}

func ValidBuildTriggerType(triggerType string) bool {
	switch triggerType {
	case BuildTriggerTypeManual, BuildTriggerTypeRetry, BuildTriggerTypeSystem:
		return true
	default:
		return false
	}
}

func ValidSecurityStatus(status string) bool {
	switch status {
	case SecurityStatusPending, SecurityStatusPassed, SecurityStatusFailed, SecurityStatusWaived:
		return true
	default:
		return false
	}
}

func ValidDeploymentUpdateStatus(status string) bool {
	switch status {
	case DeploymentUpdateStatusPending, DeploymentUpdateStatusInProgress, DeploymentUpdateStatusSucceeded, DeploymentUpdateStatusFailed:
		return true
	default:
		return false
	}
}

func ValidLifecycleEventType(eventType string) bool {
	switch eventType {
	case EventTypeApplicationCreated,
		EventTypeApplicationUpdated,
		EventTypeApplicationDeleted,
		EventTypeBuildQueued,
		EventTypeBuildRunning,
		EventTypeBuildCanceled,
		EventTypeBuildFailed,
		EventTypeBuildSucceeded,
		EventTypeSecuritySBOM,
		EventTypeSecurityScan,
		EventTypeSecuritySigning,
		EventTypeSecurityBlocked,
		EventTypeDeployStarted,
		EventTypeDeploySucceeded,
		EventTypeDeployFailed:
		return true
	default:
		return false
	}
}

func ToBuildApplicationResponse(app *BuildApplication) *BuildApplicationResponse {
	if app == nil {
		return nil
	}
	return &BuildApplicationResponse{
		ID:                          app.ID,
		TeamID:                      app.TeamID,
		Name:                        app.Name,
		Description:                 app.Description,
		Status:                      app.Status,
		RepositoryURL:               app.RepositoryURL,
		ApplicationDescriptorPath:   app.ApplicationDescriptorPath,
		RuntimeFamily:               app.RuntimeFamily,
		RuntimeDetectionMode:        app.RuntimeDetectionMode,
		DeploymentAutomationEnabled: app.DeploymentAutomationEnabled,
		CreatedAt:                   app.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                   app.UpdatedAt.Format(time.RFC3339),
	}
}

func ToBuildApplicationListResponse(applications []BuildApplication, total int64) *BuildApplicationListResponse {
	responses := make([]BuildApplicationResponse, len(applications))
	for i := range applications {
		responses[i] = *ToBuildApplicationResponse(&applications[i])
	}
	return &BuildApplicationListResponse{Applications: responses, Total: total}
}

func timeToRFC3339Ptr(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(time.RFC3339)
	return &formatted
}

func ToBuildResponse(build *Build) *BuildResponse {
	if build == nil {
		return nil
	}
	return &BuildResponse{
		ID:                      build.ID,
		ApplicationID:           build.ApplicationID,
		TeamID:                  build.TeamID,
		Status:                  build.Status,
		TriggerType:             build.TriggerType,
		SourceRevisionRequested: build.SourceRevisionRequested,
		SourceRevisionResolved:  build.SourceRevisionResolved,
		RetryOfBuildID:          build.RetryOfBuildID,
		QueuedAt:                timeToRFC3339Ptr(build.QueuedAt),
		StartedAt:               timeToRFC3339Ptr(build.StartedAt),
		FinishedAt:              timeToRFC3339Ptr(build.FinishedAt),
		FailureReason:           build.FailureReason,
	}
}

func ToBuildHistoryResponse(builds []Build, total int64) *BuildHistoryResponse {
	responses := make([]BuildResponse, len(builds))
	for i := range builds {
		responses[i] = *ToBuildResponse(&builds[i])
	}
	return &BuildHistoryResponse{Builds: responses, Total: total}
}
