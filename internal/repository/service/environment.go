package service

import (
	"context"

	svcEnvModel "github.com/davidsugianto/idp-core/internal/model/service_environment"
)

// CreateServiceEnvironment creates a new service environment deployment record
func (r *repository) CreateServiceEnvironment(ctx context.Context, se *svcEnvModel.ServiceEnvironment) error {
	return r.db.WithContext(ctx).Create(se).Error
}

// GetServiceEnvironmentByID retrieves a service environment by ID
func (r *repository) GetServiceEnvironmentByID(ctx context.Context, id string) (*svcEnvModel.ServiceEnvironment, error) {
	var se svcEnvModel.ServiceEnvironment
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&se).Error
	if err != nil {
		return nil, err
	}
	return &se, nil
}

// ListServiceEnvironmentsByVersion lists all deployments for a service version
func (r *repository) ListServiceEnvironmentsByVersion(ctx context.Context, versionID string, req *svcEnvModel.ListDeploymentsRequest) ([]svcEnvModel.ServiceEnvironment, int64, error) {
	var deployments []svcEnvModel.ServiceEnvironment
	var total int64

	query := r.db.WithContext(ctx).Model(&svcEnvModel.ServiceEnvironment{}).Where("service_version_id = ?", versionID)

	if req.EnvironmentID != "" {
		query = query.Where("environment_id = ?", req.EnvironmentID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	if err := query.Order("deployed_at DESC").Limit(limit).Offset(req.Offset).Find(&deployments).Error; err != nil {
		return nil, 0, err
	}

	return deployments, total, nil
}

// ListServiceEnvironmentsByService lists all deployments for a service (across all versions)
func (r *repository) ListServiceEnvironmentsByService(ctx context.Context, serviceID string, req *svcEnvModel.ListDeploymentsRequest) ([]svcEnvModel.ServiceEnvironment, int64, error) {
	var deployments []svcEnvModel.ServiceEnvironment
	var total int64

	// Join with service_versions to filter by service_id
	query := r.db.WithContext(ctx).Model(&svcEnvModel.ServiceEnvironment{}).
		Joins("JOIN service_versions sv ON sv.id = service_environments.service_version_id").
		Where("sv.service_id = ?", serviceID)

	if req.EnvironmentID != "" {
		query = query.Where("service_environments.environment_id = ?", req.EnvironmentID)
	}
	if req.Status != "" {
		query = query.Where("service_environments.status = ?", req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	if err := query.Order("service_environments.deployed_at DESC").Limit(limit).Offset(req.Offset).Find(&deployments).Error; err != nil {
		return nil, 0, err
	}

	return deployments, total, nil
}

// ListServiceEnvironmentsByEnvironment lists all deployments to an environment
func (r *repository) ListServiceEnvironmentsByEnvironment(ctx context.Context, environmentID string, req *svcEnvModel.ListDeploymentsRequest) ([]svcEnvModel.ServiceEnvironment, int64, error) {
	var deployments []svcEnvModel.ServiceEnvironment
	var total int64

	query := r.db.WithContext(ctx).Model(&svcEnvModel.ServiceEnvironment{}).Where("environment_id = ?", environmentID)

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	if err := query.Order("deployed_at DESC").Limit(limit).Offset(req.Offset).Find(&deployments).Error; err != nil {
		return nil, 0, err
	}

	return deployments, total, nil
}

// UpdateServiceEnvironment updates a service environment deployment
func (r *repository) UpdateServiceEnvironment(ctx context.Context, se *svcEnvModel.ServiceEnvironment) error {
	return r.db.WithContext(ctx).Save(se).Error
}

// DeleteServiceEnvironment soft deletes a service environment deployment
func (r *repository) DeleteServiceEnvironment(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&svcEnvModel.ServiceEnvironment{}).Error
}

// GetActiveDeployment retrieves the active deployment for a version in an environment
func (r *repository) GetActiveDeployment(ctx context.Context, versionID, environmentID string) (*svcEnvModel.ServiceEnvironment, error) {
	var se svcEnvModel.ServiceEnvironment
	err := r.db.WithContext(ctx).
		Where("service_version_id = ? AND environment_id = ?", versionID, environmentID).
		Where("status IN ?", []string{svcEnvModel.StatusDeployed, svcEnvModel.StatusDeploying}).
		Order("deployed_at DESC").
		First(&se).Error
	if err != nil {
		return nil, err
	}
	return &se, nil
}
