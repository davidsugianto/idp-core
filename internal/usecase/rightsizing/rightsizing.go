package rightsizing

import (
	"context"
	"fmt"
	"time"

	"github.com/davidsugianto/idp-core/internal/model/rightsizing"
	"github.com/davidsugianto/idp-core/internal/pkg/config"
	"github.com/google/uuid"
	corev1 "k8s.io/api/core/v1"
)

// GenerateRecommendations analyzes workloads and generates rightsizing recommendations
func (u *usecase) GenerateRecommendations(ctx context.Context) error {
	// Get all workloads from Kubernetes (all namespaces)
	workloads, err := u.provisionerRepo.GetWorkloads("")
	if err != nil {
		return fmt.Errorf("failed to get workloads: %w", err)
	}

	now := time.Now()
	lookbackDays := config.GetConfig().FinOps.Rightsizing.LookbackDays
	if lookbackDays == 0 {
		lookbackDays = 7
	}
	analysisStart := now.AddDate(0, 0, -lookbackDays)

	for _, workload := range workloads {
		for _, container := range workload.Spec.Template.Spec.Containers {
			// Skip if pending recommendation already exists
			exists, err := u.rightsizingRepo.ExistsPendingForContainer(ctx, workload.Namespace, workload.Name, rightsizing.WorkloadTypeDeployment, container.Name)
			if err != nil {
				continue
			}
			if exists {
				continue
			}

			// Analyze container and generate recommendation
			rec, err := u.analyzeContainer(ctx, workload.Namespace, workload.Name, rightsizing.WorkloadTypeDeployment, container, analysisStart, now)
			if err != nil {
				continue
			}

			if rec != nil {
				if err := u.rightsizingRepo.Create(ctx, rec); err != nil {
					continue
				}
			}
		}
	}

	return nil
}

// analyzeContainer analyzes a container's resource usage and generates a recommendation
func (u *usecase) analyzeContainer(ctx context.Context, namespace, workloadName, workloadType string, container corev1.Container, start, end time.Time) (*rightsizing.RightsizingRecommendation, error) {
	// Query Prometheus for CPU usage
	cpuUsageAvg, cpuUsageMax, err := u.getCPUUsage(ctx, namespace, container.Name, start, end)
	if err != nil {
		return nil, err
	}

	// Query Prometheus for memory usage
	memUsageAvg, memUsageMax, err := u.getMemoryUsage(ctx, namespace, container.Name, start, end)
	if err != nil {
		return nil, err
	}

	// Calculate recommendations
	rec := u.calculateRecommendation(namespace, workloadName, workloadType, container, cpuUsageAvg, cpuUsageMax, memUsageAvg, memUsageMax, start, end)

	return rec, nil
}

// getCPUUsage queries Prometheus for CPU usage metrics
func (u *usecase) getCPUUsage(ctx context.Context, namespace, container string, start, end time.Time) (avg, max float64, err error) {
	// avg_over_time(rate(container_cpu_usage_seconds_total{namespace="%s",container="%s"}[5m])[%dd])
	avgQuery := fmt.Sprintf(
		`avg_over_time(rate(container_cpu_usage_seconds_total{namespace="%s",container="%s"}[5m])[%dd])`,
		namespace, container, config.GetConfig().FinOps.Rightsizing.LookbackDays,
	)

	results, err := u.monitoringRepo.Query(ctx, avgQuery)
	if err != nil {
		return 0, 0, err
	}
	if len(results) > 0 {
		avg = results[0].Value
	}

	// max_over_time(rate(container_cpu_usage_seconds_total{namespace="%s",container="%s"}[5m])[%dd])
	maxQuery := fmt.Sprintf(
		`max_over_time(rate(container_cpu_usage_seconds_total{namespace="%s",container="%s"}[5m])[%dd])`,
		namespace, container, config.GetConfig().FinOps.Rightsizing.LookbackDays,
	)

	results, err = u.monitoringRepo.Query(ctx, maxQuery)
	if err != nil {
		return avg, 0, err
	}
	if len(results) > 0 {
		max = results[0].Value
	}

	return avg, max, nil
}

