package delivery_target

import (
	"context"
	"errors"
	"time"

	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	templateModel "github.com/davidsugianto/idp-core/internal/model/template"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (u *usecase) Create(ctx context.Context, req *deliveryTargetModel.CreateDeliveryTargetRequest) (*deliveryTargetModel.DeliveryTargetResponse, error) {
	if req == nil || req.Name == "" {
		return nil, ErrDeliveryTargetNameRequired
	}
	if req.ClusterName == "" {
		return nil, ErrClusterNameRequired
	}

	availabilityState := req.AvailabilityState
	if availabilityState == "" {
		availabilityState = deliveryTargetModel.AvailabilityAvailable
	}
	if !deliveryTargetModel.ValidAvailabilityState(availabilityState) {
		return nil, ErrInvalidAvailabilityState
	}

	healthState := req.HealthState
	if healthState == "" {
		healthState = deliveryTargetModel.HealthUnknown
	}
	if !deliveryTargetModel.ValidHealthState(healthState) {
		return nil, ErrInvalidHealthState
	}

	slug := templateModel.GenerateSlug(req.Name)
	exists, err := u.deliveryTargetRepo.ExistsBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDeliveryTargetExists
	}

	capacitySummary, err := deliveryTargetModel.EncodeCapacitySummary(req.CapacitySummary)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	target := &deliveryTargetModel.DeliveryTarget{
		ID:                uuid.New().String(),
		Name:              req.Name,
		Slug:              slug,
		DisplayName:       req.DisplayName,
		Description:       req.Description,
		Purpose:           req.Purpose,
		TeamID:            req.TeamID,
		ClusterName:       req.ClusterName,
		ClusterServer:     req.ClusterServer,
		AvailabilityState: availabilityState,
		HealthState:       healthState,
		CapacitySummary:   capacitySummary,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := u.deliveryTargetRepo.Create(ctx, target); err != nil {
		return nil, err
	}

	return deliveryTargetModel.ToDeliveryTargetResponse(target), nil
}

func (u *usecase) Get(ctx context.Context, id string) (*deliveryTargetModel.DeliveryTargetResponse, error) {
	target, err := u.deliveryTargetRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeliveryTargetNotFound
		}
		return nil, err
	}

	return deliveryTargetModel.ToDeliveryTargetResponse(target), nil
}

func (u *usecase) List(ctx context.Context, req *deliveryTargetModel.ListDeliveryTargetsRequest) (*deliveryTargetModel.DeliveryTargetListResponse, error) {
	targets, total, err := u.deliveryTargetRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	return deliveryTargetModel.ToDeliveryTargetListResponse(targets, total), nil
}

func (u *usecase) Update(ctx context.Context, id string, req *deliveryTargetModel.UpdateDeliveryTargetRequest) (*deliveryTargetModel.DeliveryTargetResponse, error) {
	target, err := u.deliveryTargetRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDeliveryTargetNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		if *req.Name == "" {
			return nil, ErrDeliveryTargetNameRequired
		}
		slug := templateModel.GenerateSlug(*req.Name)
		exists, err := u.deliveryTargetRepo.ExistsBySlugExcludingID(ctx, slug, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrDeliveryTargetExists
		}
		target.Name = *req.Name
		target.Slug = slug
	}
	if req.DisplayName != nil {
		target.DisplayName = *req.DisplayName
	}
	if req.Description != nil {
		target.Description = *req.Description
	}
	if req.Purpose != nil {
		target.Purpose = *req.Purpose
	}
	if req.TeamID != nil {
		target.TeamID = *req.TeamID
	}
	if req.ClusterName != nil {
		if *req.ClusterName == "" {
			return nil, ErrClusterNameRequired
		}
		target.ClusterName = *req.ClusterName
	}
	if req.ClusterServer != nil {
		target.ClusterServer = *req.ClusterServer
	}
	if req.AvailabilityState != nil {
		if !deliveryTargetModel.ValidAvailabilityState(*req.AvailabilityState) {
			return nil, ErrInvalidAvailabilityState
		}
		target.AvailabilityState = *req.AvailabilityState
	}
	if req.HealthState != nil {
		if !deliveryTargetModel.ValidHealthState(*req.HealthState) {
			return nil, ErrInvalidHealthState
		}
		target.HealthState = *req.HealthState
	}
	if req.CapacitySummary != nil {
		capacitySummary, err := deliveryTargetModel.EncodeCapacitySummary(*req.CapacitySummary)
		if err != nil {
			return nil, err
		}
		target.CapacitySummary = capacitySummary
	}

	target.UpdatedAt = time.Now()
	if err := u.deliveryTargetRepo.Update(ctx, target); err != nil {
		return nil, err
	}

	return deliveryTargetModel.ToDeliveryTargetResponse(target), nil
}

func (u *usecase) Delete(ctx context.Context, id string) error {
	_, err := u.deliveryTargetRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrDeliveryTargetNotFound
		}
		return err
	}

	if u.environmentRepo != nil {
		count, err := u.environmentRepo.CountByDeliveryTarget(ctx, id)
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrDeliveryTargetInUse
		}
	}

	if u.environmentMovementRepo != nil {
		movements, err := u.environmentMovementRepo.ListActiveByTarget(ctx, id)
		if err != nil {
			return err
		}
		if len(movements) > 0 {
			return ErrDeliveryTargetInUse
		}
	}

	return u.deliveryTargetRepo.Delete(ctx, id)
}
