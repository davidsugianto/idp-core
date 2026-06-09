package template

import (
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	StatusActive     = "active"
	StatusDeprecated = "deprecated"
	StatusArchived   = "archived"
	StatusDraft      = "draft"
	StatusStable     = "stable"

	VisibilityPublic  = "public"
	VisibilityTeam    = "team"
	VisibilityPrivate = "private"
)

var nonSlugCharRegex = regexp.MustCompile(`[^a-z0-9-]`)

type Template struct {
	ID          string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null;index:idx_templates_name" json:"name"`
	Slug        string         `gorm:"type:varchar(255);not null;index:idx_templates_slug" json:"slug"`
	Description string         `gorm:"type:text;not null" json:"description"`
	Category    string         `gorm:"type:varchar(100);index:idx_templates_category" json:"category"`
	Author      string         `gorm:"type:varchar(255);not null" json:"author"`
	AuthorEmail string         `gorm:"type:varchar(255)" json:"author_email"`
	Visibility  string         `gorm:"type:varchar(20);not null;default:'team'" json:"visibility"`
	TeamID      string         `gorm:"type:varchar(36);index:idx_templates_team" json:"team_id"`
	Status      string         `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Template) TableName() string {
	return "templates"
}

type CreateTemplateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Category    string `json:"category"`
	Author      string `json:"author" binding:"required"`
	AuthorEmail string `json:"author_email"`
	Visibility  string `json:"visibility"`
	TeamID      string `json:"team_id"`
}

type UpdateTemplateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Author      *string `json:"author"`
	AuthorEmail *string `json:"author_email"`
	Visibility  *string `json:"visibility"`
	Status      *string `json:"status"`
	TeamID      *string `json:"team_id"`
}

type ListTemplatesRequest struct {
	TeamID     string `form:"team_id"`
	Category   string `form:"category"`
	Visibility string `form:"visibility"`
	Status     string `form:"status"`
	Search     string `form:"search"`
	Limit      int    `form:"limit"`
	Offset     int    `form:"offset"`
}

type TemplateResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Author      string `json:"author"`
	AuthorEmail string `json:"author_email"`
	Visibility  string `json:"visibility"`
	TeamID      string `json:"team_id"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type TemplateListResponse struct {
	Templates []TemplateResponse `json:"templates"`
	Total     int64              `json:"total"`
}

func ToTemplateResponse(t *Template) *TemplateResponse {
	return &TemplateResponse{
		ID:          t.ID,
		Name:        t.Name,
		Slug:        t.Slug,
		Description: t.Description,
		Category:    t.Category,
		Author:      t.Author,
		AuthorEmail: t.AuthorEmail,
		Visibility:  t.Visibility,
		TeamID:      t.TeamID,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
	}
}

func ToTemplateListResponse(templates []Template, total int64) *TemplateListResponse {
	responses := make([]TemplateResponse, len(templates))
	for i, template := range templates {
		responses[i] = *ToTemplateResponse(&template)
	}

	return &TemplateListResponse{
		Templates: responses,
		Total:     total,
	}
}

func ValidStatus(status string) bool {
	return status == StatusActive || status == StatusDeprecated || status == StatusArchived || status == StatusDraft || status == StatusStable
}

func ValidVisibility(visibility string) bool {
	return visibility == VisibilityPublic || visibility == VisibilityTeam || visibility == VisibilityPrivate
}

func GenerateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = nonSlugCharRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return slug
}

type TemplateInstance struct {
	ID            string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TemplateID    string    `gorm:"type:varchar(36);not null;index" json:"template_id"`
	VersionID     string    `gorm:"type:varchar(36);not null;index" json:"version_id"`
	EnvironmentID string    `gorm:"type:varchar(36);not null;index" json:"environment_id"`
	Parameters    string    `gorm:"type:text" json:"parameters"`
	CreatedAt     time.Time `json:"created_at"`
}

func (TemplateInstance) TableName() string {
	return "template_instances"
}

