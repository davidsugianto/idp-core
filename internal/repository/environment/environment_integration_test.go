package environment

import (
	"context"
	"testing"
	"time"

	"github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	ctx := context.Background()

	pgContainer, err := postgrescontainer.Run(ctx, "postgres:15-alpine",
		postgrescontainer.WithDatabase("testdb"),
		postgrescontainer.WithUsername("test"),
		postgrescontainer.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "Failed to start postgres container")

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err, "Failed to get connection string")

	var db *gorm.DB
	var dbErr error

	// Retry connection a few times
	for i := 0; i < 5; i++ {
		db, dbErr = gorm.Open(postgres.Open(connStr), &gorm.Config{})
		if dbErr == nil {
			sqlDB, err := db.DB()
			if err == nil {
				err = sqlDB.Ping()
				if err == nil {
					break
				}
				sqlDB.Close()
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	require.NoError(t, dbErr, "Failed to connect to database after retries")

	// Run migrations
	err = db.AutoMigrate(&environment.Environment{})
	require.NoError(t, err, "Failed to run migrations")

	cleanup := func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
		testcontainers.TerminateContainer(pgContainer)
	}

	return db, cleanup
}

func TestRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := New(Dependencies{Database: db})
	ctx := context.Background()

	tests := []struct {
		name    string
		env     *environment.Environment
		wantErr bool
	}{
		{
			name: "create environment successfully",
			env: &environment.Environment{
				ID:           "test-env-1",
				TeamID:       "team-123",
				Name:         "dev",
				Namespace:    "idp-team-123-dev",
				Status:       "creating",
				GitRepoURL:   "https://github.com/org/repo.git",
				GitRevision:  "main",
				ManifestPath: "manifests",
				ArgoAppName:  "env-test-env-1",
			},
			wantErr: false,
		},
		{
			name: "create environment with resource quota",
			env: &environment.Environment{
				ID:                 "test-env-2",
				TeamID:             "team-123",
				Name:               "prod",
				Namespace:          "idp-team-123-prod",
				Status:             "ready",
				GitRepoURL:         "https://github.com/org/repo.git",
				GitRevision:        "main",
				ManifestPath:       "manifests",
				ResourceQuotaCPU:   "4",
				ResourceQuotaMemory: "8Gi",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.env)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the environment was created
				var found environment.Environment
				result := db.First(&found, "id = ?", tt.env.ID)
				assert.NoError(t, result.Error)
				assert.Equal(t, tt.env.ID, found.ID)
				assert.Equal(t, tt.env.TeamID, found.TeamID)
				assert.Equal(t, tt.env.Name, found.Name)
			}
		})
	}
}

func TestRepository_GetByIDAndTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := New(Dependencies{Database: db})
	ctx := context.Background()

	// Create test data
	env := &environment.Environment{
		ID:           "test-get-1",
		TeamID:       "team-get",
		Name:         "dev",
		Namespace:    "idp-team-get-dev",
		Status:       "ready",
		GitRepoURL:   "https://github.com/org/repo.git",
		GitRevision:  "main",
		ManifestPath: "manifests",
	}
	err := repo.Create(ctx, env)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      string
		teamID  string
		wantNil bool
		wantErr bool
	}{
		{
			name:    "get existing environment",
			id:      "test-get-1",
			teamID:  "team-get",
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "get non-existing environment",
			id:      "nonexistent",
			teamID:  "team-get",
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "get with wrong team id",
			id:      "test-get-1",
			teamID:  "wrong-team",
			wantNil: true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByIDAndTeam(ctx, tt.id, tt.teamID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantNil {
					assert.Nil(t, result)
				} else {
					assert.NotNil(t, result)
					assert.Equal(t, tt.id, result.ID)
					assert.Equal(t, tt.teamID, result.TeamID)
				}
			}
		})
	}
}

