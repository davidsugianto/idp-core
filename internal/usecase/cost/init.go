package cost

import (
	"context"

	"github.com/davidsugianto/idp-core/internal/model/cost"
	"github.com/davidsugianto/idp-core/internal/pkg/opencost"
	costRepo "github.com/davidsugianto/idp-core/internal/repository/cost"
)

// Usecase defines the interface for cost tracking business logic
type Usecase interface {
	SyncCosts(ctx context.Context) error
	List(ctx context.Context, filter cost.CostFilter) (*cost.CostListResponse, error)
	GetTeamCosts(ctx context.Context, teamID, namespace, start, end string) (*cost.CostListResponse, error)
}

// OpenCostClient defines the interface for the OpenCost API client
type OpenCostClient interface {
	GetAllocation(ctx context.Context, req opencost.AllocationRequest) (*opencost.AllocationResponse, error)
}

type usecase struct {
	repo           costRepo.Repository
	opencostClient OpenCostClient
}

// Dependencies holds the dependencies for the cost usecase
type Dependencies struct {
	Repo           costRepo.Repository
	OpenCostClient OpenCostClient
}

// New creates a new cost usecase
func New(deps Dependencies) Usecase {
	return &usecase{
		repo:           deps.Repo,
		opencostClient: deps.OpenCostClient,
	}
}