package environment_movement

import (
	"context"
	"errors"
	"time"

	environmentMovementModel "github.com/davidsugianto/idp-core/internal/model/environment_movement"
	notificationModel "github.com/davidsugianto/idp-core/internal/model/notification"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (u *usecase) Create(ctx context.Context, teamID, userID, environmentID, destinationTargetID string) (*environmentMovementModel.EnvironmentMovement, error) {
	env, err := u.environmentRepo.GetByIDAndTeam(ctx, environmentID, teamID)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, ErrEnvironmentNotFound
	}
	if env.DeliveryTargetID == destinationTargetID {
		return nil, ErrMovementTargetConflict
	}

	target, err := u.deliveryTargetRepo.GetByID(ctx, destinationTargetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeliveryTargetNotFound
		}
		return nil, err
	}
	if !deliveryTargetAllowsMovement(target, teamID) {
		return nil, ErrInvalidDestinationTarget
	}

	now := time.Now()
	movement := &environmentMovementModel.EnvironmentMovement{
		ID:                  uuid.New().String(),
		EnvironmentID:       env.ID,
		SourceTargetID:      env.DeliveryTargetID,
		DestinationTargetID: destinationTargetID,
		RequestedBy:         userID,
		Status:              environmentMovementModel.StatusPending,
		ProgressPercent:     0,
		Message:             "Movement requested",
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := u.environmentMovementRepo.Create(ctx, movement); err != nil {
		return nil, err
	}

	u.emitMovementNotification(ctx, env.ID, "Movement requested", movement.Message)
	if u.liveUpdateUC != nil {
		_ = u.liveUpdateUC.PublishProgress(ctx, &notificationModel.ProgressEventPayload{EnvironmentID: env.ID, Operation: "movement", ProgressPercent: movement.ProgressPercent, Message: movement.Message, ChangedAt: movement.UpdatedAt})
	}
	return movement, nil
}

func (u *usecase) Get(ctx context.Context, teamID, environmentID, movementID string) (*environmentMovementModel.EnvironmentMovement, error) {
	env, err := u.environmentRepo.GetByIDAndTeam(ctx, environmentID, teamID)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, ErrEnvironmentNotFound
	}

	movement, err := u.environmentMovementRepo.GetByID(ctx, movementID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMovementNotFound
		}
		return nil, err
	}
	if movement.EnvironmentID != env.ID {
		return nil, ErrMovementNotFound
	}

	return movement, nil
}

func (u *usecase) ListByEnvironment(ctx context.Context, teamID, environmentID string) ([]environmentMovementModel.EnvironmentMovement, error) {
	env, err := u.environmentRepo.GetByIDAndTeam(ctx, environmentID, teamID)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, ErrEnvironmentNotFound
	}

	return u.environmentMovementRepo.ListByEnvironment(ctx, env.ID)
}

func (u *usecase) UpdateStatus(ctx context.Context, teamID, environmentID, movementID, status string, progressPercent int, message string) (*environmentMovementModel.EnvironmentMovement, error) {
	if !environmentMovementModel.ValidStatus(status) {
		return nil, ErrInvalidMovementStatus
	}

	env, err := u.environmentRepo.GetByIDAndTeam(ctx, environmentID, teamID)
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, ErrEnvironmentNotFound
	}

	movement, err := u.environmentMovementRepo.GetByID(ctx, movementID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMovementNotFound
		}
		return nil, err
	}
	if movement.EnvironmentID != env.ID {
		return nil, ErrMovementNotFound
	}

	now := time.Now()
	movement.Status = status
	movement.ProgressPercent = progressPercent
	movement.Message = message
	movement.UpdatedAt = now
	if status == environmentMovementModel.StatusRunning && movement.StartedAt == nil {
		movement.StartedAt = &now
	}
	if status == environmentMovementModel.StatusCompleted || status == environmentMovementModel.StatusFailed || status == environmentMovementModel.StatusCancelled {
		movement.CompletedAt = &now
	}

	if err := u.environmentMovementRepo.Update(ctx, movement); err != nil {
		return nil, err
	}

	if status == environmentMovementModel.StatusCompleted {
		target, err := u.deliveryTargetRepo.GetByID(ctx, movement.DestinationTargetID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrDeliveryTargetNotFound
			}
			return nil, err
		}
		if err := u.environmentRepo.UpdateDeliveryTarget(ctx, env.ID, teamID, movement.DestinationTargetID, target.ClusterName, target.ClusterServer); err != nil {
			return nil, err
		}
		env.DeliveryTargetID = movement.DestinationTargetID
		env.ClusterName = target.ClusterName
		env.ClusterServer = target.ClusterServer
	}

	if u.liveUpdateUC != nil {
		_ = u.liveUpdateUC.PublishProgress(ctx, &notificationModel.ProgressEventPayload{EnvironmentID: env.ID, Operation: "movement", ProgressPercent: movement.ProgressPercent, Message: movement.Message, ChangedAt: movement.UpdatedAt})
	}
	u.emitMovementNotification(ctx, env.ID, "Movement status updated", movement.Message)
	return movement, nil
}
