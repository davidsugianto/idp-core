package delivery_target

import (
	"context"

	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
)

func (r *repository) Create(ctx context.Context, target *deliveryTargetModel.DeliveryTarget) error {
	return r.db.WithContext(ctx).Create(target).Error
}

func (r *repository) GetByID(ctx context.Context, id string) (*deliveryTargetModel.DeliveryTarget, error) {
	var target deliveryTargetModel.DeliveryTarget
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&target).Error; err != nil {
		return nil, err
	}
	return &target, nil
}

func (r *repository) GetControlPlaneByID(ctx context.Context, id string) (*deliveryTargetModel.TargetControlPlane, error) {
	target, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return target.ControlPlane(), nil
}

func (r *repository) Update(ctx context.Context, target *deliveryTargetModel.DeliveryTarget) error {
	return r.db.WithContext(ctx).Save(target).Error
}

func (r *repository) UpdateAvailability(ctx context.Context, id, availabilityState, healthState, capacitySummary string) error {
	updates := map[string]any{
		"availability_state": availabilityState,
		"health_state":       healthState,
		"capacity_summary":   capacitySummary,
	}

	return r.db.WithContext(ctx).
		Model(&deliveryTargetModel.DeliveryTarget{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&deliveryTargetModel.DeliveryTarget{}, "id = ?", id).Error
}

func (r *repository) List(ctx context.Context, req *deliveryTargetModel.ListDeliveryTargetsRequest) ([]deliveryTargetModel.DeliveryTarget, int64, error) {
	query := r.db.WithContext(ctx).Model(&deliveryTargetModel.DeliveryTarget{})

	if req != nil {
		if req.TeamID != "" {
			query = query.Where("team_id = ? OR team_id = '' OR team_id IS NULL", req.TeamID)
		}
		if req.Purpose != "" {
			query = query.Where("purpose = ?", req.Purpose)
		}
		if req.AvailabilityState != "" {
			query = query.Where("availability_state = ?", req.AvailabilityState)
		}
		if req.HealthState != "" {
			query = query.Where("health_state = ?", req.HealthState)
		}
		if req.Search != "" {
			like := "%" + req.Search + "%"
			query = query.Where("name ILIKE ? OR slug ILIKE ? OR display_name ILIKE ?", like, like, like)
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

	var targets []deliveryTargetModel.DeliveryTarget
	if err := result.Find(&targets).Error; err != nil {
		return nil, 0, err
	}
	return targets, total, nil
}

func (r *repository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&deliveryTargetModel.DeliveryTarget{}).
		Where("slug = ?", slug).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *repository) ExistsBySlugExcludingID(ctx context.Context, slug, id string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&deliveryTargetModel.DeliveryTarget{}).
		Where("slug = ?", slug).
		Where("id <> ?", id).
		Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
