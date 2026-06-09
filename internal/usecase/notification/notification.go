package notification

import (
	"context"
	"errors"

	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
	"gorm.io/gorm"
)

func (u *usecase) Create(ctx context.Context, notification *notificationModel.Notification) error {
	return u.notificationRepo.Create(ctx, notification)
}

func (u *usecase) Get(ctx context.Context, id string) (*notificationModel.Notification, error) {
	notification, err := u.notificationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotificationNotFound
		}
		return nil, err
	}
	return notification, nil
}

func (u *usecase) ListByEnvironment(ctx context.Context, environmentID string) ([]notificationModel.Notification, error) {
	return u.notificationRepo.ListByEnvironment(ctx, environmentID)
}

func (u *usecase) ListByUser(ctx context.Context, userID string) ([]notificationModel.Notification, error) {
	return u.notificationRepo.ListByUser(ctx, userID)
}

func (u *usecase) Update(ctx context.Context, notification *notificationModel.Notification) error {
	return u.notificationRepo.Update(ctx, notification)
}
