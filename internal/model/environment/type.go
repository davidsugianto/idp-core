package environment

import (
	"time"

	"gorm.io/gorm"
)

// Environment represents an isolated Kubernetes environment
type Environment struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TeamID       string         `gorm:"index;not null;type:varchar(36)" json:"team_id"`
	Name         string         `gorm:"not null;type:varchar(255)" json:"name"`
	Description  string         `gorm:"type:text" json:"description,omitempty"`
	Namespace    string         `gorm:"unique;not null;type:varchar(63)" json:"namespace"`
	Status       string         `gorm:"not null;type:varchar(20)" json:"status"` // creating|ready|deleting|failed

	// GitOps configuration
	GitRepoURL   string         `gorm:"type:varchar(512)" json:"git_repo_url"`
	GitRevision  string         `gorm:"default:'main';type:varchar(64)" json:"git_revision"`
	ManifestPath string         `gorm:"type:varchar(512)" json:"manifest_path"`
	ArgoAppName  string         `gorm:"type:varchar(63)" json:"argo_app_name"`

	// Cluster information
	ClusterName   string `gorm:"type:varchar(255)" json:"cluster_name,omitempty"`
	ClusterServer string `gorm:"type:varchar(512)" json:"cluster_server,omitempty"`

	// Resource quotas
	ResourceQuotaCPU    string `gorm:"type:varchar(32)" json:"resource_quota_cpu,omitempty"`
	ResourceQuotaMemory string `gorm:"type:varchar(32)" json:"resource_quota_memory,omitempty"`

	// Metadata
	Labels      string `gorm:"type:text" json:"labels,omitempty"`       // JSON encoded labels
	Annotations string `gorm:"type:text" json:"annotations,omitempty"` // JSON encoded annotations

	// Ownership and lifecycle
	OwnerID     string     `gorm:"type:varchar(36)" json:"owner_id,omitempty"`
	ExpiresAt   *time.Time `gorm:"index" json:"expires_at,omitempty"`
	LastSyncAt  *time.Time `json:"last_sync_at,omitempty"`

	// Error tracking
	LastError    string `gorm:"type:text" json:"last_error,omitempty"`
	ErrorCount   int    `gorm:"default:0" json:"error_count"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Environment) TableName() string {
	return "environments"
}

// CreateEnvironmentRequest is the request body for creating an environment
type CreateEnvironmentRequest struct {
	Name         string `json:"name" binding:"required"`
	Description  string `json:"description"`
	GitRepoURL   string `json:"git_repo_url" binding:"required"`
	ManifestPath string `json:"manifest_path" binding:"required"`
	GitRevision  string `json:"git_revision"` // optional, defaults to "main"

	// Optional cluster override
	ClusterName string `json:"cluster_name"`

	// Optional resource quotas
	ResourceQuotaCPU    string `json:"resource_quota_cpu"`
	ResourceQuotaMemory string `json:"resource_quota_memory"`

	// Optional labels
	Labels map[string]string `json:"labels"`

	// Optional expiration
	ExpiresAt *time.Time `json:"expires_at"`
}

// UpdateEnvironmentRequest is the request body for updating an environment
type UpdateEnvironmentRequest struct {
	Description         string            `json:"description"`
	GitRepoURL          string            `json:"git_repo_url"`
	GitRevision         string            `json:"git_revision"`
	ManifestPath        string            `json:"manifest_path"`
	ResourceQuotaCPU    string            `json:"resource_quota_cpu"`
	ResourceQuotaMemory string            `json:"resource_quota_memory"`
	Labels              map[string]string `json:"labels"`
	ExpiresAt           *time.Time        `json:"expires_at"`
}

// EnvironmentResponse is the response for environment endpoints
type EnvironmentResponse struct {
	ID           string    `json:"id"`
	TeamID       string    `json:"team_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Namespace    string    `json:"namespace"`
	Status       string    `json:"status"`
	GitRepoURL   string    `json:"git_repo_url"`
	GitRevision  string    `json:"git_revision"`
	ManifestPath string    `json:"manifest_path"`
	ArgoAppName  string    `json:"argo_app_name,omitempty"`

	// Cluster info
	ClusterName   string `json:"cluster_name,omitempty"`
	ClusterServer string `json:"cluster_server,omitempty"`

	// Resource quotas
	ResourceQuotaCPU    string `json:"resource_quota_cpu,omitempty"`
	ResourceQuotaMemory string `json:"resource_quota_memory,omitempty"`

	// Metadata
	Labels map[string]string `json:"labels,omitempty"`

	// Lifecycle
	OwnerID    string     `json:"owner_id,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	LastSyncAt *time.Time `json:"last_sync_at,omitempty"`

	// Error tracking
	LastError  string `json:"last_error,omitempty"`
	ErrorCount int    `json:"error_count"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EnvironmentStatusResponse includes live K8s and ArgoCD status
type EnvironmentStatusResponse struct {
	EnvironmentResponse
	PodSummary        PodSummary        `json:"pod_summary"`
	DeploymentSummary DeploymentSummary `json:"deployment_summary"`
	ArgoStatus        ArgoStatus        `json:"argo_status"`
}

// PodSummary contains pod status counts
type PodSummary struct {
	Total   int `json:"total"`
	Running int `json:"running"`
	Pending int `json:"pending"`
	Failed  int `json:"failed"`
}

// DeploymentSummary contains deployment status
type DeploymentSummary struct {
	Desired   int `json:"desired"`
	Ready     int `json:"ready"`
	Updated   int `json:"updated"`
	Available int `json:"available"`
}

// ArgoStatus contains ArgoCD application status
type ArgoStatus struct {
	SyncStatus   string `json:"sync_status"`
	HealthStatus string `json:"health_status"`
	Revision     string `json:"revision"`
	Message      string `json:"message,omitempty"`
}

func ToEnvironmentResponse(env *Environment) *EnvironmentResponse {
	var labels map[string]string
	if env.Labels != "" {
		// Parse JSON labels if needed
		labels = make(map[string]string)
	}

	return &EnvironmentResponse{
		ID:                  env.ID,
		TeamID:              env.TeamID,
		Name:                env.Name,
		Description:         env.Description,
		Namespace:           env.Namespace,
		Status:              env.Status,
		GitRepoURL:          env.GitRepoURL,
		GitRevision:         env.GitRevision,
		ManifestPath:        env.ManifestPath,
		ArgoAppName:         env.ArgoAppName,
		ClusterName:         env.ClusterName,
		ClusterServer:       env.ClusterServer,
		ResourceQuotaCPU:    env.ResourceQuotaCPU,
		ResourceQuotaMemory: env.ResourceQuotaMemory,
		Labels:              labels,
		OwnerID:             env.OwnerID,
		ExpiresAt:           env.ExpiresAt,
		LastSyncAt:          env.LastSyncAt,
		LastError:           env.LastError,
		ErrorCount:          env.ErrorCount,
		CreatedAt:           env.CreatedAt,
		UpdatedAt:            env.UpdatedAt,
	}
}
