package role

import (
	"time"

	"gorm.io/gorm"
)

// UserRole represents a user's assigned role
type UserRole struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID    string         `gorm:"type:varchar(36);not null;uniqueIndex:idx_user_role_team" json:"user_id"`
	RoleID    string         `gorm:"type:varchar(36);not null;uniqueIndex:idx_user_role_team" json:"role_id"`
	TeamID    string         `gorm:"type:varchar(36);uniqueIndex:idx_user_role_team" json:"team_id,omitempty"`
	GrantedBy string         `gorm:"type:varchar(36)" json:"granted_by,omitempty"`
	GrantedAt time.Time      `gorm:"not null" json:"granted_at"`
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Role Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`
}

// TableName returns the table name for UserRole
func (UserRole) TableName() string {
	return "user_roles"
}

// AssignRoleRequest represents the request to assign a role to a user
type AssignRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required"`
	TeamID string `json:"team_id"` // Optional for team-scoped roles
}

// RevokeRoleRequest represents the request to revoke a role from a user
type RevokeRoleRequest struct {
	UserID string `json:"user_id" binding:"required"`
	RoleID string `json:"role_id" binding:"required"`
	TeamID string `json:"team_id"` // Optional for team-scoped roles
}

// UserRoleResponse represents a user role assignment in API responses
type UserRoleResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	RoleID    string    `json:"role_id"`
	RoleName  string    `json:"role_name"`
	TeamID    string    `json:"team_id,omitempty"`
	GrantedBy string    `json:"granted_by,omitempty"`
	GrantedAt time.Time `json:"granted_at"`
}

// UserRoleListResponse represents a list of user role assignments
type UserRoleListResponse struct {
	UserRoles []UserRoleResponse `json:"user_roles"`
	Total     int64              `json:"total"`
}

// ToUserRoleResponse converts UserRole to UserRoleResponse
func ToUserRoleResponse(ur *UserRole) *UserRoleResponse {
	if ur == nil {
		return nil
	}
	return &UserRoleResponse{
		ID:        ur.ID,
		UserID:    ur.UserID,
		RoleID:    ur.RoleID,
		RoleName:  ur.Role.Name,
		TeamID:    ur.TeamID,
		GrantedBy: ur.GrantedBy,
		GrantedAt: ur.GrantedAt,
	}
}
