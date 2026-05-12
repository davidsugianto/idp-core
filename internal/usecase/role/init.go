package role

import (
	"context"

	roleModel "github.com/davidsugianto/idp-core/internal/model/role"
	permissionRepo "github.com/davidsugianto/idp-core/internal/repository/permission"
	roleRepo "github.com/davidsugianto/idp-core/internal/repository/role"
)

// Usecase defines the interface for role business logic
type Usecase interface {
	// Role CRUD
	Create(ctx context.Context, req roleModel.CreateRoleRequest) (*roleModel.Role, error)
	Get(ctx context.Context, id string) (*roleModel.Role, error)
	GetByName(ctx context.Context, name string) (*roleModel.Role, error)
	List(ctx context.Context, limit, offset int) (*roleModel.RoleListResponse, error)
	Update(ctx context.Context, id string, req roleModel.UpdateRoleRequest) (*roleModel.Role, error)
	Delete(ctx context.Context, id string) error

	// Role permissions
	AddPermission(ctx context.Context, roleID, permissionID string) error
	RemovePermission(ctx context.Context, roleID, permissionID string) error
	SetPermissions(ctx context.Context, roleID string, permissionIDs []string) error

	// User role assignments
	AssignRole(ctx context.Context, req roleModel.AssignRoleRequest, grantedBy string) (*roleModel.UserRole, error)
	RevokeRole(ctx context.Context, req roleModel.RevokeRoleRequest) error
	GetUserRoles(ctx context.Context, userID string) (*roleModel.UserRoleListResponse, error)
	GetUserRolesByTeam(ctx context.Context, userID, teamID string) (*roleModel.UserRoleListResponse, error)

	// Permission checks
	HasPermission(ctx context.Context, userID, resource, action string) (bool, error)
	HasTeamPermission(ctx context.Context, userID, teamID, resource, action string) (bool, error)
	GetUserPermissions(ctx context.Context, userID string) ([]string, error)
	GetUserPermissionsByTeam(ctx context.Context, userID, teamID string) ([]string, error)
}

type usecase struct {
	roleRepo       roleRepo.Repository
	permissionRepo permissionRepo.Repository
}

// Dependencies holds usecase dependencies
type Dependencies struct {
	RoleRepo       roleRepo.Repository
	PermissionRepo permissionRepo.Repository
}

// New creates a new role usecase
func New(deps Dependencies) Usecase {
	return &usecase{
		roleRepo:       deps.RoleRepo,
		permissionRepo: deps.PermissionRepo,
	}
}
