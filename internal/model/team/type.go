package team

import (
	"time"

	"gorm.io/gorm"
)

// Team represents a group of users
type Team struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null;type:varchar(255)" json:"name"`
	Slug        string         `gorm:"uniqueIndex;not null;type:varchar(63)" json:"slug"`
	Description string         `gorm:"type:text" json:"description,omitempty"`

	// Settings (JSON encoded)
	Settings string `gorm:"type:text" json:"settings,omitempty"`

	// Status
	Status string `gorm:"not null;default:'active';type:varchar(20)" json:"status"` // active, disabled

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Team) TableName() string {
	return "teams"
}

// TeamMember represents user-team membership
type TeamMember struct {
	ID        string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TeamID    string    `gorm:"index;not null;type:varchar(36)" json:"team_id"`
	UserID    string    `gorm:"index;not null;type:varchar(36)" json:"user_id"`
	Role      string    `gorm:"not null;default:'member';type:varchar(20)" json:"role"` // owner, admin, member
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (TeamMember) TableName() string {
	return "team_members"
}

// TeamMemberRole constants
const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

// CreateTeamRequest is the request body for creating a team
type CreateTeamRequest struct {
	Name        string `json:"name" binding:"required"`
	Slug        string `json:"slug" binding:"required"`
	Description string `json:"description"`
}

// UpdateTeamRequest is the request body for updating a team
type UpdateTeamRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"` // active, disabled
}

// AddTeamMemberRequest is the request body for adding a team member
type AddTeamMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required,oneof=owner admin member"`
}

// UpdateTeamMemberRequest is the request body for updating a team member
type UpdateTeamMemberRequest struct {
	Role string `json:"role" binding:"required,oneof=owner admin member"`
}

// TeamResponse is the response for team endpoints
type TeamResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TeamWithMembersResponse includes team details with members
type TeamWithMembersResponse struct {
	TeamResponse
	Members []TeamMemberResponse `json:"members"`
}

// TeamMemberResponse is the response for team member
type TeamMemberResponse struct {
	ID        string          `json:"id"`
	User      *UserBasicInfo `json:"user"`
	Role      string         `json:"role"`
	CreatedAt time.Time      `json:"created_at"`
}

// UserBasicInfo contains basic user info for nested responses
type UserBasicInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// TeamListResponse is the response for listing teams
type TeamListResponse struct {
	Teams []TeamResponse `json:"teams"`
	Total int64          `json:"total"`
}

// ToTeamResponse converts Team to TeamResponse
func ToTeamResponse(t *Team) *TeamResponse {
	return &TeamResponse{
		ID:          t.ID,
		Name:        t.Name,
		Slug:        t.Slug,
		Description: t.Description,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// ToTeamListResponse converts slice of Team to TeamListResponse
func ToTeamListResponse(teams []Team, total int64) *TeamListResponse {
	responses := make([]TeamResponse, len(teams))
	for i, t := range teams {
		responses[i] = *ToTeamResponse(&t)
	}
	return &TeamListResponse{
		Teams: responses,
		Total: total,
	}
}
