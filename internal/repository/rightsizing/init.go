package rightsizing

import (
	"context"

	rightsizingModel "github.com/davidsugianto/idp-core/internal/model/rightsizing"
	"gorm.io/gorm"
)

type Repository interface {
	// Recommendation CRUD
	Create(ctx context.Context, recommendation *rightsizingModel.RightsizingRecommendation) error
	GetByID(ctx context.Context, id string) (*rightsizingModel.RightsizingRecommendation, error)
	List(ctx context.Context, req *rightsizingModel.ListRecommendationsRequest) ([]rightsizingModel.RightsizingRecommendation, int64, error)
	Update(ctx context.Context, recommendation *rightsizingModel.RightsizingRecommendation) error
	Delete(ctx context.Context, id string) error

	// Batch operations for recommendation generation
	DeletePendingByWorkload(ctx context.Context, namespace, workloadName, workloadType string) error
	ListPendingByWorkload(ctx context.Context, namespace, workloadName, workloadType string) ([]rightsizingModel.RightsizingRecommendation, error)

	// Utility
	ExistsPendingForContainer(ctx context.Context, namespace, workloadName, workloadType, containerName string) (bool, error)
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
