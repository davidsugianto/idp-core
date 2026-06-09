package live_update

import (
	"context"

	liveSubscriptionModel "github.com/davidsugianto/idp-core/internal/model/live_subscription"
	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
)

func (u *usecase) Subscribe(ctx context.Context, subscription *liveSubscriptionModel.LiveSubscription) error {
	return nil
}

func (u *usecase) Unsubscribe(ctx context.Context, subscriptionID string) error {
	return nil
}

func (u *usecase) PublishNotification(ctx context.Context, notification *notificationModel.Notification) error {
	return nil
}
