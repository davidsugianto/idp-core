package environment_movement

import (
	"context"
	"errors"

	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	environmentMovementModel "github.com/davidsugianto/idp-core/internal/model/environment_movement"
	deliveryTargetRepo "github.com/davidsugianto/idp-core/internal/repository/delivery_target"
	environmentRepo "github.com/davidsugianto/idp-core/internal/repository/environment"
	environmentMovementRepo "github.com/davidsugianto/idp-core/internal/repository/environment_movement"
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
}

type Dependencies struct {
	EnvironmentRepo         environmentRepo.Repository
	DeliveryTargetRepo      deliveryTargetRepo.Repository
	EnvironmentMovementRepo environmentMovementRepo.Repository
}

func New(deps Dependencies) Usecase {
	return &usecase{
		environmentRepo:         deps.EnvironmentRepo,
		deliveryTargetRepo:      deps.DeliveryTargetRepo,
		environmentMovementRepo: deps.EnvironmentMovementRepo,
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
