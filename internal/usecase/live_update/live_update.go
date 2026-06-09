package live_update

import (
	"context"
	"time"

	liveSubscriptionModel "github.com/davidsugianto/idp-core/internal/model/live_subscription"
	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
)

func (u *usecase) Subscribe(ctx context.Context, subscription *liveSubscriptionModel.LiveSubscription) error {
	if subscription == nil {
		return ErrSubscriptionRequired
	}
	if subscription.ID == "" {
		return ErrSubscriptionRequired
	}
	if subscription.Channel != "" && !liveSubscriptionModel.ValidChannel(subscription.Channel) {
		return ErrInvalidChannel
	}
	if len(subscription.Channels) > 0 {
		subscription.Channels = liveSubscriptionModel.FilterValidChannels(subscription.Channels)
		if len(subscription.Channels) == 0 {
			return ErrInvalidChannel
		}
	}
	if subscription.ExpiresAt != nil && subscription.ExpiresAt.Before(time.Now()) {
		return ErrStreamExpired
	}

	now := time.Now()
	subscription.Status = liveSubscriptionModel.SessionStatusActive
	subscription.SubscribedAt = &now
	subscription.LastHeartbeatAt = &now

	u.mu.Lock()
	defer u.mu.Unlock()

	state, exists := u.subscriptions[subscription.ID]
	if exists {
		state.subscription = subscription
		return nil
	}

	u.subscriptions[subscription.ID] = &subscriptionState{
		subscription: subscription,
		events:       make(chan notificationModel.StreamEvent, 32),
	}
	return nil
}

func (u *usecase) Unsubscribe(ctx context.Context, subscriptionID string) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	state, exists := u.subscriptions[subscriptionID]
	if !exists {
		return ErrSubscriptionNotFound
	}

	now := time.Now()
	state.subscription.Status = liveSubscriptionModel.SessionStatusClosed
	state.subscription.ClosedAt = &now
	close(state.events)
	delete(u.subscriptions, subscriptionID)
	return nil
}

func (u *usecase) ListNotifications(ctx context.Context, req *notificationModel.ListNotificationsRequest) (*notificationModel.NotificationListResponse, error) {
	if u.notificationUC == nil {
		return notificationModel.ToNotificationListResponse(nil, 0), nil
	}
	return u.notificationUC.List(ctx, req)
}

func (u *usecase) StreamEvents(ctx context.Context, subscription *liveSubscriptionModel.LiveSubscription) (<-chan notificationModel.StreamEvent, error) {
	if err := u.Subscribe(ctx, subscription); err != nil {
		return nil, err
	}

	u.mu.RLock()
	defer u.mu.RUnlock()
	state := u.subscriptions[subscription.ID]
	return state.events, nil
}

func (u *usecase) PublishNotification(ctx context.Context, notification *notificationModel.Notification) error {
	if notification == nil {
		return nil
	}
	return u.publishEvent(notificationModel.EventNotification, notificationModel.ToNotificationEventPayload(notification), func(subscription *liveSubscriptionModel.LiveSubscription) bool {
		return matchesEnvironmentSubscription(subscription, notification.EnvironmentID) && matchesChannelSubscription(subscription, liveSubscriptionModel.ChannelNotification)
	})
}

func (u *usecase) PublishStatus(ctx context.Context, payload *notificationModel.StatusEventPayload) error {
	if payload == nil {
		return nil
	}
	return u.publishEvent(notificationModel.EventStatus, payload, func(subscription *liveSubscriptionModel.LiveSubscription) bool {
		return matchesEnvironmentSubscription(subscription, payload.EnvironmentID) && matchesChannelSubscription(subscription, liveSubscriptionModel.ChannelStatus)
	})
}

func (u *usecase) PublishProgress(ctx context.Context, payload *notificationModel.ProgressEventPayload) error {
	if payload == nil {
		return nil
	}
	return u.publishEvent(notificationModel.EventProgress, payload, func(subscription *liveSubscriptionModel.LiveSubscription) bool {
		return matchesEnvironmentSubscription(subscription, payload.EnvironmentID) && matchesChannelSubscription(subscription, liveSubscriptionModel.ChannelProgress)
	})
}

func (u *usecase) publishEvent(eventType string, data any, matches func(subscription *liveSubscriptionModel.LiveSubscription) bool) error {
	event := notificationModel.StreamEvent{
		Event:     eventType,
		Data:      data,
		Timestamp: time.Now(),
	}

	u.mu.RLock()
	states := make([]*subscriptionState, 0, len(u.subscriptions))
	for _, state := range u.subscriptions {
		if matches == nil || matches(state.subscription) {
			states = append(states, state)
		}
	}
	u.mu.RUnlock()

	for _, state := range states {
		state.events <- event
	}
	return nil
}

func matchesEnvironmentSubscription(subscription *liveSubscriptionModel.LiveSubscription, environmentID string) bool {
	if subscription == nil {
		return false
	}
	return subscription.EnvironmentID == "" || subscription.EnvironmentID == environmentID
}

func matchesChannelSubscription(subscription *liveSubscriptionModel.LiveSubscription, channel string) bool {
	if subscription == nil {
		return false
	}
	if subscription.Channel != "" {
		return subscription.Channel == channel
	}
	for _, value := range subscription.Channels {
		if value == channel {
			return true
		}
	}
	return false
}
