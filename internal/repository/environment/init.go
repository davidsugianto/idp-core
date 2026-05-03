package environment

import (
	"context"
	"time"

	"github.com/davidsugianto/idp-core/internal/model/environment"
	"gorm.io/gorm"
)

// Repository defines the interface for environment persistence operations
type Repository interface {
	Create(ctx context.Context, env *environment.Environment) error
	GetByID(ctx context.Context, id string) (*environment.Environment, error)
	GetByIDAndTeam(ctx context.Context, id, teamID string) (*environment.Environment, error)
	GetByNamespace(ctx context.Context, namespace string) (*environment.Environment, error)
	ListByTeam(ctx context.Context, teamID string) ([]environment.Environment, error)
	ListByStatus(ctx context.Context, status string) ([]environment.Environment, error)
	ListExpired(ctx context.Context) ([]environment.Environment, error)
	Update(ctx context.Context, env *environment.Environment) error
	UpdateStatus(ctx context.Context, id, teamID, status, lastError string) error
	UpdateArgoAppName(ctx context.Context, id, teamID, argoAppName string) error
	UpdateLastSync(ctx context.Context, id string, syncedAt time.Time) error
	IncrementErrorCount(ctx context.Context, id string) error
	ResetErrorCount(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id, teamID string) error
	HardDelete(ctx context.Context, id string) error
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
