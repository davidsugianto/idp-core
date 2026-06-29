package build_application

import (
	"context"
	"time"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (r *repository) CreateBuild(ctx context.Context, build *buildApplicationModel.Build) error {
	return r.db.WithContext(ctx).Create(build).Error
}

func (r *repository) GetBuildByID(ctx context.Context, buildID string) (*buildApplicationModel.Build, error) {
	var build buildApplicationModel.Build
	err := r.db.WithContext(ctx).Where("id = ?", buildID).First(&build).Error
	if err != nil {
		return nil, err
	}
	return &build, nil
}

func (r *repository) GetBuildByIDAndTeam(ctx context.Context, buildID, teamID string) (*buildApplicationModel.Build, error) {
	var build buildApplicationModel.Build
	err := r.db.WithContext(ctx).
		Where("id = ? AND team_id = ?", buildID, teamID).
		First(&build).Error
	if err != nil {
		return nil, err
	}
	return &build, nil
}

func (r *repository) GetBuildByApplicationAndIdempotencyKey(ctx context.Context, applicationID, idempotencyKey string) (*buildApplicationModel.Build, error) {
	var build buildApplicationModel.Build
	err := r.db.WithContext(ctx).
		Where("application_id = ? AND idempotency_key = ?", applicationID, idempotencyKey).
		First(&build).Error
	if err != nil {
		return nil, err
	}
	return &build, nil
}

func (r *repository) GetLatestBuildByApplicationID(ctx context.Context, applicationID string) (*buildApplicationModel.Build, error) {
	var build buildApplicationModel.Build
	err := r.db.WithContext(ctx).
		Where("application_id = ?", applicationID).
		Order("sequence_number DESC").
		First(&build).Error
	if err != nil {
		return nil, err
	}
	return &build, nil
}

func (r *repository) ListBuildsByApplication(ctx context.Context, applicationID, teamID string, limit, offset int) ([]buildApplicationModel.Build, int64, error) {
	query := r.db.WithContext(ctx).Model(&buildApplicationModel.Build{}).
		Where("application_id = ? AND team_id = ?", applicationID, teamID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("sequence_number DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var builds []buildApplicationModel.Build
	if err := query.Find(&builds).Error; err != nil {
		return nil, 0, err
	}
	return builds, total, nil
}

func (r *repository) ListBuildsByTeam(ctx context.Context, teamID string, limit, offset int) ([]buildApplicationModel.Build, int64, error) {
	query := r.db.WithContext(ctx).Model(&buildApplicationModel.Build{}).
		Where("team_id = ?", teamID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var builds []buildApplicationModel.Build
	if err := query.Find(&builds).Error; err != nil {
		return nil, 0, err
	}
	return builds, total, nil
}

func (r *repository) GetLatestBuildSequenceByApplicationID(ctx context.Context, applicationID string) (int64, error) {
	var latest buildApplicationModel.Build
	err := r.db.WithContext(ctx).
		Where("application_id = ?", applicationID).
		Order("sequence_number DESC").
		Select("sequence_number").
		First(&latest).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return latest.SequenceNumber, nil
}

func (r *repository) ClaimNextQueuedBuild(ctx context.Context, workerID string, leaseUntil time.Time) (*buildApplicationModel.Build, error) {
	var claimed *buildApplicationModel.Build

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var build buildApplicationModel.Build
		err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where("status = ?", buildApplicationModel.BuildStatusQueued).
			Order("queued_at ASC NULLS LAST").
			Order("created_at ASC").
			First(&build).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil
			}
			return err
		}

		now := time.Now().UTC()
		build.Status = buildApplicationModel.BuildStatusRunning
		build.StartedAt = &now
		build.ExecutionWorkerID = workerID
		build.ExecutionClaimedAt = &now
		build.ExecutionLeaseExpiresAt = &leaseUntil
		build.ExecutionAttempts++
		build.UpdatedAt = now

		if err := tx.Save(&build).Error; err != nil {
			return err
		}

		claimed = &build
		return nil
	})
	if err != nil {
		return nil, err
	}

	return claimed, nil
}

func (r *repository) UpdateBuild(ctx context.Context, build *buildApplicationModel.Build) error {
	return r.db.WithContext(ctx).Save(build).Error
}

func (r *repository) UpdateBuildStatus(ctx context.Context, buildID, status, failureReason string) error {
	updates := map[string]any{
		"status": status,
	}
	if failureReason != "" {
		updates["failure_reason"] = failureReason
	}

	return r.db.WithContext(ctx).
		Model(&buildApplicationModel.Build{}).
		Where("id = ?", buildID).
		Updates(updates).Error
}

