package environment_movement

import "time"

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
)

type EnvironmentMovement struct {
	ID                  string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	EnvironmentID       string     `gorm:"type:varchar(36);not null;index" json:"environment_id"`
	SourceTargetID      string     `gorm:"type:varchar(36);index" json:"source_target_id,omitempty"`
	DestinationTargetID string     `gorm:"type:varchar(36);not null;index" json:"destination_target_id"`
	RequestedBy         string     `gorm:"type:varchar(36);index" json:"requested_by,omitempty"`
	Status              string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ProgressPercent     int        `gorm:"not null;default:0" json:"progress_percent"`
	Message             string     `gorm:"type:text" json:"message,omitempty"`
	StartedAt           *time.Time `json:"started_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (EnvironmentMovement) TableName() string {
	return "environment_movements"
}

func ValidStatus(value string) bool {
	switch value {
	case StatusPending, StatusRunning, StatusCompleted, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}
