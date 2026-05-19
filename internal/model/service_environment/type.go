package service_environment

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Deployment status constants
const (
	StatusDeployed    = "deployed"
	StatusDeploying   = "deploying"
	StatusFailed      = "failed"
	StatusRolledBack  = "rolled_back"
)

// Metadata is a JSONB type for deployment metadata
type Metadata map[string]interface{}

// Value implements driver.Valuer for Metadata
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements sql.Scanner for Metadata
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan Metadata: expected []byte")
	}
	return json.Unmarshal(bytes, m)
}

// ServiceEnvironment represents a deployment of a service version to an environment
type ServiceEnvironment struct {
	ID                string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	ServiceVersionID  string         `gorm:"type:varchar(36);not null;index:idx_service_environments_version" json:"service_version_id"`
	EnvironmentID     string         `gorm:"type:varchar(36);not null;index:idx_service_environments_environment" json:"environment_id"`
	DeployedBy        string         `gorm:"type:varchar(36)" json:"deployed_by"`
	Status            string         `gorm:"type:varchar(20);not null;default:'deployed';index:idx_service_environments_status" json:"status"`
	DeploymentMetadata Metadata       `gorm:"type:jsonb" json:"deployment_metadata"`
	DeployedAt        time.Time      `json:"deployed_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for ServiceEnvironment
func (ServiceEnvironment) TableName() string {
	return "service_environments"
}

// DeployRequest is the request body for deploying a version to an environment
type DeployRequest struct {
	EnvironmentID     string   `json:"environment_id" binding:"required"`
	DeploymentMetadata Metadata `json:"deployment_metadata"`
}

// UpdateDeploymentRequest is the request body for updating a deployment
type UpdateDeploymentRequest struct {
	Status            *string   `json:"status"`
	DeploymentMetadata *Metadata `json:"deployment_metadata"`
}

// ListDeploymentsRequest is the query params for listing deployments
type ListDeploymentsRequest struct {
	EnvironmentID string `form:"environment_id"`
	Status        string `form:"status"`
	Limit         int    `form:"limit"`
	Offset        int    `form:"offset"`
}

// ServiceEnvironmentResponse is the response body for a single deployment
type ServiceEnvironmentResponse struct {
	ID                string   `json:"id"`
	ServiceVersionID  string   `json:"service_version_id"`
	EnvironmentID     string   `json:"environment_id"`
	DeployedBy        string   `json:"deployed_by"`
	Status            string   `json:"status"`
	DeploymentMetadata Metadata `json:"deployment_metadata"`
	DeployedAt        string   `json:"deployed_at"`
	CreatedAt         string   `json:"created_at"`
	UpdatedAt         string   `json:"updated_at"`
}

// ServiceEnvironmentListResponse is the response body for listing deployments
type ServiceEnvironmentListResponse struct {
	Deployments []ServiceEnvironmentResponse `json:"deployments"`
	Total       int64                        `json:"total"`
}

// EnvironmentServiceResponse is the response for listing services deployed to an environment
type EnvironmentServiceResponse struct {
	ServiceID      string `json:"service_id"`
	ServiceName    string `json:"service_name"`
	VersionID      string `json:"version_id"`
	Version        string `json:"version"`
	DeploymentID   string `json:"deployment_id"`
	DeploymentStatus string `json:"deployment_status"`
	DeployedAt     string `json:"deployed_at"`
}

// EnvironmentServiceListResponse is the response for listing services in an environment
type EnvironmentServiceListResponse struct {
	Services []EnvironmentServiceResponse `json:"services"`
	Total    int64                        `json:"total"`
}

// ToServiceEnvironmentResponse converts a ServiceEnvironment to ServiceEnvironmentResponse
func ToServiceEnvironmentResponse(d *ServiceEnvironment) *ServiceEnvironmentResponse {
	return &ServiceEnvironmentResponse{
		ID:                d.ID,
		ServiceVersionID:  d.ServiceVersionID,
		EnvironmentID:     d.EnvironmentID,
		DeployedBy:        d.DeployedBy,
		Status:            d.Status,
		DeploymentMetadata: d.DeploymentMetadata,
		DeployedAt:        d.DeployedAt.Format(time.RFC3339),
		CreatedAt:         d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         d.UpdatedAt.Format(time.RFC3339),
	}
}

// ToServiceEnvironmentListResponse converts a slice of ServiceEnvironment to ServiceEnvironmentListResponse
func ToServiceEnvironmentListResponse(deployments []ServiceEnvironment, total int64) *ServiceEnvironmentListResponse {
	responses := make([]ServiceEnvironmentResponse, len(deployments))
	for i, d := range deployments {
		responses[i] = *ToServiceEnvironmentResponse(&d)
	}
	return &ServiceEnvironmentListResponse{
		Deployments: responses,
		Total:       total,
	}
}

// ValidStatus checks if the deployment status is valid
func ValidStatus(s string) bool {
	return s == StatusDeployed || s == StatusDeploying || s == StatusFailed || s == StatusRolledBack
}
