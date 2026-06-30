package build_application

import (
	"testing"
	"time"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
	buildApplicationRepo "github.com/davidsugianto/idp-core/internal/repository/build_application"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBuildApplicationUsecaseForTest(t *testing.T) (Usecase, buildApplicationRepo.Repository) {
	t.Helper()

	dsn := "file:" + uuid.NewString() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(
		&buildApplicationModel.BuildApplication{},
		&buildApplicationModel.Build{},
		&buildApplicationModel.BuildArtifact{},
		&buildApplicationModel.SecurityVerification{},
		&buildApplicationModel.DeploymentUpdate{},
		&buildApplicationModel.LifecycleEvent{},
		&buildApplicationModel.BuildLog{},
	))
	assert.NoError(t, db.Exec("CREATE INDEX IF NOT EXISTS idx_builds_queue_claim ON builds(status, queued_at, created_at)").Error)

	repo := buildApplicationRepo.New(buildApplicationRepo.Dependencies{Database: db})
	uc := New(Dependencies{BuildApplicationRepo: repo})
	return uc, repo
}

func TestDrainQueuedBuildsTransitionsQueuedBuildAndAppendsLogs(t *testing.T) {
	uc, repo := setupBuildApplicationUsecaseForTest(t)
	now := time.Now().UTC()
	queuedAt := now.Add(-time.Minute)

	app := &buildApplicationModel.BuildApplication{
		ID:                        "app-1",
		TeamID:                    "team-1",
		Name:                      "worker-app",
		Status:                    buildApplicationModel.ApplicationStatusActive,
		RepositoryURL:             "https://example.com/repo.git",
		ApplicationDescriptorPath: "application.yaml",
		DefaultBranch:             "main",
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}
	assert.NoError(t, repo.CreateApplication(t.Context(), app))

	build := &buildApplicationModel.Build{
		ID:             "build-1",
		ApplicationID:  app.ID,
		TeamID:         app.TeamID,
		SequenceNumber: 1,
		Status:         buildApplicationModel.BuildStatusQueued,
		TriggerType:    buildApplicationModel.BuildTriggerTypeManual,
		QueuedAt:       &queuedAt,
		CreatedAt:      queuedAt,
		UpdatedAt:      queuedAt,
	}
	assert.NoError(t, repo.CreateBuild(t.Context(), build))

	assert.NoError(t, uc.DrainQueuedBuilds(t.Context()))

	persisted, err := repo.GetBuildByID(t.Context(), build.ID)
	assert.NoError(t, err)
	assert.Equal(t, buildApplicationModel.BuildStatusDeploymentReady, persisted.Status)
	assert.Equal(t, "main", persisted.SourceRevisionResolved)
	assert.NotNil(t, persisted.StartedAt)
	assert.NotNil(t, persisted.FinishedAt)
	assert.Equal(t, 1, persisted.ExecutionAttempts)

	artifact, err := repo.GetBuildArtifactByBuildID(t.Context(), build.ID)
	assert.NoError(t, err)
	assert.NotEmpty(t, artifact.PublishedImageReference)

	verification, err := repo.GetSecurityVerificationByBuildID(t.Context(), build.ID)
	assert.NoError(t, err)
	assert.Equal(t, buildApplicationModel.SecurityStatusPassed, verification.Status)
	assert.Equal(t, buildApplicationModel.SecurityStatusPassed, verification.PolicyGateStatus)

	logs, err := repo.ListBuildLogs(t.Context(), build.ID, 0, 20)
	assert.NoError(t, err)
	assert.Len(t, logs, 5)
	assert.Equal(t, int64(1), logs[0].Sequence)
	assert.Equal(t, "build claimed by worker", logs[0].Line)
	assert.Equal(t, "build completed with status deployment_ready", logs[len(logs)-1].Line)
}

func TestCancelBuildCancelsQueuedBuildImmediately(t *testing.T) {
	uc, repo := setupBuildApplicationUsecaseForTest(t)
	now := time.Now().UTC()

	app := &buildApplicationModel.BuildApplication{
		ID:                        "app-1",
		TeamID:                    "team-1",
		Name:                      "cancel-app",
		Status:                    buildApplicationModel.ApplicationStatusActive,
		RepositoryURL:             "https://example.com/repo.git",
		ApplicationDescriptorPath: "application.yaml",
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}
	assert.NoError(t, repo.CreateApplication(t.Context(), app))

	build := &buildApplicationModel.Build{
		ID:             "build-queued",
		ApplicationID:  app.ID,
		TeamID:         app.TeamID,
		SequenceNumber: 1,
		Status:         buildApplicationModel.BuildStatusQueued,
		TriggerType:    buildApplicationModel.BuildTriggerTypeManual,
		QueuedAt:       &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	assert.NoError(t, repo.CreateBuild(t.Context(), build))

	response, err := uc.CancelBuild(t.Context(), app.TeamID, "user-1", build.ID)
	assert.NoError(t, err)
	assert.Equal(t, buildApplicationModel.BuildStatusCanceled, response.Build.Status)
	assert.NotNil(t, response.Build.FinishedAt)

	persisted, err := repo.GetBuildByID(t.Context(), build.ID)
	assert.NoError(t, err)
	assert.Equal(t, buildApplicationModel.BuildStatusCanceled, persisted.Status)
	assert.Equal(t, "user-1", persisted.CancelRequestedBy)
	assert.NotNil(t, persisted.FinishedAt)
}
