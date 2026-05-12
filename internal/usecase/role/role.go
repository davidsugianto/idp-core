package role

import (
	"context"
	"errors"
	"time"

	roleModel "github.com/davidsugianto/idp-core/internal/model/role"
)

var (
	ErrRoleNotFound         = errors.New("role not found")
	ErrRoleAlreadyExists    = errors.New("role already exists")
	ErrInvalidRoleID        = errors.New("invalid role id")
	ErrInvalidRoleScope     = errors.New("invalid role scope")
	ErrPermissionNotFound   = errors.New("permission not found")
	ErrCannotModifySystem   = errors.New("cannot modify system role")
	ErrUserRoleNotFound     = errors.New("user role assignment not found")
	ErrCannotRevokeLastRole = errors.New("cannot revoke the last role")
)

// validScopes contains all valid role scopes
var validScopes = map[string]bool{
	roleModel.ScopePlatform:    true,
	roleModel.ScopeTeam:        true,
	roleModel.ScopeEnvironment: true,
}

// Create creates a new role
func (u *usecase) Create(ctx context.Context, req roleModel.CreateRoleRequest) (*roleModel.Role, error) {
	// Validate scope
	if !validScopes[req.Scope] {
		return nil, ErrInvalidRoleScope
	}

	// Check if role already exists
	existing, err := u.roleRepo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrRoleAlreadyExists
	}

	// Create role
	role := &roleModel.Role{
		Name:        req.Name,
		Description: req.Description,
		Scope:       req.Scope,
	}

	if err := u.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	// Add permissions if provided
	if len(req.Permissions) > 0 {
		if err := u.roleRepo.SetPermissions(ctx, role.ID, req.Permissions); err != nil {
			return nil, err
		}
	}

	return u.roleRepo.GetByID(ctx, role.ID)
}

// Get retrieves a role by ID
func (u *usecase) Get(ctx context.Context, id string) (*roleModel.Role, error) {
	if id == "" {
		return nil, ErrInvalidRoleID
	}

	role, err := u.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrRoleNotFound
	}

	return role, nil
}

// GetByName retrieves a role by name
func (u *usecase) GetByName(ctx context.Context, name string) (*roleModel.Role, error) {
	role, err := u.roleRepo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrRoleNotFound
	}
	return role, nil
}

// List retrieves a paginated list of roles
func (u *usecase) List(ctx context.Context, limit, offset int) (*roleModel.RoleListResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	roles, total, err := u.roleRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return roleModel.ToRoleListResponse(roles, total), nil
}

// Update updates a role
func (u *usecase) Update(ctx context.Context, id string, req roleModel.UpdateRoleRequest) (*roleModel.Role, error) {
	role, err := u.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		// Check if new name already exists
		existing, err := u.roleRepo.GetByName(ctx, req.Name)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != id {
			return nil, ErrRoleAlreadyExists
		}
		role.Name = req.Name
	}

	if req.Description != "" {
		role.Description = req.Description
	}

	if err := u.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}

	// Update permissions if provided
	if len(req.Permissions) > 0 {
		if err := u.roleRepo.SetPermissions(ctx, role.ID, req.Permissions); err != nil {
			return nil, err
		}
	}

	return u.roleRepo.GetByID(ctx, id)
}

// Delete soft deletes a role
func (u *usecase) Delete(ctx context.Context, id string) error {
	_, err := u.Get(ctx, id)
	if err != nil {
		return err
	}

	return u.roleRepo.SoftDelete(ctx, id)
}

// AddPermission adds a permission to a role
func (u *usecase) AddPermission(ctx context.Context, roleID, permissionID string) error {
	_, err := u.Get(ctx, roleID)
	if err != nil {
		return err
	}

	perm, err := u.permissionRepo.GetByID(ctx, permissionID)
	if err != nil {
		return err
	}
	if perm == nil {
		return ErrPermissionNotFound
	}

	return u.roleRepo.AddPermission(ctx, roleID, permissionID)
}

