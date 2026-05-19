package service

import (
	"context"
	"errors"

	endpointModel "github.com/davidsugianto/idp-core/internal/model/service_endpoint"
	serviceModel "github.com/davidsugianto/idp-core/internal/model/service"
	depModel "github.com/davidsugianto/idp-core/internal/model/service_dependency"
	svcEnvModel "github.com/davidsugianto/idp-core/internal/model/service_environment"
	versionModel "github.com/davidsugianto/idp-core/internal/model/service_version"
	envRepo "github.com/davidsugianto/idp-core/internal/repository/environment"
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

	// Dependency errors
	ErrDependencyNotFound      = errors.New("dependency not found")
	ErrDependencyAlreadyExists = errors.New("dependency already exists")
	ErrCircularDependency      = errors.New("circular dependency detected")
	ErrSelfDependency          = errors.New("service cannot depend on itself")
	ErrInvalidDependencyType   = errors.New("invalid dependency type")

	// Deployment errors
	ErrDeploymentNotFound      = errors.New("deployment not found")
	ErrEnvironmentNotFound     = errors.New("environment not found")
	ErrInvalidDeploymentStatus = errors.New("invalid deployment status")
	ErrDeploymentAlreadyExists = errors.New("active deployment already exists for this version and environment")
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

	// Dependency management
	AddDependency(ctx context.Context, serviceID string, req *depModel.CreateDependencyRequest) (*depModel.DependencyResponse, error)
	GetDependency(ctx context.Context, depID string) (*depModel.DependencyResponse, error)
	ListDependencies(ctx context.Context, serviceID string, req *depModel.ListDependenciesRequest) (*depModel.DependencyListResponse, error)
	ListDependents(ctx context.Context, serviceID string, req *depModel.ListDependenciesRequest) (*depModel.DependencyListResponse, error)
	UpdateDependency(ctx context.Context, depID string, req *depModel.UpdateDependencyRequest) (*depModel.DependencyResponse, error)
	RemoveDependency(ctx context.Context, depID string) error
	GetDependencyGraph(ctx context.Context, serviceID string) (*depModel.DependencyGraphResponse, error)

	// Deployment management
	DeployToEnvironment(ctx context.Context, versionID string, req *svcEnvModel.DeployRequest, deployedBy string) (*svcEnvModel.ServiceEnvironmentResponse, error)
	GetDeployment(ctx context.Context, deploymentID string) (*svcEnvModel.ServiceEnvironmentResponse, error)
	ListDeploymentsByService(ctx context.Context, serviceID string, req *svcEnvModel.ListDeploymentsRequest) (*svcEnvModel.ServiceEnvironmentListResponse, error)
	ListDeploymentsByVersion(ctx context.Context, versionID string, req *svcEnvModel.ListDeploymentsRequest) (*svcEnvModel.ServiceEnvironmentListResponse, error)
	UpdateDeployment(ctx context.Context, deploymentID string, req *svcEnvModel.UpdateDeploymentRequest) (*svcEnvModel.ServiceEnvironmentResponse, error)
	ListEnvironmentServices(ctx context.Context, environmentID string, req *svcEnvModel.ListDeploymentsRequest) (*svcEnvModel.EnvironmentServiceListResponse, error)
}

type usecase struct {
	serviceRepo   serviceRepo.Repository
	environmentRepo envRepo.Repository
}

// Dependencies holds the dependencies for the usecase
type Dependencies struct {
	ServiceRepo   serviceRepo.Repository
	EnvironmentRepo envRepo.Repository
}

// New creates a new service usecase
func New(deps Dependencies) Usecase {
	return &usecase{
		serviceRepo:   deps.ServiceRepo,
		environmentRepo: deps.EnvironmentRepo,
	}
}
