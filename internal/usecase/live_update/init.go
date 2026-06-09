package live_update

import (
	"context"

	liveSubscriptionModel "github.com/davidsugianto/idp-core/internal/model/live_subscription"
	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
)

type Usecase interface {
	Subscribe(ctx context.Context, subscription *liveSubscriptionModel.LiveSubscription) error
	Unsubscribe(ctx context.Context, subscriptionID string) error
	PublishNotification(ctx context.Context, notification *notificationModel.Notification) error
}

type usecase struct{}

type Dependencies struct{}

func New(deps Dependencies) Usecase {
	return &usecase{}
}
