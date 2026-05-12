package role

import (
	"time"

	"github.com/davidsugianto/idp-core/internal/model/permission"
	"gorm.io/gorm"
)

// Role scopes
const (
	ScopePlatform   = "platform"
	ScopeTeam       = "team"
	ScopeEnvironment = "environment"
)

// Role represents a RBAC role
type Role struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name        string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description,omitempty"`
	Scope       string         `gorm:"type:varchar(20);not null;default:'team'" json:"scope"`
	Permissions []permission.Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
	CreatedAt   time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for Role
func (Role) TableName() string {
	return "roles"
}

// CreateRoleRequest represents the request to create a role
type CreateRoleRequest struct {
	Name        string   `json:"name" binding:"required,min=2,max=50"`
	Description string   `json:"description"`
	Scope       string   `json:"scope" binding:"required,oneof=platform team environment"`
	Permissions []string `json:"permissions"` // permission IDs
}

// UpdateRoleRequest represents the request to update a role
type UpdateRoleRequest struct {
	Name        string   `json:"name" binding:"omitempty,min=2,max=50"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"` // permission IDs to replace
}

// RoleResponse represents a role in API responses
type RoleResponse struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Scope       string              `json:"scope"`
	Permissions []PermissionSummary `json:"permissions,omitempty"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

// RoleListResponse represents a list of roles
type RoleListResponse struct {
	Roles []RoleResponse `json:"roles"`
	Total int64          `json:"total"`
}

// PermissionSummary represents a brief permission info
type PermissionSummary struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Resource string `json:"resource"`
	Action   string `json:"action"`
}

// ToRoleResponse converts Role to RoleResponse
func ToRoleResponse(r *Role) *RoleResponse {
	if r == nil {
		return nil
	}

	perms := make([]PermissionSummary, len(r.Permissions))
	for i, p := range r.Permissions {
		perms[i] = PermissionSummary{
			ID:       p.ID,
			Name:     p.Name,
			Resource: p.Resource,
			Action:   p.Action,
		}
	}

	return &RoleResponse{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		Scope:       r.Scope,
		Permissions: perms,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

// ToRoleListResponse converts slice of Role to RoleListResponse
func ToRoleListResponse(roles []Role, total int64) *RoleListResponse {
	responses := make([]RoleResponse, len(roles))
	for i, r := range roles {
		responses[i] = *ToRoleResponse(&r)
	}
	return &RoleListResponse{
		Roles: responses,
		Total: total,
	}
}
