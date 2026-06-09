package environment_movement

import (
	"context"
	"errors"
	"time"

	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	environmentMovementModel "github.com/davidsugianto/idp-core/internal/model/environment_movement"
	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
	deliveryTargetRepo "github.com/davidsugianto/idp-core/internal/repository/delivery_target"
	environmentRepo "github.com/davidsugianto/idp-core/internal/repository/environment"
	environmentMovementRepo "github.com/davidsugianto/idp-core/internal/repository/environment_movement"
	liveUpdateUsecase "github.com/davidsugianto/idp-core/internal/usecase/live_update"
	notificationUsecase "github.com/davidsugianto/idp-core/internal/usecase/notification"
)

var (
	ErrEnvironmentNotFound      = errors.New("environment not found")
	ErrMovementNotFound         = errors.New("environment movement not found")
	ErrDeliveryTargetNotFound   = errors.New("delivery target not found")
	ErrInvalidDestinationTarget = errors.New("delivery target is not available for movement")
	ErrMovementTargetConflict   = errors.New("movement destination must differ from current target")
	ErrInvalidMovementStatus    = errors.New("invalid movement status")
)

type Usecase interface {
	Create(ctx context.Context, teamID, userID, environmentID, destinationTargetID string) (*environmentMovementModel.EnvironmentMovement, error)
	Get(ctx context.Context, teamID, environmentID, movementID string) (*environmentMovementModel.EnvironmentMovement, error)
	ListByEnvironment(ctx context.Context, teamID, environmentID string) ([]environmentMovementModel.EnvironmentMovement, error)
	UpdateStatus(ctx context.Context, teamID, environmentID, movementID, status string, progressPercent int, message string) (*environmentMovementModel.EnvironmentMovement, error)
}

type usecase struct {
	environmentRepo         environmentRepo.Repository
	deliveryTargetRepo      deliveryTargetRepo.Repository
	environmentMovementRepo environmentMovementRepo.Repository
	notificationUC          notificationUsecase.Usecase
	liveUpdateUC            liveUpdateUsecase.Usecase
}

type Dependencies struct {
	EnvironmentRepo         environmentRepo.Repository
	DeliveryTargetRepo      deliveryTargetRepo.Repository
	EnvironmentMovementRepo environmentMovementRepo.Repository
	NotificationUC          notificationUsecase.Usecase
	LiveUpdateUC            liveUpdateUsecase.Usecase
}

func New(deps Dependencies) Usecase {
	return &usecase{
		environmentRepo:         deps.EnvironmentRepo,
		deliveryTargetRepo:      deps.DeliveryTargetRepo,
		environmentMovementRepo: deps.EnvironmentMovementRepo,
		notificationUC:          deps.NotificationUC,
		liveUpdateUC:            deps.LiveUpdateUC,
	}
}

func deliveryTargetAllowsMovement(target *deliveryTargetModel.DeliveryTarget, teamID string) bool {
	if target == nil {
		return false
	}
	if target.AvailabilityState != deliveryTargetModel.AvailabilityAvailable {
		return false
	}
	if target.TeamID == "" {
		return true
	}
	return target.TeamID == teamID
}

func (u *usecase) emitMovementNotification(ctx context.Context, environmentID, title, message string) {
	if u.notificationUC == nil {
		return
	}
	notification := &notificationModel.Notification{
		EnvironmentID: environmentID,
		Kind:          notificationModel.KindMovement,
		Severity:      notificationModel.SeverityInfo,
		Title:         title,
		Message:       message,
		CreatedAt:     time.Now(),
	}
	_ = u.notificationUC.Create(ctx, notification)
	if u.liveUpdateUC != nil {
		_ = u.liveUpdateUC.PublishNotification(ctx, notification)
	}
}
