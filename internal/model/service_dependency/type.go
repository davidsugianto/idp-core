package service_dependency

import (
	"time"

	"gorm.io/gorm"
)

// Dependency type constants
const (
	TypeRuntime = "runtime" // Required at runtime (e.g., database, cache)
	TypeBuild   = "build"   // Required at build time (e.g., base image)
	TypeData    = "data"    // Data dependency (e.g., shared dataset)
	TypeAPI     = "api"     // API dependency (e.g., external service)
)

// ServiceDependency represents a dependency relationship between services
type ServiceDependency struct {
	ID                string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	ServiceID         string         `gorm:"type:varchar(36);not null;index:idx_service_dependencies_service" json:"service_id"`
	DependsOnServiceID string        `gorm:"type:varchar(36);not null;index:idx_service_dependencies_depends_on" json:"depends_on_service_id"`
	DependencyType    string         `gorm:"type:varchar(20);not null;default:'runtime'" json:"dependency_type"`
	Description       string         `gorm:"type:text" json:"description"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for ServiceDependency
func (ServiceDependency) TableName() string {
	return "service_dependencies"
}

// CreateDependencyRequest is the request body for creating a dependency
type CreateDependencyRequest struct {
	DependsOnServiceID string `json:"depends_on_service_id" binding:"required"`
	DependencyType     string `json:"dependency_type"`
	Description        string `json:"description"`
}

// UpdateDependencyRequest is the request body for updating a dependency
type UpdateDependencyRequest struct {
	DependencyType *string `json:"dependency_type"`
	Description    *string `json:"description"`
}

// ListDependenciesRequest is the query params for listing dependencies
type ListDependenciesRequest struct {
	DependencyType string `form:"dependency_type"`
	Limit          int    `form:"limit"`
	Offset         int    `form:"offset"`
}

// DependencyResponse is the response body for a single dependency
type DependencyResponse struct {
	ID                 string `json:"id"`
	ServiceID          string `json:"service_id"`
	DependsOnServiceID string `json:"depends_on_service_id"`
	DependencyType     string `json:"dependency_type"`
	Description        string `json:"description"`
	CreatedAt          string `json:"created_at"`
	UpdatedAt          string `json:"updated_at"`
}

// DependencyListResponse is the response body for listing dependencies
type DependencyListResponse struct {
	Dependencies []DependencyResponse `json:"dependencies"`
	Total        int64                `json:"total"`
}

// DependencyGraphResponse is the response for the dependency graph visualization
type DependencyGraphResponse struct {
	ServiceID   string       `json:"service_id"`
	ServiceName string       `json:"service_name"`
	Nodes       []GraphNode  `json:"nodes"`
	Edges       []GraphEdge  `json:"edges"`
}

// GraphNode represents a node in the dependency graph
type GraphNode struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // "root", "dependency", "dependent"
}

// GraphEdge represents an edge in the dependency graph
type GraphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // dependency type: "runtime", "build", "data", "api"
}

// ToDependencyResponse converts a ServiceDependency to DependencyResponse
func ToDependencyResponse(d *ServiceDependency) *DependencyResponse {
	return &DependencyResponse{
		ID:                 d.ID,
		ServiceID:          d.ServiceID,
		DependsOnServiceID: d.DependsOnServiceID,
		DependencyType:     d.DependencyType,
		Description:        d.Description,
		CreatedAt:          d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          d.UpdatedAt.Format(time.RFC3339),
	}
}

// ToDependencyListResponse converts a slice of ServiceDependency to DependencyListResponse
func ToDependencyListResponse(deps []ServiceDependency, total int64) *DependencyListResponse {
	responses := make([]DependencyResponse, len(deps))
	for i, d := range deps {
		responses[i] = *ToDependencyResponse(&d)
	}
	return &DependencyListResponse{
		Dependencies: responses,
		Total:        total,
	}
}

// ValidDependencyType checks if the dependency type is valid
func ValidDependencyType(t string) bool {
	return t == TypeRuntime || t == TypeBuild || t == TypeData || t == TypeAPI
}
