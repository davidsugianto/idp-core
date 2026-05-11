package user

import (
	"time"

	"gorm.io/gorm"
)

// User represents an authenticated user
type User struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Email     string         `gorm:"uniqueIndex;not null;type:varchar(255)" json:"email"`
	Name      string         `gorm:"not null;type:varchar(255)" json:"name"`

	// Authentication provider
	Provider   string `gorm:"not null;default:'local';type:varchar(50)" json:"provider"` // oidc, local
	ProviderID string `gorm:"type:varchar(255)" json:"provider_id,omitempty"`

	// Profile
	AvatarURL string `gorm:"type:varchar(512)" json:"avatar_url,omitempty"`

	// Status
	Status      string     `gorm:"not null;default:'active';type:varchar(20)" json:"status"` // active, disabled, pending
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// CreateUserRequest is the request body for creating a user
type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Name      string `json:"name" binding:"required"`
	Provider  string `json:"provider"`  // defaults to 'local'
	AvatarURL string `json:"avatar_url"`
}

// UpdateUserRequest is the request body for updating a user
type UpdateUserRequest struct {
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Status    string `json:"status"` // active, disabled
}

// UserResponse is the response for user endpoints
type UserResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
	Provider    string     `json:"provider"`
	AvatarURL   string     `json:"avatar_url,omitempty"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// UserListResponse is the response for listing users
type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total"`
}

// ToUserResponse converts User to UserResponse
func ToUserResponse(u *User) *UserResponse {
	return &UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		Name:        u.Name,
		Provider:    u.Provider,
		AvatarURL:   u.AvatarURL,
		Status:      u.Status,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// ToUserListResponse converts slice of User to UserListResponse
func ToUserListResponse(users []User, total int64) *UserListResponse {
	responses := make([]UserResponse, len(users))
	for i, u := range users {
		responses[i] = *ToUserResponse(&u)
	}
	return &UserListResponse{
		Users: responses,
		Total: total,
	}
}
