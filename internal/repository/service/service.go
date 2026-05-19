package service

import (
	"context"

	serviceModel "github.com/davidsugianto/idp-core/internal/model/service"
)

// Create creates a new service
func (r *repository) Create(ctx context.Context, svc *serviceModel.Service) error {
	return r.db.WithContext(ctx).Create(svc).Error
}

// GetByID returns a service by ID
func (r *repository) GetByID(ctx context.Context, id string) (*serviceModel.Service, error) {
	var svc serviceModel.Service
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&svc).Error
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

// GetByIDIncludingDeleted returns a service by ID including soft-deleted records
func (r *repository) GetByIDIncludingDeleted(ctx context.Context, id string) (*serviceModel.Service, error) {
	var svc serviceModel.Service
	err := r.db.WithContext(ctx).Unscoped().Where("id = ?", id).First(&svc).Error
	if err != nil {
		return nil, err
	}
	return &svc, nil
}

// List returns a paginated list of services with optional filters
func (r *repository) List(ctx context.Context, req *serviceModel.ListServicesRequest) ([]serviceModel.Service, int64, error) {
	query := r.db.WithContext(ctx).Model(&serviceModel.Service{})

	if req.TeamID != "" {
		query = query.Where("team_id = ?", req.TeamID)
	}
	if req.Visibility != "" {
		query = query.Where("visibility = ?", req.Visibility)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var services []serviceModel.Service
	query = query.Order("created_at DESC")
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	err := query.Find(&services).Error
	return services, total, err
}

// Update updates a service
func (r *repository) Update(ctx context.Context, svc *serviceModel.Service) error {
	return r.db.WithContext(ctx).Save(svc).Error
}

// Delete soft-deletes a service
func (r *repository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&serviceModel.Service{}).Error
}

// SearchServices searches services by name or description
func (r *repository) SearchServices(ctx context.Context, query string) ([]serviceModel.Service, error) {
	var services []serviceModel.Service
	searchPattern := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern).
		Where("visibility != ?", serviceModel.VisibilityPrivate).
		Order("created_at DESC").
		Find(&services).Error
	return services, err
}

// ExistsByName checks if a service with the given name exists in a team
func (r *repository) ExistsByName(ctx context.Context, name string, teamID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&serviceModel.Service{}).
		Where("name = ? AND team_id = ?", name, teamID).
		Count(&count).Error
	return count > 0, err
}
