package environment_movement

import (
	"context"

	environmentMovementModel "github.com/davidsugianto/idp-core/internal/model/environment_movement"
)

func (r *repository) Create(ctx context.Context, movement *environmentMovementModel.EnvironmentMovement) error {
	return r.db.WithContext(ctx).Create(movement).Error
}

func (r *repository) GetByID(ctx context.Context, id string) (*environmentMovementModel.EnvironmentMovement, error) {
	var movement environmentMovementModel.EnvironmentMovement
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&movement).Error; err != nil {
		return nil, err
	}
	return &movement, nil
}

func (r *repository) ListByEnvironment(ctx context.Context, environmentID string) ([]environmentMovementModel.EnvironmentMovement, error) {
	var movements []environmentMovementModel.EnvironmentMovement
	if err := r.db.WithContext(ctx).Where("environment_id = ?", environmentID).Order("created_at DESC").Find(&movements).Error; err != nil {
		return nil, err
	}
	return movements, nil
}

func (r *repository) ListActiveByTarget(ctx context.Context, targetID string) ([]environmentMovementModel.EnvironmentMovement, error) {
	var movements []environmentMovementModel.EnvironmentMovement
	if err := r.db.WithContext(ctx).
		Where("destination_target_id = ?", targetID).
		Where("status IN ?", []string{environmentMovementModel.StatusPending, environmentMovementModel.StatusRunning}).
		Order("created_at DESC").
		Find(&movements).Error; err != nil {
		return nil, err
	}
	return movements, nil
}

func (r *repository) Update(ctx context.Context, movement *environmentMovementModel.EnvironmentMovement) error {
	return r.db.WithContext(ctx).Save(movement).Error
}

func (r *repository) UpdateStatus(ctx context.Context, id, status string, progressPercent int, message string) error {
	updates := map[string]any{
		"status":           status,
		"progress_percent": progressPercent,
		"message":          message,
	}

	return r.db.WithContext(ctx).
		Model(&environmentMovementModel.EnvironmentMovement{}).
		Where("id = ?", id).
		Updates(updates).Error
}
