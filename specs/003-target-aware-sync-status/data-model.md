# Data Model: Target-Aware Sync and Status Resolution

## Overview

This feature extends the existing environment and delivery-target model so sync and GitOps status operations can resolve the correct management control plane for each environment instead of relying on one process-wide client.

## Entities

### Environment
- **Purpose**: Existing environment lifecycle record whose assigned delivery target determines which control plane should handle sync and GitOps status operations.
- **Existing code**: `internal/model/environment/type.go`
- **Relevant existing fields**:
  - `id`
  - `team_id`
  - `name`
  - `namespace`
  - `argo_app_name`
  - `cluster_name`
  - `cluster_server`
  - `delivery_target_id`
  - `last_sync_at`
  - `last_error`
  - `error_count`
- **Validation rules**:
  - `delivery_target_id` remains the source link to the target used for control-plane resolution.
  - Sync and status requests require a non-empty `argo_app_name`.
  - Team-scoped access must be validated before any target-specific control-plane call is attempted.
- **Relationships**:
  - Belongs to one `DeliveryTarget` through `delivery_target_id`.
  - Has many audit and notification records created by sync and status operations.

### DeliveryTarget
- **Purpose**: Approved placement target extended with the metadata needed to resolve the management control plane for sync and status work.
- **Existing code**: `internal/model/delivery_target/type.go`
- **Existing fields retained**:
  - `id`
  - `name`
  - `slug`
  - `team_id`
  - `cluster_name`
  - `cluster_server`
  - `availability_state`
  - `health_state`
- **New planned fields**:
  - `control_plane_name` — stable operator-facing identifier for the management control plane.
  - `control_plane_type` — distinguishes shapes such as shared-cluster ArgoCD vs dedicated management cluster if needed.
  - `kubeconfig_context` — Kubernetes context used to reach the control plane when running outside the cluster.
  - `argocd_namespace` — namespace containing the Application CRs for this target’s control plane.
  - `argocd_server` — logical or network identifier for the ArgoCD instance when separate instances exist.
  - `credential_reference` or equivalent secret/config reference — safe indirection to any non-default credentials required for the target.
- **Validation rules**:
  - Placement eligibility still depends on `availability_state` and team scope.
  - Sync/status eligibility additionally requires sufficient control-plane metadata to resolve a target-scoped client.
  - Secret values must not be stored directly in logs or exposed by read APIs.
- **Relationships**:
  - Referenced by `Environment` as the source of truth for target-aware control-plane selection.

### TargetControlPlane
- **Purpose**: Resolved, runtime-safe view of the control-plane metadata needed to choose target-scoped repositories for a request.
- **Planned code**: likely a new model under `internal/model/delivery_target` or `internal/model/environment`
- **Core fields**:
  - `delivery_target_id`
  - `control_plane_name`
  - `kubeconfig_context`
  - `cluster_server`
  - `argocd_namespace`
  - `argocd_server`
  - `uses_default_control_plane`
- **Validation rules**:
  - Must be derivable from the target and any referenced configuration without exposing raw credentials.
  - Must distinguish default single-target behavior from explicit per-target control-plane selection.
- **Relationships**:
  - Derived from `DeliveryTarget` and consumed by provider interfaces in the usecase flow.

### TargetResolutionAttempt
- **Purpose**: Audit-safe record of the control-plane selection attempt for a sync or status request.
- **Planned shape**: may be represented through existing audit logs and notification payloads rather than a new table in the first increment.
- **Core fields**:
  - `environment_id`
  - `delivery_target_id`
  - `operation` (`sync`, `gitops_status`, optional `environment_status` if Argo status is included)
  - `resolved_control_plane_name`
  - `outcome` (`resolved`, `rejected`, `unavailable`, `not_found`, `application_missing`)
  - `message`
  - `created_at`
- **Validation rules**:
  - Must not include tokens, kubeconfigs, or raw secret material.
  - Must remain understandable to operators diagnosing multi-target failures.
- **Relationships**:
  - References an `Environment` and a `DeliveryTarget`.
  - Can be surfaced through existing audit/logging mechanisms.

### TargetScopedGitOpsRepository
- **Purpose**: Runtime repository instance bound to the control plane selected for one request.
- **Planned code**: provider-backed repository under `internal/repository/gitops`
- **Core behavior**:
  - `GetApplicationStatus`
  - `SyncApplication`
  - optional `CreateApplication` and `DeleteApplication` if create/delete flows later need the same provider path
- **Validation rules**:
  - Must operate only against the control plane resolved for the current request.
  - Must surface target-specific failures in a way the usecase can wrap safely.

### TargetScopedProvisionerRepository
- **Purpose**: Runtime Kubernetes-facing repository instance bound to the resolved target when environment status or related workload views need target-aware access.
- **Planned code**: provider-backed repository under `internal/repository/provisioner`
- **Core behavior**:
  - `GetPodSummary`
  - `GetDeploymentSummary`
  - workload and log access methods already exposed by the existing provisioner repository interface
- **Validation rules**:
  - Must be selectable per environment target and avoid cross-target cache contamination.

## Relationships Summary
- `DeliveryTarget 1 -> N Environment`
- `DeliveryTarget 1 -> N TargetControlPlane` derivations over time as metadata changes
- `Environment 1 -> N TargetResolutionAttempt`
- `TargetControlPlane 1 -> N TargetScopedGitOpsRepository` runtime resolutions
- `TargetControlPlane 1 -> N TargetScopedProvisionerRepository` runtime resolutions

## Migration Impact
- Extend `delivery_targets` with explicit control-plane selection metadata for target-aware sync and status.
- Preserve existing `cluster_name` and `cluster_server` fields so placement behavior and historical data remain intact.
- Reuse existing audit and notification persistence unless implementation proves a dedicated resolution-attempt table is necessary.

## State Notes
- Existing single-target deployments remain valid when delivery targets resolve to the default shared control plane.
- A target can be valid for placement but invalid for sync/status if required control-plane metadata is missing; the feature must surface that distinction clearly.
