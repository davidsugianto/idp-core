package notification

import (
	"context"
	"testing"
	"time"

	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
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
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	var db *gorm.DB
	var dbErr error
	for i := 0; i < 5; i++ {
		db, dbErr = gorm.Open(postgres.Open(connStr), &gorm.Config{})
		if dbErr == nil {
			sqlDB, err := db.DB()
			if err == nil {
				err = sqlDB.Ping()
				if err == nil {
					break
				}
				_ = sqlDB.Close()
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	require.NoError(t, dbErr)

	err = db.AutoMigrate(&notificationModel.Notification{})
	require.NoError(t, err)

	cleanup := func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
		testcontainers.TerminateContainer(pgContainer)
	}

	return db, cleanup
}

func TestRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := New(Dependencies{Database: db})
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	fixtures := []notificationModel.Notification{
		{ID: "notif-1", UserID: "user-1", TeamID: "team-1", EnvironmentID: "env-1", Kind: notificationModel.KindEnvironment, Severity: notificationModel.SeverityInfo, Title: "Created", Message: "Environment created", CreatedAt: now.Add(-2 * time.Minute)},
		{ID: "notif-2", UserID: "user-1", TeamID: "team-1", EnvironmentID: "env-2", Kind: notificationModel.KindMovement, Severity: notificationModel.SeverityWarning, Title: "Moving", Message: "Environment moving", CreatedAt: now.Add(-1 * time.Minute)},
		{ID: "notif-3", UserID: "user-2", TeamID: "team-2", EnvironmentID: "env-1", Kind: notificationModel.KindTarget, Severity: notificationModel.SeverityError, Title: "Target degraded", Message: "Cluster unhealthy", CreatedAt: now},
	}

	for i := range fixtures {
		err := repo.Create(ctx, &fixtures[i])
		require.NoError(t, err)
	}

	t.Run("filters by environment and kind with total count", func(t *testing.T) {
		result, total, err := repo.List(ctx, &notificationModel.ListNotificationsRequest{
			EnvironmentID: "env-1",
			Kind:          notificationModel.KindEnvironment,
		})
		require.NoError(t, err)
		assert.EqualValues(t, 1, total)
		require.Len(t, result, 1)
		assert.Equal(t, "notif-1", result[0].ID)
	})

	t.Run("orders newest first and paginates", func(t *testing.T) {
		result, total, err := repo.List(ctx, &notificationModel.ListNotificationsRequest{
			TeamID: "team-1",
			Limit:  1,
		})
		require.NoError(t, err)
		assert.EqualValues(t, 2, total)
		require.Len(t, result, 1)
		assert.Equal(t, "notif-2", result[0].ID)
	})
}

func TestRepository_ListByEnvironmentAndUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := New(Dependencies{Database: db})
	ctx := context.Background()
	createdAt := time.Now().UTC().Truncate(time.Second)

	fixtures := []notificationModel.Notification{
		{ID: "notif-env-1", UserID: "user-1", TeamID: "team-1", EnvironmentID: "env-1", Kind: notificationModel.KindEnvironment, Severity: notificationModel.SeverityInfo, Title: "One", Message: "One", CreatedAt: createdAt.Add(-1 * time.Minute)},
		{ID: "notif-env-2", UserID: "user-1", TeamID: "team-1", EnvironmentID: "env-1", Kind: notificationModel.KindMovement, Severity: notificationModel.SeveritySuccess, Title: "Two", Message: "Two", CreatedAt: createdAt},
		{ID: "notif-env-3", UserID: "user-2", TeamID: "team-2", EnvironmentID: "env-2", Kind: notificationModel.KindTarget, Severity: notificationModel.SeverityWarning, Title: "Three", Message: "Three", CreatedAt: createdAt.Add(-2 * time.Minute)},
	}

	for i := range fixtures {
		err := repo.Create(ctx, &fixtures[i])
		require.NoError(t, err)
	}

	byEnvironment, err := repo.ListByEnvironment(ctx, "env-1")
	require.NoError(t, err)
	require.Len(t, byEnvironment, 2)
	assert.Equal(t, "notif-env-2", byEnvironment[0].ID)
	assert.Equal(t, "notif-env-1", byEnvironment[1].ID)

	byUser, err := repo.ListByUser(ctx, "user-1")
	require.NoError(t, err)
	require.Len(t, byUser, 2)
	assert.Equal(t, "notif-env-2", byUser[0].ID)
	assert.Equal(t, "notif-env-1", byUser[1].ID)
}