func TestRepository_ListByTeam(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := New(Dependencies{Database: db})
	ctx := context.Background()

	// Create test data
	teamID := "team-list"
	envs := []*environment.Environment{
		{
			ID:           "list-1",
			TeamID:       teamID,
			Name:         "dev",
			Namespace:    "idp-team-list-dev",
			Status:       "ready",
			GitRepoURL:   "https://github.com/org/repo.git",
			GitRevision:  "main",
			ManifestPath: "manifests",
		},
		{
			ID:           "list-2",
			TeamID:       teamID,
			Name:         "staging",
			Namespace:    "idp-team-list-staging",
			Status:       "ready",
			GitRepoURL:   "https://github.com/org/repo.git",
			GitRevision:  "main",
			ManifestPath: "manifests",
		},
		{
			ID:           "list-3",
			TeamID:       "other-team",
			Name:         "prod",
			Namespace:    "idp-other-team-prod",
			Status:       "ready",
			GitRepoURL:   "https://github.com/org/repo.git",
			GitRevision:  "main",
			ManifestPath: "manifests",
		},
	}

	for _, env := range envs {
		err := repo.Create(ctx, env)
		require.NoError(t, err)
	}

	tests := []struct {
		name    string
		teamID  string
		wantLen int
		wantErr bool
	}{
		{
			name:    "list environments for team",
			teamID:  teamID,
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "list environments for empty team",
			teamID:  "empty-team",
			wantLen: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.ListByTeam(ctx, tt.teamID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.wantLen)
			}
		})
	}
}

func TestRepository_UpdateStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := New(Dependencies{Database: db})
	ctx := context.Background()

	// Create test data
	env := &environment.Environment{
		ID:           "update-status-1",
		TeamID:       "team-update",
		Name:         "dev",
		Namespace:    "idp-team-update-dev",
		Status:       "creating",
		GitRepoURL:   "https://github.com/org/repo.git",
		GitRevision:  "main",
		ManifestPath: "manifests",
	}
	err := repo.Create(ctx, env)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        string
		teamID    string
		status    string
		lastError string
		wantErr   bool
	}{
		{
			name:      "update status successfully",
			id:        "update-status-1",
			teamID:    "team-update",
			status:    "ready",
			lastError: "",
			wantErr:   false,
		},
		{
			name:      "update status with error",
			id:        "update-status-1",
			teamID:    "team-update",
			status:    "failed",
			lastError: "something went wrong",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.UpdateStatus(ctx, tt.id, tt.teamID, tt.status, tt.lastError)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the status was updated
				var found environment.Environment
				result := db.First(&found, "id = ? AND team_id = ?", tt.id, tt.teamID)
				assert.NoError(t, result.Error)
				assert.Equal(t, tt.status, found.Status)
				assert.Equal(t, tt.lastError, found.LastError)
			}
		})
	}
}

func TestRepository_SoftDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := New(Dependencies{Database: db})
	ctx := context.Background()

	// Create test data
	env := &environment.Environment{
		ID:           "soft-delete-1",
		TeamID:       "team-delete",
		Name:         "dev",
		Namespace:    "idp-team-delete-dev",
		Status:       "ready",
		GitRepoURL:   "https://github.com/org/repo.git",
		GitRevision:  "main",
		ManifestPath: "manifests",
	}
	err := repo.Create(ctx, env)
	require.NoError(t, err)

	// Soft delete
	err = repo.SoftDelete(ctx, "soft-delete-1", "team-delete")
	assert.NoError(t, err)

	// Verify it's soft deleted (not found in normal query)
	result, err := repo.GetByIDAndTeam(ctx, "soft-delete-1", "team-delete")
	assert.NoError(t, err)
	assert.Nil(t, result)

	// Verify it still exists in database (with deleted_at set)
	var deletedEnv environment.Environment
	db.Unscoped().First(&deletedEnv, "id = ?", "soft-delete-1")
	assert.NotEmpty(t, deletedEnv.ID)
	assert.NotNil(t, deletedEnv.DeletedAt)
}

func TestRepository_UniqueNamespace(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := New(Dependencies{Database: db})
	ctx := context.Background()

	// Create first environment
	env1 := &environment.Environment{
		ID:           "unique-ns-1",
		TeamID:       "team-unique",
		Name:         "dev",
		Namespace:    "idp-unique-namespace",
		Status:       "ready",
		GitRepoURL:   "https://github.com/org/repo.git",
		GitRevision:  "main",
		ManifestPath: "manifests",
	}
	err := repo.Create(ctx, env1)
	require.NoError(t, err)

	// Try to create another environment with the same namespace
	env2 := &environment.Environment{
		ID:           "unique-ns-2",
		TeamID:       "team-unique-2",
		Name:         "staging",
		Namespace:    "idp-unique-namespace", // Same namespace
		Status:       "ready",
		GitRepoURL:   "https://github.com/org/repo.git",
		GitRevision:  "main",
		ManifestPath: "manifests",
	}
	err = repo.Create(ctx, env2)
	assert.Error(t, err) // Should fail due to unique constraint
}
