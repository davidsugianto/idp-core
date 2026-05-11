package user

import (
	"context"

	"github.com/davidsugianto/idp-core/internal/model/user"
	"gorm.io/gorm"
)

// Repository defines the interface for user persistence operations
type Repository interface {
	Create(ctx context.Context, u *user.User) error
	GetByID(ctx context.Context, id string) (*user.User, error)
	GetByEmail(ctx context.Context, email string) (*user.User, error)
	GetByProviderID(ctx context.Context, provider, providerID string) (*user.User, error)
	List(ctx context.Context, limit, offset int) ([]user.User, int64, error)
	ListByStatus(ctx context.Context, status string) ([]user.User, error)
	Update(ctx context.Context, u *user.User) error
	UpdateStatus(ctx context.Context, id, status string) error
	UpdateLastLogin(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error
	HardDelete(ctx context.Context, id string) error
}

type repository struct {
	db *gorm.DB
}

// Dependencies holds repository dependencies
type Dependencies struct {
	Database *gorm.DB
}

// New creates a new user repository
func New(deps Dependencies) Repository {
	return &repository{db: deps.Database}
}
