package environment

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/davidsugianto/idp-core/internal/model/environment"
	"github.com/davidsugianto/idp-core/internal/model/workload"
	"github.com/davidsugianto/idp-core/internal/pkg/argocd"
	"github.com/google/uuid"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	StatusCreating = "creating"
	StatusReady    = "ready"
	StatusDeleting = "deleting"
	StatusFailed   = "failed"
)

// Sentinel errors
var (
	ErrEnvironmentNotFound = errors.New("environment not found")
	ErrNoArgoApp           = errors.New("environment has no ArgoCD application")
	ErrGitOpsNotConfigured = errors.New("GitOps integration not configured")
	ErrK8sNotConfigured    = errors.New("Kubernetes integration not configured")
	ErrWorkloadNotFound    = errors.New("workload not found")
)

func (u *usecase) Create(ctx context.Context, teamID string, req environment.CreateEnvironmentRequest) (*environment.Environment, error) {
	// Generate UUID
	id := uuid.New().String()

	// Generate namespace name (DNS-1123 compliant)
	namespace := generateNamespace(teamID, req.Name)

	// Generate ArgoCD app name
	argAppName := fmt.Sprintf("env-%s", id[:8])

	// Set default git revision
	gitRevision := req.GitRevision
	if gitRevision == "" {
		gitRevision = "main"
	}

	env := &environment.Environment{
		ID:           id,
		TeamID:       teamID,
		Name:         req.Name,
		Namespace:    namespace,
		Status:       StatusCreating,
		GitRepoURL:   req.GitRepoURL,
		GitRevision:  gitRevision,
		ManifestPath: req.ManifestPath,
		ArgoAppName:  argAppName,
	}

	// Persist environment record
	if err := u.environmentRepo.Create(ctx, env); err != nil {
		return nil, fmt.Errorf("failed to create environment: %w", err)
	}

	// Create Kubernetes namespace via provisioner repository
	if u.provisionerRepo != nil {
		labels := map[string]string{
			"idp-core/team-id":        teamID,
			"idp-core/environment-id": id,
			"idp-core/managed-by":     "idp-core",
		}

		if err := u.provisionerRepo.CreateNamespace(ctx, namespace, labels); err != nil {
			u.environmentRepo.UpdateStatus(ctx, id, teamID, StatusFailed, err.Error())
			return nil, fmt.Errorf("failed to create namespace: %w", err)
		}

		// Create ResourceQuota if specified
		if req.ResourceQuotaCPU != "" || req.ResourceQuotaMemory != "" {
			if err := u.provisionerRepo.CreateResourceQuota(ctx, namespace, "idp-quota", req.ResourceQuotaCPU, req.ResourceQuotaMemory); err != nil {
				// Log but don't fail - quota is optional
			}
		}

		// Create NetworkPolicy for namespace isolation
		if err := u.provisionerRepo.CreateNetworkPolicy(ctx, namespace, "idp-isolation", labels); err != nil {
			// Log but don't fail - network policy is optional
		}
	}

	// Create ArgoCD Application via gitops repository
	if u.gitopsRepo != nil {
		appSpec := argocd.ApplicationSpec{
			Name:      argAppName,
			Namespace: namespace,
			RepoURL:   req.GitRepoURL,
			Revision:  gitRevision,
			Path:      req.ManifestPath,
			ServerURL: "https://kubernetes.default.svc",
		}

		if err := u.gitopsRepo.CreateApplication(ctx, appSpec); err != nil {
			u.environmentRepo.UpdateStatus(ctx, id, teamID, StatusFailed, fmt.Sprintf("failed to create ArgoCD Application: %v", err))
			return nil, fmt.Errorf("failed to create ArgoCD Application: %w", err)
		}
	}

	// Mark as ready
	if err := u.environmentRepo.UpdateStatus(ctx, id, teamID, StatusReady, ""); err != nil {
		return nil, fmt.Errorf("failed to update environment status: %w", err)
	}

	env.Status = StatusReady
	return env, nil
}

// generateNamespace creates a DNS-1123 compliant namespace name
func generateNamespace(teamID, envName string) string {
	// Format: idp-{teamSlug}-{envSlug}
	teamSlug := sanitizeSlug(teamID, 20)
	envSlug := sanitizeSlug(envName, 30)

	ns := fmt.Sprintf("idp-%s-%s", teamSlug, envSlug)

	// Ensure max length of 63 characters
	if len(ns) > 63 {
		ns = ns[:63]
	}

	// Trim trailing hyphens
	ns = strings.TrimRight(ns, "-")

	return ns
}

var nonDNSCharRegex = regexp.MustCompile(`[^a-z0-9-]`)

func sanitizeSlug(s string, maxLen int) string {
	s = strings.ToLower(s)
	s = nonDNSCharRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")

	// Collapse multiple hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	if len(s) > maxLen {
		s = s[:maxLen]
	}

	return strings.TrimRight(s, "-")
}

