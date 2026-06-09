package notification

import (
	"context"
	"errors"

	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
	notificationRepo "github.com/davidsugianto/idp-core/internal/repository/notification"
)

var ErrNotificationNotFound = errors.New("notification not found")

type Usecase interface {
	Create(ctx context.Context, notification *notificationModel.Notification) error
	Get(ctx context.Context, id string) (*notificationModel.Notification, error)
	List(ctx context.Context, req *notificationModel.ListNotificationsRequest) (*notificationModel.NotificationListResponse, error)
	ListByEnvironment(ctx context.Context, environmentID string) ([]notificationModel.Notification, error)
	ListByUser(ctx context.Context, userID string) ([]notificationModel.Notification, error)
	Update(ctx context.Context, notification *notificationModel.Notification) error
}

type usecase struct {
	notificationRepo notificationRepo.Repository
}

type Dependencies struct {
	NotificationRepo notificationRepo.Repository
}

func New(deps Dependencies) Usecase {
	return &usecase{notificationRepo: deps.NotificationRepo}
}
