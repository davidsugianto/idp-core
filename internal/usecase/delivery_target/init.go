package delivery_target

import (
	"context"
	"errors"

	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	deliveryTargetRepo "github.com/davidsugianto/idp-core/internal/repository/delivery_target"
	environmentRepo "github.com/davidsugianto/idp-core/internal/repository/environment"
	environmentMovementRepo "github.com/davidsugianto/idp-core/internal/repository/environment_movement"
)

var (
	ErrDeliveryTargetNotFound     = errors.New("delivery target not found")
	ErrDeliveryTargetExists       = errors.New("delivery target already exists")
	ErrDeliveryTargetInUse        = errors.New("delivery target is referenced by existing environments or active movements")
	ErrDeliveryTargetNameRequired = errors.New("delivery target name is required")
	ErrClusterNameRequired        = errors.New("cluster name is required")
	ErrInvalidAvailabilityState   = errors.New("invalid availability state")
	ErrInvalidHealthState         = errors.New("invalid health state")
	ErrIncompleteControlPlaneMetadata = errors.New("control plane metadata is incomplete")
)

type Usecase interface {
	Create(ctx context.Context, req *deliveryTargetModel.CreateDeliveryTargetRequest) (*deliveryTargetModel.DeliveryTargetResponse, error)
	Get(ctx context.Context, id string) (*deliveryTargetModel.DeliveryTargetResponse, error)
	List(ctx context.Context, req *deliveryTargetModel.ListDeliveryTargetsRequest) (*deliveryTargetModel.DeliveryTargetListResponse, error)
	Update(ctx context.Context, id string, req *deliveryTargetModel.UpdateDeliveryTargetRequest) (*deliveryTargetModel.DeliveryTargetResponse, error)
	Delete(ctx context.Context, id string) error
}

type usecase struct {
	deliveryTargetRepo      deliveryTargetRepo.Repository
	environmentRepo         environmentRepo.Repository
	environmentMovementRepo environmentMovementRepo.Repository
}

type Dependencies struct {
	DeliveryTargetRepo      deliveryTargetRepo.Repository
	EnvironmentRepo         environmentRepo.Repository
	EnvironmentMovementRepo environmentMovementRepo.Repository
}

func New(deps Dependencies) Usecase {
	return &usecase{
		deliveryTargetRepo:      deps.DeliveryTargetRepo,
		environmentRepo:         deps.EnvironmentRepo,
		environmentMovementRepo: deps.EnvironmentMovementRepo,
	}
}
