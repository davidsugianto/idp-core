package role

import (
	"context"

	roleModel "github.com/davidsugianto/idp-core/internal/model/role"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RolePermission represents the join table for role_permissions
type RolePermission struct {
	RoleID       string `gorm:"primaryKey;type:varchar(36)"`
	PermissionID string `gorm:"primaryKey;type:varchar(36)"`
}

// TableName returns the table name for RolePermission
func (RolePermission) TableName() string {
	return "role_permissions"
}

// AddPermission adds a permission to a role
func (r *repository) AddPermission(ctx context.Context, roleID, permissionID string) error {
	rp := RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}
	return r.db.WithContext(ctx).Create(&rp).Error
}

// RemovePermission removes a permission from a role
func (r *repository) RemovePermission(ctx context.Context, roleID, permissionID string) error {
	return r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&RolePermission{}).Error
}

// GetRolePermissions retrieves all permissions for a role
func (r *repository) GetRolePermissions(ctx context.Context, roleID string) ([]roleModel.Role, error) {
	var role roleModel.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		First(&role, "id = ?", roleID).Error
	if err != nil {
		return nil, err
	}
	return []roleModel.Role{role}, nil
}

// SetPermissions replaces all permissions for a role
func (r *repository) SetPermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Remove existing permissions
		if err := tx.Where("role_id = ?", roleID).Delete(&RolePermission{}).Error; err != nil {
			return err
		}

		// Add new permissions
		for _, permID := range permissionIDs {
			rp := RolePermission{
				RoleID:       roleID,
				PermissionID: permID,
			}
			if err := tx.Create(&rp).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// AssignRole assigns a role to a user
func (r *repository) AssignRole(ctx context.Context, userRole *roleModel.UserRole) error {
	if userRole.ID == "" {
		userRole.ID = uuid.New().String()
	}
	return r.db.WithContext(ctx).Create(userRole).Error
}

// RevokeRole revokes a role from a user
func (r *repository) RevokeRole(ctx context.Context, userID, roleID, teamID string) error {
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID)

	if teamID != "" {
		query = query.Where("team_id = ?", teamID)
	} else {
		query = query.Where("team_id IS NULL")
	}

	return query.Delete(&roleModel.UserRole{}).Error
}

// GetUserRoles retrieves all roles assigned to a user
func (r *repository) GetUserRoles(ctx context.Context, userID string) ([]roleModel.UserRole, error) {
	var userRoles []roleModel.UserRole
	err := r.db.WithContext(ctx).
		Preload("Role.Permissions").
		Where("user_id = ?", userID).
		Find(&userRoles).Error
	return userRoles, err
}

// GetUserRolesByTeam retrieves all roles assigned to a user for a specific team
func (r *repository) GetUserRolesByTeam(ctx context.Context, userID, teamID string) ([]roleModel.UserRole, error) {
	var userRoles []roleModel.UserRole
	err := r.db.WithContext(ctx).
		Preload("Role.Permissions").
		Where("user_id = ? AND (team_id = ? OR team_id IS NULL)", userID, teamID).
		Find(&userRoles).Error
	return userRoles, err
}

// GetUserPermissions retrieves all permission names for a user (platform-level)
func (r *repository) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	var permissions []string
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("DISTINCT permissions.name").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ? AND user_roles.team_id IS NULL", userID).
		Pluck("permissions.name", &permissions).Error
	return permissions, err
}

// GetUserPermissionsByTeam retrieves all permission names for a user in a team context
func (r *repository) GetUserPermissionsByTeam(ctx context.Context, userID, teamID string) ([]string, error) {
	var permissions []string
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("DISTINCT permissions.name").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ? AND (user_roles.team_id = ? OR user_roles.team_id IS NULL)", userID, teamID).
		Pluck("permissions.name", &permissions).Error
	return permissions, err
}

// HasPermission checks if a user has a specific permission
func (r *repository) HasPermission(ctx context.Context, userID, resource, action string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Where("permissions.resource = ?", resource).
		Where("permissions.action = ? OR permissions.action = ?", action, "manage").
		Where("user_roles.deleted_at IS NULL").
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// HasTeamPermission checks if a user has a specific permission for a team
func (r *repository) HasTeamPermission(ctx context.Context, userID, teamID, resource, action string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON role_permissions.role_id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Where("user_roles.team_id = ? OR user_roles.team_id IS NULL", teamID).
		Where("permissions.resource = ?", resource).
		Where("permissions.action = ? OR permissions.action = ?", action, "manage").
		Where("user_roles.deleted_at IS NULL").
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}
