package role

import (
	"context"

	roleModel "github.com/davidsugianto/idp-core/internal/model/role"
	"gorm.io/gorm"
)

// Repository defines the interface for role data access
type Repository interface {
	// Role CRUD
	Create(ctx context.Context, role *roleModel.Role) error
	GetByID(ctx context.Context, id string) (*roleModel.Role, error)
	GetByName(ctx context.Context, name string) (*roleModel.Role, error)
	List(ctx context.Context, limit, offset int) ([]roleModel.Role, int64, error)
	Update(ctx context.Context, role *roleModel.Role) error
	SoftDelete(ctx context.Context, id string) error

	// Role permissions
	AddPermission(ctx context.Context, roleID, permissionID string) error
	RemovePermission(ctx context.Context, roleID, permissionID string) error
	GetRolePermissions(ctx context.Context, roleID string) ([]roleModel.Role, error)
	SetPermissions(ctx context.Context, roleID string, permissionIDs []string) error

	// User roles
	AssignRole(ctx context.Context, userRole *roleModel.UserRole) error
	RevokeRole(ctx context.Context, userID, roleID, teamID string) error
	GetUserRoles(ctx context.Context, userID string) ([]roleModel.UserRole, error)
	GetUserRolesByTeam(ctx context.Context, userID, teamID string) ([]roleModel.UserRole, error)
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
	GetUserPermissionsByTeam(ctx context.Context, userID, teamID string) ([]string, error)

	// Permission checks
	HasPermission(ctx context.Context, userID, resource, action string) (bool, error)
	HasTeamPermission(ctx context.Context, userID, teamID, resource, action string) (bool, error)
}

type repository struct {
	db *gorm.DB
}

// Dependencies holds repository dependencies
type Dependencies struct {
	Database *gorm.DB
}

// New creates a new role repository
func New(deps Dependencies) Repository {
	return &repository{
		db: deps.Database,
	}
}
