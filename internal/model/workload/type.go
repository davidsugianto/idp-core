package workload

import (
	"time"

	"gorm.io/gorm"
)

// WorkloadStatus represents cached workload status for an environment
type WorkloadStatus struct {
	ID            string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	EnvironmentID string         `gorm:"index;not null;type:varchar(36)" json:"environment_id"`
	Namespace     string         `gorm:"index;not null;type:varchar(63)" json:"namespace"`

	// Workload identification
	Name        string `gorm:"not null;type:varchar(255)" json:"name"`
	Kind        string `gorm:"not null;type:varchar(32)" json:"kind"` // Deployment, StatefulSet, DaemonSet
	APIVersion  string `gorm:"type:varchar(64)" json:"api_version"`

	// Replicas
	DesiredReplicas   int `gorm:"default:0" json:"desired_replicas"`
	CurrentReplicas   int `gorm:"default:0" json:"current_replicas"`
	ReadyReplicas     int `gorm:"default:0" json:"ready_replicas"`
	UpdatedReplicas   int `gorm:"default:0" json:"updated_replicas"`
	AvailableReplicas int `gorm:"default:0" json:"available_replicas"`

	// Status
	Status       string `gorm:"type:varchar(32)" json:"status"` // Running, Progressing, Degraded, Failed
	StatusReason string `gorm:"type:text" json:"status_reason"`

	// Image info
	Image         string `gorm:"type:varchar(512)" json:"image"`
	LatestImage   string `gorm:"type:varchar(512)" json:"latest_image"`

	// Metadata
	Labels      string `gorm:"type:text" json:"labels"`       // JSON encoded
	Annotations string `gorm:"type:text" json:"annotations"` // JSON encoded

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (WorkloadStatus) TableName() string {
	return "workload_statuses"
}

// PodStatus represents cached pod status
type PodStatus struct {
	ID            string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	EnvironmentID string         `gorm:"index;not null;type:varchar(36)" json:"environment_id"`
	Namespace     string         `gorm:"index;not null;type:varchar(63)" json:"namespace"`

	// Pod identification
	Name        string `gorm:"not null;type:varchar(255)" json:"name"`
	OwnerName   string `gorm:"type:varchar(255)" json:"owner_name"`   // Parent workload name
	OwnerKind   string `gorm:"type:varchar(32)" json:"owner_kind"`   // Deployment, StatefulSet, etc.

	// Status
	Phase       string `gorm:"not null;type:varchar(32)" json:"phase"` // Pending, Running, Succeeded, Failed, Unknown
	PodIP       string `gorm:"type:varchar(64)" json:"pod_ip"`
	NodeName    string `gorm:"type:varchar(255)" json:"node_name"`

	// Conditions
	Ready           bool  `gorm:"default:false" json:"ready"`
	Initialized     bool  `gorm:"default:false" json:"initialized"`
	ContainersReady bool  `gorm:"default:false" json:"containers_ready"`
	Scheduled       bool  `gorm:"default:false" json:"scheduled"`

	// Container info
	ContainerCount int `gorm:"default:0" json:"container_count"`
	InitContainerCount int `gorm:"default:0" json:"init_container_count"`

	// Restart count
	RestartCount int `gorm:"default:0" json:"restart_count"`

	// Resource requests/limits
	CPURequest    string `gorm:"type:varchar(32)" json:"cpu_request"`
	CPULimit      string `gorm:"type:varchar(32)" json:"cpu_limit"`
	MemoryRequest string `gorm:"type:varchar(32)" json:"memory_request"`
	MemoryLimit   string `gorm:"type:varchar(32)" json:"memory_limit"`

	// Timestamps
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	StartedAt     *time.Time     `json:"started_at,omitempty"`
}

func (PodStatus) TableName() string {
	return "pod_statuses"
}

// WorkloadSummary aggregates workload status for an environment
type WorkloadSummary struct {
	TotalWorkloads   int `json:"total_workloads"`
	HealthyWorkloads int `json:"healthy_workloads"`
	DegradedWorkloads int `json:"degraded_workloads"`
	TotalPods        int `json:"total_pods"`
	RunningPods      int `json:"running_pods"`
	PendingPods      int `json:"pending_pods"`
	FailedPods       int `json:"failed_pods"`
}

// WorkloadStatusResponse is the API response for workload status
type WorkloadStatusResponse struct {
	EnvironmentID string            `json:"environment_id"`
	Namespace     string            `json:"namespace"`
	Summary       WorkloadSummary   `json:"summary"`
	Workloads     []WorkloadInfo    `json:"workloads"`
}

// WorkloadInfo contains detailed workload information
type WorkloadInfo struct {
	Name              string `json:"name"`
	Kind              string `json:"kind"`
	Status            string `json:"status"`
	DesiredReplicas   int    `json:"desired_replicas"`
	ReadyReplicas     int    `json:"ready_replicas"`
	Image             string `json:"image"`
}

func ToWorkloadStatusResponse(statuses []WorkloadStatus, pods []PodStatus) *WorkloadStatusResponse {
	if len(statuses) == 0 {
		return &WorkloadStatusResponse{}
	}

	response := &WorkloadStatusResponse{
		EnvironmentID: statuses[0].EnvironmentID,
		Namespace:     statuses[0].Namespace,
		Workloads:     make([]WorkloadInfo, len(statuses)),
	}

	for i, w := range statuses {
		response.Workloads[i] = WorkloadInfo{
			Name:            w.Name,
			Kind:            w.Kind,
			Status:          w.Status,
			DesiredReplicas: w.DesiredReplicas,
			ReadyReplicas:   w.ReadyReplicas,
			Image:           w.Image,
		}

		response.Summary.TotalWorkloads++
		if w.Status == "Running" || w.Status == "Available" {
			response.Summary.HealthyWorkloads++
		} else if w.Status == "Degraded" || w.Status == "Failed" {
			response.Summary.DegradedWorkloads++
		}
	}

	for _, p := range pods {
		response.Summary.TotalPods++
		switch p.Phase {
		case "Running":
			response.Summary.RunningPods++
		case "Pending":
			response.Summary.PendingPods++
		case "Failed":
			response.Summary.FailedPods++
		}
	}

	return response
}
