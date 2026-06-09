package live_update

import (
	"context"
	"errors"
	"sync"

	liveSubscriptionModel "github.com/davidsugianto/idp-core/internal/model/live_subscription"
	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
	environmentRepo "github.com/davidsugianto/idp-core/internal/repository/environment"
	provisionerRepo "github.com/davidsugianto/idp-core/internal/repository/provisioner"
	notificationUsecase "github.com/davidsugianto/idp-core/internal/usecase/notification"
)

var (
	ErrSubscriptionRequired = errors.New("live subscription is required")
	ErrSubscriptionNotFound = errors.New("live subscription not found")
	ErrInvalidChannel       = errors.New("invalid live subscription channel")
	ErrStreamExpired        = errors.New("live subscription expired")
	ErrWorkloadRequired     = errors.New("workload name is required")
	ErrEnvironmentNotFound  = errors.New("environment not found")
	ErrWorkloadNotFound     = errors.New("workload not found")
	ErrLogStreamingDisabled = errors.New("log streaming is not configured")
)

type Usecase interface {
	Subscribe(ctx context.Context, subscription *liveSubscriptionModel.LiveSubscription) error
	Unsubscribe(ctx context.Context, subscriptionID string) error
	ListNotifications(ctx context.Context, req *notificationModel.ListNotificationsRequest) (*notificationModel.NotificationListResponse, error)
	StreamEvents(ctx context.Context, subscription *liveSubscriptionModel.LiveSubscription) (<-chan notificationModel.StreamEvent, error)
	StreamLogs(ctx context.Context, req *liveSubscriptionModel.StreamLogsRequest, subscription *liveSubscriptionModel.LiveSubscription) (<-chan notificationModel.StreamEvent, error)
	PublishNotification(ctx context.Context, notification *notificationModel.Notification) error
	PublishStatus(ctx context.Context, payload *notificationModel.StatusEventPayload) error
	PublishProgress(ctx context.Context, payload *notificationModel.ProgressEventPayload) error
}

type subscriptionState struct {
	subscription *liveSubscriptionModel.LiveSubscription
	events       chan notificationModel.StreamEvent
}

type usecase struct {
	notificationUC  notificationUsecase.Usecase
	environmentRepo environmentRepo.Repository
	provisionerRepo provisionerRepo.Repository
	mu              sync.RWMutex
	subscriptions   map[string]*subscriptionState
}

type Dependencies struct {
	NotificationUC  notificationUsecase.Usecase
	EnvironmentRepo environmentRepo.Repository
	ProvisionerRepo provisionerRepo.Repository
}

func New(deps Dependencies) Usecase {
	return &usecase{
		notificationUC:  deps.NotificationUC,
		environmentRepo: deps.EnvironmentRepo,
		provisionerRepo: deps.ProvisionerRepo,
		subscriptions:   make(map[string]*subscriptionState),
	}
}
