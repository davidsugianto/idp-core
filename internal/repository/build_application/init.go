package build_application

import (
	"context"
	"time"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
	"gorm.io/gorm"
)

type Repository interface {
	CreateApplication(ctx context.Context, app *buildApplicationModel.BuildApplication) error
	GetApplicationByID(ctx context.Context, id string) (*buildApplicationModel.BuildApplication, error)
	GetApplicationByIDAndTeam(ctx context.Context, id, teamID string) (*buildApplicationModel.BuildApplication, error)
	GetApplicationByNameAndTeam(ctx context.Context, name, teamID string) (*buildApplicationModel.BuildApplication, error)
	ListApplicationsByTeam(ctx context.Context, teamID string, req *buildApplicationModel.ListBuildApplicationsRequest) ([]buildApplicationModel.BuildApplication, int64, error)
	UpdateApplication(ctx context.Context, app *buildApplicationModel.BuildApplication) error
	SoftDeleteApplication(ctx context.Context, id, teamID string) error

	CreateBuild(ctx context.Context, build *buildApplicationModel.Build) error
	GetBuildByID(ctx context.Context, buildID string) (*buildApplicationModel.Build, error)
	GetBuildByIDAndTeam(ctx context.Context, buildID, teamID string) (*buildApplicationModel.Build, error)
	GetBuildByApplicationAndIdempotencyKey(ctx context.Context, applicationID, idempotencyKey string) (*buildApplicationModel.Build, error)
	GetLatestBuildByApplicationID(ctx context.Context, applicationID string) (*buildApplicationModel.Build, error)
	ListBuildsByApplication(ctx context.Context, applicationID, teamID string, limit, offset int) ([]buildApplicationModel.Build, int64, error)
	ListBuildsByTeam(ctx context.Context, teamID string, limit, offset int) ([]buildApplicationModel.Build, int64, error)
	GetLatestBuildSequenceByApplicationID(ctx context.Context, applicationID string) (int64, error)
	ClaimNextQueuedBuild(ctx context.Context, workerID string, leaseUntil time.Time) (*buildApplicationModel.Build, error)
	UpdateBuild(ctx context.Context, build *buildApplicationModel.Build) error
	UpdateBuildStatus(ctx context.Context, buildID, status, failureReason string) error

	CreateBuildArtifact(ctx context.Context, artifact *buildApplicationModel.BuildArtifact) error
	GetBuildArtifactByBuildID(ctx context.Context, buildID string) (*buildApplicationModel.BuildArtifact, error)
	UpdateBuildArtifact(ctx context.Context, artifact *buildApplicationModel.BuildArtifact) error

	CreateSecurityVerification(ctx context.Context, verification *buildApplicationModel.SecurityVerification) error
	GetSecurityVerificationByBuildID(ctx context.Context, buildID string) (*buildApplicationModel.SecurityVerification, error)
	UpdateSecurityVerification(ctx context.Context, verification *buildApplicationModel.SecurityVerification) error

	CreateDeploymentUpdate(ctx context.Context, update *buildApplicationModel.DeploymentUpdate) error
	GetDeploymentUpdateByBuildID(ctx context.Context, buildID string) (*buildApplicationModel.DeploymentUpdate, error)
	UpdateDeploymentUpdate(ctx context.Context, update *buildApplicationModel.DeploymentUpdate) error

	CreateLifecycleEvent(ctx context.Context, event *buildApplicationModel.LifecycleEvent) error
	ListLifecycleEventsByBuildID(ctx context.Context, buildID string) ([]buildApplicationModel.LifecycleEvent, error)
	ListLifecycleEventsByApplicationID(ctx context.Context, applicationID string) ([]buildApplicationModel.LifecycleEvent, error)

	AppendBuildLog(ctx context.Context, log *buildApplicationModel.BuildLog) error
	ListBuildLogs(ctx context.Context, buildID string, afterSequence int64, limit int) ([]buildApplicationModel.BuildLog, error)
	GetLatestBuildLogSequence(ctx context.Context, buildID string) (int64, error)
	GetBuildLogSummary(ctx context.Context, buildID string) (string, error)
}

type repository struct {
	db *gorm.DB
}

type Dependencies struct {
	Database *gorm.DB
}

func New(deps Dependencies) Repository {
	return &repository{db: deps.Database}
}
