package service

import (
	"context"
	"time"

	endpointModel "github.com/davidsugianto/idp-core/internal/model/service_endpoint"
	serviceModel "github.com/davidsugianto/idp-core/internal/model/service"
	versionModel "github.com/davidsugianto/idp-core/internal/model/service_version"
	"github.com/google/uuid"
)

// Register creates a new service
func (u *usecase) Register(ctx context.Context, req *serviceModel.CreateServiceRequest) (*serviceModel.ServiceResponse, error) {
	if req.Name == "" {
		return nil, ErrServiceNameRequired
	}

	// Check if service already exists in team
	exists, err := u.serviceRepo.ExistsByName(ctx, req.Name, req.TeamID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrServiceAlreadyExists
	}

	// Set defaults
	visibility := req.Visibility
	if visibility == "" {
		visibility = serviceModel.VisibilityTeam
	}
	if !serviceModel.ValidVisibility(visibility) {
		return nil, ErrInvalidVisibility
	}

	now := time.Now()
	svc := &serviceModel.Service{
		ID:         uuid.New().String(),
		Name:       req.Name,
		Description: req.Description,
		TeamID:     req.TeamID,
		Visibility: visibility,
		Status:     serviceModel.StatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := u.serviceRepo.Create(ctx, svc); err != nil {
		return nil, err
	}

	return serviceModel.ToServiceResponse(svc), nil
}

// Get returns a service by ID
func (u *usecase) Get(ctx context.Context, id string) (*serviceModel.ServiceResponse, error) {
	svc, err := u.serviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrServiceNotFound
	}
	return serviceModel.ToServiceResponse(svc), nil
}

// List returns a paginated list of services
func (u *usecase) List(ctx context.Context, req *serviceModel.ListServicesRequest) (*serviceModel.ServiceListResponse, error) {
	services, total, err := u.serviceRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}
	return serviceModel.ToServiceListResponse(services, total), nil
}

// Update updates a service
func (u *usecase) Update(ctx context.Context, id string, req *serviceModel.UpdateServiceRequest) (*serviceModel.ServiceResponse, error) {
	svc, err := u.serviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrServiceNotFound
	}

	if req.Name != nil {
		svc.Name = *req.Name
	}
	if req.Description != nil {
		svc.Description = *req.Description
	}
	if req.Visibility != nil {
		if !serviceModel.ValidVisibility(*req.Visibility) {
			return nil, ErrInvalidVisibility
		}
		svc.Visibility = *req.Visibility
	}
	if req.Status != nil {
		if !serviceModel.ValidStatus(*req.Status) {
			return nil, ErrInvalidStatus
		}
		svc.Status = *req.Status
	}

	svc.UpdatedAt = time.Now()

	if err := u.serviceRepo.Update(ctx, svc); err != nil {
		return nil, err
	}

	return serviceModel.ToServiceResponse(svc), nil
}

// Deregister soft-deletes a service
func (u *usecase) Deregister(ctx context.Context, id string) error {
	_, err := u.serviceRepo.GetByID(ctx, id)
	if err != nil {
		return ErrServiceNotFound
	}
	return u.serviceRepo.Delete(ctx, id)
}

// RegisterVersion creates a new service version
func (u *usecase) RegisterVersion(ctx context.Context, serviceID string, req *versionModel.CreateServiceVersionRequest) (*versionModel.ServiceVersionResponse, error) {
	// Verify service exists
	_, err := u.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return nil, ErrServiceNotFound
	}

	// Check for duplicate version
	existing, err := u.serviceRepo.GetVersionByServiceAndVersion(ctx, serviceID, req.Version)
	if err == nil && existing != nil {
		return nil, ErrVersionAlreadyExists
	}

	now := time.Now()
	v := &versionModel.ServiceVersion{
		ID:        uuid.New().String(),
		ServiceID: serviceID,
		Version:   req.Version,
		GitRef:    req.GitRef,
		Changelog: req.Changelog,
		Status:    versionModel.StatusActive,
		CreatedAt: now,
	}

	if err := u.serviceRepo.CreateVersion(ctx, v); err != nil {
		return nil, err
	}

	return versionModel.ToServiceVersionResponse(v), nil
}

// GetVersion returns a version by ID
func (u *usecase) GetVersion(ctx context.Context, versionID string) (*versionModel.ServiceVersionResponse, error) {
	v, err := u.serviceRepo.GetVersionByID(ctx, versionID)
	if err != nil {
		return nil, ErrVersionNotFound
	}
	return versionModel.ToServiceVersionResponse(v), nil
}

// ListVersions returns a paginated list of versions for a service
func (u *usecase) ListVersions(ctx context.Context, serviceID string, req *versionModel.ListServiceVersionsRequest) (*versionModel.ServiceVersionListResponse, error) {
	versions, total, err := u.serviceRepo.ListVersionsByService(ctx, serviceID, req)
	if err != nil {
		return nil, err
	}
	return versionModel.ToServiceVersionListResponse(versions, total), nil
}

