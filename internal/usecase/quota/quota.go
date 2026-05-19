package quota

import (
	"context"
	"fmt"
	"time"

	quotaModel "github.com/davidsugianto/idp-core/internal/model/resourcequota"
	quotaRepo "github.com/davidsugianto/idp-core/internal/repository/quota"
	provisionerRepo "github.com/davidsugianto/idp-core/internal/repository/provisioner"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/api/resource"
)

type Usecase interface {
	// Quota management
	CreateQuota(ctx context.Context, req *quotaModel.CreateResourceQuotaRequest) (*quotaModel.ResourceQuotaResponse, error)
	GetQuota(ctx context.Context, id string) (*quotaModel.ResourceQuotaResponse, error)
	GetQuotaByNamespace(ctx context.Context, namespace string) (*quotaModel.ResourceQuotaResponse, error)
	ListQuotas(ctx context.Context, req *quotaModel.ListResourceQuotasRequest) (*quotaModel.ResourceQuotaListResponse, error)
	UpdateQuota(ctx context.Context, id string, req *quotaModel.UpdateResourceQuotaRequest) (*quotaModel.ResourceQuotaResponse, error)
	DeleteQuota(ctx context.Context, id string) error

	// Usage tracking
	GetUsage(ctx context.Context, namespace string) (*quotaModel.UsageResponse, error)
	RefreshUsage(ctx context.Context, namespace string) error
	RefreshAllUsage(ctx context.Context) error

	// Quota enforcement
	CheckQuota(ctx context.Context, req *quotaModel.QuotaCheckRequest) (*quotaModel.QuotaCheckResponse, error)
	IsQuotaExceeded(ctx context.Context, namespace string) (bool, []quotaModel.QuotaExceededReason, error)
}

type usecase struct {
	quotaRepo      quotaRepo.Repository
	provisionerRepo provisionerRepo.Repository
}

type Dependencies struct {
	QuotaRepo       quotaRepo.Repository
	ProvisionerRepo provisionerRepo.Repository
}

func New(deps Dependencies) Usecase {
	return &usecase{
		quotaRepo:       deps.QuotaRepo,
		provisionerRepo: deps.ProvisionerRepo,
	}
}

// CreateQuota creates a new resource quota
func (u *usecase) CreateQuota(ctx context.Context, req *quotaModel.CreateResourceQuotaRequest) (*quotaModel.ResourceQuotaResponse, error) {
	// Check if quota already exists for namespace
	exists, err := u.quotaRepo.ExistsForNamespace(ctx, req.Namespace)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("quota already exists for namespace %s", req.Namespace)
	}

	now := time.Now()
	quota := &quotaModel.ResourceQuota{
		ID:                   uuid.New().String(),
		Namespace:            req.Namespace,
		TeamID:               req.TeamID,
		EnvironmentID:        req.EnvironmentID,
		CPURequestLimit:      req.CPURequestLimit,
		CPULimitLimit:        req.CPULimitLimit,
		MemoryRequestLimit:   req.MemoryRequestLimit,
		MemoryLimitLimit:     req.MemoryLimitLimit,
		StorageRequestLimit:  req.StorageRequestLimit,
		PodCountLimit:        req.PodCountLimit,
		ConfigMapCountLimit:  req.ConfigMapCountLimit,
		SecretCountLimit:     req.SecretCountLimit,
		PVCCountLimit:        req.PVCCountLimit,
		Enforce:              req.Enforce,
		GracePeriodHours:     req.GracePeriodHours,
		Description:          req.Description,
		Status:               quotaModel.StatusActive,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := u.quotaRepo.Create(ctx, quota); err != nil {
		return nil, err
	}

	// Refresh usage to get current state
	_ = u.RefreshUsage(ctx, req.Namespace)

	return quotaModel.ToResourceQuotaResponse(quota), nil
}

// GetQuota returns a quota by ID
func (u *usecase) GetQuota(ctx context.Context, id string) (*quotaModel.ResourceQuotaResponse, error) {
	quota, err := u.quotaRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return quotaModel.ToResourceQuotaResponse(quota), nil
}

// GetQuotaByNamespace returns a quota by namespace
func (u *usecase) GetQuotaByNamespace(ctx context.Context, namespace string) (*quotaModel.ResourceQuotaResponse, error) {
	quota, err := u.quotaRepo.GetByNamespace(ctx, namespace)
	if err != nil {
		return nil, err
	}
	return quotaModel.ToResourceQuotaResponse(quota), nil
}

// ListQuotas returns a paginated list of quotas
func (u *usecase) ListQuotas(ctx context.Context, req *quotaModel.ListResourceQuotasRequest) (*quotaModel.ResourceQuotaListResponse, error) {
	quotas, total, err := u.quotaRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}
	return quotaModel.ToResourceQuotaListResponse(quotas, total), nil
}

