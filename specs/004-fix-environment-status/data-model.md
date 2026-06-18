# Data Model: Environment Status Accuracy

## Environment

**Purpose**: Represents the persisted team-scoped environment record that anchors all operational endpoint lookups.

**Existing fields used by this feature**:
- `id`
- `team_id`
- `name`
- `namespace`
- `status`
- `argo_app_name`
- `cluster_name`
- `cluster_server`
- `delivery_target_id`
- `last_error`
- `error_count`
- `created_at`
- `updated_at`

**Relationships**:
- Belongs to one team.
- Resolves to zero or one delivery target context.
- Owns one namespace whose live workloads are queried for operational status.

**Validation rules**:
- Environment lookups for operational endpoints must remain scoped by `team_id` and `id`.
- `namespace` is required to resolve workload-derived status.
- `delivery_target_id` determines which cluster context is used for workload and GitOps reads when target-aware providers are configured.

## Delivery Target Context

**Purpose**: Supplies the resolved cluster/control-plane context used to read live workload and GitOps state for an environment.

**Relevant attributes**:
- Delivery target identifier
- Placement availability for the requesting team
- Resolved Kubernetes query context
- Resolved ArgoCD query context

**Relationships**:
- One delivery target context may be referenced by many environments.
- One environment resolves through at most one delivery target context per request.

**Validation rules**:
- The environment's assigned delivery target must be resolvable before workload-derived status is read.
- Access checks must ensure the delivery target remains valid for the environment's team.

## Environment Status Summary

**Purpose**: User-visible aggregate view returned by `/v1/environments/{id}/status`.

**Fields**:
- Environment base response fields
- `pod_summary.total`
- `pod_summary.running`
- `pod_summary.pending`
- `pod_summary.failed`
- `deployment_summary.desired`
- `deployment_summary.ready`
- `deployment_summary.updated`
- `deployment_summary.available`
- `argo_status.sync_status`
- `argo_status.health_status`
- `argo_status.revision`
- `argo_status.message`

**Relationships**:
- Derived from one environment record plus live namespace workload state and optional GitOps application state.

**Validation rules**:
- Summary counts must describe the same namespace resolved for the environment request.
- Zero-valued summaries are valid only when the resolved namespace truly has no workloads.
- Retrieval failures must not be represented as successful zero-valued summaries.

## Environment Workload Summary

**Purpose**: Aggregate view returned by `/v1/environments/{id}/workloads`.

**Fields**:
- `environment_id`
- `namespace`
- `summary.total_workloads`
- `summary.healthy_workloads`
- `summary.degraded_workloads`
- `summary.total_pods`
- `summary.running_pods`
- `summary.pending_pods`
- `summary.failed_pods`
- `workloads[]`

**Relationships**:
- Derived from one environment and the set of workload and pod records currently discoverable in its namespace.

**Validation rules**:
- `environment_id` and `namespace` must always identify the requested environment, including empty namespaces.
- `workloads` must be an empty list when no workloads exist, not a null or blank context response.
- Summary totals must align with the `workloads` and pods returned from the same namespace read.

## Workload

**Purpose**: User-visible operational view of a deployable unit in an environment namespace.

**Fields**:
- `name`
- `kind`
- `status`
- `desired_replicas`
- `ready_replicas`
- `image`

**Relationships**:
- Exists within one environment namespace.
- May own one or more pods.
- May be selected for detail lookup through `/v1/environments/{id}/workloads/{name}`.

**Validation rules**:
- Workload detail must resolve from the same namespace and target context as the list endpoint.
- A workload name absent from the resolved namespace must produce a not-found outcome.

## Pod Status Snapshot

**Purpose**: Derived runtime view used to calculate environment and workload summaries.

**Fields**:
- Pod name
- Owner workload name and kind
- Phase
- Readiness signals
- Restart count
- Node and IP data when available

**Relationships**:
- Belongs to one namespace and is associated with zero or one workload owner.

**Validation rules**:
- Pod counts must only include pods returned from the resolved environment namespace.
- Pod phases must feed the environment and workload summary counters consistently across endpoints.
