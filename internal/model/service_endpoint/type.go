package service_endpoint

import (
	"time"
)

// Type constants
const (
	TypeHTTP = "http"
	TypeGRPC = "grpc"
)

// Status constants
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
)

// ServiceEndpoint represents an endpoint for a service version
type ServiceEndpoint struct {
	ID               string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	ServiceVersionID string    `gorm:"type:varchar(36);not null;index:idx_service_endpoints_version" json:"service_version_id"`
	URL              string    `gorm:"type:varchar(2048);not null" json:"url"`
	Type             string    `gorm:"type:varchar(20);not null;default:'http'" json:"type"`
	Status           string    `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TableName returns the table name for ServiceEndpoint
func (ServiceEndpoint) TableName() string {
	return "service_endpoints"
}

// CreateServiceEndpointRequest is the request body for creating an endpoint
type CreateServiceEndpointRequest struct {
	URL  string `json:"url" binding:"required"`
	Type string `json:"type"`
}

// UpdateServiceEndpointRequest is the request body for updating an endpoint
type UpdateServiceEndpointRequest struct {
	URL    *string `json:"url"`
	Type   *string `json:"type"`
	Status *string `json:"status"`
}

// ServiceEndpointResponse is the response body for a single endpoint
type ServiceEndpointResponse struct {
	ID               string `json:"id"`
	ServiceVersionID string `json:"service_version_id"`
	URL              string `json:"url"`
	Type             string `json:"type"`
	Status           string `json:"status"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

// ServiceEndpointListResponse is the response body for listing endpoints
type ServiceEndpointListResponse struct {
	Endpoints []ServiceEndpointResponse `json:"endpoints"`
	Total     int64                     `json:"total"`
}

// ToServiceEndpointResponse converts a ServiceEndpoint to ServiceEndpointResponse
func ToServiceEndpointResponse(e *ServiceEndpoint) *ServiceEndpointResponse {
	return &ServiceEndpointResponse{
		ID:               e.ID,
		ServiceVersionID: e.ServiceVersionID,
		URL:              e.URL,
		Type:             e.Type,
		Status:           e.Status,
		CreatedAt:        e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        e.UpdatedAt.Format(time.RFC3339),
	}
}

// ToServiceEndpointListResponse converts a slice of ServiceEndpoint to ServiceEndpointListResponse
func ToServiceEndpointListResponse(endpoints []ServiceEndpoint) *ServiceEndpointListResponse {
	responses := make([]ServiceEndpointResponse, len(endpoints))
	for i, e := range endpoints {
		responses[i] = *ToServiceEndpointResponse(&e)
	}
	return &ServiceEndpointListResponse{
		Endpoints: responses,
		Total:     int64(len(endpoints)),
	}
}

// ValidType checks if the type is valid
func ValidType(t string) bool {
	return t == TypeHTTP || t == TypeGRPC
}

// ValidStatus checks if the status is valid
func ValidStatus(s string) bool {
	return s == StatusActive || s == StatusInactive
}