func (u *usecase) List(ctx context.Context, teamID string) ([]environment.Environment, error) {
	return u.environmentRepo.ListByTeam(ctx, teamID)
}

func (u *usecase) Get(ctx context.Context, teamID, id string) (*environment.Environment, error) {
	env, err := u.environmentRepo.GetByIDAndTeam(ctx, id, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}
	if env == nil {
		return nil, ErrEnvironmentNotFound
	}
	return env, nil
}

func (u *usecase) GetStatus(ctx context.Context, teamID, id string) (*environment.EnvironmentStatusResponse, error) {
	env, err := u.environmentRepo.GetByIDAndTeam(ctx, id, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}
	if env == nil {
		return nil, ErrEnvironmentNotFound
	}

	response := &environment.EnvironmentStatusResponse{
		EnvironmentResponse: *environment.ToEnvironmentResponse(env),
	}

	// Get K8s status from provisioner repository
	if u.provisionerRepo != nil {
		if podSummary, ok := u.provisionerRepo.GetPodSummary(env.Namespace); ok {
			response.PodSummary = podSummary
		}

		if deploySummary, ok := u.provisionerRepo.GetDeploymentSummary(env.Namespace); ok {
			response.DeploymentSummary = deploySummary
		}
	}

	// Get ArgoCD status from gitops repository
	if u.gitopsRepo != nil && env.ArgoAppName != "" {
		if argoStatus, err := u.gitopsRepo.GetApplicationStatus(ctx, env.ArgoAppName); err == nil {
			response.ArgoStatus = *argoStatus
		}
	}

	return response, nil
}

func (u *usecase) Delete(ctx context.Context, teamID, id string) error {
	env, err := u.environmentRepo.GetByIDAndTeam(ctx, id, teamID)
	if err != nil {
		return fmt.Errorf("failed to get environment: %w", err)
	}
	if env == nil {
		return ErrEnvironmentNotFound
	}

	// Update status to deleting
	if err := u.environmentRepo.UpdateStatus(ctx, id, teamID, StatusDeleting, ""); err != nil {
		return fmt.Errorf("failed to update environment status: %w", err)
	}

	// Delete ArgoCD Application via gitops repository
	if u.gitopsRepo != nil && env.ArgoAppName != "" {
		if err := u.gitopsRepo.DeleteApplication(ctx, env.ArgoAppName); err != nil {
			// Log but continue
		}
	}

	// Delete Kubernetes namespace via provisioner repository
	if u.provisionerRepo != nil {
		if err := u.provisionerRepo.DeleteNamespace(ctx, env.Namespace); err != nil {
			// Log error but continue with DB deletion
			u.environmentRepo.UpdateStatus(ctx, id, teamID, StatusFailed, fmt.Sprintf("failed to delete namespace: %v", err))
		}
	}

	// Soft delete the record
	if err := u.environmentRepo.SoftDelete(ctx, id, teamID); err != nil {
		return fmt.Errorf("failed to delete environment: %w", err)
	}

	return nil
}

// TriggerSync triggers a manual sync of the ArgoCD Application
func (u *usecase) TriggerSync(ctx context.Context, teamID, id string) error {
	env, err := u.environmentRepo.GetByIDAndTeam(ctx, id, teamID)
	if err != nil {
		return fmt.Errorf("failed to get environment: %w", err)
	}
	if env == nil {
		return ErrEnvironmentNotFound
	}

	if env.ArgoAppName == "" {
		return ErrNoArgoApp
	}

	if u.gitopsRepo == nil {
		return ErrGitOpsNotConfigured
	}

	if err := u.gitopsRepo.SyncApplication(ctx, env.ArgoAppName); err != nil {
		return fmt.Errorf("failed to trigger sync: %w", err)
	}

	// Update last sync time
	now := time.Now()
	env.LastSyncAt = &now

	return nil
}

// GetGitOpsStatus fetches sync and health status from ArgoCD
func (u *usecase) GetGitOpsStatus(ctx context.Context, teamID, id string) (*environment.ArgoStatus, error) {
	env, err := u.environmentRepo.GetByIDAndTeam(ctx, id, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}
	if env == nil {
		return nil, ErrEnvironmentNotFound
	}

	if env.ArgoAppName == "" {
		return nil, ErrNoArgoApp
	}

	if u.gitopsRepo == nil {
		return nil, ErrGitOpsNotConfigured
	}

	return u.gitopsRepo.GetApplicationStatus(ctx, env.ArgoAppName)
}

