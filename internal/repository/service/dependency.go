package service

import (
	"context"

	depModel "github.com/davidsugianto/idp-core/internal/model/service_dependency"
)

// CreateDependency creates a new service dependency
func (r *repository) CreateDependency(ctx context.Context, dep *depModel.ServiceDependency) error {
	return r.db.WithContext(ctx).Create(dep).Error
}

// GetDependencyByID retrieves a dependency by ID
func (r *repository) GetDependencyByID(ctx context.Context, id string) (*depModel.ServiceDependency, error) {
	var dep depModel.ServiceDependency
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&dep).Error
	if err != nil {
		return nil, err
	}
	return &dep, nil
}

// ListDependenciesByService lists all dependencies for a service
func (r *repository) ListDependenciesByService(ctx context.Context, serviceID string, req *depModel.ListDependenciesRequest) ([]depModel.ServiceDependency, int64, error) {
	var deps []depModel.ServiceDependency
	var total int64

	query := r.db.WithContext(ctx).Model(&depModel.ServiceDependency{}).Where("service_id = ?", serviceID)

	if req.DependencyType != "" {
		query = query.Where("dependency_type = ?", req.DependencyType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(req.Offset).Find(&deps).Error; err != nil {
		return nil, 0, err
	}

	return deps, total, nil
}

// ListDependentsByService lists all services that depend on a given service
func (r *repository) ListDependentsByService(ctx context.Context, serviceID string, req *depModel.ListDependenciesRequest) ([]depModel.ServiceDependency, int64, error) {
	var deps []depModel.ServiceDependency
	var total int64

	query := r.db.WithContext(ctx).Model(&depModel.ServiceDependency{}).Where("depends_on_service_id = ?", serviceID)

	if req.DependencyType != "" {
		query = query.Where("dependency_type = ?", req.DependencyType)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	if err := query.Order("created_at DESC").Limit(limit).Offset(req.Offset).Find(&deps).Error; err != nil {
		return nil, 0, err
	}

	return deps, total, nil
}

// UpdateDependency updates a dependency
func (r *repository) UpdateDependency(ctx context.Context, dep *depModel.ServiceDependency) error {
	return r.db.WithContext(ctx).Save(dep).Error
}

// DeleteDependency soft deletes a dependency
func (r *repository) DeleteDependency(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&depModel.ServiceDependency{}).Error
}

// ExistsDependency checks if a dependency exists between two services
func (r *repository) ExistsDependency(ctx context.Context, serviceID, dependsOnServiceID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&depModel.ServiceDependency{}).
		Where("service_id = ? AND depends_on_service_id = ?", serviceID, dependsOnServiceID).
		Count(&count).Error
	return count > 0, err
}

// ListAllDependencies retrieves all dependencies for graph traversal
func (r *repository) ListAllDependencies(ctx context.Context) ([]depModel.ServiceDependency, error) {
	var deps []depModel.ServiceDependency
	err := r.db.WithContext(ctx).Find(&deps).Error
	return deps, err
}