// getMemoryUsage queries Prometheus for memory usage metrics
func (u *usecase) getMemoryUsage(ctx context.Context, namespace, container string, start, end time.Time) (avg, max float64, err error) {
	// avg_over_time(container_memory_working_set_bytes{namespace="%s",container="%s"}[%dd])
	avgQuery := fmt.Sprintf(
		`avg_over_time(container_memory_working_set_bytes{namespace="%s",container="%s"}[%dd])`,
		namespace, container, config.GetConfig().FinOps.Rightsizing.LookbackDays,
	)

	results, err := u.monitoringRepo.Query(ctx, avgQuery)
	if err != nil {
		return 0, 0, err
	}
	if len(results) > 0 {
		avg = results[0].Value
	}

	// max_over_time(container_memory_working_set_bytes{namespace="%s",container="%s"}[%dd])
	maxQuery := fmt.Sprintf(
		`max_over_time(container_memory_working_set_bytes{namespace="%s",container="%s"}[%dd])`,
		namespace, container, config.GetConfig().FinOps.Rightsizing.LookbackDays,
	)

	results, err = u.monitoringRepo.Query(ctx, maxQuery)
	if err != nil {
		return avg, 0, err
	}
	if len(results) > 0 {
		max = results[0].Value
	}

	return avg, max, nil
}

