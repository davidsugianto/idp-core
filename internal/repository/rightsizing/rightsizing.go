package rightsizing

import (
	"context"
	"time"

	rightsizingModel "github.com/davidsugianto/idp-core/internal/model/rightsizing"
)

func (r *repository) Create(ctx context.Context, rec *rightsizingModel.RightsizingRecommendation) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *repository) GetByID(ctx context.Context, id string) (*rightsizingModel.RightsizingRecommendation, error) {
	var rec rightsizingModel.RightsizingRecommendation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *repository) List(ctx context.Context, req *rightsizingModel.ListRecommendationsRequest) ([]rightsizingModel.RightsizingRecommendation, int64, error) {
	query := r.db.WithContext(ctx).Model(&rightsizingModel.RightsizingRecommendation{})

	if req.Namespace != "" {
		query = query.Where("namespace = ?", req.Namespace)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.TeamID != "" {
		query = query.Where("team_id = ?", req.TeamID)
	}
	if req.WorkloadType != "" {
		query = query.Where("workload_type = ?", req.WorkloadType)
	}
	if req.RecommendationType != "" {
		query = query.Where("recommendation_type = ?", req.RecommendationType)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var recs []rightsizingModel.RightsizingRecommendation
	query = query.Order("created_at DESC")
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	err := query.Find(&recs).Error
	return recs, total, err
}

func (r *repository) Update(ctx context.Context, rec *rightsizingModel.RightsizingRecommendation) error {
	return r.db.WithContext(ctx).Save(rec).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&rightsizingModel.RightsizingRecommendation{}).Error
}

func (r *repository) DeletePendingByWorkload(ctx context.Context, namespace, workloadName, workloadType string) error {
	return r.db.WithContext(ctx).
		Where("namespace = ? AND workload_name = ? AND workload_type = ? AND status = ?",
			namespace, workloadName, workloadType, rightsizingModel.StatusPending).
		Delete(&rightsizingModel.RightsizingRecommendation{}).Error
}

func (r *repository) ListPendingByWorkload(ctx context.Context, namespace, workloadName, workloadType string) ([]rightsizingModel.RightsizingRecommendation, error) {
	var recs []rightsizingModel.RightsizingRecommendation
	err := r.db.WithContext(ctx).
		Where("namespace = ? AND workload_name = ? AND workload_type = ? AND status = ?",
			namespace, workloadName, workloadType, rightsizingModel.StatusPending).
		Find(&recs).Error
	return recs, err
}

func (r *repository) ExistsPendingForContainer(ctx context.Context, namespace, workloadName, workloadType, containerName string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&rightsizingModel.RightsizingRecommendation{}).
		Where("namespace = ? AND workload_name = ? AND workload_type = ? AND container_name = ? AND status = ?",
			namespace, workloadName, workloadType, containerName, rightsizingModel.StatusPending).
		Count(&count).Error
	return count > 0, err
}

// CleanupOldRecommendations removes recommendations older than the specified duration
func (r *repository) CleanupOldRecommendations(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", rightsizingModel.StatusDismissed, cutoff).
		Delete(&rightsizingModel.RightsizingRecommendation{})
	return result.RowsAffected, result.Error
}
