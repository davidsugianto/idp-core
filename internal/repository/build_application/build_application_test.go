package build_application

import (
	"testing"
	"time"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBuildApplicationRepoForTest(t *testing.T) Repository {
	dsn := "file:" + uuid.New().String() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	assert.NoError(t, err)
	assert.NoError(t, db.AutoMigrate(&buildApplicationModel.BuildApplication{}))
	return New(Dependencies{Database: db})
}

func TestListApplicationsByTeamFiltersDeploymentAutomation(t *testing.T) {
	repo := setupBuildApplicationRepoForTest(t)
	now := time.Now()

	appEnabled := &buildApplicationModel.BuildApplication{ID: uuid.New().String(), TeamID: "team-1", Name: "app-enabled", Status: buildApplicationModel.ApplicationStatusActive, RepositoryURL: "https://example.com/a.git", ApplicationDescriptorPath: "application.yaml", DeploymentAutomationEnabled: true, CreatedAt: now, UpdatedAt: now}
	appDisabled := &buildApplicationModel.BuildApplication{ID: uuid.New().String(), TeamID: "team-1", Name: "app-disabled", Status: buildApplicationModel.ApplicationStatusActive, RepositoryURL: "https://example.com/b.git", ApplicationDescriptorPath: "application.yaml", DeploymentAutomationEnabled: false, CreatedAt: now, UpdatedAt: now}

	assert.NoError(t, repo.CreateApplication(t.Context(), appEnabled))
	assert.NoError(t, repo.CreateApplication(t.Context(), appDisabled))

	enabled := true
	resp, total, err := repo.ListApplicationsByTeam(t.Context(), "team-1", &buildApplicationModel.ListBuildApplicationsRequest{DeploymentAutomationEnabled: &enabled})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, resp, 1)
	assert.Equal(t, appEnabled.ID, resp[0].ID)
}

func TestSoftDeleteApplicationRemovesFromList(t *testing.T) {
	repo := setupBuildApplicationRepoForTest(t)
	now := time.Now()
	app := &buildApplicationModel.BuildApplication{ID: uuid.New().String(), TeamID: "team-1", Name: "to-delete", Status: buildApplicationModel.ApplicationStatusActive, RepositoryURL: "https://example.com/a.git", ApplicationDescriptorPath: "application.yaml", CreatedAt: now, UpdatedAt: now}
	assert.NoError(t, repo.CreateApplication(t.Context(), app))

	assert.NoError(t, repo.SoftDeleteApplication(t.Context(), app.ID, "team-1"))

	resp, total, err := repo.ListApplicationsByTeam(t.Context(), "team-1", &buildApplicationModel.ListBuildApplicationsRequest{})
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, resp, 0)
}