// calculateRecommendation determines the recommendation type and calculates recommended resources
func (u *usecase) calculateRecommendation(namespace, workloadName, workloadType string, container corev1.Container, cpuAvg, cpuMax, memAvg, memMax float64, start, end time.Time) *rightsizing.RightsizingRecommendation {
	// Extract current resource values from Kubernetes container
	cpuRequest := getResourceString(container.Resources.Requests, corev1.ResourceCPU)
	cpuLimit := getResourceString(container.Resources.Limits, corev1.ResourceCPU)
	memRequest := getResourceString(container.Resources.Requests, corev1.ResourceMemory)
	memLimit := getResourceString(container.Resources.Limits, corev1.ResourceMemory)

	rec := &rightsizing.RightsizingRecommendation{
		ID:                   uuid.New().String(),
		Namespace:            namespace,
		WorkloadName:         workloadName,
		WorkloadType:         workloadType,
		ContainerName:        container.Name,
		CurrentCPURequest:    cpuRequest,
		CurrentCPULimit:      cpuLimit,
		CurrentMemoryRequest: memRequest,
		CurrentMemoryLimit:   memLimit,
		CPUUsageAvg:          formatCPU(cpuAvg),
		CPUUsageMax:          formatCPU(cpuMax),
		MemoryUsageAvg:       formatMemory(memAvg),
		MemoryUsageMax:       formatMemory(memMax),
		AnalysisPeriodStart:  start,
		AnalysisPeriodEnd:    end,
		Status:               rightsizing.StatusPending,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Parse current requests for utilization calculation
	cpuRequestCores := parseCPU(cpuRequest)
	memRequestBytes := parseMemory(memRequest)

	// Calculate utilization
	var cpuUtilization, memUtilization float64
	if cpuRequestCores > 0 {
		cpuUtilization = cpuAvg / cpuRequestCores
	}
	if memRequestBytes > 0 {
		memUtilization = memAvg / memRequestBytes
	}

	// Determine recommendation type
	cpuUnderutilThreshold := config.GetConfig().FinOps.Rightsizing.CPUUnderutilThreshold
	if cpuUnderutilThreshold == 0 {
		cpuUnderutilThreshold = 0.5
	}
	memUnderutilThreshold := config.GetConfig().FinOps.Rightsizing.MemoryUnderutilThreshold
	if memUnderutilThreshold == 0 {
		memUnderutilThreshold = 0.5
	}
	cpuOverutilThreshold := config.GetConfig().FinOps.Rightsizing.CPUOverutilThreshold
	if cpuOverutilThreshold == 0 {
		cpuOverutilThreshold = 0.9
	}
	memOverutilThreshold := config.GetConfig().FinOps.Rightsizing.MemoryOverutilThreshold
	if memOverutilThreshold == 0 {
		memOverutilThreshold = 0.9
	}

	// Safety buffers
	safetyBufferCPU := config.GetConfig().FinOps.Rightsizing.SafetyBufferCPU
	if safetyBufferCPU == 0 {
		safetyBufferCPU = 1.2
	}
	safetyBufferMemory := config.GetConfig().FinOps.Rightsizing.SafetyBufferMemory
	if safetyBufferMemory == 0 {
		safetyBufferMemory = 1.3
	}

	minConfidence := config.GetConfig().FinOps.Rightsizing.MinConfidenceScore
	if minConfidence == 0 {
		minConfidence = 70
	}

	// Calculate confidence score based on data availability
	confidence := u.calculateConfidence(cpuAvg, cpuMax, memAvg, memMax)

	if cpuUtilization < cpuUnderutilThreshold && memUtilization < memUnderutilThreshold && confidence >= minConfidence {
		// Scale down recommendation
		rec.RecommendationType = rightsizing.RecommendationTypeScaleDown
		rec.RecommendedCPURequest = formatCPU(cpuAvg * safetyBufferCPU)
		rec.RecommendedCPULimit = formatCPU(cpuMax * safetyBufferCPU)
		rec.RecommendedMemoryRequest = formatMemory(memAvg * safetyBufferMemory)
		rec.RecommendedMemoryLimit = formatMemory(memMax * safetyBufferMemory)
		rec.SavingsPotential = u.calculateSavingsPotential(cpuRequest, memRequest, cpuAvg, memAvg)
		rec.ConfidenceScore = confidence
	} else if cpuUtilization > cpuOverutilThreshold || memUtilization > memOverutilThreshold {
		// Scale up recommendation
		rec.RecommendationType = rightsizing.RecommendationTypeScaleUp
		rec.RecommendedCPURequest = formatCPU(cpuMax * 1.5)
		rec.RecommendedCPULimit = formatCPU(cpuMax * 2.0)
		rec.RecommendedMemoryRequest = formatMemory(memMax * 1.5)
		rec.RecommendedMemoryLimit = formatMemory(memMax * 2.0)
		rec.SavingsPotential = 0
		rec.ConfidenceScore = confidence
	} else {
		// Optimal - no recommendation needed
		return nil
	}

	return rec
}

// calculateConfidence calculates a confidence score (0-100) for the recommendation
func (u *usecase) calculateConfidence(cpuAvg, cpuMax, memAvg, memMax float64) float64 {
	score := 100.0

	// Reduce confidence if data is missing or zero
	if cpuAvg == 0 {
		score -= 25
	}
	if cpuMax == 0 {
		score -= 15
	}
	if memAvg == 0 {
		score -= 25
	}
	if memMax == 0 {
		score -= 15
	}

	// Reduce confidence if there's high variance (max >> avg)
	if cpuAvg > 0 && cpuMax/cpuAvg > 3.0 {
		score -= 10
	}
	if memAvg > 0 && memMax/memAvg > 3.0 {
		score -= 10
	}

	return score
}

// calculateSavingsPotential estimates monthly cost savings
func (u *usecase) calculateSavingsPotential(cpuRequest, memRequest string, cpuAvg, memAvg float64) float64 {
	// Simplified calculation: estimate % reduction in resources
	cpuCurrent := parseCPU(cpuRequest)
	memCurrent := parseMemory(memRequest)

	safetyBufferCPU := config.GetConfig().FinOps.Rightsizing.SafetyBufferCPU
	if safetyBufferCPU == 0 {
		safetyBufferCPU = 1.2
	}
	safetyBufferMemory := config.GetConfig().FinOps.Rightsizing.SafetyBufferMemory
	if safetyBufferMemory == 0 {
		safetyBufferMemory = 1.3
	}

	var cpuSavings, memSavings float64
	if cpuCurrent > 0 {
		cpuSavings = (cpuCurrent - cpuAvg*safetyBufferCPU) / cpuCurrent
	}
	if memCurrent > 0 {
		memSavings = (memCurrent - memAvg*safetyBufferMemory) / memCurrent
	}

	// Average savings percentage (simplified)
	return (cpuSavings + memSavings) / 2 * 100
}

// getResourceString extracts a resource value as string from a ResourceList
func getResourceString(resources corev1.ResourceList, name corev1.ResourceName) string {
	if val, ok := resources[name]; ok {
		return val.String()
	}
	return ""
}

// ListRecommendations returns a paginated list of recommendations
func (u *usecase) ListRecommendations(ctx context.Context, req *rightsizing.ListRecommendationsRequest) (*rightsizing.RecommendationListResponse, error) {
	recs, total, err := u.rightsizingRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}
	return rightsizing.ToRecommendationListResponse(recs, total), nil
}

// GetRecommendation returns a single recommendation by ID
func (u *usecase) GetRecommendation(ctx context.Context, id string) (*rightsizing.RecommendationResponse, error) {
	rec, err := u.rightsizingRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return rightsizing.ToRecommendationResponse(rec), nil
}

