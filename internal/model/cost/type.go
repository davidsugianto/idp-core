package cost

import (
	"time"
)

// CostRecord represents a cost tracking record stored in PostgreSQL
type CostRecord struct {
	ID            string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TeamID        string    `gorm:"type:varchar(36);index:idx_cost_records_team_period" json:"team_id,omitempty"`
	EnvironmentID string    `gorm:"type:varchar(36)" json:"environment_id,omitempty"`
	Namespace     string    `gorm:"type:varchar(255);not null;index:idx_cost_records_namespace_period" json:"namespace"`
	PeriodStart   time.Time `gorm:"not null;index:idx_cost_records_team_period,priority:2;index:idx_cost_records_namespace_period,priority:2;index:idx_cost_records_period_start" json:"period_start"`
	PeriodEnd     time.Time `gorm:"not null" json:"period_end"`
	CPUCost       float64   `gorm:"type:numeric(12,4);not null;default:0" json:"cpu_cost"`
	RAMCost       float64   `gorm:"type:numeric(12,4);not null;default:0" json:"ram_cost"`
	PVCost        float64   `gorm:"type:numeric(12,4);not null;default:0" json:"pv_cost"`
	NetworkCost   float64   `gorm:"type:numeric(12,4);not null;default:0" json:"network_cost"`
	TotalCost     float64   `gorm:"type:numeric(12,4);not null;default:0" json:"total_cost"`
	RawData       string    `gorm:"type:jsonb" json:"raw_data,omitempty"`
	CreatedAt     time.Time `gorm:"autoCreateTime;not null" json:"created_at"`
}

// TableName returns the table name for CostRecord
func (CostRecord) TableName() string {
	return "cost_records"
}

// CostRecordResponse represents a cost record in API responses
type CostRecordResponse struct {
	ID            string  `json:"id"`
	TeamID        string  `json:"team_id,omitempty"`
	EnvironmentID string  `json:"environment_id,omitempty"`
	Namespace     string  `json:"namespace"`
	PeriodStart   string  `json:"period_start"`
	PeriodEnd     string  `json:"period_end"`
	CPUCost       float64 `json:"cpu_cost"`
	RAMCost       float64 `json:"ram_cost"`
	PVCost        float64 `json:"pv_cost"`
	NetworkCost   float64 `json:"network_cost"`
	TotalCost     float64 `json:"total_cost"`
	CreatedAt     string  `json:"created_at"`
}

// CostListResponse represents a list of cost records in API responses
type CostListResponse struct {
	CostRecords []CostRecordResponse `json:"cost_records"`
	Total       int64                `json:"total"`
}

// CostFilter represents filter options for listing cost records
type CostFilter struct {
	TeamID        string
	EnvironmentID string
	Namespace     string
	StartDate     *time.Time
	EndDate       *time.Time
	Limit         int
	Offset        int
}

// ToCostRecordResponse converts CostRecord to CostRecordResponse
func ToCostRecordResponse(record *CostRecord) *CostRecordResponse {
	if record == nil {
		return nil
	}
	return &CostRecordResponse{
		ID:            record.ID,
		TeamID:        record.TeamID,
		EnvironmentID: record.EnvironmentID,
		Namespace:     record.Namespace,
		PeriodStart:   record.PeriodStart.Format(time.RFC3339),
		PeriodEnd:     record.PeriodEnd.Format(time.RFC3339),
		CPUCost:       record.CPUCost,
		RAMCost:       record.RAMCost,
		PVCost:        record.PVCost,
		NetworkCost:   record.NetworkCost,
		TotalCost:     record.TotalCost,
		CreatedAt:     record.CreatedAt.Format(time.RFC3339),
	}
}

// ToCostListResponse converts a slice of CostRecord to CostListResponse
func ToCostListResponse(records []CostRecord, total int64) *CostListResponse {
	responses := make([]CostRecordResponse, len(records))
	for i, r := range records {
		responses[i] = *ToCostRecordResponse(&r)
	}
	return &CostListResponse{
		CostRecords: responses,
		Total:       total,
	}
}