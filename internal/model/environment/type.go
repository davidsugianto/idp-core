package environment

import (
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	urlPattern      = regexp.MustCompile(`https?://[^\s"']+`)
	endpointPattern = regexp.MustCompile(`\b(?:localhost|(?:\d{1,3}\.){3}\d{1,3}|\[[0-9a-fA-F:]+\]|(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,})(?::\d{1,5})\b`)
)

// Environment represents an isolated Kubernetes environment
type Environment struct {
	ID          string `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TeamID      string `gorm:"index;not null;type:varchar(36)" json:"team_id"`
	Name        string `gorm:"not null;type:varchar(255)" json:"name"`
	Description string `gorm:"type:text" json:"description,omitempty"`
	Namespace   string `gorm:"unique;not null;type:varchar(63)" json:"namespace"`
	Status      string `gorm:"not null;type:varchar(20)" json:"status"` // creating|ready|deleting|failed

	// GitOps configuration
	GitRepoURL   string `gorm:"type:varchar(512)" json:"git_repo_url"`
	GitRevision  string `gorm:"default:'main';type:varchar(64)" json:"git_revision"`
	ManifestPath string `gorm:"type:varchar(512)" json:"manifest_path"`
	ArgoAppName  string `gorm:"type:varchar(63)" json:"argo_app_name"`

	// Cluster information
	ClusterName      string `gorm:"type:varchar(255)" json:"cluster_name,omitempty"`
	ClusterServer    string `gorm:"type:varchar(512)" json:"cluster_server,omitempty"`
	DeliveryTargetID string `gorm:"type:varchar(36);index" json:"delivery_target_id,omitempty"`

	// Template provenance
	TemplateInstanceID string `gorm:"type:varchar(36);index" json:"template_instance_id,omitempty"`

	// Resource quotas
	ResourceQuotaCPU    string `gorm:"type:varchar(32)" json:"resource_quota_cpu,omitempty"`
	ResourceQuotaMemory string `gorm:"type:varchar(32)" json:"resource_quota_memory,omitempty"`

	// Metadata
	Labels      string `gorm:"type:text" json:"labels,omitempty"`      // JSON encoded labels
	Annotations string `gorm:"type:text" json:"annotations,omitempty"` // JSON encoded annotations

	// Ownership and lifecycle
	OwnerID    string     `gorm:"type:varchar(36)" json:"owner_id,omitempty"`
	ExpiresAt  *time.Time `gorm:"index" json:"expires_at,omitempty"`
	LastSyncAt *time.Time `json:"last_sync_at,omitempty"`

	// Error tracking
	LastError  string `gorm:"type:text" json:"last_error,omitempty"`
	ErrorCount int    `gorm:"default:0" json:"error_count"`

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

	// Optional template provisioning
	TemplateVersionID string         `json:"template_version_id"`
	TemplateInputs    map[string]any `json:"template_inputs"`

	// Optional target placement
	DeliveryTargetID string `json:"delivery_target_id"`

	// Optional cluster override
	ClusterName   string `json:"cluster_name"`
	ClusterServer string `json:"cluster_server"`

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
	ID           string `json:"id"`
	TeamID       string `json:"team_id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Namespace    string `json:"namespace"`
	Status       string `json:"status"`
	GitRepoURL   string `json:"git_repo_url"`
	GitRevision  string `json:"git_revision"`
	ManifestPath string `json:"manifest_path"`
	ArgoAppName  string `json:"argo_app_name,omitempty"`

	// Cluster info
	ClusterName        string `json:"cluster_name,omitempty"`
	ClusterServer      string `json:"cluster_server,omitempty"`
	DeliveryTargetID   string `json:"delivery_target_id,omitempty"`
	TemplateInstanceID string `json:"template_instance_id,omitempty"`

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

func (s PodSummary) IsZero() bool {
	return s.Total == 0 && s.Running == 0 && s.Pending == 0 && s.Failed == 0
}

// DeploymentSummary contains deployment status
type DeploymentSummary struct {
	Desired   int `json:"desired"`
	Ready     int `json:"ready"`
	Updated   int `json:"updated"`
	Available int `json:"available"`
}

func (s DeploymentSummary) IsZero() bool {
	return s.Desired == 0 && s.Ready == 0 && s.Updated == 0 && s.Available == 0
}

func NewEnvironmentStatusResponse(env *Environment, podSummary PodSummary, deploymentSummary DeploymentSummary) *EnvironmentStatusResponse {
	return &EnvironmentStatusResponse{
		EnvironmentResponse: *ToEnvironmentResponse(env),
		PodSummary:          podSummary,
		DeploymentSummary:   deploymentSummary,
	}
}

// ArgoStatus contains ArgoCD application status
type ArgoStatus struct {
	SyncStatus   string `json:"sync_status"`
	HealthStatus string `json:"health_status"`
	Revision     string `json:"revision"`
	Message      string `json:"message,omitempty"`
}

type TargetResolutionOutcome struct {
	Operation               string `json:"operation"`
	Outcome                 string `json:"outcome"`
	EnvironmentID           string `json:"environment_id,omitempty"`
	DeliveryTargetID        string `json:"delivery_target_id,omitempty"`
	ControlPlaneName        string `json:"control_plane_name,omitempty"`
	UsesDefaultControlPlane bool   `json:"uses_default_control_plane"`
	Error                   string `json:"error,omitempty"`
}

func SanitizeOperationError(err error) string {
	if err == nil {
		return ""
	}

	message := err.Error()
	message = urlPattern.ReplaceAllString(message, "[redacted]")
	message = endpointPattern.ReplaceAllString(message, "[redacted]")
	for _, secret := range []string{"token", "bearer", "kubeconfig", "client-key", "client-certificate", "password", "authorization"} {
		message = strings.ReplaceAll(message, secret, "[redacted]")
		message = strings.ReplaceAll(message, strings.ToUpper(secret[:1])+secret[1:], "[redacted]")
		message = strings.ReplaceAll(message, strings.ToUpper(secret), "[redacted]")
	}

	return message
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
		DeliveryTargetID:    env.DeliveryTargetID,
		TemplateInstanceID:  env.TemplateInstanceID,
		ResourceQuotaCPU:    env.ResourceQuotaCPU,
		ResourceQuotaMemory: env.ResourceQuotaMemory,
		Labels:              labels,
		OwnerID:             env.OwnerID,
		ExpiresAt:           env.ExpiresAt,
		LastSyncAt:          env.LastSyncAt,
		LastError:           env.LastError,
		ErrorCount:          env.ErrorCount,
		CreatedAt:           env.CreatedAt,
		UpdatedAt:           env.UpdatedAt,
	}
}
