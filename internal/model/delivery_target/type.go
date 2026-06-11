package delivery_target

import (
	"encoding/json"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	AvailabilityAvailable   = "available"
	AvailabilityMaintenance = "maintenance"
	AvailabilityDisabled    = "disabled"

	HealthHealthy   = "healthy"
	HealthDegraded  = "degraded"
	HealthUnhealthy = "unhealthy"
	HealthUnknown   = "unknown"
)

type DeliveryTarget struct {
	ID                  string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name                string         `gorm:"type:varchar(255);not null" json:"name"`
	Slug                string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"slug"`
	DisplayName         string         `gorm:"type:varchar(255)" json:"display_name,omitempty"`
	Description         string         `gorm:"type:text" json:"description,omitempty"`
	Purpose             string         `gorm:"type:varchar(50)" json:"purpose,omitempty"`
	TeamID              string         `gorm:"type:varchar(36);index" json:"team_id,omitempty"`
	ClusterName         string         `gorm:"type:varchar(255);not null" json:"cluster_name"`
	ClusterServer       string         `gorm:"type:varchar(512)" json:"cluster_server,omitempty"`
	ControlPlaneName    string         `gorm:"type:varchar(255)" json:"control_plane_name,omitempty"`
	ControlPlaneType    string         `gorm:"type:varchar(50)" json:"control_plane_type,omitempty"`
	KubeconfigContext   string         `gorm:"type:varchar(255)" json:"kubeconfig_context,omitempty"`
	ArgoCDNamespace     string         `gorm:"type:varchar(255)" json:"argocd_namespace,omitempty"`
	ArgoCDServer        string         `gorm:"type:varchar(512)" json:"argocd_server,omitempty"`
	CredentialReference string         `gorm:"type:varchar(255)" json:"credential_reference,omitempty"`
	AvailabilityState   string         `gorm:"type:varchar(20);not null;default:'available'" json:"availability_state"`
	HealthState         string         `gorm:"type:varchar(20);not null;default:'unknown'" json:"health_state"`
	CapacitySummary     string         `gorm:"type:text" json:"capacity_summary,omitempty"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DeliveryTarget) TableName() string {
	return "delivery_targets"
}

func ValidAvailabilityState(value string) bool {
	switch value {
	case AvailabilityAvailable, AvailabilityMaintenance, AvailabilityDisabled:
		return true
	default:
		return false
	}
}

func ValidHealthState(value string) bool {
	switch value {
	case HealthHealthy, HealthDegraded, HealthUnhealthy, HealthUnknown:
		return true
	default:
		return false
	}
}

type CapacitySummary map[string]string

type TargetControlPlaneDefaults struct {
	InCluster           bool
	KubeconfigPath      string
	KubeconfigContext   string
	ControlPlaneName    string
	ControlPlaneType    string
	ArgoCDNamespace     string
	ArgoCDServer        string
	CredentialReference string
}

type TargetControlPlane struct {
	DeliveryTargetID        string `json:"delivery_target_id,omitempty"`
	ClusterServer           string `json:"cluster_server,omitempty"`
	ControlPlaneName        string `json:"control_plane_name,omitempty"`
	ControlPlaneType        string `json:"control_plane_type,omitempty"`
	KubeconfigContext       string `json:"kubeconfig_context,omitempty"`
	KubeconfigPath          string `json:"-"`
	InCluster               bool   `json:"-"`
	ArgoCDNamespace         string `json:"argocd_namespace,omitempty"`
	ArgoCDServer            string `json:"argocd_server,omitempty"`
	CredentialReference     string `json:"credential_reference,omitempty"`
	UsesDefaultControlPlane bool   `json:"uses_default_control_plane"`
}

func (t *DeliveryTarget) ControlPlane() *TargetControlPlane {
	if t == nil {
		return nil
	}

	return &TargetControlPlane{
		DeliveryTargetID:    t.ID,
		ClusterServer:       t.ClusterServer,
		ControlPlaneName:    t.ControlPlaneName,
		ControlPlaneType:    t.ControlPlaneType,
		KubeconfigContext:   t.KubeconfigContext,
		ArgoCDNamespace:     t.ArgoCDNamespace,
		ArgoCDServer:        t.ArgoCDServer,
		CredentialReference: t.CredentialReference,
	}
}

