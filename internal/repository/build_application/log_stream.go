package build_application

import (
	"context"

	buildApplicationModel "github.com/davidsugianto/idp-core/internal/model/build_application"
)

func (r *repository) AppendBuildLog(ctx context.Context, log *buildApplicationModel.BuildLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *repository) ListBuildLogs(ctx context.Context, buildID string, afterSequence int64, limit int) ([]buildApplicationModel.BuildLog, error) {
	query := r.db.WithContext(ctx).Model(&buildApplicationModel.BuildLog{}).
		Where("build_id = ?", buildID)
	if afterSequence > 0 {
		query = query.Where("sequence > ?", afterSequence)
	}
	query = query.Order("sequence ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var logs []buildApplicationModel.BuildLog
	if err := query.Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *repository) GetLatestBuildLogSequence(ctx context.Context, buildID string) (int64, error) {
	var latestSequence int64
	if err := r.db.WithContext(ctx).Model(&buildApplicationModel.BuildLog{}).
		Where("build_id = ?", buildID).
		Select("COALESCE(MAX(sequence), 0)").
		Scan(&latestSequence).Error; err != nil {
		return 0, err
	}
	return latestSequence, nil
}

func (r *repository) GetBuildLogSummary(ctx context.Context, buildID string) (string, error) {
	var build buildApplicationModel.Build
	err := r.db.WithContext(ctx).Where("id = ?", buildID).Select("status", "failure_reason").First(&build).Error
	if err != nil {
		return "", err
	}

	if build.FailureReason != "" {
		return build.FailureReason, nil
	}
	return build.Status, nil
}
