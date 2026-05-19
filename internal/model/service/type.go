package service

import (
	"time"

	"gorm.io/gorm"
)

// Status constants
const (
	StatusActive     = "active"
	StatusInactive   = "inactive"
	StatusDeprecated = "deprecated"
)

// Visibility constants
const (
	VisibilityPublic  = "public"
	VisibilityTeam    = "team"
	VisibilityPrivate = "private"
)

// Service represents a registered service in the catalog
type Service struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null;index:idx_services_name" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	TeamID      string         `gorm:"type:varchar(36);not null;index:idx_services_team" json:"team_id"`
	Visibility  string         `gorm:"type:varchar(20);not null;default:'team'" json:"visibility"`
	Status      string         `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for Service
func (Service) TableName() string {
	return "services"
}

// CreateServiceRequest is the request body for creating a service
type CreateServiceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	TeamID      string `json:"team_id" binding:"required"`
	Visibility  string `json:"visibility"`
}

// UpdateServiceRequest is the request body for updating a service
type UpdateServiceRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Visibility  *string `json:"visibility"`
	Status      *string `json:"status"`
}

// ListServicesRequest is the query params for listing services
type ListServicesRequest struct {
	TeamID     string `form:"team_id"`
	Visibility string `form:"visibility"`
	Status     string `form:"status"`
	Search     string `form:"search"`
	Limit      int    `form:"limit"`
	Offset     int    `form:"offset"`
}

// ServiceResponse is the response body for a single service
type ServiceResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	TeamID      string `json:"team_id"`
	Visibility  string `json:"visibility"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ServiceListResponse is the response body for listing services
type ServiceListResponse struct {
	Services []ServiceResponse `json:"services"`
	Total    int64             `json:"total"`
}

// ToServiceResponse converts a Service to ServiceResponse
func ToServiceResponse(s *Service) *ServiceResponse {
	return &ServiceResponse{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		TeamID:      s.TeamID,
		Visibility:  s.Visibility,
		Status:      s.Status,
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   s.UpdatedAt.Format(time.RFC3339),
	}
}

// ToServiceListResponse converts a slice of Service to ServiceListResponse
func ToServiceListResponse(services []Service, total int64) *ServiceListResponse {
	responses := make([]ServiceResponse, len(services))
	for i, s := range services {
		responses[i] = *ToServiceResponse(&s)
	}
	return &ServiceListResponse{
		Services: responses,
		Total:    total,
	}
}

// ValidStatus checks if the status is valid
func ValidStatus(s string) bool {
	return s == StatusActive || s == StatusInactive || s == StatusDeprecated
}

// ValidVisibility checks if the visibility is valid
func ValidVisibility(v string) bool {
	return v == VisibilityPublic || v == VisibilityTeam || v == VisibilityPrivate
}
