package build_application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
	buildApplicationRepo "github.com/davidsugianto/idp-core/internal/repository/build_application"
	deliveryTargetRepo "github.com/davidsugianto/idp-core/internal/repository/delivery_target"
	gitopsRepo "github.com/davidsugianto/idp-core/internal/repository/gitops"
	notificationUsecase "github.com/davidsugianto/idp-core/internal/usecase/notification"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrApplicationNotFound          = errors.New("build application not found")
	ErrApplicationAlreadyExists     = errors.New("build application already exists")
	ErrBuildNotFound                = errors.New("build not found")
	ErrBuildNotCancelable           = errors.New("build is already terminal")
	ErrBuildRetryNotAllowed         = errors.New("build retry is not allowed from current state")
	ErrBuildIdempotencyConflict     = errors.New("conflicting build idempotency key")
	ErrBuildUnavailable             = errors.New("build state unavailable")
	ErrUnauthorizedTeamScope        = errors.New("unauthorized team scope")
	ErrInvalidRuntimeFamily         = errors.New("unsupported runtime family")
	ErrInvalidRegistryType          = errors.New("unsupported registry type")
	ErrInvalidApplicationStatus     = errors.New("invalid application status")
	ErrInvalidBuildStatus           = errors.New("invalid build status")
	ErrInvalidBuildTriggerType      = errors.New("invalid build trigger type")
	ErrInvalidLifecycleEventType    = errors.New("invalid lifecycle event type")
	ErrInvalidDeploymentUpdateState = errors.New("invalid deployment update status")
	ErrInvalidSecurityStatus        = errors.New("invalid security verification status")
)

type Usecase interface {
	CreateApplication(ctx context.Context, teamID, actorID string, req *buildApplicationModel.CreateBuildApplicationRequest) (*buildApplicationModel.BuildApplicationResponse, error)
	ListApplications(ctx context.Context, teamID string, req *buildApplicationModel.ListBuildApplicationsRequest) (*buildApplicationModel.BuildApplicationListResponse, error)
	GetApplication(ctx context.Context, teamID, applicationID string) (*buildApplicationModel.BuildApplicationResponse, error)
	UpdateApplication(ctx context.Context, teamID, actorID, applicationID string, req *buildApplicationModel.UpdateBuildApplicationRequest) (*buildApplicationModel.BuildApplicationResponse, error)
	DeleteApplication(ctx context.Context, teamID, actorID, applicationID string) error

	TriggerBuild(ctx context.Context, teamID, actorID, applicationID string, req *buildApplicationModel.TriggerBuildRequest) (*buildApplicationModel.BuildActionResponse, error)
	GetBuild(ctx context.Context, teamID, buildID string) (*buildApplicationModel.BuildResponse, error)
	ListBuilds(ctx context.Context, teamID, applicationID string, limit, offset int) (*buildApplicationModel.BuildHistoryResponse, error)
	RetryBuild(ctx context.Context, teamID, actorID, buildID string) (*buildApplicationModel.BuildActionResponse, error)
	CancelBuild(ctx context.Context, teamID, actorID, buildID string) (*buildApplicationModel.BuildActionResponse, error)
	StreamBuildLogs(ctx context.Context, teamID, buildID string, afterSequence int64, limit int) (*buildApplicationModel.BuildLogStreamResponse, error)
	DrainQueuedBuilds(ctx context.Context) error
}

type GitopsProvider func(ctx context.Context, target *deliveryTargetModel.TargetControlPlane) (gitopsRepo.Repository, error)

type usecase struct {
	repo               buildApplicationRepo.Repository
	deliveryTargetRepo deliveryTargetRepo.Repository
	gitopsRepo         gitopsRepo.Repository
	gitopsProvider     GitopsProvider
	notificationUC     notificationUsecase.Usecase
}

type Dependencies struct {
	BuildApplicationRepo buildApplicationRepo.Repository
	DeliveryTargetRepo   deliveryTargetRepo.Repository
	GitopsRepo           gitopsRepo.Repository
	GitopsProvider       GitopsProvider
	NotificationUC       notificationUsecase.Usecase
}

func New(deps Dependencies) Usecase {
	return &usecase{
		repo:               deps.BuildApplicationRepo,
		deliveryTargetRepo: deps.DeliveryTargetRepo,
		gitopsRepo:         deps.GitopsRepo,
		gitopsProvider:     deps.GitopsProvider,
		notificationUC:     deps.NotificationUC,
	}
}

func nowPtr() *time.Time {
	now := time.Now()
	return &now
}

func nextSequence(current int64) int64 {
	return current + 1
}

func normalizeRuntimeFamily(runtime string) string {
	return strings.ToLower(strings.TrimSpace(runtime))
}

func normalizeRegistryType(registryType string) string {
	return strings.ToLower(strings.TrimSpace(registryType))
}

func validateRuntimeFamily(runtime string) error {
	if err := buildApplicationModel.ValidateRuntimeFamily(normalizeRuntimeFamily(runtime)); err != nil {
		return ErrInvalidRuntimeFamily
	}
	return nil
}

func validateRegistryType(registryType string) error {
	if err := buildApplicationModel.ValidateRegistryType(normalizeRegistryType(registryType)); err != nil {
		return ErrInvalidRegistryType
	}
	return nil
}

