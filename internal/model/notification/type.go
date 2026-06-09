package notification

import "time"

const (
	KindEnvironment = "environment"
	KindMovement    = "movement"
	KindTarget      = "target"
	KindTemplate    = "template"

	SeverityInfo    = "info"
	SeverityWarning = "warning"
	SeverityError   = "error"
	SeveritySuccess = "success"
)

type Notification struct {
	ID            string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID        string     `gorm:"type:varchar(36);index" json:"user_id,omitempty"`
	TeamID        string     `gorm:"type:varchar(36);index" json:"team_id,omitempty"`
	EnvironmentID string     `gorm:"type:varchar(36);index" json:"environment_id,omitempty"`
	Kind          string     `gorm:"type:varchar(50);not null" json:"kind"`
	Severity      string     `gorm:"type:varchar(20);not null;default:'info'" json:"severity"`
	Title         string     `gorm:"type:varchar(255);not null" json:"title"`
	Message       string     `gorm:"type:text;not null" json:"message"`
	Payload       string     `gorm:"type:text" json:"payload,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
}

func (Notification) TableName() string {
	return "notifications"
}

func ValidKind(value string) bool {
	switch value {
	case KindEnvironment, KindMovement, KindTarget, KindTemplate:
		return true
	default:
		return false
	}
}

func ValidSeverity(value string) bool {
	switch value {
	case SeverityInfo, SeverityWarning, SeverityError, SeveritySuccess:
		return true
	default:
		return false
	}
}
