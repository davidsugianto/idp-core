package live_update

import (
	"context"
	"testing"
	"time"

	liveSubscriptionModel "github.com/davidsugianto/idp-core/internal/model/live_subscription"
	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
	notificationUsecase "github.com/davidsugianto/idp-core/internal/usecase/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type liveUpdateNotificationStub struct {
	notificationUsecase.Usecase
	listReq    *notificationModel.ListNotificationsRequest
	listResult *notificationModel.NotificationListResponse
	listErr    error
}

func (s *liveUpdateNotificationStub) List(ctx context.Context, req *notificationModel.ListNotificationsRequest) (*notificationModel.NotificationListResponse, error) {
	s.listReq = req
	return s.listResult, s.listErr
}

func TestListNotificationsDelegatesToNotificationUsecase(t *testing.T) {
	stub := &liveUpdateNotificationStub{
		listResult: &notificationModel.NotificationListResponse{Total: 1, Notifications: []notificationModel.Notification{{ID: "notif-1"}}},
	}
	uc := New(Dependencies{NotificationUC: stub})

	req := &notificationModel.ListNotificationsRequest{EnvironmentID: "env-1", Limit: 5}
	result, err := uc.ListNotifications(context.Background(), req)
	require.NoError(t, err)
	assert.Same(t, req, stub.listReq)
	require.NotNil(t, result)
	assert.EqualValues(t, 1, result.Total)
	assert.Len(t, result.Notifications, 1)
}

func TestStreamEventsPublishesMatchingNotifications(t *testing.T) {
	uc := New(Dependencies{})
	subscription := &liveSubscriptionModel.LiveSubscription{
		ID:            "sub-1",
		EnvironmentID: "env-1",
		Channels:      []string{liveSubscriptionModel.ChannelNotification},
	}

	stream, err := uc.StreamEvents(context.Background(), subscription)
	require.NoError(t, err)

	err = uc.PublishNotification(context.Background(), &notificationModel.Notification{
		ID:            "notif-1",
		EnvironmentID: "env-1",
		Kind:          notificationModel.KindEnvironment,
		Severity:      notificationModel.SeverityInfo,
		Title:         "Created",
		Message:       "Environment created",
		CreatedAt:     time.Now(),
	})
	require.NoError(t, err)

	select {
	case event := <-stream:
		assert.Equal(t, notificationModel.EventNotification, event.Event)
		payload, ok := event.Data.(*notificationModel.NotificationEventPayload)
		require.True(t, ok)
		assert.Equal(t, "notif-1", payload.NotificationID)
	case <-time.After(2 * time.Second):
		t.Fatal("expected notification event")
	}
}

func TestStreamLogsRequiresWorkloadName(t *testing.T) {
	uc := New(Dependencies{})
	_, err := uc.StreamLogs(context.Background(), &liveSubscriptionModel.StreamLogsRequest{EnvironmentID: "env-1"}, &liveSubscriptionModel.LiveSubscription{ID: "sub-log"})
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrWorkloadRequired)
}

func TestUnsubscribeRemovesSubscription(t *testing.T) {
	uc := New(Dependencies{})
	subscription := &liveSubscriptionModel.LiveSubscription{ID: "sub-1", Channel: liveSubscriptionModel.ChannelStatus}
	_, err := uc.StreamEvents(context.Background(), subscription)
	require.NoError(t, err)

	err = uc.Unsubscribe(context.Background(), "sub-1")
	require.NoError(t, err)

	err = uc.Unsubscribe(context.Background(), "sub-1")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSubscriptionNotFound)
}