func validateApplicationStatus(status string) error {
	if status == "" {
		return nil
	}
	if !buildApplicationModel.ValidApplicationStatus(status) {
		return ErrInvalidApplicationStatus
	}
	return nil
}

func validateBuildStatus(status string) error {
	if status == "" {
		return nil
	}
	if !buildApplicationModel.ValidBuildStatus(status) {
		return ErrInvalidBuildStatus
	}
	return nil
}

func validateBuildTriggerType(triggerType string) error {
	if triggerType == "" {
		return nil
	}
	if !buildApplicationModel.ValidBuildTriggerType(triggerType) {
		return ErrInvalidBuildTriggerType
	}
	return nil
}

func validateSecurityStatus(status string) error {
	if status == "" {
		return nil
	}
	if !buildApplicationModel.ValidSecurityStatus(status) {
		return ErrInvalidSecurityStatus
	}
	return nil
}

func validateDeploymentUpdateStatus(status string) error {
	if status == "" {
		return nil
	}
	if !buildApplicationModel.ValidDeploymentUpdateStatus(status) {
		return ErrInvalidDeploymentUpdateState
	}
	return nil
}

func validateLifecycleEventType(eventType string) error {
	if !buildApplicationModel.ValidLifecycleEventType(eventType) {
		return ErrInvalidLifecycleEventType
	}
	return nil
}

func isBuildTerminal(status string) bool {
	switch status {
	case buildApplicationModel.BuildStatusCanceled,
		buildApplicationModel.BuildStatusFailed,
		buildApplicationModel.BuildStatusBlocked,
		buildApplicationModel.BuildStatusDeploymentReady:
		return true
	default:
		return false
	}
}

func isRetryAllowed(status string) bool {
	switch status {
	case buildApplicationModel.BuildStatusFailed, buildApplicationModel.BuildStatusCanceled:
		return true
	default:
		return false
	}
}

func sanitizeError(err error) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	for _, secret := range []string{"token", "password", "authorization", "credential", "secret", "kubeconfig"} {
		message = strings.ReplaceAll(message, secret, "[redacted]")
		message = strings.ReplaceAll(message, strings.ToUpper(secret), "[redacted]")
	}
	return errors.New(message)
}

func (u *usecase) assertApplicationAccess(ctx context.Context, teamID, applicationID string) (*buildApplicationModel.BuildApplication, error) {
	app, err := u.repo.GetApplicationByIDAndTeam(ctx, applicationID, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrApplicationNotFound
		}
		return nil, err
	}
	if app == nil {
		return nil, ErrApplicationNotFound
	}
	if app.TeamID != teamID {
		return nil, ErrUnauthorizedTeamScope
	}
	return app, nil
}

func (u *usecase) assertBuildAccess(ctx context.Context, teamID, buildID string) (*buildApplicationModel.Build, error) {
	build, err := u.repo.GetBuildByIDAndTeam(ctx, buildID, teamID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrBuildNotFound
		}
		return nil, err
	}
	if build == nil {
		return nil, ErrBuildNotFound
	}
	if build.TeamID != teamID {
		return nil, ErrUnauthorizedTeamScope
	}
	return build, nil
}

func (u *usecase) createLifecycleEvent(ctx context.Context, teamID, applicationID, buildID, eventType, source, summary string) error {
	if err := validateLifecycleEventType(eventType); err != nil {
		return err
	}
	now := time.Now()
	event := &buildApplicationModel.LifecycleEvent{
		ID:             uuid.New().String(),
		TeamID:         teamID,
		ApplicationID:  applicationID,
		BuildID:        buildID,
		EventType:      eventType,
		EventSource:    source,
		PayloadSummary: summary,
		OccurredAt:     now,
		CreatedAt:      now,
	}
	if err := u.repo.CreateLifecycleEvent(ctx, event); err != nil {
		return err
	}

	u.notify(ctx, teamID, applicationID, lifecycleNotificationTitle(eventType), summary, lifecycleNotificationSeverity(eventType))
	return nil
}

func (u *usecase) notify(ctx context.Context, teamID, applicationID, title, message, severity string) {
	if u.notificationUC == nil {
		return
	}
	_ = u.notificationUC.Create(ctx, &notificationModel.Notification{
		ID:            uuid.New().String(),
		TeamID:        teamID,
		EnvironmentID: applicationID,
		Kind:          notificationModel.KindEnvironment,
		Severity:      severity,
		Title:         title,
		Message:       message,
		CreatedAt:     time.Now(),
	})
}

func (u *usecase) resolveRegistryType(ctx context.Context, registryTargetID string) (string, error) {
	if registryTargetID == "" || u.deliveryTargetRepo == nil {
		return "", nil
	}
	target, err := u.deliveryTargetRepo.GetByID(ctx, registryTargetID)
	if err != nil {
		return "", err
	}
	if target == nil {
		return "", fmt.Errorf("delivery target not found")
	}
	registryType := strings.ToLower(strings.TrimSpace(target.ControlPlaneType))
	if registryType == "" {
		registryType = strings.ToLower(strings.TrimSpace(target.Purpose))
	}
	if registryType == "" {
		registryType = strings.ToLower(strings.TrimSpace(target.Name))
	}
	return registryType, nil
}

func toBuildActionResponse(build *buildApplicationModel.Build) *buildApplicationModel.BuildActionResponse {
	return &buildApplicationModel.BuildActionResponse{
		Build: *buildApplicationModel.ToBuildResponse(build),
	}
}
