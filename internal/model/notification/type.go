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

	EventNotification = "notification"
	EventStatus       = "status"
	EventProgress     = "progress"
	EventLog          = "log"
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

type ListNotificationsRequest struct {
	UserID        string `form:"user_id" json:"user_id,omitempty"`
	TeamID        string `form:"team_id" json:"team_id,omitempty"`
	EnvironmentID string `form:"environment_id" json:"environment_id,omitempty"`
	Kind          string `form:"kind" json:"kind,omitempty"`
	Limit         int    `form:"limit" json:"limit,omitempty"`
	Offset        int    `form:"offset" json:"offset,omitempty"`
}

type NotificationListResponse struct {
	Notifications []Notification `json:"notifications"`
	Total         int64          `json:"total"`
}

type NotificationEventPayload struct {
	NotificationID string    `json:"notification_id"`
	EnvironmentID  string    `json:"environment_id,omitempty"`
	Kind           string    `json:"kind"`
	Severity       string    `json:"severity"`
	Title          string    `json:"title"`
	Message        string    `json:"message"`
	CreatedAt      time.Time `json:"created_at"`
}

type StatusEventPayload struct {
	EnvironmentID string    `json:"environment_id"`
	Status        string    `json:"status"`
	ChangedAt     time.Time `json:"changed_at"`
}

type ProgressEventPayload struct {
	EnvironmentID   string    `json:"environment_id"`
	Operation       string    `json:"operation,omitempty"`
	ProgressPercent int       `json:"progress_percent"`
	Message         string    `json:"message,omitempty"`
	ChangedAt       time.Time `json:"changed_at"`
}

type LogEventPayload struct {
	EnvironmentID string    `json:"environment_id,omitempty"`
	Workload      string    `json:"workload"`
	Container     string    `json:"container,omitempty"`
	Line          string    `json:"line"`
	Timestamp     time.Time `json:"timestamp"`
}

type StreamEvent struct {
	ID        string    `json:"id,omitempty"`
	Event     string    `json:"event"`
	Data      any       `json:"data"`
	Timestamp time.Time `json:"timestamp"`
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

func ValidEventType(value string) bool {
	switch value {
	case EventNotification, EventStatus, EventProgress, EventLog:
		return true
	default:
		return false
	}
}

func ToNotificationEventPayload(notification *Notification) *NotificationEventPayload {
	if notification == nil {
		return nil
	}

	return &NotificationEventPayload{
		NotificationID: notification.ID,
		EnvironmentID:  notification.EnvironmentID,
		Kind:           notification.Kind,
		Severity:       notification.Severity,
		Title:          notification.Title,
		Message:        notification.Message,
		CreatedAt:      notification.CreatedAt,
	}
}

func ToNotificationListResponse(notifications []Notification, total int64) *NotificationListResponse {
	return &NotificationListResponse{
		Notifications: notifications,
		Total:         total,
	}
}
