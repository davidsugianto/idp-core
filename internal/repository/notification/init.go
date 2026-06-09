package notification

import (
	"context"

	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, notification *notificationModel.Notification) error
	GetByID(ctx context.Context, id string) (*notificationModel.Notification, error)
	List(ctx context.Context, req *notificationModel.ListNotificationsRequest) ([]notificationModel.Notification, int64, error)
	ListByEnvironment(ctx context.Context, environmentID string) ([]notificationModel.Notification, error)
	ListByUser(ctx context.Context, userID string) ([]notificationModel.Notification, error)
	Update(ctx context.Context, notification *notificationModel.Notification) error
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
