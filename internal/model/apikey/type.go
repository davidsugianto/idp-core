package apikey

import (
	"time"

	"gorm.io/gorm"
)

// APIKey represents an API key for service-to-service authentication
type APIKey struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Key         string         `gorm:"uniqueIndex;not null;type:varchar(64)" json:"key"`
	Name        string         `gorm:"not null;type:varchar(255)" json:"name"`
	Description string         `gorm:"type:text" json:"description"`

	// Ownership
	TeamID    string `gorm:"index;type:varchar(36)" json:"team_id"`
	CreatedBy string `gorm:"type:varchar(36)" json:"created_by"`

	// Permissions
	Scopes     string `gorm:"type:text" json:"scopes"` // JSON encoded array of scopes
	IsAdmin    bool   `gorm:"default:false" json:"is_admin"`
	IsReadOnly bool   `gorm:"default:false" json:"is_read_only"`

	// Rate limiting
	RateLimit int `gorm:"default:100" json:"rate_limit"` // requests per minute

	// Status
	IsActive    bool       `gorm:"default:true;index" json:"is_active"`
	ExpiresAt   *time.Time `gorm:"index" json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	UsageCount  int64      `gorm:"default:0" json:"usage_count"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (APIKey) TableName() string {
	return "api_keys"
}

// CreateAPIKeyRequest is the request body for creating an API key
type CreateAPIKeyRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	TeamID      string   `json:"team_id"`
	Scopes      []string `json:"scopes"`
	IsAdmin     bool     `json:"is_admin"`
	IsReadOnly  bool     `json:"is_read_only"`
	RateLimit   int      `json:"rate_limit"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// APIKeyResponse is the response for API key operations
type APIKeyResponse struct {
	ID          string    `json:"id"`
	Key         string    `json:"key,omitempty"` // Only returned on creation
	Name        string    `json:"name"`
	Description string    `json:"description"`
	TeamID      string    `json:"team_id"`
	Scopes      []string  `json:"scopes"`
	IsAdmin     bool      `json:"is_admin"`
	IsReadOnly  bool      `json:"is_read_only"`
	RateLimit   int       `json:"rate_limit"`
	IsActive    bool      `json:"is_active"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	UsageCount  int64     `json:"usage_count"`
	CreatedAt   time.Time `json:"created_at"`
}

// Scope constants
const (
	ScopeReadEnvironments   = "environments:read"
	ScopeWriteEnvironments  = "environments:write"
	ScopeDeleteEnvironments = "environments:delete"
	ScopeReadWorkloads      = "workloads:read"
	ScopeReadGitOps         = "gitops:read"
	ScopeWriteGitOps        = "gitops:write"
	ScopeAdmin              = "admin"
)

// DefaultScopes returns default scopes for a regular API key
func DefaultScopes() []string {
	return []string{
		ScopeReadEnvironments,
		ScopeWriteEnvironments,
		ScopeDeleteEnvironments,
		ScopeReadWorkloads,
		ScopeReadGitOps,
	}
}

// AdminScopes returns all scopes including admin
func AdminScopes() []string {
	return []string{
		ScopeAdmin,
		ScopeReadEnvironments,
		ScopeWriteEnvironments,
		ScopeDeleteEnvironments,
		ScopeReadWorkloads,
		ScopeReadGitOps,
		ScopeWriteGitOps,
	}
}

// ReadOnlyScopes returns read-only scopes
func ReadOnlyScopes() []string {
	return []string{
		ScopeReadEnvironments,
		ScopeReadWorkloads,
		ScopeReadGitOps,
	}
}

func ToAPIKeyResponse(key *APIKey, includeKey bool) *APIKeyResponse {
	var scopes []string
	// Parse scopes from JSON if needed

	response := &APIKeyResponse{
		ID:          key.ID,
		Name:        key.Name,
		Description: key.Description,
		TeamID:      key.TeamID,
		Scopes:      scopes,
		IsAdmin:     key.IsAdmin,
		IsReadOnly:  key.IsReadOnly,
		RateLimit:   key.RateLimit,
		IsActive:    key.IsActive,
		ExpiresAt:   key.ExpiresAt,
		LastUsedAt:  key.LastUsedAt,
		UsageCount:  key.UsageCount,
		CreatedAt:   key.CreatedAt,
	}

	if includeKey {
		response.Key = key.Key
	}

	return response
}
