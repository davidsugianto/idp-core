package resourcequota

import (
	"time"
)

// Status constants
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusExceeded = "exceeded"
)

// ResourceQuota represents resource limits for a namespace
type ResourceQuota struct {
	ID            string `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Namespace     string `gorm:"type:varchar(255);not null;uniqueIndex" json:"namespace"`
	TeamID        string `gorm:"type:varchar(36);not null;index:idx_resource_quotas_team" json:"team_id"`
	EnvironmentID string `gorm:"type:varchar(36);index:idx_resource_quotas_environment" json:"environment_id"`

	// CPU limits
	CPURequestLimit string `gorm:"type:varchar(50)" json:"cpu_request_limit"`
	CPULimitLimit   string `gorm:"type:varchar(50)" json:"cpu_limit_limit"`

	// Memory limits
	MemoryRequestLimit string `gorm:"type:varchar(50)" json:"memory_request_limit"`
	MemoryLimitLimit   string `gorm:"type:varchar(50)" json:"memory_limit_limit"`

	// Storage limits
	StorageRequestLimit string `gorm:"type:varchar(50)" json:"storage_request_limit"`

	// Pod count limit
	PodCountLimit *int `gorm:"type:integer" json:"pod_count_limit"`

	// Object count limits
	ConfigMapCountLimit *int `gorm:"type:integer" json:"configmap_count_limit"`
	SecretCountLimit    *int `gorm:"type:integer" json:"secret_count_limit"`
	PVCCountLimit       *int `gorm:"type:integer" json:"pvc_count_limit"`

	// Current usage (cached)
	CurrentCPURequest     string `gorm:"type:varchar(50)" json:"current_cpu_request"`
	CurrentCPULimit       string `gorm:"type:varchar(50)" json:"current_cpu_limit"`
	CurrentMemoryRequest  string `gorm:"type:varchar(50)" json:"current_memory_request"`
	CurrentMemoryLimit    string `gorm:"type:varchar(50)" json:"current_memory_limit"`
	CurrentStorageRequest string `gorm:"type:varchar(50)" json:"current_storage_request"`
	CurrentPodCount       *int   `gorm:"type:integer" json:"current_pod_count"`

	// Enforcement settings
	Enforce          bool  `gorm:"not null;default:true" json:"enforce"`
	GracePeriodHours *int  `gorm:"type:integer" json:"grace_period_hours"`

	// Status
	Status      string `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	Description string `gorm:"type:text" json:"description"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ResourceQuota) TableName() string {
	return "resource_quotas"
}

// Request/Response DTOs

type CreateResourceQuotaRequest struct {
	Namespace          string `json:"namespace" binding:"required"`
	TeamID             string `json:"team_id" binding:"required"`
	EnvironmentID      string `json:"environment_id"`
	CPURequestLimit    string `json:"cpu_request_limit"`
	CPULimitLimit      string `json:"cpu_limit_limit"`
	MemoryRequestLimit string `json:"memory_request_limit"`
	MemoryLimitLimit   string `json:"memory_limit_limit"`
	StorageRequestLimit string `json:"storage_request_limit"`
	PodCountLimit      *int   `json:"pod_count_limit"`
	ConfigMapCountLimit *int  `json:"configmap_count_limit"`
	SecretCountLimit   *int   `json:"secret_count_limit"`
	PVCCountLimit      *int   `json:"pvc_count_limit"`
	Enforce            bool   `json:"enforce"`
	GracePeriodHours   *int   `json:"grace_period_hours"`
	Description        string `json:"description"`
}

type UpdateResourceQuotaRequest struct {
	CPURequestLimit     *string `json:"cpu_request_limit"`
	CPULimitLimit       *string `json:"cpu_limit_limit"`
	MemoryRequestLimit  *string `json:"memory_request_limit"`
	MemoryLimitLimit    *string `json:"memory_limit_limit"`
	StorageRequestLimit *string `json:"storage_request_limit"`
	PodCountLimit       *int    `json:"pod_count_limit"`
	ConfigMapCountLimit *int    `json:"configmap_count_limit"`
	SecretCountLimit    *int    `json:"secret_count_limit"`
	PVCCountLimit       *int    `json:"pvc_count_limit"`
	Enforce             *bool   `json:"enforce"`
	GracePeriodHours    *int    `json:"grace_period_hours"`
	Description         *string `json:"description"`
}

type ListResourceQuotasRequest struct {
	TeamID       string `form:"team_id"`
	EnvironmentID string `form:"environment_id"`
	Namespace    string `form:"namespace"`
	Status       string `form:"status"`
	Limit        int    `form:"limit"`
	Offset       int    `form:"offset"`
}

type ResourceQuotaResponse struct {
	ID                   string    `json:"id"`
	Namespace            string    `json:"namespace"`
	TeamID               string    `json:"team_id"`
	EnvironmentID        string    `json:"environment_id"`
	CPURequestLimit      string    `json:"cpu_request_limit"`
	CPULimitLimit        string    `json:"cpu_limit_limit"`
	MemoryRequestLimit   string    `json:"memory_request_limit"`
	MemoryLimitLimit     string    `json:"memory_limit_limit"`
	StorageRequestLimit  string    `json:"storage_request_limit"`
	PodCountLimit        *int      `json:"pod_count_limit"`
	ConfigMapCountLimit  *int      `json:"configmap_count_limit"`
	SecretCountLimit     *int      `json:"secret_count_limit"`
	PVCCountLimit        *int      `json:"pvc_count_limit"`
	CurrentCPURequest    string    `json:"current_cpu_request"`
	CurrentCPULimit      string    `json:"current_cpu_limit"`
	CurrentMemoryRequest string    `json:"current_memory_request"`
	CurrentMemoryLimit   string    `json:"current_memory_limit"`
	CurrentStorageRequest string   `json:"current_storage_request"`
	CurrentPodCount      *int      `json:"current_pod_count"`
	Enforce              bool      `json:"enforce"`
	GracePeriodHours     *int      `json:"grace_period_hours"`
	Status               string    `json:"status"`
	Description          string    `json:"description"`
	CreatedAt            string    `json:"created_at"`
	UpdatedAt            string    `json:"updated_at"`

	// Computed utilization percentages
	CPURequestUtilization    float64 `json:"cpu_request_utilization"`
	MemoryRequestUtilization float64 `json:"memory_request_utilization"`
	PodCountUtilization      float64 `json:"pod_count_utilization"`
}

type ResourceQuotaListResponse struct {
	Quotas []ResourceQuotaResponse `json:"quotas"`
	Total  int64                   `json:"total"`
}

type UsageResponse struct {
	Namespace            string    `json:"namespace"`
	CPURequest           string    `json:"cpu_request"`
	CPULimit             string    `json:"cpu_limit"`
	MemoryRequest        string    `json:"memory_request"`
	MemoryLimit          string    `json:"memory_limit"`
	StorageRequest       string    `json:"storage_request"`
	PodCount             int       `json:"pod_count"`
	ConfigMapCount       int       `json:"configmap_count"`
	SecretCount          int       `json:"secret_count"`
	PVCCount             int       `json:"pvc_count"`
	LastUpdated          string    `json:"last_updated"`
}

type QuotaCheckRequest struct {
	Namespace       string `json:"namespace" binding:"required"`
	CPURequest      string `json:"cpu_request"`
	CPULimit        string `json:"cpu_limit"`
	MemoryRequest   string `json:"memory_request"`
	MemoryLimit     string `json:"memory_limit"`
	StorageRequest  string `json:"storage_request"`
	PodDelta        int    `json:"pod_delta"`        // +1 for create, -1 for delete
	ConfigMapDelta  int    `json:"configmap_delta"`
	SecretDelta     int    `json:"secret_delta"`
	PVCDelta        int    `json:"pvc_delta"`
}

type QuotaCheckResponse struct {
	Allowed bool                    `json:"allowed"`
	Reasons []QuotaExceededReason   `json:"reasons,omitempty"`
}

type QuotaExceededReason struct {
	ResourceType string  `json:"resource_type"`
	Requested    string  `json:"requested"`
	Limit        string  `json:"limit"`
	Current      string  `json:"current"`
	Utilization  float64 `json:"utilization"`
}

// Helper functions

func ToResourceQuotaResponse(q *ResourceQuota) *ResourceQuotaResponse {
	resp := &ResourceQuotaResponse{
		ID:                    q.ID,
		Namespace:             q.Namespace,
		TeamID:                q.TeamID,
		EnvironmentID:         q.EnvironmentID,
		CPURequestLimit:       q.CPURequestLimit,
		CPULimitLimit:         q.CPULimitLimit,
		MemoryRequestLimit:    q.MemoryRequestLimit,
		MemoryLimitLimit:      q.MemoryLimitLimit,
		StorageRequestLimit:   q.StorageRequestLimit,
		PodCountLimit:         q.PodCountLimit,
		ConfigMapCountLimit:   q.ConfigMapCountLimit,
		SecretCountLimit:      q.SecretCountLimit,
		PVCCountLimit:         q.PVCCountLimit,
		CurrentCPURequest:     q.CurrentCPURequest,
		CurrentCPULimit:       q.CurrentCPULimit,
		CurrentMemoryRequest:  q.CurrentMemoryRequest,
		CurrentMemoryLimit:    q.CurrentMemoryLimit,
		CurrentStorageRequest: q.CurrentStorageRequest,
		CurrentPodCount:       q.CurrentPodCount,
		Enforce:               q.Enforce,
		GracePeriodHours:      q.GracePeriodHours,
		Status:                q.Status,
		Description:           q.Description,
		CreatedAt:             q.CreatedAt.Format(time.RFC3339),
		UpdatedAt:             q.UpdatedAt.Format(time.RFC3339),
	}

	// Calculate utilization percentages
	resp.CPURequestUtilization = calculateUtilization(q.CurrentCPURequest, q.CPURequestLimit)
	resp.MemoryRequestUtilization = calculateUtilization(q.CurrentMemoryRequest, q.MemoryRequestLimit)
	if q.PodCountLimit != nil && *q.PodCountLimit > 0 && q.CurrentPodCount != nil {
		resp.PodCountUtilization = float64(*q.CurrentPodCount) / float64(*q.PodCountLimit) * 100
	}

	return resp
}

func ToResourceQuotaListResponse(quotas []ResourceQuota, total int64) *ResourceQuotaListResponse {
	responses := make([]ResourceQuotaResponse, len(quotas))
	for i, q := range quotas {
		responses[i] = *ToResourceQuotaResponse(&q)
	}
	return &ResourceQuotaListResponse{
		Quotas: responses,
		Total:  total,
	}
}

// calculateUtilization parses resource strings and returns percentage
func calculateUtilization(current, limit string) float64 {
	if current == "" || limit == "" {
		return 0
	}
	currentVal := parseResource(current)
	limitVal := parseResource(limit)
	if limitVal == 0 {
		return 0
	}
	return currentVal / limitVal * 100
}

// parseResource parses Kubernetes resource string to float64
func parseResource(s string) float64 {
	// Simple implementation - full implementation would use resource.Quantity
	// This is a placeholder that handles basic cases
	return 0 // Will be implemented with proper resource parsing
}

// ValidStatus checks if the status is valid
func ValidStatus(s string) bool {
	return s == StatusActive || s == StatusInactive || s == StatusExceeded
}
