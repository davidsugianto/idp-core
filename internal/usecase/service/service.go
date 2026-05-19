package service

import (
	"context"
	"time"

	endpointModel "github.com/davidsugianto/idp-core/internal/model/service_endpoint"
	serviceModel "github.com/davidsugianto/idp-core/internal/model/service"
	depModel "github.com/davidsugianto/idp-core/internal/model/service_dependency"
	svcEnvModel "github.com/davidsugianto/idp-core/internal/model/service_environment"
	versionModel "github.com/davidsugianto/idp-core/internal/model/service_version"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

// ========== Dependency Methods ==========

// AddDependency creates a new dependency between services
func (u *usecase) AddDependency(ctx context.Context, serviceID string, req *depModel.CreateDependencyRequest) (*depModel.DependencyResponse, error) {
	// Verify source service exists
	_, err := u.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return nil, ErrServiceNotFound
	}

	// Verify target service exists
	_, err = u.serviceRepo.GetByID(ctx, req.DependsOnServiceID)
	if err != nil {
		return nil, ErrServiceNotFound
	}

	// Check for self-dependency
	if serviceID == req.DependsOnServiceID {
		return nil, ErrSelfDependency
	}

	// Check for duplicate
	exists, err := u.serviceRepo.ExistsDependency(ctx, serviceID, req.DependsOnServiceID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDependencyAlreadyExists
	}

	// Check for circular dependency
	if err := u.checkCircularDependency(ctx, serviceID, req.DependsOnServiceID); err != nil {
		return nil, err
	}

	// Set defaults
	depType := req.DependencyType
	if depType == "" {
		depType = depModel.TypeRuntime
	}
	if !depModel.ValidDependencyType(depType) {
		return nil, ErrInvalidDependencyType
	}

	now := time.Now()
	dep := &depModel.ServiceDependency{
		ID:                 uuid.New().String(),
		ServiceID:          serviceID,
		DependsOnServiceID: req.DependsOnServiceID,
		DependencyType:     depType,
		Description:        req.Description,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := u.serviceRepo.CreateDependency(ctx, dep); err != nil {
		return nil, err
	}

	return depModel.ToDependencyResponse(dep), nil
}

// GetDependency returns a dependency by ID
func (u *usecase) GetDependency(ctx context.Context, depID string) (*depModel.DependencyResponse, error) {
	dep, err := u.serviceRepo.GetDependencyByID(ctx, depID)
	if err != nil {
		return nil, ErrDependencyNotFound
	}
	return depModel.ToDependencyResponse(dep), nil
}

// ListDependencies returns all dependencies for a service
func (u *usecase) ListDependencies(ctx context.Context, serviceID string, req *depModel.ListDependenciesRequest) (*depModel.DependencyListResponse, error) {
	deps, total, err := u.serviceRepo.ListDependenciesByService(ctx, serviceID, req)
	if err != nil {
		return nil, err
	}
	return depModel.ToDependencyListResponse(deps, total), nil
}

// ListDependents returns all services that depend on a given service
func (u *usecase) ListDependents(ctx context.Context, serviceID string, req *depModel.ListDependenciesRequest) (*depModel.DependencyListResponse, error) {
	deps, total, err := u.serviceRepo.ListDependentsByService(ctx, serviceID, req)
	if err != nil {
		return nil, err
	}
	return depModel.ToDependencyListResponse(deps, total), nil
}

// UpdateDependency updates a dependency
func (u *usecase) UpdateDependency(ctx context.Context, depID string, req *depModel.UpdateDependencyRequest) (*depModel.DependencyResponse, error) {
	dep, err := u.serviceRepo.GetDependencyByID(ctx, depID)
	if err != nil {
		return nil, ErrDependencyNotFound
	}

	if req.DependencyType != nil {
		if !depModel.ValidDependencyType(*req.DependencyType) {
			return nil, ErrInvalidDependencyType
		}
		dep.DependencyType = *req.DependencyType
	}
	if req.Description != nil {
		dep.Description = *req.Description
	}

	dep.UpdatedAt = time.Now()

	if err := u.serviceRepo.UpdateDependency(ctx, dep); err != nil {
		return nil, err
	}

	return depModel.ToDependencyResponse(dep), nil
}

// RemoveDependency soft-deletes a dependency
func (u *usecase) RemoveDependency(ctx context.Context, depID string) error {
	_, err := u.serviceRepo.GetDependencyByID(ctx, depID)
	if err != nil {
		return ErrDependencyNotFound
	}
	return u.serviceRepo.DeleteDependency(ctx, depID)
}

// GetDependencyGraph returns the dependency graph for visualization
func (u *usecase) GetDependencyGraph(ctx context.Context, serviceID string) (*depModel.DependencyGraphResponse, error) {
	// Get the root service
	rootSvc, err := u.serviceRepo.GetByID(ctx, serviceID)
	if err != nil {
		return nil, ErrServiceNotFound
	}

	// Build nodes and edges via BFS
	nodes := make(map[string]*depModel.GraphNode)
	edges := []depModel.GraphEdge{}
	visited := make(map[string]bool)
	queue := []string{serviceID}

	// Add root node
	nodes[serviceID] = &depModel.GraphNode{
		ID:   serviceID,
		Name: rootSvc.Name,
		Type: "root",
	}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		// Get dependencies (outgoing edges)
		deps, _, err := u.serviceRepo.ListDependenciesByService(ctx, currentID, &depModel.ListDependenciesRequest{Limit: 1000})
		if err != nil {
			continue
		}

		for _, dep := range deps {
			// Add edge
			edges = append(edges, depModel.GraphEdge{
				From: dep.ServiceID,
				To:   dep.DependsOnServiceID,
				Type: dep.DependencyType,
			})

			// Add target node if not exists
			if _, exists := nodes[dep.DependsOnServiceID]; !exists {
				targetSvc, err := u.serviceRepo.GetByID(ctx, dep.DependsOnServiceID)
				if err == nil {
					nodes[dep.DependsOnServiceID] = &depModel.GraphNode{
						ID:   dep.DependsOnServiceID,
						Name: targetSvc.Name,
						Type: "dependency",
					}
				}
			}

			// Add to queue for traversal
			if !visited[dep.DependsOnServiceID] {
				queue = append(queue, dep.DependsOnServiceID)
			}
		}

		// Get dependents (incoming edges)
		dependents, _, err := u.serviceRepo.ListDependentsByService(ctx, currentID, &depModel.ListDependenciesRequest{Limit: 1000})
		if err != nil {
			continue
		}

		for _, dep := range dependents {
			// Add edge (reversed for dependents)
			edges = append(edges, depModel.GraphEdge{
				From: dep.ServiceID,
				To:   dep.DependsOnServiceID,
				Type: dep.DependencyType,
			})

			// Add source node if not exists
			if _, exists := nodes[dep.ServiceID]; !exists {
				sourceSvc, err := u.serviceRepo.GetByID(ctx, dep.ServiceID)
				if err == nil {
					nodes[dep.ServiceID] = &depModel.GraphNode{
						ID:   dep.ServiceID,
						Name: sourceSvc.Name,
						Type: "dependent",
					}
				}
			}

			// Add to queue for traversal
			if !visited[dep.ServiceID] {
				queue = append(queue, dep.ServiceID)
			}
		}
	}

	// Convert map to slice
	nodeSlice := make([]depModel.GraphNode, 0, len(nodes))
	for _, node := range nodes {
		nodeSlice = append(nodeSlice, *node)
	}

	return &depModel.DependencyGraphResponse{
		ServiceID:   serviceID,
		ServiceName: rootSvc.Name,
		Nodes:       nodeSlice,
		Edges:       edges,
	}, nil
}

