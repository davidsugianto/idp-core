package quota

import (
	"context"

	quotaModel "github.com/davidsugianto/idp-core/internal/model/resourcequota"
	"gorm.io/gorm"
)

type Repository interface {
	// Quota CRUD
	Create(ctx context.Context, quota *quotaModel.ResourceQuota) error
	GetByID(ctx context.Context, id string) (*quotaModel.ResourceQuota, error)
	GetByNamespace(ctx context.Context, namespace string) (*quotaModel.ResourceQuota, error)
	List(ctx context.Context, req *quotaModel.ListResourceQuotasRequest) ([]quotaModel.ResourceQuota, int64, error)
	Update(ctx context.Context, quota *quotaModel.ResourceQuota) error
	Delete(ctx context.Context, id string) error

	// Usage tracking
	UpdateUsage(ctx context.Context, namespace string, usage *quotaModel.UsageResponse) error

	// Quota enforcement
	GetActiveByNamespace(ctx context.Context, namespace string) (*quotaModel.ResourceQuota, error)
	ListActiveByTeam(ctx context.Context, teamID string) ([]quotaModel.ResourceQuota, error)

	// Utility
	ExistsForNamespace(ctx context.Context, namespace string) (bool, error)
}

type repository struct {
	db *gorm.DB
}

type Dependencies struct {
	Database *gorm.DB
}

func New(deps Dependencies) Repository {
	return &repository{db: deps.Database}
}

func (r *repository) Create(ctx context.Context, quota *quotaModel.ResourceQuota) error {
	return r.db.WithContext(ctx).Create(quota).Error
}

func (r *repository) GetByID(ctx context.Context, id string) (*quotaModel.ResourceQuota, error) {
	var quota quotaModel.ResourceQuota
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&quota).Error
	if err != nil {
		return nil, err
	}
	return &quota, nil
}

func (r *repository) GetByNamespace(ctx context.Context, namespace string) (*quotaModel.ResourceQuota, error) {
	var quota quotaModel.ResourceQuota
	err := r.db.WithContext(ctx).Where("namespace = ?", namespace).First(&quota).Error
	if err != nil {
		return nil, err
	}
	return &quota, nil
}

func (r *repository) List(ctx context.Context, req *quotaModel.ListResourceQuotasRequest) ([]quotaModel.ResourceQuota, int64, error) {
	query := r.db.WithContext(ctx).Model(&quotaModel.ResourceQuota{})

	if req.TeamID != "" {
		query = query.Where("team_id = ?", req.TeamID)
	}
	if req.EnvironmentID != "" {
		query = query.Where("environment_id = ?", req.EnvironmentID)
	}
	if req.Namespace != "" {
		query = query.Where("namespace = ?", req.Namespace)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var quotas []quotaModel.ResourceQuota
	query = query.Order("created_at DESC")
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	err := query.Find(&quotas).Error
	return quotas, total, err
}

func (r *repository) Update(ctx context.Context, quota *quotaModel.ResourceQuota) error {
	return r.db.WithContext(ctx).Save(quota).Error
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&quotaModel.ResourceQuota{}).Error
}

func (r *repository) UpdateUsage(ctx context.Context, namespace string, usage *quotaModel.UsageResponse) error {
	podCount := usage.PodCount
	return r.db.WithContext(ctx).Model(&quotaModel.ResourceQuota{}).
		Where("namespace = ?", namespace).
		Updates(map[string]interface{}{
			"current_cpu_request":     usage.CPURequest,
			"current_cpu_limit":       usage.CPULimit,
			"current_memory_request":  usage.MemoryRequest,
			"current_memory_limit":    usage.MemoryLimit,
			"current_storage_request": usage.StorageRequest,
			"current_pod_count":       podCount,
			"updated_at":              usage.LastUpdated,
		}).Error
}

func (r *repository) GetActiveByNamespace(ctx context.Context, namespace string) (*quotaModel.ResourceQuota, error) {
	var quota quotaModel.ResourceQuota
	err := r.db.WithContext(ctx).
		Where("namespace = ? AND status = ? AND enforce = ?", namespace, quotaModel.StatusActive, true).
		First(&quota).Error
	if err != nil {
		return nil, err
	}
	return &quota, nil
}

func (r *repository) ListActiveByTeam(ctx context.Context, teamID string) ([]quotaModel.ResourceQuota, error) {
	var quotas []quotaModel.ResourceQuota
	err := r.db.WithContext(ctx).
		Where("team_id = ? AND status = ? AND enforce = ?", teamID, quotaModel.StatusActive, true).
		Find(&quotas).Error
	return quotas, err
}

func (r *repository) ExistsForNamespace(ctx context.Context, namespace string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&quotaModel.ResourceQuota{}).
		Where("namespace = ?", namespace).
		Count(&count).Error
	return count > 0, err
}
