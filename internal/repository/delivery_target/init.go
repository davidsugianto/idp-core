package delivery_target

import (
	"context"

	deliveryTargetModel "github.com/davidsugianto/idp-core/internal/model/delivery_target"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, target *deliveryTargetModel.DeliveryTarget) error
	GetByID(ctx context.Context, id string) (*deliveryTargetModel.DeliveryTarget, error)
	Update(ctx context.Context, target *deliveryTargetModel.DeliveryTarget) error
	UpdateAvailability(ctx context.Context, id, availabilityState, healthState, capacitySummary string) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, req *deliveryTargetModel.ListDeliveryTargetsRequest) ([]deliveryTargetModel.DeliveryTarget, int64, error)
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	ExistsBySlugExcludingID(ctx context.Context, slug, id string) (bool, error)
}

type repository struct {
	db *gorm.DB
}

type Dependencies struct {
	Database *gorm.DB
}

func New(deps Dependencies) Repository {
	return &repository{db: deps.Database}
}