func (cp *TargetControlPlane) HasExplicitConfiguration() bool {
	if cp == nil {
		return false
	}

	return cp.ControlPlaneName != "" ||
		cp.ControlPlaneType != "" ||
		cp.KubeconfigContext != "" ||
		cp.ArgoCDNamespace != "" ||
		cp.ArgoCDServer != "" ||
		cp.CredentialReference != ""
}

func (cp *TargetControlPlane) Resolve(defaults TargetControlPlaneDefaults) *TargetControlPlane {
	resolved := &TargetControlPlane{
		InCluster:      defaults.InCluster,
		KubeconfigPath: defaults.KubeconfigPath,
	}

	if cp != nil {
		resolved.DeliveryTargetID = cp.DeliveryTargetID
		resolved.ClusterServer = cp.ClusterServer
	}

	if cp == nil || !cp.HasExplicitConfiguration() {
		resolved.ControlPlaneName = defaults.ControlPlaneName
		resolved.ControlPlaneType = defaults.ControlPlaneType
		resolved.KubeconfigContext = defaults.KubeconfigContext
		resolved.ArgoCDNamespace = defaults.ArgoCDNamespace
		resolved.ArgoCDServer = defaults.ArgoCDServer
		resolved.CredentialReference = defaults.CredentialReference
		resolved.UsesDefaultControlPlane = true
		return resolved
	}

	resolved.ControlPlaneName = cp.ControlPlaneName
	resolved.ControlPlaneType = cp.ControlPlaneType
	resolved.KubeconfigContext = cp.KubeconfigContext
	resolved.ArgoCDNamespace = cp.ArgoCDNamespace
	resolved.ArgoCDServer = cp.ArgoCDServer
	resolved.CredentialReference = cp.CredentialReference
	return resolved
}

func (cp *TargetControlPlane) CacheKey() string {
	if cp == nil {
		return ""
	}

	parts := []string{
		cp.DeliveryTargetID,
		cp.ControlPlaneName,
		cp.ControlPlaneType,
		cp.KubeconfigContext,
		cp.KubeconfigPath,
		cp.ArgoCDNamespace,
		cp.ArgoCDServer,
		cp.CredentialReference,
		cp.ClusterServer,
	}

	return strings.Join(parts, "|")
}

type CreateDeliveryTargetRequest struct {
	Name                string          `json:"name" binding:"required"`
	DisplayName         string          `json:"display_name"`
	Description         string          `json:"description"`
	Purpose             string          `json:"purpose"`
	TeamID              string          `json:"team_id"`
	ClusterName         string          `json:"cluster_name" binding:"required"`
	ClusterServer       string          `json:"cluster_server"`
	ControlPlaneName    string          `json:"control_plane_name"`
	ControlPlaneType    string          `json:"control_plane_type"`
	KubeconfigContext   string          `json:"kubeconfig_context"`
	ArgoCDNamespace     string          `json:"argocd_namespace"`
	ArgoCDServer        string          `json:"argocd_server"`
	CredentialReference string          `json:"credential_reference"`
	AvailabilityState   string          `json:"availability_state"`
	HealthState         string          `json:"health_state"`
	CapacitySummary     CapacitySummary `json:"capacity_summary"`
}

type UpdateDeliveryTargetRequest struct {
	Name                *string          `json:"name"`
	DisplayName         *string          `json:"display_name"`
	Description         *string          `json:"description"`
	Purpose             *string          `json:"purpose"`
	TeamID              *string          `json:"team_id"`
	ClusterName         *string          `json:"cluster_name"`
	ClusterServer       *string          `json:"cluster_server"`
	ControlPlaneName    *string          `json:"control_plane_name"`
	ControlPlaneType    *string          `json:"control_plane_type"`
	KubeconfigContext   *string          `json:"kubeconfig_context"`
	ArgoCDNamespace     *string          `json:"argocd_namespace"`
	ArgoCDServer        *string          `json:"argocd_server"`
	CredentialReference *string          `json:"credential_reference"`
	AvailabilityState   *string          `json:"availability_state"`
	HealthState         *string          `json:"health_state"`
	CapacitySummary     *CapacitySummary `json:"capacity_summary"`
}

