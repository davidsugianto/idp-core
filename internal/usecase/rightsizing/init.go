package rightsizing

import (
	"context"

	rightsizingModel "github.com/davidsugianto/idp-core/internal/model/rightsizing"
	monitoringRepo "github.com/davidsugianto/idp-core/internal/repository/monitoring"
	provisionerRepo "github.com/davidsugianto/idp-core/internal/repository/provisioner"
	rightsizingRepo "github.com/davidsugianto/idp-core/internal/repository/rightsizing"
)

type Usecase interface {
	// Recommendation management
	GenerateRecommendations(ctx context.Context) error
	ListRecommendations(ctx context.Context, req *rightsizingModel.ListRecommendationsRequest) (*rightsizingModel.RecommendationListResponse, error)
	GetRecommendation(ctx context.Context, id string) (*rightsizingModel.RecommendationResponse, error)

	// Actions
	ApplyRecommendation(ctx context.Context, id, userID string) error
	RollbackRecommendation(ctx context.Context, id, userID string) error
	DismissRecommendation(ctx context.Context, id string, reason string) error
}

type usecase struct {
	rightsizingRepo rightsizingRepo.Repository
	provisionerRepo provisionerRepo.Repository
	monitoringRepo  monitoringRepo.Repository
}

type Dependencies struct {
	RightsizingRepo rightsizingRepo.Repository
	ProvisionerRepo provisionerRepo.Repository
	MonitoringRepo  monitoringRepo.Repository
}

func New(deps Dependencies) Usecase {
	return &usecase{
		rightsizingRepo: deps.RightsizingRepo,
		provisionerRepo: deps.ProvisionerRepo,
		monitoringRepo:  deps.MonitoringRepo,
	}
}