// GetWorkloads fetches live workload data from Kubernetes for a specific environment
func (u *usecase) GetWorkloads(ctx context.Context, teamID, id string) (*workload.WorkloadStatusResponse, error) {
	env, err := u.environmentRepo.GetByIDAndTeam(ctx, id, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}
	if env == nil {
		return nil, ErrEnvironmentNotFound
	}

	if u.provisionerRepo == nil {
		return nil, ErrK8sNotConfigured
	}

	// Get deployments from cache
	deployments, err := u.provisionerRepo.GetWorkloads(env.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get workloads: %w", err)
	}

	// Get pods from cache
	pods, err := u.provisionerRepo.GetPods(env.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %w", err)
	}

	// Convert to workload status
	workloadStatuses := make([]workload.WorkloadStatus, len(deployments))
	for i, d := range deployments {
		workloadStatuses[i] = convertDeploymentToWorkloadStatus(d, id, env.Namespace)
	}

	// Convert pods
	podStatuses := make([]workload.PodStatus, len(pods))
	for i, p := range pods {
		podStatuses[i] = convertPodToPodStatus(p, id, env.Namespace)
	}

	return workload.ToWorkloadStatusResponse(workloadStatuses, podStatuses), nil
}

// convertDeploymentToWorkloadStatus converts a K8s Deployment to WorkloadStatus
func convertDeploymentToWorkloadStatus(d *appsv1.Deployment, envID, namespace string) workload.WorkloadStatus {
	status := "Progressing"
	if d.Status.ReadyReplicas == d.Status.Replicas && d.Status.Replicas > 0 {
		status = "Running"
	} else if d.Status.UnavailableReplicas > 0 {
		status = "Degraded"
	}

	image := ""
	if len(d.Spec.Template.Spec.Containers) > 0 {
		image = d.Spec.Template.Spec.Containers[0].Image
	}

	desired := int32(0)
	if d.Spec.Replicas != nil {
		desired = *d.Spec.Replicas
	}

	return workload.WorkloadStatus{
		ID:                string(d.UID),
		EnvironmentID:     envID,
		Namespace:         namespace,
		Name:              d.Name,
		Kind:              "Deployment",
		DesiredReplicas:   int(desired),
		CurrentReplicas:   int(d.Status.Replicas),
		ReadyReplicas:     int(d.Status.ReadyReplicas),
		UpdatedReplicas:   int(d.Status.UpdatedReplicas),
		AvailableReplicas: int(d.Status.AvailableReplicas),
		Status:            status,
		Image:             image,
	}
}

// convertPodToPodStatus converts a K8s Pod to PodStatus
func convertPodToPodStatus(p *corev1.Pod, envID, namespace string) workload.PodStatus {
	ownerName := ""
	ownerKind := ""
	for _, owner := range p.OwnerReferences {
		ownerName = owner.Name
		ownerKind = owner.Kind
		break
	}

	ready := false
	initialized := false
	containersReady := false
	scheduled := false
	restartCount := 0

	for _, cond := range p.Status.Conditions {
		switch cond.Type {
		case corev1.PodReady:
			ready = cond.Status == corev1.ConditionTrue
		case corev1.PodInitialized:
			initialized = cond.Status == corev1.ConditionTrue
		case corev1.ContainersReady:
			containersReady = cond.Status == corev1.ConditionTrue
		case corev1.PodScheduled:
			scheduled = cond.Status == corev1.ConditionTrue
		}
	}

	for _, cs := range p.Status.ContainerStatuses {
		restartCount += int(cs.RestartCount)
	}

	return workload.PodStatus{
		ID:                string(p.UID),
		EnvironmentID:     envID,
		Namespace:         namespace,
		Name:              p.Name,
		OwnerName:         ownerName,
		OwnerKind:         ownerKind,
		Phase:             string(p.Status.Phase),
		PodIP:             p.Status.PodIP,
		NodeName:          p.Spec.NodeName,
		Ready:             ready,
		Initialized:       initialized,
		ContainersReady:   containersReady,
		Scheduled:         scheduled,
		ContainerCount:    len(p.Spec.Containers),
		InitContainerCount: len(p.Spec.InitContainers),
		RestartCount:      restartCount,
	}
}

// GetWorkloadDetails fetches a specific workload and its pods' status
func (u *usecase) GetWorkloadDetails(ctx context.Context, teamID, id, workloadName string) (*workload.WorkloadInfo, error) {
	env, err := u.environmentRepo.GetByIDAndTeam(ctx, id, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment: %w", err)
	}
	if env == nil {
		return nil, ErrEnvironmentNotFound
	}

	if u.provisionerRepo == nil {
		return nil, ErrK8sNotConfigured
	}

	// Get all deployments from cache
	deployments, err := u.provisionerRepo.GetWorkloads(env.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get workloads: %w", err)
	}

	// Find the specific deployment
	var targetDeployment *appsv1.Deployment
	for _, d := range deployments {
		if d.Name == workloadName {
			targetDeployment = d
			break
		}
	}

	if targetDeployment == nil {
		return nil, ErrWorkloadNotFound
	}

	// Convert deployment to workload info
	ws := convertDeploymentToWorkloadStatus(targetDeployment, id, env.Namespace)

	return &workload.WorkloadInfo{
		Name:            ws.Name,
		Kind:            ws.Kind,
		Status:          ws.Status,
		DesiredReplicas: ws.DesiredReplicas,
		ReadyReplicas:   ws.ReadyReplicas,
		Image:           ws.Image,
	}, nil
}
