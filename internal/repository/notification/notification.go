package notification

import (
	"context"

	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
)

func (r *repository) Create(ctx context.Context, notification *notificationModel.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *repository) GetByID(ctx context.Context, id string) (*notificationModel.Notification, error) {
	var notification notificationModel.Notification
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&notification).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *repository) ListByEnvironment(ctx context.Context, environmentID string) ([]notificationModel.Notification, error) {
	var notifications []notificationModel.Notification
	if err := r.db.WithContext(ctx).Where("environment_id = ?", environmentID).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *repository) ListByUser(ctx context.Context, userID string) ([]notificationModel.Notification, error) {
	var notifications []notificationModel.Notification
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *repository) Update(ctx context.Context, notification *notificationModel.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}
