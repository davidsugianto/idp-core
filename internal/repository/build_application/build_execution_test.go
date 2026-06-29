package build_application

import (
	"testing"
	"time"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func setupBuildExecutionRepoForTest(t *testing.T) Repository {
	dsn := "file:" + uuid.New().String() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(&buildApplicationModel.Build{}))
	assert.NoError(t, db.Exec("CREATE INDEX IF NOT EXISTS idx_builds_queue_claim ON builds(status, queued_at, created_at)").Error)
	return New(Dependencies{Database: db})
}

func TestGetLatestBuildSequenceByApplicationID(t *testing.T) {
	repo := setupBuildExecutionRepoForTest(t)

	seq, err := repo.GetLatestBuildSequenceByApplicationID(t.Context(), "app-1")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), seq)

	now := time.Now()
	assert.NoError(t, repo.CreateBuild(t.Context(), &buildApplicationModel.Build{ID: uuid.New().String(), ApplicationID: "app-1", TeamID: "team-1", SequenceNumber: 1, Status: buildApplicationModel.BuildStatusQueued, TriggerType: buildApplicationModel.BuildTriggerTypeManual, CreatedAt: now, UpdatedAt: now}))
	assert.NoError(t, repo.CreateBuild(t.Context(), &buildApplicationModel.Build{ID: uuid.New().String(), ApplicationID: "app-1", TeamID: "team-1", SequenceNumber: 2, Status: buildApplicationModel.BuildStatusRunning, TriggerType: buildApplicationModel.BuildTriggerTypeManual, CreatedAt: now, UpdatedAt: now}))

	seq, err = repo.GetLatestBuildSequenceByApplicationID(t.Context(), "app-1")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), seq)
}

func TestListBuildsByApplicationOrdersBySequenceDesc(t *testing.T) {
	repo := setupBuildExecutionRepoForTest(t)
	now := time.Now()

	assert.NoError(t, repo.CreateBuild(t.Context(), &buildApplicationModel.Build{ID: "b1", ApplicationID: "app-1", TeamID: "team-1", SequenceNumber: 1, Status: buildApplicationModel.BuildStatusQueued, TriggerType: buildApplicationModel.BuildTriggerTypeManual, CreatedAt: now, UpdatedAt: now}))
	assert.NoError(t, repo.CreateBuild(t.Context(), &buildApplicationModel.Build{ID: "b2", ApplicationID: "app-1", TeamID: "team-1", SequenceNumber: 2, Status: buildApplicationModel.BuildStatusRunning, TriggerType: buildApplicationModel.BuildTriggerTypeManual, CreatedAt: now, UpdatedAt: now}))

	builds, total, err := repo.ListBuildsByApplication(t.Context(), "app-1", "team-1", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, builds, 2)
	assert.Equal(t, int64(2), builds[0].SequenceNumber)
	assert.Equal(t, int64(1), builds[1].SequenceNumber)
}

func TestClaimNextQueuedBuildClaimsOldestQueuedBuild(t *testing.T) {
	repo := setupBuildExecutionRepoForTest(t)
	dbRepo := repo.(*repository)
	now := time.Now().UTC()
	earlier := now.Add(-2 * time.Minute)
	later := now.Add(-1 * time.Minute)

	assert.NoError(t, repo.CreateBuild(t.Context(), &buildApplicationModel.Build{
		ID:             "queued-oldest",
		ApplicationID:  "app-1",
		TeamID:         "team-1",
		SequenceNumber: 1,
		Status:         buildApplicationModel.BuildStatusQueued,
		TriggerType:    buildApplicationModel.BuildTriggerTypeManual,
		QueuedAt:       &earlier,
		CreatedAt:      earlier,
		UpdatedAt:      earlier,
	}))
	assert.NoError(t, repo.CreateBuild(t.Context(), &buildApplicationModel.Build{
		ID:             "queued-newer",
		ApplicationID:  "app-1",
		TeamID:         "team-1",
		SequenceNumber: 2,
		Status:         buildApplicationModel.BuildStatusQueued,
		TriggerType:    buildApplicationModel.BuildTriggerTypeManual,
		QueuedAt:       &later,
		CreatedAt:      later,
		UpdatedAt:      later,
	}))

	leaseUntil := now.Add(5 * time.Minute)
	claimed, err := repo.ClaimNextQueuedBuild(t.Context(), "worker-1", leaseUntil)
	assert.NoError(t, err)
	assert.NotNil(t, claimed)
	assert.Equal(t, "queued-oldest", claimed.ID)
	assert.Equal(t, buildApplicationModel.BuildStatusRunning, claimed.Status)
	assert.Equal(t, "worker-1", claimed.ExecutionWorkerID)
	assert.Equal(t, 1, claimed.ExecutionAttempts)
	assert.NotNil(t, claimed.StartedAt)
	assert.NotNil(t, claimed.ExecutionClaimedAt)
	assert.NotNil(t, claimed.ExecutionLeaseExpiresAt)
	assert.WithinDuration(t, leaseUntil, *claimed.ExecutionLeaseExpiresAt, time.Second)

	persisted, err := repo.GetBuildByID(t.Context(), "queued-oldest")
	assert.NoError(t, err)
	assert.Equal(t, buildApplicationModel.BuildStatusRunning, persisted.Status)
	assert.Equal(t, "worker-1", persisted.ExecutionWorkerID)
	assert.Equal(t, 1, persisted.ExecutionAttempts)

	var stillQueued buildApplicationModel.Build
	assert.NoError(t, dbRepo.db.WithContext(t.Context()).Clauses(clause.Locking{Strength: "SHARE"}).Where("id = ?", "queued-newer").First(&stillQueued).Error)
	assert.Equal(t, buildApplicationModel.BuildStatusQueued, stillQueued.Status)
}

func TestClaimNextQueuedBuildReturnsNilWhenQueueEmpty(t *testing.T) {
	repo := setupBuildExecutionRepoForTest(t)
	claimed, err := repo.ClaimNextQueuedBuild(t.Context(), "worker-1", time.Now().UTC().Add(5*time.Minute))
	assert.NoError(t, err)
	assert.Nil(t, claimed)
}
