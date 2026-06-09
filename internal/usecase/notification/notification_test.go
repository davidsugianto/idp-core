package notification

import (
	"context"
	"errors"
	"testing"

	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
	notificationRepo "github.com/davidsugianto/idp-core/internal/repository/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type notificationTestRepository struct {
	notificationRepo.Repository
	listReq    *notificationModel.ListNotificationsRequest
	listResult []notificationModel.Notification
	listTotal  int64
	listErr    error
}

func (r *notificationTestRepository) List(ctx context.Context, req *notificationModel.ListNotificationsRequest) ([]notificationModel.Notification, int64, error) {
	r.listReq = req
	return r.listResult, r.listTotal, r.listErr
}

func TestListUsesRepositoryFilters(t *testing.T) {
	repo := &notificationTestRepository{
		listResult: []notificationModel.Notification{{ID: "notif-1"}, {ID: "notif-2"}},
		listTotal:  2,
	}
	uc := New(Dependencies{NotificationRepo: repo})

	req := &notificationModel.ListNotificationsRequest{
		TeamID:        "team-1",
		EnvironmentID: "env-1",
		Kind:          notificationModel.KindEnvironment,
		Limit:         10,
		Offset:        20,
	}

	result, err := uc.List(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Same(t, req, repo.listReq)
	assert.EqualValues(t, 2, result.Total)
	assert.Len(t, result.Notifications, 2)
	assert.Equal(t, "notif-1", result.Notifications[0].ID)
}

func TestListReturnsRepositoryError(t *testing.T) {
	repo := &notificationTestRepository{listErr: errors.New("db error")}
	uc := New(Dependencies{NotificationRepo: repo})

	result, err := uc.List(context.Background(), &notificationModel.ListNotificationsRequest{TeamID: "team-1"})
	require.Error(t, err)
	assert.Nil(t, result)
}
