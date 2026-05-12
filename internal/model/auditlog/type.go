package auditlog

import (
	"time"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           string     `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID       string     `gorm:"type:varchar(36);index" json:"user_id,omitempty"`
	UserEmail    string     `gorm:"type:varchar(255)" json:"user_email,omitempty"`
	ActorType    string     `gorm:"type:varchar(20);not null" json:"actor_type"`
	Action       string     `gorm:"type:varchar(100);not null;index" json:"action"`
	ResourceType string     `gorm:"type:varchar(50);not null;index" json:"resource_type"`
	ResourceID   string     `gorm:"type:varchar(36)" json:"resource_id,omitempty"`
	TeamID       string     `gorm:"type:varchar(36);index" json:"team_id,omitempty"`
	EnvironmentID string    `gorm:"type:varchar(36)" json:"environment_id,omitempty"`
	IPAddress    string     `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent    string     `gorm:"type:text" json:"user_agent,omitempty"`
	RequestMethod string    `gorm:"type:varchar(10)" json:"request_method,omitempty"`
	RequestPath  string     `gorm:"type:text" json:"request_path,omitempty"`
	RequestID    string     `gorm:"type:varchar(36)" json:"request_id,omitempty"`
	OldValues    Map        `gorm:"type:jsonb" json:"old_values,omitempty"`
	NewValues    Map        `gorm:"type:jsonb" json:"new_values,omitempty"`
	Status       string     `gorm:"type:varchar(20);not null;default:'success';index" json:"status"`
	ErrorMessage string     `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt    time.Time  `gorm:"not null;index" json:"created_at"`
}

// Map is a generic map for JSONB fields
type Map map[string]interface{}

// TableName returns the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}

// Actor type constants
const (
	ActorTypeUser   = "user"
	ActorTypeAPIKey = "api_key"
	ActorTypeSystem = "system"
)

// Status constants
const (
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusDenied  = "denied"
)

// Action constants
const (
	ActionCreate    = "create"
	ActionUpdate    = "update"
	ActionDelete    = "delete"
	ActionLogin     = "login"
	ActionLogout    = "logout"
	ActionAssign    = "assign"
	ActionRevoke    = "revoke"
	ActionSync      = "sync"
	ActionExport    = "export"
	ActionRotate    = "rotate"
)

// Resource type constants
const (
	ResourceTypeEnvironment = "environment"
	ResourceTypeTeam        = "team"
	ResourceTypeUser        = "user"
	ResourceTypeRole        = "role"
	ResourceTypeAPIKey      = "api_key"
	ResourceTypeBudget      = "budget"
	ResourceTypeService     = "service"
)

// CreateAuditLogRequest represents the request to create an audit log
type CreateAuditLogRequest struct {
	UserID        string
	UserEmail     string
	ActorType     string
	Action        string
	ResourceType  string
	ResourceID    string
	TeamID        string
	EnvironmentID string
	IPAddress     string
	UserAgent     string
	RequestMethod string
	RequestPath   string
	RequestID     string
	OldValues     Map
	NewValues     Map
	Status        string
	ErrorMessage  string
}

// AuditLogResponse represents an audit log in API responses
type AuditLogResponse struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id,omitempty"`
	UserEmail     string    `json:"user_email,omitempty"`
	ActorType     string    `json:"actor_type"`
	Action        string    `json:"action"`
	ResourceType  string    `json:"resource_type"`
	ResourceID    string    `json:"resource_id,omitempty"`
	TeamID        string    `json:"team_id,omitempty"`
	IPAddress     string    `json:"ip_address,omitempty"`
	RequestMethod string    `json:"request_method,omitempty"`
	RequestPath   string    `json:"request_path,omitempty"`
	Status        string    `json:"status"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	CreatedAt     string    `json:"created_at"`
}

// AuditLogListResponse represents a list of audit logs
type AuditLogListResponse struct {
	AuditLogs []AuditLogResponse `json:"audit_logs"`
	Total     int64              `json:"total"`
}

// AuditLogFilter represents filter options for listing audit logs
type AuditLogFilter struct {
	UserID       string
	TeamID       string
	Action       string
	ResourceType string
	ResourceID   string
	Status       string
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}

// ToAuditLogResponse converts AuditLog to AuditLogResponse
func ToAuditLogResponse(log *AuditLog) *AuditLogResponse {
	if log == nil {
		return nil
	}

	return &AuditLogResponse{
		ID:            log.ID,
		UserID:        log.UserID,
		UserEmail:     log.UserEmail,
		ActorType:     log.ActorType,
		Action:        log.Action,
		ResourceType:  log.ResourceType,
		ResourceID:    log.ResourceID,
		TeamID:        log.TeamID,
		IPAddress:     log.IPAddress,
		RequestMethod: log.RequestMethod,
		RequestPath:   log.RequestPath,
		Status:        log.Status,
		ErrorMessage:  log.ErrorMessage,
		CreatedAt:     log.CreatedAt.Format(time.RFC3339),
	}
}

// ToAuditLogListResponse converts a slice of AuditLog to AuditLogListResponse
func ToAuditLogListResponse(logs []AuditLog, total int64) *AuditLogListResponse {
	responses := make([]AuditLogResponse, len(logs))
	for i, l := range logs {
		responses[i] = *ToAuditLogResponse(&l)
	}
	return &AuditLogListResponse{
		AuditLogs: responses,
		Total:     total,
	}
}