type TemplateParameter struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TemplateID  string    `gorm:"type:varchar(36);not null;index" json:"template_id"`
	VersionID   string    `gorm:"type:varchar(36);not null;index" json:"version_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	DisplayName string    `gorm:"type:varchar(255);not null" json:"display_name"`
	Description string    `gorm:"type:text" json:"description"`
	Type        string    `gorm:"type:varchar(50);not null" json:"type"`
	Default     string    `gorm:"type:text" json:"default"`
	Required    bool      `gorm:"not null;default:false" json:"required"`
	Validation  string    `gorm:"type:text" json:"validation"`
	Order       int       `gorm:"not null;default:0" json:"order"`
	CreatedAt   time.Time `json:"created_at"`
}

func (TemplateParameter) TableName() string {
	return "template_parameters"
}

type TemplateResource struct {
	ID         string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TemplateID string    `gorm:"type:varchar(36);not null;index" json:"template_id"`
	VersionID  string    `gorm:"type:varchar(36);not null;index" json:"version_id"`
	Name       string    `gorm:"type:varchar(255);not null" json:"name"`
	Type       string    `gorm:"type:varchar(50);not null" json:"type"`
	Content    string    `gorm:"type:text;not null" json:"content"`
	Order      int       `gorm:"not null;default:0" json:"order"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (TemplateResource) TableName() string {
	return "template_resources"
}

type TemplateVersion struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TemplateID  string    `gorm:"type:varchar(36);not null;index:idx_template_versions_template" json:"template_id"`
	Version     string    `gorm:"type:varchar(100);not null" json:"version"`
	Description string    `gorm:"type:text" json:"description"`
	Changelog   string    `gorm:"type:text" json:"changelog"`
	IsLatest    bool      `gorm:"not null;default:false" json:"is_latest"`
	IsStable    bool      `gorm:"not null;default:false" json:"is_stable"`
	Status      string    `gorm:"type:varchar(20);not null;default:'draft'" json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (TemplateVersion) TableName() string {
	return "template_versions"
}

type CreateTemplateVersionRequest struct {
	Version     string `json:"version" binding:"required"`
	Description string `json:"description"`
	Changelog   string `json:"changelog"`
	IsLatest    *bool  `json:"is_latest"`
	IsStable    bool   `json:"is_stable"`
	Status      string `json:"status"`
}

type UpdateTemplateVersionRequest struct {
	Description *string `json:"description"`
	Changelog   *string `json:"changelog"`
	IsLatest    *bool   `json:"is_latest"`
	IsStable    *bool   `json:"is_stable"`
	Status      *string `json:"status"`
}

type ListTemplateVersionsRequest struct {
	Status string `form:"status"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

type TemplateVersionResponse struct {
	ID          string `json:"id"`
	TemplateID  string `json:"template_id"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Changelog   string `json:"changelog"`
	IsLatest    bool   `json:"is_latest"`
	IsStable    bool   `json:"is_stable"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type TemplateVersionListResponse struct {
	Versions []TemplateVersionResponse `json:"versions"`
	Total    int64                     `json:"total"`
}

type ReplaceTemplateParametersRequest struct {
	Parameters []TemplateParameter `json:"parameters" binding:"required"`
}

type ReplaceTemplateResourcesRequest struct {
	Resources []TemplateResource `json:"resources" binding:"required"`
}

type ValidateTemplateVersionRequest struct {
	Inputs map[string]any `json:"inputs"`
}

type ValidateTemplateVersionResponse struct {
	Valid  bool     `json:"valid"`
	Errors []string `json:"errors"`
}

func ToTemplateVersionResponse(v *TemplateVersion) *TemplateVersionResponse {
	return &TemplateVersionResponse{
		ID:          v.ID,
		TemplateID:  v.TemplateID,
		Version:     v.Version,
		Description: v.Description,
		Changelog:   v.Changelog,
		IsLatest:    v.IsLatest,
		IsStable:    v.IsStable,
		Status:      v.Status,
		CreatedAt:   v.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   v.UpdatedAt.Format(time.RFC3339),
	}
}

func ToTemplateVersionListResponse(versions []TemplateVersion, total int64) *TemplateVersionListResponse {
	responses := make([]TemplateVersionResponse, len(versions))
	for i, version := range versions {
		responses[i] = *ToTemplateVersionResponse(&version)
	}

	return &TemplateVersionListResponse{
		Versions: responses,
		Total:    total,
	}
}
