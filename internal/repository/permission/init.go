package permission

import (
	"context"

	permissionModel "github.com/davidsugianto/idp-core/internal/model/permission"
	"gorm.io/gorm"
)

// Repository defines the interface for permission data access
type Repository interface {
	Create(ctx context.Context, permission *permissionModel.Permission) error
	GetByID(ctx context.Context, id string) (*permissionModel.Permission, error)
	GetByName(ctx context.Context, name string) (*permissionModel.Permission, error)
	GetByResourceAction(ctx context.Context, resource, action string) (*permissionModel.Permission, error)
	List(ctx context.Context, limit, offset int) ([]permissionModel.Permission, int64, error)
	Update(ctx context.Context, permission *permissionModel.Permission) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	db *gorm.DB
}

// Dependencies holds repository dependencies
type Dependencies struct {
	Database *gorm.DB
}

// New creates a new permission repository
func New(deps Dependencies) Repository {
	return &repository{
		db: deps.Database,
	}
}
