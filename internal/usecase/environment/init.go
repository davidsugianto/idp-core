package environment

import (
	"context"

	"github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/davidsugianto/idp-core/internal/model/workload"
	environmentRepo "github.com/davidsugianto/idp-core/internal/repository/environment"
	gitopsRepo "github.com/davidsugianto/idp-core/internal/repository/gitops"
	provisionerRepo "github.com/davidsugianto/idp-core/internal/repository/provisioner"
)

type Usecase interface {
	// Environment CRUD
	Create(ctx context.Context, teamID string, req environment.CreateEnvironmentRequest) (*environment.Environment, error)
	List(ctx context.Context, teamID string) ([]environment.Environment, error)
	Get(ctx context.Context, teamID, id string) (*environment.Environment, error)
	Delete(ctx context.Context, teamID, id string) error

	// Status
	GetStatus(ctx context.Context, teamID, id string) (*environment.EnvironmentStatusResponse, error)

	// GitOps operations
	TriggerSync(ctx context.Context, teamID, id string) error
	GetGitOpsStatus(ctx context.Context, teamID, id string) (*environment.ArgoStatus, error)

	// Workload operations
	GetWorkloads(ctx context.Context, teamID, id string) (*workload.WorkloadStatusResponse, error)
	GetWorkloadDetails(ctx context.Context, teamID, id, workloadName string) (*workload.WorkloadInfo, error)
}

type usecase struct {
	environmentRepo environmentRepo.Repository
	provisionerRepo provisionerRepo.Repository
	gitopsRepo      gitopsRepo.Repository
}

type Dependencies struct {
	EnvironmentRepo  environmentRepo.Repository
	ProvisionerRepo  provisionerRepo.Repository
	GitopsRepo       gitopsRepo.Repository
}

func New(deps Dependencies) Usecase {
	return &usecase{
		environmentRepo:  deps.EnvironmentRepo,
		provisionerRepo:  deps.ProvisionerRepo,
		gitopsRepo:       deps.GitopsRepo,
	}
}
