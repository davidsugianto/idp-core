package environment

import (
	"context"

	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	"github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/davidsugianto/idp-core/internal/model/workload"
	deliveryTargetRepo "github.com/davidsugianto/idp-core/internal/repository/delivery_target"
	environmentRepo "github.com/davidsugianto/idp-core/internal/repository/environment"
	gitopsRepo "github.com/davidsugianto/idp-core/internal/repository/gitops"
	provisionerRepo "github.com/davidsugianto/idp-core/internal/repository/provisioner"
	templateRepo "github.com/davidsugianto/idp-core/internal/repository/template"
	liveUpdateUsecase "github.com/davidsugianto/idp-core/internal/usecase/live_update"
	notificationUsecase "github.com/davidsugianto/idp-core/internal/usecase/notification"
)

type Usecase interface {
	Create(ctx context.Context, teamID string, req environment.CreateEnvironmentRequest) (*environment.Environment, error)
	List(ctx context.Context, teamID string) ([]environment.Environment, error)
	Get(ctx context.Context, teamID, id string) (*environment.Environment, error)
	Delete(ctx context.Context, teamID, id string) error
	GetStatus(ctx context.Context, teamID, id string) (*environment.EnvironmentStatusResponse, error)
	TriggerSync(ctx context.Context, teamID, id string) error
	GetGitOpsStatus(ctx context.Context, teamID, id string) (*environment.ArgoStatus, error)
	GetWorkloads(ctx context.Context, teamID, id string) (*workload.WorkloadStatusResponse, error)
	GetWorkloadDetails(ctx context.Context, teamID, id, workloadName string) (*workload.WorkloadInfo, error)
}

func deliveryTargetAllowsPlacement(target *deliveryTargetModel.DeliveryTarget, teamID string) bool {
	if target == nil {
		return false
	}
	if target.AvailabilityState != deliveryTargetModel.AvailabilityAvailable {
		return false
	}
	if target.TeamID == "" {
		return true
	}
	return target.TeamID == teamID
}

type usecase struct {
	environmentRepo    environmentRepo.Repository
	deliveryTargetRepo deliveryTargetRepo.Repository
	provisionerRepo    provisionerRepo.Repository
	gitopsRepo         gitopsRepo.Repository
	templateRepo       templateRepo.Repository
	notificationUC     notificationUsecase.Usecase
	liveUpdateUC       liveUpdateUsecase.Usecase
}

type Dependencies struct {
	EnvironmentRepo    environmentRepo.Repository
	DeliveryTargetRepo deliveryTargetRepo.Repository
	ProvisionerRepo    provisionerRepo.Repository
	GitopsRepo         gitopsRepo.Repository
	TemplateRepo       templateRepo.Repository
	NotificationUC     notificationUsecase.Usecase
	LiveUpdateUC       liveUpdateUsecase.Usecase
}

func New(deps Dependencies) Usecase {
	return &usecase{
		environmentRepo:    deps.EnvironmentRepo,
		deliveryTargetRepo: deps.DeliveryTargetRepo,
		provisionerRepo:    deps.ProvisionerRepo,
		gitopsRepo:         deps.GitopsRepo,
		templateRepo:       deps.TemplateRepo,
		notificationUC:     deps.NotificationUC,
		liveUpdateUC:       deps.LiveUpdateUC,
	}
}
