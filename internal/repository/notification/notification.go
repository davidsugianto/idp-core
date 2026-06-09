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

func (r *repository) List(ctx context.Context, req *notificationModel.ListNotificationsRequest) ([]notificationModel.Notification, int64, error) {
	query := r.db.WithContext(ctx).Model(&notificationModel.Notification{})

	if req != nil {
		if req.UserID != "" {
			query = query.Where("user_id = ?", req.UserID)
		}
		if req.TeamID != "" {
			query = query.Where("team_id = ?", req.TeamID)
		}
		if req.EnvironmentID != "" {
			query = query.Where("environment_id = ?", req.EnvironmentID)
		}
		if req.Kind != "" {
			query = query.Where("kind = ?", req.Kind)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	result := query.Order("created_at DESC")
	if req != nil {
		if req.Offset > 0 {
			result = result.Offset(req.Offset)
		}
		if req.Limit > 0 {
			result = result.Limit(req.Limit)
		}
	}

	var notifications []notificationModel.Notification
	if err := result.Find(&notifications).Error; err != nil {
		return nil, 0, err
	}

	return notifications, total, nil
}

func (r *repository) ListByEnvironment(ctx context.Context, environmentID string) ([]notificationModel.Notification, error) {
	notifications, _, err := r.List(ctx, &notificationModel.ListNotificationsRequest{EnvironmentID: environmentID})
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *repository) ListByUser(ctx context.Context, userID string) ([]notificationModel.Notification, error) {
	notifications, _, err := r.List(ctx, &notificationModel.ListNotificationsRequest{UserID: userID})
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *repository) Update(ctx context.Context, notification *notificationModel.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}