// RemovePermission removes a permission from a role
func (u *usecase) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	_, err := u.Get(ctx, roleID)
	if err != nil {
		return err
	}

	return u.roleRepo.RemovePermission(ctx, roleID, permissionID)
}

// SetPermissions sets all permissions for a role
func (u *usecase) SetPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	_, err := u.Get(ctx, roleID)
	if err != nil {
		return err
	}

	// Validate all permissions exist
	for _, permID := range permissionIDs {
		perm, err := u.permissionRepo.GetByID(ctx, permID)
		if err != nil {
			return err
		}
		if perm == nil {
			return ErrPermissionNotFound
		}
	}

	return u.roleRepo.SetPermissions(ctx, roleID, permissionIDs)
}

// AssignRole assigns a role to a user
func (u *usecase) AssignRole(ctx context.Context, req roleModel.AssignRoleRequest, grantedBy string) (*roleModel.UserRole, error) {
	// Verify role exists
	role, err := u.roleRepo.GetByID(ctx, req.RoleID)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, ErrRoleNotFound
	}

	// For team-scoped roles, verify team context
	if role.Scope == roleModel.ScopeTeam && req.TeamID == "" {
		return nil, errors.New("team-scoped roles require a team ID")
	}

	userRole := &roleModel.UserRole{
		UserID:    req.UserID,
		RoleID:    req.RoleID,
		TeamID:    req.TeamID,
		GrantedBy: grantedBy,
		GrantedAt: time.Now(),
	}

	if err := u.roleRepo.AssignRole(ctx, userRole); err != nil {
		return nil, err
	}

	return userRole, nil
}

// RevokeRole revokes a role from a user
func (u *usecase) RevokeRole(ctx context.Context, req roleModel.RevokeRoleRequest) error {
	return u.roleRepo.RevokeRole(ctx, req.UserID, req.RoleID, req.TeamID)
}

// GetUserRoles retrieves all roles assigned to a user
func (u *usecase) GetUserRoles(ctx context.Context, userID string) (*roleModel.UserRoleListResponse, error) {
	userRoles, err := u.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}

	responses := make([]roleModel.UserRoleResponse, len(userRoles))
	for i, ur := range userRoles {
		responses[i] = *roleModel.ToUserRoleResponse(&ur)
	}

	return &roleModel.UserRoleListResponse{
		UserRoles: responses,
		Total:     int64(len(responses)),
	}, nil
}

// GetUserRolesByTeam retrieves all roles assigned to a user for a specific team
func (u *usecase) GetUserRolesByTeam(ctx context.Context, userID, teamID string) (*roleModel.UserRoleListResponse, error) {
	userRoles, err := u.roleRepo.GetUserRolesByTeam(ctx, userID, teamID)
	if err != nil {
		return nil, err
	}

	responses := make([]roleModel.UserRoleResponse, len(userRoles))
	for i, ur := range userRoles {
		responses[i] = *roleModel.ToUserRoleResponse(&ur)
	}

	return &roleModel.UserRoleListResponse{
		UserRoles: responses,
		Total:     int64(len(responses)),
	}, nil
}

// HasPermission checks if a user has a specific permission
func (u *usecase) HasPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	return u.roleRepo.HasPermission(ctx, userID, resource, action)
}

// HasTeamPermission checks if a user has a specific permission for a team
func (u *usecase) HasTeamPermission(ctx context.Context, userID, teamID, resource, action string) (bool, error) {
	return u.roleRepo.HasTeamPermission(ctx, userID, teamID, resource, action)
}

// GetUserPermissions retrieves all permissions for a user
func (u *usecase) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	return u.roleRepo.GetUserPermissions(ctx, userID)
}

// GetUserPermissionsByTeam retrieves all permissions for a user in a team context
func (u *usecase) GetUserPermissionsByTeam(ctx context.Context, userID, teamID string) ([]string, error) {
	return u.roleRepo.GetUserPermissionsByTeam(ctx, userID, teamID)
}