type ListDeliveryTargetsRequest struct {
	TeamID            string `form:"team_id"`
	Purpose           string `form:"purpose"`
	AvailabilityState string `form:"availability_state"`
	HealthState       string `form:"health_state"`
	Search            string `form:"search"`
	Limit             int    `form:"limit"`
	Offset            int    `form:"offset"`
}

type DeliveryTargetResponse struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	Slug                string          `json:"slug"`
	DisplayName         string          `json:"display_name,omitempty"`
	Description         string          `json:"description,omitempty"`
	Purpose             string          `json:"purpose,omitempty"`
	TeamID              string          `json:"team_id,omitempty"`
	ClusterName         string          `json:"cluster_name"`
	ClusterServer       string          `json:"cluster_server,omitempty"`
	ControlPlaneName    string          `json:"control_plane_name,omitempty"`
	ControlPlaneType    string          `json:"control_plane_type,omitempty"`
	KubeconfigContext   string          `json:"kubeconfig_context,omitempty"`
	ArgoCDNamespace     string          `json:"argocd_namespace,omitempty"`
	ArgoCDServer        string          `json:"argocd_server,omitempty"`
	CredentialReference string          `json:"credential_reference,omitempty"`
	AvailabilityState   string          `json:"availability_state"`
	HealthState         string          `json:"health_state"`
	CapacitySummary     CapacitySummary `json:"capacity_summary,omitempty"`
	CreatedAt           string          `json:"created_at"`
	UpdatedAt           string          `json:"updated_at"`
}

type DeliveryTargetListResponse struct {
	Targets []DeliveryTargetResponse `json:"targets"`
	Total   int64                    `json:"total"`
}

func ToDeliveryTargetResponse(target *DeliveryTarget) *DeliveryTargetResponse {
	return &DeliveryTargetResponse{
		ID:                  target.ID,
		Name:                target.Name,
		Slug:                target.Slug,
		DisplayName:         target.DisplayName,
		Description:         target.Description,
		Purpose:             target.Purpose,
		TeamID:              target.TeamID,
		ClusterName:         target.ClusterName,
		ClusterServer:       target.ClusterServer,
		ControlPlaneName:    target.ControlPlaneName,
		ControlPlaneType:    target.ControlPlaneType,
		KubeconfigContext:   target.KubeconfigContext,
		ArgoCDNamespace:     target.ArgoCDNamespace,
		ArgoCDServer:        target.ArgoCDServer,
		CredentialReference: target.CredentialReference,
		AvailabilityState:   target.AvailabilityState,
		HealthState:         target.HealthState,
		CapacitySummary:     ParseCapacitySummary(target.CapacitySummary),
		CreatedAt:           target.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           target.UpdatedAt.Format(time.RFC3339),
	}
}

func ToDeliveryTargetListResponse(targets []DeliveryTarget, total int64) *DeliveryTargetListResponse {
	responses := make([]DeliveryTargetResponse, len(targets))
	for i, target := range targets {
		responses[i] = *ToDeliveryTargetResponse(&target)
	}

	return &DeliveryTargetListResponse{
		Targets: responses,
		Total:   total,
	}
}

func ParseCapacitySummary(value string) CapacitySummary {
	if value == "" {
		return nil
	}

	var summary CapacitySummary
	if err := json.Unmarshal([]byte(value), &summary); err != nil {
		return nil
	}

	return summary
}

func EncodeCapacitySummary(summary CapacitySummary) (string, error) {
	if len(summary) == 0 {
		return "", nil
	}

	payload, err := json.Marshal(summary)
	if err != nil {
		return "", err
	}

	return string(payload), nil
}