// checkCircularDependency checks if adding a dependency would create a cycle
func (u *usecase) checkCircularDependency(ctx context.Context, serviceID, dependsOnServiceID string) error {
	// BFS from dependsOnServiceID to see if we can reach serviceID
	visited := make(map[string]bool)
	queue := []string{dependsOnServiceID}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		if currentID == serviceID {
			return ErrCircularDependency
		}

		if visited[currentID] {
			continue
		}
		visited[currentID] = true

		// Get dependencies of current service
		deps, _, err := u.serviceRepo.ListDependenciesByService(ctx, currentID, &depModel.ListDependenciesRequest{Limit: 1000})
		if err != nil {
			continue
		}

		for _, dep := range deps {
			if !visited[dep.DependsOnServiceID] {
				queue = append(queue, dep.DependsOnServiceID)
			}
		}
	}

	return nil
}

// ========== Deployment Methods ==========

// DeployToEnvironment deploys a service version to an environment
func (u *usecase) DeployToEnvironment(ctx context.Context, versionID string, req *svcEnvModel.DeployRequest, deployedBy string) (*svcEnvModel.ServiceEnvironmentResponse, error) {
	// Verify version exists
	version, err := u.serviceRepo.GetVersionByID(ctx, versionID)
	if err != nil {
		return nil, ErrVersionNotFound
	}

	// Verify environment exists
	_, err = u.environmentRepo.GetByID(ctx, req.EnvironmentID)
	if err != nil {
		return nil, ErrEnvironmentNotFound
	}

	// Check for existing active deployment (optional - can be removed if multiple deployments allowed)
	existing, err := u.serviceRepo.GetActiveDeployment(ctx, versionID, req.EnvironmentID)
	if err == nil && existing != nil {
		// Mark existing as rolled back
		existing.Status = svcEnvModel.StatusRolledBack
		existing.UpdatedAt = time.Now()
		u.serviceRepo.UpdateServiceEnvironment(ctx, existing)
	}

	now := time.Now()
	se := &svcEnvModel.ServiceEnvironment{
		ID:                uuid.New().String(),
		ServiceVersionID:  versionID,
		EnvironmentID:     req.EnvironmentID,
		DeployedBy:        deployedBy,
		Status:            svcEnvModel.StatusDeployed,
		DeploymentMetadata: req.DeploymentMetadata,
		DeployedAt:        now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if err := u.serviceRepo.CreateServiceEnvironment(ctx, se); err != nil {
		return nil, err
	}

	// Enrich response with service info
	resp := svcEnvModel.ToServiceEnvironmentResponse(se)
	_ = version // version used for potential enrichment
	return resp, nil
}

// GetDeployment returns a deployment by ID
func (u *usecase) GetDeployment(ctx context.Context, deploymentID string) (*svcEnvModel.ServiceEnvironmentResponse, error) {
	se, err := u.serviceRepo.GetServiceEnvironmentByID(ctx, deploymentID)
	if err != nil {
		return nil, ErrDeploymentNotFound
	}
	return svcEnvModel.ToServiceEnvironmentResponse(se), nil
}

// ListDeploymentsByService returns all deployments for a service
func (u *usecase) ListDeploymentsByService(ctx context.Context, serviceID string, req *svcEnvModel.ListDeploymentsRequest) (*svcEnvModel.ServiceEnvironmentListResponse, error) {
	deployments, total, err := u.serviceRepo.ListServiceEnvironmentsByService(ctx, serviceID, req)
	if err != nil {
		return nil, err
	}
	return svcEnvModel.ToServiceEnvironmentListResponse(deployments, total), nil
}

// ListDeploymentsByVersion returns all deployments for a version
func (u *usecase) ListDeploymentsByVersion(ctx context.Context, versionID string, req *svcEnvModel.ListDeploymentsRequest) (*svcEnvModel.ServiceEnvironmentListResponse, error) {
	deployments, total, err := u.serviceRepo.ListServiceEnvironmentsByVersion(ctx, versionID, req)
	if err != nil {
		return nil, err
	}
	return svcEnvModel.ToServiceEnvironmentListResponse(deployments, total), nil
}

// UpdateDeployment updates a deployment status
func (u *usecase) UpdateDeployment(ctx context.Context, deploymentID string, req *svcEnvModel.UpdateDeploymentRequest) (*svcEnvModel.ServiceEnvironmentResponse, error) {
	se, err := u.serviceRepo.GetServiceEnvironmentByID(ctx, deploymentID)
	if err != nil {
		return nil, ErrDeploymentNotFound
	}

	if req.Status != nil {
		if !svcEnvModel.ValidStatus(*req.Status) {
			return nil, ErrInvalidDeploymentStatus
		}
		se.Status = *req.Status
	}
	if req.DeploymentMetadata != nil {
		se.DeploymentMetadata = *req.DeploymentMetadata
	}

	se.UpdatedAt = time.Now()

	if err := u.serviceRepo.UpdateServiceEnvironment(ctx, se); err != nil {
		return nil, err
	}

	return svcEnvModel.ToServiceEnvironmentResponse(se), nil
}

// ListEnvironmentServices returns all services deployed to an environment
func (u *usecase) ListEnvironmentServices(ctx context.Context, environmentID string, req *svcEnvModel.ListDeploymentsRequest) (*svcEnvModel.EnvironmentServiceListResponse, error) {
	deployments, total, err := u.serviceRepo.ListServiceEnvironmentsByEnvironment(ctx, environmentID, req)
	if err != nil {
		return nil, err
	}

	// Enrich with service and version names
	services := make([]svcEnvModel.EnvironmentServiceResponse, len(deployments))
	for i, dep := range deployments {
		// Get version info
		version, err := u.serviceRepo.GetVersionByID(ctx, dep.ServiceVersionID)
		if err != nil {
			continue
		}

		// Get service info
		svc, err := u.serviceRepo.GetByID(ctx, version.ServiceID)
		if err != nil {
			continue
		}

		services[i] = svcEnvModel.EnvironmentServiceResponse{
			ServiceID:        svc.ID,
			ServiceName:      svc.Name,
			VersionID:        version.ID,
			Version:          version.Version,
			DeploymentID:     dep.ID,
			DeploymentStatus: dep.Status,
			DeployedAt:       dep.DeployedAt.Format(time.RFC3339),
		}
	}

	return &svcEnvModel.EnvironmentServiceListResponse{
		Services: services,
		Total:    total,
	}, nil
}

// Helper to handle gorm errors
func isNotFoundErr(err error) bool {
	return err == gorm.ErrRecordNotFound
}