// ApplyRecommendation applies a recommendation to the Kubernetes workload
func (u *usecase) ApplyRecommendation(ctx context.Context, id, userID string) error {
	rec, err := u.rightsizingRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if rec.Status != rightsizing.StatusPending {
		return fmt.Errorf("recommendation is not in pending status")
	}

	// Store previous state for rollback
	previousState := &rightsizing.PreviousResourceState{
		CPURequest:    rec.CurrentCPURequest,
		CPULimit:      rec.CurrentCPULimit,
		MemoryRequest: rec.CurrentMemoryRequest,
		MemoryLimit:   rec.CurrentMemoryLimit,
	}
	if err := rec.SetPreviousState(previousState); err != nil {
		return err
	}

	// Apply to Kubernetes
	var updateErr error
	switch rec.WorkloadType {
	case rightsizing.WorkloadTypeDeployment:
		updateErr = u.provisionerRepo.UpdateDeploymentResources(
			ctx, rec.Namespace, rec.WorkloadName, rec.ContainerName,
			rec.RecommendedCPURequest, rec.RecommendedCPULimit,
			rec.RecommendedMemoryRequest, rec.RecommendedMemoryLimit,
		)
	case rightsizing.WorkloadTypeStatefulSet:
		updateErr = u.provisionerRepo.UpdateStatefulSetResources(
			ctx, rec.Namespace, rec.WorkloadName, rec.ContainerName,
			rec.RecommendedCPURequest, rec.RecommendedCPULimit,
			rec.RecommendedMemoryRequest, rec.RecommendedMemoryLimit,
		)
	default:
		return fmt.Errorf("unsupported workload type: %s", rec.WorkloadType)
	}

	if updateErr != nil {
		rec.Status = rightsizing.StatusFailed
		u.rightsizingRepo.Update(ctx, rec)
		return fmt.Errorf("failed to apply recommendation: %w", updateErr)
	}

	// Update status
	now := time.Now()
	rec.Status = rightsizing.StatusApplied
	rec.AppliedAt = &now
	rec.AppliedBy = userID
	rec.UpdatedAt = now

	return u.rightsizingRepo.Update(ctx, rec)
}

// RollbackRecommendation rolls back an applied recommendation
func (u *usecase) RollbackRecommendation(ctx context.Context, id, userID string) error {
	rec, err := u.rightsizingRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if rec.Status != rightsizing.StatusApplied {
		return fmt.Errorf("recommendation is not in applied status")
	}

	previousState, err := rec.GetPreviousState()
	if err != nil {
		return fmt.Errorf("failed to get previous state: %w", err)
	}
	if previousState == nil {
		return fmt.Errorf("no previous state stored for rollback")
	}

	// Rollback in Kubernetes
	var updateErr error
	switch rec.WorkloadType {
	case rightsizing.WorkloadTypeDeployment:
		updateErr = u.provisionerRepo.UpdateDeploymentResources(
			ctx, rec.Namespace, rec.WorkloadName, rec.ContainerName,
			previousState.CPURequest, previousState.CPULimit,
			previousState.MemoryRequest, previousState.MemoryLimit,
		)
	case rightsizing.WorkloadTypeStatefulSet:
		updateErr = u.provisionerRepo.UpdateStatefulSetResources(
			ctx, rec.Namespace, rec.WorkloadName, rec.ContainerName,
			previousState.CPURequest, previousState.CPULimit,
			previousState.MemoryRequest, previousState.MemoryLimit,
		)
	}

	if updateErr != nil {
		return fmt.Errorf("failed to rollback recommendation: %w", updateErr)
	}

	// Update status back to pending
	rec.Status = rightsizing.StatusPending
	rec.AppliedAt = nil
	rec.AppliedBy = ""
	rec.UpdatedAt = time.Now()

	return u.rightsizingRepo.Update(ctx, rec)
}

// DismissRecommendation dismisses a pending recommendation
func (u *usecase) DismissRecommendation(ctx context.Context, id string, reason string) error {
	rec, err := u.rightsizingRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if rec.Status != rightsizing.StatusPending {
		return fmt.Errorf("recommendation is not in pending status")
	}

	rec.Status = rightsizing.StatusDismissed
	rec.UpdatedAt = time.Now()

	return u.rightsizingRepo.Update(ctx, rec)
}