// UpdateVersion updates a service version
func (u *usecase) UpdateVersion(ctx context.Context, versionID string, req *versionModel.UpdateServiceVersionRequest) (*versionModel.ServiceVersionResponse, error) {
	v, err := u.serviceRepo.GetVersionByID(ctx, versionID)
	if err != nil {
		return nil, ErrVersionNotFound
	}

	if req.GitRef != nil {
		v.GitRef = *req.GitRef
	}
	if req.Changelog != nil {
		v.Changelog = *req.Changelog
	}
	if req.Status != nil {
		if !versionModel.ValidStatus(*req.Status) {
			return nil, ErrInvalidStatus
		}
		v.Status = *req.Status
	}

	if err := u.serviceRepo.UpdateVersion(ctx, v); err != nil {
		return nil, err
	}

	return versionModel.ToServiceVersionResponse(v), nil
}

// AddEndpoint creates a new service endpoint
func (u *usecase) AddEndpoint(ctx context.Context, versionID string, req *endpointModel.CreateServiceEndpointRequest) (*endpointModel.ServiceEndpointResponse, error) {
	// Verify version exists
	_, err := u.serviceRepo.GetVersionByID(ctx, versionID)
	if err != nil {
		return nil, ErrVersionNotFound
	}

	// Set defaults
	endpointType := req.Type
	if endpointType == "" {
		endpointType = endpointModel.TypeHTTP
	}
	if !endpointModel.ValidType(endpointType) {
		return nil, ErrEndpointTypeInvalid
	}

	now := time.Now()
	ep := &endpointModel.ServiceEndpoint{
		ID:               uuid.New().String(),
		ServiceVersionID: versionID,
		URL:              req.URL,
		Type:             endpointType,
		Status:           endpointModel.StatusActive,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := u.serviceRepo.CreateEndpoint(ctx, ep); err != nil {
		return nil, err
	}

	return endpointModel.ToServiceEndpointResponse(ep), nil
}

// GetEndpoint returns an endpoint by ID
func (u *usecase) GetEndpoint(ctx context.Context, endpointID string) (*endpointModel.ServiceEndpointResponse, error) {
	ep, err := u.serviceRepo.GetEndpointByID(ctx, endpointID)
	if err != nil {
		return nil, ErrEndpointNotFound
	}
	return endpointModel.ToServiceEndpointResponse(ep), nil
}

// ListEndpoints returns all endpoints for a version
func (u *usecase) ListEndpoints(ctx context.Context, versionID string) (*endpointModel.ServiceEndpointListResponse, error) {
	endpoints, err := u.serviceRepo.ListEndpointsByVersion(ctx, versionID)
	if err != nil {
		return nil, err
	}
	return endpointModel.ToServiceEndpointListResponse(endpoints), nil
}

// UpdateEndpoint updates a service endpoint
func (u *usecase) UpdateEndpoint(ctx context.Context, endpointID string, req *endpointModel.UpdateServiceEndpointRequest) (*endpointModel.ServiceEndpointResponse, error) {
	ep, err := u.serviceRepo.GetEndpointByID(ctx, endpointID)
	if err != nil {
		return nil, ErrEndpointNotFound
	}

	if req.URL != nil {
		ep.URL = *req.URL
	}
	if req.Type != nil {
		if !endpointModel.ValidType(*req.Type) {
			return nil, ErrEndpointTypeInvalid
		}
		ep.Type = *req.Type
	}
	if req.Status != nil {
		if !endpointModel.ValidStatus(*req.Status) {
			return nil, ErrInvalidStatus
		}
		ep.Status = *req.Status
	}

	ep.UpdatedAt = time.Now()

	if err := u.serviceRepo.UpdateEndpoint(ctx, ep); err != nil {
		return nil, err
	}

	return endpointModel.ToServiceEndpointResponse(ep), nil
}

// RemoveEndpoint deletes an endpoint
func (u *usecase) RemoveEndpoint(ctx context.Context, endpointID string) error {
	_, err := u.serviceRepo.GetEndpointByID(ctx, endpointID)
	if err != nil {
		return ErrEndpointNotFound
	}
	return u.serviceRepo.DeleteEndpoint(ctx, endpointID)
}

// Discover searches services by name or description
func (u *usecase) Discover(ctx context.Context, query string) (*serviceModel.ServiceListResponse, error) {
	services, err := u.serviceRepo.SearchServices(ctx, query)
	if err != nil {
		return nil, err
	}
	return serviceModel.ToServiceListResponse(services, int64(len(services))), nil
}

// DiscoverByType returns all endpoints of a specific type
func (u *usecase) DiscoverByType(ctx context.Context, endpointType string) (*endpointModel.ServiceEndpointListResponse, error) {
	if !endpointModel.ValidType(endpointType) {
		return nil, ErrEndpointTypeInvalid
	}
	endpoints, err := u.serviceRepo.ListEndpointsByType(ctx, endpointType)
	if err != nil {
		return nil, err
	}
	return endpointModel.ToServiceEndpointListResponse(endpoints), nil
}
