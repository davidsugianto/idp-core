package service

import (
	"context"

	endpointModel "github.com/davidsugianto/idp-core/internal/model/service_endpoint"
	serviceModel "github.com/davidsugianto/idp-core/internal/model/service"
	versionModel "github.com/davidsugianto/idp-core/internal/model/service_version"
	"gorm.io/gorm"
)

// Repository defines the interface for service catalog data access
type Repository interface {
	// Service CRUD
	Create(ctx context.Context, svc *serviceModel.Service) error
	GetByID(ctx context.Context, id string) (*serviceModel.Service, error)
	GetByIDIncludingDeleted(ctx context.Context, id string) (*serviceModel.Service, error)
	List(ctx context.Context, req *serviceModel.ListServicesRequest) ([]serviceModel.Service, int64, error)
	Update(ctx context.Context, svc *serviceModel.Service) error
	Delete(ctx context.Context, id string) error

	// Version operations
	CreateVersion(ctx context.Context, v *versionModel.ServiceVersion) error
	GetVersionByID(ctx context.Context, id string) (*versionModel.ServiceVersion, error)
	GetVersionByServiceAndVersion(ctx context.Context, serviceID, version string) (*versionModel.ServiceVersion, error)
	ListVersionsByService(ctx context.Context, serviceID string, req *versionModel.ListServiceVersionsRequest) ([]versionModel.ServiceVersion, int64, error)
	UpdateVersion(ctx context.Context, v *versionModel.ServiceVersion) error
	GetActiveVersionsByService(ctx context.Context, serviceID string) ([]versionModel.ServiceVersion, error)

	// Endpoint operations
	CreateEndpoint(ctx context.Context, ep *endpointModel.ServiceEndpoint) error
	GetEndpointByID(ctx context.Context, id string) (*endpointModel.ServiceEndpoint, error)
	ListEndpointsByVersion(ctx context.Context, versionID string) ([]endpointModel.ServiceEndpoint, error)
	UpdateEndpoint(ctx context.Context, ep *endpointModel.ServiceEndpoint) error
	DeleteEndpoint(ctx context.Context, id string) error

	// Service discovery queries
	SearchServices(ctx context.Context, query string) ([]serviceModel.Service, error)
	ListEndpointsByType(ctx context.Context, endpointType string) ([]endpointModel.ServiceEndpoint, error)
	ExistsByName(ctx context.Context, name string, teamID string) (bool, error)
}

type repository struct {
	db *gorm.DB
}

// Dependencies holds the dependencies for the repository
type Dependencies struct {
	Database *gorm.DB
}

// New creates a new service repository
func New(deps Dependencies) Repository {
	return &repository{db: deps.Database}
}