// UpdateQuota updates a resource quota
func (u *usecase) UpdateQuota(ctx context.Context, id string, req *quotaModel.UpdateResourceQuotaRequest) (*quotaModel.ResourceQuotaResponse, error) {
	quota, err := u.quotaRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if req.CPURequestLimit != nil {
		quota.CPURequestLimit = *req.CPURequestLimit
	}
	if req.CPULimitLimit != nil {
		quota.CPULimitLimit = *req.CPULimitLimit
	}
	if req.MemoryRequestLimit != nil {
		quota.MemoryRequestLimit = *req.MemoryRequestLimit
	}
	if req.MemoryLimitLimit != nil {
		quota.MemoryLimitLimit = *req.MemoryLimitLimit
	}
	if req.StorageRequestLimit != nil {
		quota.StorageRequestLimit = *req.StorageRequestLimit
	}
	if req.PodCountLimit != nil {
		quota.PodCountLimit = req.PodCountLimit
	}
	if req.ConfigMapCountLimit != nil {
		quota.ConfigMapCountLimit = req.ConfigMapCountLimit
	}
	if req.SecretCountLimit != nil {
		quota.SecretCountLimit = req.SecretCountLimit
	}
	if req.PVCCountLimit != nil {
		quota.PVCCountLimit = req.PVCCountLimit
	}
	if req.Enforce != nil {
		quota.Enforce = *req.Enforce
	}
	if req.GracePeriodHours != nil {
		quota.GracePeriodHours = req.GracePeriodHours
	}
	if req.Description != nil {
		quota.Description = *req.Description
	}

	quota.UpdatedAt = time.Now()

	if err := u.quotaRepo.Update(ctx, quota); err != nil {
		return nil, err
	}

	return quotaModel.ToResourceQuotaResponse(quota), nil
}

// DeleteQuota deletes a resource quota
func (u *usecase) DeleteQuota(ctx context.Context, id string) error {
	return u.quotaRepo.Delete(ctx, id)
}

// GetUsage returns current resource usage for a namespace
func (u *usecase) GetUsage(ctx context.Context, namespace string) (*quotaModel.UsageResponse, error) {
	usage, err := u.calculateUsage(ctx, namespace)
	if err != nil {
		return nil, err
	}
	return usage, nil
}

// RefreshUsage updates cached usage for a namespace
func (u *usecase) RefreshUsage(ctx context.Context, namespace string) error {
	usage, err := u.calculateUsage(ctx, namespace)
	if err != nil {
		return err
	}
	return u.quotaRepo.UpdateUsage(ctx, namespace, usage)
}

// RefreshAllUsage updates cached usage for all quotas
func (u *usecase) RefreshAllUsage(ctx context.Context) error {
	quotas, _, err := u.quotaRepo.List(ctx, &quotaModel.ListResourceQuotasRequest{})
	if err != nil {
		return err
	}

	for _, quota := range quotas {
		_ = u.RefreshUsage(ctx, quota.Namespace)
	}

	return nil
}

// CheckQuota checks if a resource request would exceed quota limits
func (u *usecase) CheckQuota(ctx context.Context, req *quotaModel.QuotaCheckRequest) (*quotaModel.QuotaCheckResponse, error) {
	quota, err := u.quotaRepo.GetActiveByNamespace(ctx, req.Namespace)
	if err != nil {
		// No quota defined - allow
		return &quotaModel.QuotaCheckResponse{Allowed: true}, nil
	}

	if !quota.Enforce {
		return &quotaModel.QuotaCheckResponse{Allowed: true}, nil
	}

	var reasons []quotaModel.QuotaExceededReason

	// Check CPU request
	if req.CPURequest != "" && quota.CPURequestLimit != "" {
		if exceeded, reason := u.checkResourceLimit(
			"cpu_request",
			req.CPURequest,
			quota.CPURequestLimit,
			quota.CurrentCPURequest,
		); exceeded {
			reasons = append(reasons, reason)
		}
	}

	// Check memory request
	if req.MemoryRequest != "" && quota.MemoryRequestLimit != "" {
		if exceeded, reason := u.checkResourceLimit(
			"memory_request",
			req.MemoryRequest,
			quota.MemoryRequestLimit,
			quota.CurrentMemoryRequest,
		); exceeded {
			reasons = append(reasons, reason)
		}
	}

	// Check pod count
	if req.PodDelta > 0 && quota.PodCountLimit != nil {
		currentPods := 0
		if quota.CurrentPodCount != nil {
			currentPods = *quota.CurrentPodCount
		}
		if currentPods+req.PodDelta > *quota.PodCountLimit {
			reasons = append(reasons, quotaModel.QuotaExceededReason{
				ResourceType: "pods",
				Requested:    fmt.Sprintf("%d", currentPods+req.PodDelta),
				Limit:        fmt.Sprintf("%d", *quota.PodCountLimit),
				Current:      fmt.Sprintf("%d", currentPods),
				Utilization:  float64(currentPods+req.PodDelta) / float64(*quota.PodCountLimit) * 100,
			})
		}
	}

	return &quotaModel.QuotaCheckResponse{
		Allowed: len(reasons) == 0,
		Reasons: reasons,
	}, nil
}

