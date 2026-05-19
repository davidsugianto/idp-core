package service_version

import (
	"time"
)

// Status constants
const (
	StatusActive     = "active"
	StatusDeprecated = "deprecated"
	StatusSuperseded = "superseded"
)

// ServiceVersion represents a version of a service
type ServiceVersion struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	ServiceID string    `gorm:"type:varchar(36);not null;index:idx_service_versions_service" json:"service_id"`
	Version   string    `gorm:"type:varchar(100);not null" json:"version"`
	GitRef    string    `gorm:"type:varchar(255)" json:"git_ref"`
	Changelog string    `gorm:"type:text" json:"changelog"`
	Status    string    `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName returns the table name for ServiceVersion
func (ServiceVersion) TableName() string {
	return "service_versions"
}

// CreateServiceVersionRequest is the request body for creating a version
type CreateServiceVersionRequest struct {
	Version   string `json:"version" binding:"required"`
	GitRef    string `json:"git_ref"`
	Changelog string `json:"changelog"`
}

// UpdateServiceVersionRequest is the request body for updating a version
type UpdateServiceVersionRequest struct {
	GitRef    *string `json:"git_ref"`
	Changelog *string `json:"changelog"`
	Status    *string `json:"status"`
}

// ListServiceVersionsRequest is the query params for listing versions
type ListServiceVersionsRequest struct {
	Status string `form:"status"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

// ServiceVersionResponse is the response body for a single version
type ServiceVersionResponse struct {
	ID        string `json:"id"`
	ServiceID string `json:"service_id"`
	Version   string `json:"version"`
	GitRef    string `json:"git_ref"`
	Changelog string `json:"changelog"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// ServiceVersionListResponse is the response body for listing versions
type ServiceVersionListResponse struct {
	Versions []ServiceVersionResponse `json:"versions"`
	Total    int64                    `json:"total"`
}

// ToServiceVersionResponse converts a ServiceVersion to ServiceVersionResponse
func ToServiceVersionResponse(v *ServiceVersion) *ServiceVersionResponse {
	return &ServiceVersionResponse{
		ID:        v.ID,
		ServiceID: v.ServiceID,
		Version:   v.Version,
		GitRef:    v.GitRef,
		Changelog: v.Changelog,
		Status:    v.Status,
		CreatedAt: v.CreatedAt.Format(time.RFC3339),
	}
}

// ToServiceVersionListResponse converts a slice of ServiceVersion to ServiceVersionListResponse
func ToServiceVersionListResponse(versions []ServiceVersion, total int64) *ServiceVersionListResponse {
	responses := make([]ServiceVersionResponse, len(versions))
	for i, v := range versions {
		responses[i] = *ToServiceVersionResponse(&v)
	}
	return &ServiceVersionListResponse{
		Versions: responses,
		Total:    total,
	}
}

// ValidStatus checks if the status is valid
func ValidStatus(s string) bool {
	return s == StatusActive || s == StatusDeprecated || s == StatusSuperseded
}
