package auth

import (
	"context"

	roleModel "github.com/davidsugianto/idp-core/internal/model/role"
	roleRepo "github.com/davidsugianto/idp-core/internal/repository/role"
)

// RBACEngine provides role-based access control functionality
type RBACEngine struct {
	roleRepo roleRepo.Repository
}

// NewRBACEngine creates a new RBAC engine
func NewRBACEngine(roleRepo roleRepo.Repository) *RBACEngine {
	return &RBACEngine{
		roleRepo: roleRepo,
	}
}

// CheckPermission checks if a user has permission for a resource and action
func (e *RBACEngine) CheckPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	// Check for platform-level permissions first
	hasPermission, err := e.roleRepo.HasPermission(ctx, userID, resource, action)
	if err != nil {
		return false, err
	}

	if hasPermission {
		return true, nil
	}

	// Check for manage permission (full access)
	hasManage, err := e.roleRepo.HasPermission(ctx, userID, resource, roleModel.ScopePlatform)
	if err != nil {
		return false, err
	}

	return hasManage, nil
}

// CheckTeamPermission checks if a user has permission for a team resource
func (e *RBACEngine) CheckTeamPermission(ctx context.Context, userID, teamID, resource, action string) (bool, error) {
	// Check for team-specific permissions
	hasPermission, err := e.roleRepo.HasTeamPermission(ctx, userID, teamID, resource, action)
	if err != nil {
		return false, err
	}

	if hasPermission {
		return true, nil
	}

	// Check for manage permission (full access)
	hasManage, err := e.roleRepo.HasTeamPermission(ctx, userID, teamID, resource, "manage")
	if err != nil {
		return false, err
	}

	return hasManage, nil
}

// GetUserPermissions returns all permissions for a user
func (e *RBACEngine) GetUserPermissions(ctx context.Context, userID string) (map[string]bool, error) {
	perms, err := e.roleRepo.GetUserPermissions(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool, len(perms))
	for _, p := range perms {
		result[p] = true
	}

	return result, nil
}

// GetUserTeamPermissions returns all permissions for a user in a team context
func (e *RBACEngine) GetUserTeamPermissions(ctx context.Context, userID, teamID string) (map[string]bool, error) {
	perms, err := e.roleRepo.GetUserPermissionsByTeam(ctx, userID, teamID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]bool, len(perms))
	for _, p := range perms {
		result[p] = true
	}

	return result, nil
}

// RequirePermission is a helper that returns error if user lacks permission
func (e *RBACEngine) RequirePermission(ctx context.Context, userID, resource, action string) error {
	has, err := e.CheckPermission(ctx, userID, resource, action)
	if err != nil {
		return err
	}
	if !has {
		return ErrPermissionDenied
	}
	return nil
}

// RequireTeamPermission is a helper that returns error if user lacks team permission
func (e *RBACEngine) RequireTeamPermission(ctx context.Context, userID, teamID, resource, action string) error {
	has, err := e.CheckTeamPermission(ctx, userID, teamID, resource, action)
	if err != nil {
		return err
	}
	if !has {
		return ErrPermissionDenied
	}
	return nil
}
