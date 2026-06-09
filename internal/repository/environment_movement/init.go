package environment_movement

import (
	"context"

	environmentMovementModel "github.com/davidsugianto/idp-core/internal/model/environment_movement"
	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, movement *environmentMovementModel.EnvironmentMovement) error
	GetByID(ctx context.Context, id string) (*environmentMovementModel.EnvironmentMovement, error)
	ListByEnvironment(ctx context.Context, environmentID string) ([]environmentMovementModel.EnvironmentMovement, error)
	ListActiveByTarget(ctx context.Context, targetID string) ([]environmentMovementModel.EnvironmentMovement, error)
	Update(ctx context.Context, movement *environmentMovementModel.EnvironmentMovement) error
	UpdateStatus(ctx context.Context, id, status string, progressPercent int, message string) error
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
