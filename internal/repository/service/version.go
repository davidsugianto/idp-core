package service

import (
	"context"

	endpointModel "github.com/davidsugianto/idp-core/internal/model/service_endpoint"
	versionModel "github.com/davidsugianto/idp-core/internal/model/service_version"
)

// CreateVersion creates a new service version
func (r *repository) CreateVersion(ctx context.Context, v *versionModel.ServiceVersion) error {
	return r.db.WithContext(ctx).Create(v).Error
}

// GetVersionByID returns a version by ID
func (r *repository) GetVersionByID(ctx context.Context, id string) (*versionModel.ServiceVersion, error) {
	var v versionModel.ServiceVersion
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// GetVersionByServiceAndVersion returns a version by service ID and version string
func (r *repository) GetVersionByServiceAndVersion(ctx context.Context, serviceID, version string) (*versionModel.ServiceVersion, error) {
	var v versionModel.ServiceVersion
	err := r.db.WithContext(ctx).
		Where("service_id = ? AND version = ?", serviceID, version).
		First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// ListVersionsByService returns a paginated list of versions for a service
func (r *repository) ListVersionsByService(ctx context.Context, serviceID string, req *versionModel.ListServiceVersionsRequest) ([]versionModel.ServiceVersion, int64, error) {
	query := r.db.WithContext(ctx).Model(&versionModel.ServiceVersion{}).Where("service_id = ?", serviceID)

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var versions []versionModel.ServiceVersion
	query = query.Order("created_at DESC")
	if req.Limit > 0 {
		query = query.Limit(req.Limit)
	}
	if req.Offset > 0 {
		query = query.Offset(req.Offset)
	}

	err := query.Find(&versions).Error
	return versions, total, err
}

// UpdateVersion updates a service version
func (r *repository) UpdateVersion(ctx context.Context, v *versionModel.ServiceVersion) error {
	return r.db.WithContext(ctx).Save(v).Error
}

// GetActiveVersionsByService returns all active versions for a service
func (r *repository) GetActiveVersionsByService(ctx context.Context, serviceID string) ([]versionModel.ServiceVersion, error) {
	var versions []versionModel.ServiceVersion
	err := r.db.WithContext(ctx).
		Where("service_id = ? AND status = ?", serviceID, versionModel.StatusActive).
		Order("created_at DESC").
		Find(&versions).Error
	return versions, err
}

// CreateEndpoint creates a new service endpoint
func (r *repository) CreateEndpoint(ctx context.Context, ep *endpointModel.ServiceEndpoint) error {
	return r.db.WithContext(ctx).Create(ep).Error
}

// GetEndpointByID returns an endpoint by ID
func (r *repository) GetEndpointByID(ctx context.Context, id string) (*endpointModel.ServiceEndpoint, error) {
	var ep endpointModel.ServiceEndpoint
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&ep).Error
	if err != nil {
		return nil, err
	}
	return &ep, nil
}

// ListEndpointsByVersion returns all endpoints for a version
func (r *repository) ListEndpointsByVersion(ctx context.Context, versionID string) ([]endpointModel.ServiceEndpoint, error) {
	var endpoints []endpointModel.ServiceEndpoint
	err := r.db.WithContext(ctx).
		Where("service_version_id = ?", versionID).
		Order("created_at DESC").
		Find(&endpoints).Error
	return endpoints, err
}

// UpdateEndpoint updates a service endpoint
func (r *repository) UpdateEndpoint(ctx context.Context, ep *endpointModel.ServiceEndpoint) error {
	return r.db.WithContext(ctx).Save(ep).Error
}

// DeleteEndpoint deletes an endpoint
func (r *repository) DeleteEndpoint(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&endpointModel.ServiceEndpoint{}).Error
}

// ListEndpointsByType returns all endpoints of a specific type
func (r *repository) ListEndpointsByType(ctx context.Context, endpointType string) ([]endpointModel.ServiceEndpoint, error) {
	var endpoints []endpointModel.ServiceEndpoint
	err := r.db.WithContext(ctx).
		Where("type = ? AND status = ?", endpointType, endpointModel.StatusActive).
		Order("created_at DESC").
		Find(&endpoints).Error
	return endpoints, err
}