// IsQuotaExceeded checks if a namespace has exceeded its quota
func (u *usecase) IsQuotaExceeded(ctx context.Context, namespace string) (bool, []quotaModel.QuotaExceededReason, error) {
	quota, err := u.quotaRepo.GetActiveByNamespace(ctx, namespace)
	if err != nil {
		return false, nil, nil
	}

	var reasons []quotaModel.QuotaExceededReason

	// Check CPU request
	if quota.CPURequestLimit != "" && quota.CurrentCPURequest != "" {
		if exceeded, reason := u.checkResourceLimit(
			"cpu_request",
			quota.CurrentCPURequest,
			quota.CPURequestLimit,
			quota.CurrentCPURequest,
		); exceeded {
			reasons = append(reasons, reason)
		}
	}

	// Check memory request
	if quota.MemoryRequestLimit != "" && quota.CurrentMemoryRequest != "" {
		if exceeded, reason := u.checkResourceLimit(
			"memory_request",
			quota.CurrentMemoryRequest,
			quota.MemoryRequestLimit,
			quota.CurrentMemoryRequest,
		); exceeded {
			reasons = append(reasons, reason)
		}
	}

	// Check pod count
	if quota.PodCountLimit != nil && quota.CurrentPodCount != nil {
		if *quota.CurrentPodCount > *quota.PodCountLimit {
			reasons = append(reasons, quotaModel.QuotaExceededReason{
				ResourceType: "pods",
				Requested:    fmt.Sprintf("%d", *quota.CurrentPodCount),
				Limit:        fmt.Sprintf("%d", *quota.PodCountLimit),
				Current:      fmt.Sprintf("%d", *quota.CurrentPodCount),
				Utilization:  float64(*quota.CurrentPodCount) / float64(*quota.PodCountLimit) * 100,
			})
		}
	}

	return len(reasons) > 0, reasons, nil
}

// calculateUsage calculates current resource usage for a namespace
func (u *usecase) calculateUsage(ctx context.Context, namespace string) (*quotaModel.UsageResponse, error) {
	pods, err := u.provisionerRepo.GetPods(namespace)
	if err != nil {
		return nil, err
	}

	var totalCPURequest, totalCPULimit int64
	var totalMemRequest, totalMemLimit int64
	var totalStorageRequest int64

	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			if container.Resources.Requests != nil {
				if cpu := container.Resources.Requests.Cpu(); cpu != nil {
					totalCPURequest += cpu.Value()
				}
				if mem := container.Resources.Requests.Memory(); mem != nil {
					totalMemRequest += mem.Value()
				}
				if storage := container.Resources.Requests.Storage(); storage != nil {
					totalStorageRequest += storage.Value()
				}
			}
			if container.Resources.Limits != nil {
				if cpu := container.Resources.Limits.Cpu(); cpu != nil {
					totalCPULimit += cpu.Value()
				}
				if mem := container.Resources.Limits.Memory(); mem != nil {
					totalMemLimit += mem.Value()
				}
			}
		}
	}

	return &quotaModel.UsageResponse{
		Namespace:      namespace,
		CPURequest:     resource.NewQuantity(totalCPURequest, resource.DecimalSI).String(),
		CPULimit:       resource.NewQuantity(totalCPULimit, resource.DecimalSI).String(),
		MemoryRequest:  resource.NewQuantity(totalMemRequest, resource.BinarySI).String(),
		MemoryLimit:    resource.NewQuantity(totalMemLimit, resource.BinarySI).String(),
		StorageRequest: resource.NewQuantity(totalStorageRequest, resource.BinarySI).String(),
		PodCount:       len(pods),
		LastUpdated:    time.Now().Format(time.RFC3339),
	}, nil
}

// checkResourceLimit checks if adding a resource would exceed the limit
func (u *usecase) checkResourceLimit(resourceType, requested, limit, current string) (bool, quotaModel.QuotaExceededReason) {
	requestedQty, err := resource.ParseQuantity(requested)
	if err != nil {
		return false, quotaModel.QuotaExceededReason{}
	}

	limitQty, err := resource.ParseQuantity(limit)
	if err != nil {
		return false, quotaModel.QuotaExceededReason{}
	}

	currentQty := resource.MustParse("0")
	if current != "" {
		currentQty, _ = resource.ParseQuantity(current)
	}

	// Add requested to current
	total := currentQty.DeepCopy()
	total.Add(requestedQty)

	if total.Cmp(limitQty) > 0 {
		utilization := float64(total.Value()) / float64(limitQty.Value()) * 100
		return true, quotaModel.QuotaExceededReason{
			ResourceType: resourceType,
			Requested:    requested,
			Limit:        limit,
			Current:      current,
			Utilization:  utilization,
		}
	}

	return false, quotaModel.QuotaExceededReason{}
}
