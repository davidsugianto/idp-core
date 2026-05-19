package service

import (
	"context"
	"errors"

	endpointModel "github.com/davidsugianto/idp-core/internal/model/service_endpoint"
	serviceModel "github.com/davidsugianto/idp-core/internal/model/service"
	versionModel "github.com/davidsugianto/idp-core/internal/model/service_version"
	serviceRepo "github.com/davidsugianto/idp-core/internal/repository/service"
)

var (
	ErrServiceNotFound      = errors.New("service not found")
	ErrServiceNameRequired  = errors.New("service name is required")
	ErrServiceAlreadyExists = errors.New("service already exists")
	ErrInvalidVisibility     = errors.New("invalid visibility")
	ErrInvalidStatus         = errors.New("invalid status")
	ErrVersionNotFound       = errors.New("version not found")
	ErrVersionAlreadyExists  = errors.New("version already exists")
	ErrEndpointNotFound      = errors.New("endpoint not found")
	ErrEndpointTypeInvalid   = errors.New("invalid endpoint type")
)

// Usecase defines the interface for service catalog business logic
type Usecase interface {
	// Service management
	Register(ctx context.Context, req *serviceModel.CreateServiceRequest) (*serviceModel.ServiceResponse, error)
	Get(ctx context.Context, id string) (*serviceModel.ServiceResponse, error)
	List(ctx context.Context, req *serviceModel.ListServicesRequest) (*serviceModel.ServiceListResponse, error)
	Update(ctx context.Context, id string, req *serviceModel.UpdateServiceRequest) (*serviceModel.ServiceResponse, error)
	Deregister(ctx context.Context, id string) error

	// Version management
	RegisterVersion(ctx context.Context, serviceID string, req *versionModel.CreateServiceVersionRequest) (*versionModel.ServiceVersionResponse, error)
	GetVersion(ctx context.Context, versionID string) (*versionModel.ServiceVersionResponse, error)
	ListVersions(ctx context.Context, serviceID string, req *versionModel.ListServiceVersionsRequest) (*versionModel.ServiceVersionListResponse, error)
	UpdateVersion(ctx context.Context, versionID string, req *versionModel.UpdateServiceVersionRequest) (*versionModel.ServiceVersionResponse, error)

	// Endpoint management
	AddEndpoint(ctx context.Context, versionID string, req *endpointModel.CreateServiceEndpointRequest) (*endpointModel.ServiceEndpointResponse, error)
	GetEndpoint(ctx context.Context, endpointID string) (*endpointModel.ServiceEndpointResponse, error)
	ListEndpoints(ctx context.Context, versionID string) (*endpointModel.ServiceEndpointListResponse, error)
	UpdateEndpoint(ctx context.Context, endpointID string, req *endpointModel.UpdateServiceEndpointRequest) (*endpointModel.ServiceEndpointResponse, error)
	RemoveEndpoint(ctx context.Context, endpointID string) error

	// Service discovery
	Discover(ctx context.Context, query string) (*serviceModel.ServiceListResponse, error)
	DiscoverByType(ctx context.Context, endpointType string) (*endpointModel.ServiceEndpointListResponse, error)
}

type usecase struct {
	serviceRepo serviceRepo.Repository
}

// Dependencies holds the dependencies for the usecase
type Dependencies struct {
	ServiceRepo serviceRepo.Repository
}

// New creates a new service usecase
func New(deps Dependencies) Usecase {
	return &usecase{
		serviceRepo: deps.ServiceRepo,
	}
}
