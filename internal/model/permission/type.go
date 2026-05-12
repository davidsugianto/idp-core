package permission

import (
	"time"
)

// Resource types
const (
	ResourceEnvironment  = "environment"
	ResourceTeam         = "team"
	ResourceUser         = "user"
	ResourceRole         = "role"
	ResourceAPIKey       = "api_key"
	ResourceCost         = "cost"
	ResourceBudget       = "budget"
	ResourceRightsizing  = "rightsizing"
	ResourceService      = "service"
	ResourceAuditLog     = "audit_log"
)

// Action types
const (
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionManage = "manage" // full access
)

// Permission represents a specific permission in the system
type Permission struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name        string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Resource    string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_resource_action" json:"resource"`
	Action      string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_resource_action" json:"action"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`
}

// TableName returns the table name for Permission
func (Permission) TableName() string {
	return "permissions"
}

// CreatePermissionRequest represents the request to create a permission
type CreatePermissionRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description"`
	Resource    string `json:"resource" binding:"required"`
	Action      string `json:"action" binding:"required,oneof=create read update delete manage"`
}

// PermissionResponse represents a permission in API responses
type PermissionResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	CreatedAt   string `json:"created_at"`
}

// PermissionListResponse represents a list of permissions
type PermissionListResponse struct {
	Permissions []PermissionResponse `json:"permissions"`
	Total       int64                `json:"total"`
}

// ToPermissionResponse converts Permission to PermissionResponse
func ToPermissionResponse(p *Permission) *PermissionResponse {
	if p == nil {
		return nil
	}
	return &PermissionResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		Resource:    p.Resource,
		Action:      p.Action,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ToPermissionListResponse converts slice of Permission to PermissionListResponse
func ToPermissionListResponse(perms []Permission, total int64) *PermissionListResponse {
	responses := make([]PermissionResponse, len(perms))
	for i, p := range perms {
		responses[i] = *ToPermissionResponse(&p)
	}
	return &PermissionListResponse{
		Permissions: responses,
		Total:       total,
	}
}
