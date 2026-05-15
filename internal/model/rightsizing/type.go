package rightsizing

import (
	"encoding/json"
	"time"
)

// Workload type constants
const (
	WorkloadTypeDeployment  = "Deployment"
	WorkloadTypeStatefulSet = "StatefulSet"
)

// Recommendation type constants
const (
	RecommendationTypeScaleDown = "scale_down"
	RecommendationTypeScaleUp   = "scale_up"
	RecommendationTypeOptimal   = "optimal"
)

// Status constants
const (
	StatusPending   = "pending"
	StatusApplied   = "applied"
	StatusDismissed = "dismissed"
	StatusFailed    = "failed"
)

// RightsizingRecommendation represents a resource rightsizing recommendation
type RightsizingRecommendation struct {
	ID            string `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Namespace     string `gorm:"type:varchar(255);not null;index:idx_rightsizing_namespace" json:"namespace"`
	WorkloadName  string `gorm:"type:varchar(255);not null;index:idx_rightsizing_workload" json:"workload_name"`
	WorkloadType  string `gorm:"type:varchar(20);not null" json:"workload_type"`
	ContainerName string `gorm:"type:varchar(255);not null" json:"container_name"`

	// Current resources
	CurrentCPURequest    string `gorm:"type:varchar(50)" json:"current_cpu_request"`
	CurrentCPULimit      string `gorm:"type:varchar(50)" json:"current_cpu_limit"`
	CurrentMemoryRequest string `gorm:"type:varchar(50)" json:"current_memory_request"`
	CurrentMemoryLimit   string `gorm:"type:varchar(50)" json:"current_memory_limit"`

	// Recommended resources
	RecommendedCPURequest    string `gorm:"type:varchar(50)" json:"recommended_cpu_request"`
	RecommendedCPULimit      string `gorm:"type:varchar(50)" json:"recommended_cpu_limit"`
	RecommendedMemoryRequest string `gorm:"type:varchar(50)" json:"recommended_memory_request"`
	RecommendedMemoryLimit   string `gorm:"type:varchar(50)" json:"recommended_memory_limit"`

	// Usage metrics
	CPUUsageAvg    string `gorm:"type:varchar(50)" json:"cpu_usage_avg"`
	CPUUsageMax    string `gorm:"type:varchar(50)" json:"cpu_usage_max"`
	MemoryUsageAvg string `gorm:"type:varchar(50)" json:"memory_usage_avg"`
	MemoryUsageMax string `gorm:"type:varchar(50)" json:"memory_usage_max"`

	// Recommendation details
	RecommendationType string  `gorm:"type:varchar(20);not null" json:"recommendation_type"`
	SavingsPotential   float64 `gorm:"type:numeric(12,4);default:0" json:"savings_potential"`
	ConfidenceScore    float64 `gorm:"type:numeric(5,2);default:0" json:"confidence_score"`

	// Status
	Status      string     `gorm:"type:varchar(20);not null;default:pending" json:"status"`
	AppliedAt   *time.Time `json:"applied_at"`
	AppliedBy   string     `gorm:"type:varchar(36)" json:"applied_by"`
	PreviousState string    `gorm:"type:jsonb" json:"previous_state"`

	// Analysis period
	AnalysisPeriodStart time.Time `gorm:"not null" json:"analysis_period_start"`
	AnalysisPeriodEnd   time.Time `gorm:"not null" json:"analysis_period_end"`

	// Metadata
	TeamID        string `gorm:"type:varchar(36)" json:"team_id"`
	EnvironmentID string `gorm:"type:varchar(36)" json:"environment_id"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (RightsizingRecommendation) TableName() string {
	return "rightsizing_recommendations"
}

// PreviousResourceState for rollback
type PreviousResourceState struct {
	CPURequest    string `json:"cpu_request"`
	CPULimit      string `json:"cpu_limit"`
	MemoryRequest string `json:"memory_request"`
	MemoryLimit   string `json:"memory_limit"`
}

func (r *RightsizingRecommendation) GetPreviousState() (*PreviousResourceState, error) {
	if r.PreviousState == "" {
		return nil, nil
	}
	var state PreviousResourceState
	err := json.Unmarshal([]byte(r.PreviousState), &state)
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *RightsizingRecommendation) SetPreviousState(state *PreviousResourceState) error {
	if state == nil {
		r.PreviousState = ""
		return nil
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	r.PreviousState = string(data)
	return nil
}

// Request/Response DTOs

type ListRecommendationsRequest struct {
	Namespace     string `form:"namespace"`
	Status        string `form:"status"`
	TeamID        string `form:"team_id"`
	WorkloadType  string `form:"workload_type"`
	RecommendationType string `form:"recommendation_type"`
	Limit         int    `form:"limit"`
	Offset        int    `form:"offset"`
}

type ApplyRecommendationRequest struct {
	// Empty - ID from path param
}

type DismissRecommendationRequest struct {
	Reason string `json:"reason"` // Optional dismissal reason
}

type RecommendationResponse struct {
	ID                       string  `json:"id"`
	Namespace                string  `json:"namespace"`
	WorkloadName             string  `json:"workload_name"`
	WorkloadType             string  `json:"workload_type"`
	ContainerName            string  `json:"container_name"`
	CurrentCPURequest        string  `json:"current_cpu_request"`
	CurrentCPULimit          string  `json:"current_cpu_limit"`
	CurrentMemoryRequest     string  `json:"current_memory_request"`
	CurrentMemoryLimit       string  `json:"current_memory_limit"`
	RecommendedCPURequest    string  `json:"recommended_cpu_request"`
	RecommendedCPULimit      string  `json:"recommended_cpu_limit"`
	RecommendedMemoryRequest string  `json:"recommended_memory_request"`
	RecommendedMemoryLimit   string  `json:"recommended_memory_limit"`
	CPUUsageAvg              string  `json:"cpu_usage_avg"`
	CPUUsageMax              string  `json:"cpu_usage_max"`
	MemoryUsageAvg           string  `json:"memory_usage_avg"`
	MemoryUsageMax           string  `json:"memory_usage_max"`
	RecommendationType       string  `json:"recommendation_type"`
	SavingsPotential         float64 `json:"savings_potential"`
	ConfidenceScore          float64 `json:"confidence_score"`
	Status                   string  `json:"status"`
	AppliedAt                string  `json:"applied_at"`
	AppliedBy                string  `json:"applied_by"`
	AnalysisPeriodStart      string  `json:"analysis_period_start"`
	AnalysisPeriodEnd        string  `json:"analysis_period_end"`
	TeamID                   string  `json:"team_id"`
	EnvironmentID            string  `json:"environment_id"`
	CreatedAt                string  `json:"created_at"`
	UpdatedAt                string  `json:"updated_at"`
}

type RecommendationListResponse struct {
	Recommendations []RecommendationResponse `json:"recommendations"`
	Total           int64                    `json:"total"`
}

// Helper functions

func ToRecommendationResponse(r *RightsizingRecommendation) *RecommendationResponse {
	resp := &RecommendationResponse{
		ID:                       r.ID,
		Namespace:                r.Namespace,
		WorkloadName:             r.WorkloadName,
		WorkloadType:             r.WorkloadType,
		ContainerName:            r.ContainerName,
		CurrentCPURequest:        r.CurrentCPURequest,
		CurrentCPULimit:          r.CurrentCPULimit,
		CurrentMemoryRequest:     r.CurrentMemoryRequest,
		CurrentMemoryLimit:       r.CurrentMemoryLimit,
		RecommendedCPURequest:    r.RecommendedCPURequest,
		RecommendedCPULimit:      r.RecommendedCPULimit,
		RecommendedMemoryRequest: r.RecommendedMemoryRequest,
		RecommendedMemoryLimit:   r.RecommendedMemoryLimit,
		CPUUsageAvg:              r.CPUUsageAvg,
		CPUUsageMax:              r.CPUUsageMax,
		MemoryUsageAvg:           r.MemoryUsageAvg,
		MemoryUsageMax:           r.MemoryUsageMax,
		RecommendationType:       r.RecommendationType,
		SavingsPotential:         r.SavingsPotential,
		ConfidenceScore:          r.ConfidenceScore,
		Status:                   r.Status,
		AppliedBy:                r.AppliedBy,
		TeamID:                   r.TeamID,
		EnvironmentID:            r.EnvironmentID,
		AnalysisPeriodStart:      r.AnalysisPeriodStart.Format(time.RFC3339),
		AnalysisPeriodEnd:        r.AnalysisPeriodEnd.Format(time.RFC3339),
		CreatedAt:                r.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                r.UpdatedAt.Format(time.RFC3339),
	}
	if r.AppliedAt != nil {
		resp.AppliedAt = r.AppliedAt.Format(time.RFC3339)
	}
	return resp
}

func ToRecommendationListResponse(recs []RightsizingRecommendation, total int64) *RecommendationListResponse {
	responses := make([]RecommendationResponse, len(recs))
	for i, r := range recs {
		responses[i] = *ToRecommendationResponse(&r)
	}
	return &RecommendationListResponse{
		Recommendations: responses,
		Total:           total,
	}
}

// ValidWorkloadType checks if the workload type is valid
func ValidWorkloadType(wt string) bool {
	return wt == WorkloadTypeDeployment || wt == WorkloadTypeStatefulSet
}

// ValidRecommendationType checks if the recommendation type is valid
func ValidRecommendationType(rt string) bool {
	return rt == RecommendationTypeScaleDown || rt == RecommendationTypeScaleUp || rt == RecommendationTypeOptimal
}

// ValidStatus checks if the status is valid
func ValidStatus(s string) bool {
	return s == StatusPending || s == StatusApplied || s == StatusDismissed || s == StatusFailed
}
