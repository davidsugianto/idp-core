# Research: Target-Aware Sync and Status Resolution

## Decision 1: Resolve delivery target in the environment usecase for each sync and status request

**Decision**: The environment usecase will load the environment, resolve its assigned delivery target, validate target availability and team scope, and then delegate sync or status work through target-scoped repository/provider interfaces.

**Rationale**: The current flow in `internal/usecase/environment/environment.go` reads the environment and directly calls a single injected `gitopsRepo` for `TriggerSync`, `GetGitOpsStatus`, and the ArgoCD portion of `GetStatus`. That shape cannot safely support different control planes per environment. Resolving the target in the usecase preserves the repository → usecase → handler boundary and keeps transport concerns out of handlers.

**Alternatives considered**:
- Resolve target selection in the HTTP handler — rejected because handlers should remain transport-only and should not decide infrastructure routing.
- Keep a single global GitOps repository and infer the target only from the application name — rejected because different environments can share the same service instance while requiring different Kubernetes or ArgoCD control planes.

## Decision 2: Introduce provider interfaces that return target-scoped GitOps and provisioner repositories

**Decision**: Add provider/factory interfaces that can build or cache target-scoped GitOps and Kubernetes-facing repositories from resolved delivery-target metadata, instead of injecting one process-wide repository instance for all operations.

**Rationale**: `cmd/http/main.go` currently constructs a single Kubernetes client and a single ArgoCD client from global config, then injects singleton `provisionerRepo` and `gitopsRepo` instances into the environment usecase. Target-aware operations need a seam that lets the usecase ask for the correct repository at request time without importing client construction details. A provider keeps client instantiation and caching in wiring or repository-adjacent packages while leaving the usecase focused on orchestration.

**Alternatives considered**:
- Construct raw Kubernetes or ArgoCD clients directly inside the usecase — rejected because it violates clean architecture and scatters configuration logic into business code.
- Create one repository per known target during process startup — rejected because targets are persisted data and may change without requiring a code-level registry of all possible clients.

## Decision 3: Extend delivery target metadata to describe control-plane selection, not only workload destination cluster fields

**Decision**: Extend delivery target records with explicit control-plane metadata sufficient to identify the Kubernetes context and ArgoCD control plane that manage sync and status operations for environments assigned to that target.

**Rationale**: Existing delivery target records only store `cluster_name` and `cluster_server`, which are enough for environment placement and ArgoCD Application destination fields but not enough to choose which ArgoCD instance or Kubernetes context contains the Application CR being synced or read. Multi-control-plane support needs stored metadata that distinguishes workload destination from management control plane.

**Alternatives considered**:
- Reuse only the existing `cluster_name` and `cluster_server` fields — rejected because they cannot represent multiple ArgoCD instances, namespaces, or kubeconfig contexts reliably.
- Keep target-to-control-plane mappings only in process config — rejected because the feature requires persisted, auditable, target-aware behavior tied to delivery targets.

## Decision 4: Preserve current single-target behavior through an explicit default resolution path

**Decision**: If a delivery target resolves to the same shared control plane used in current single-target deployments, the provider will return repositories backed by that default control plane so existing deployments continue to behave as they do today.

**Rationale**: The specification requires preserving current behavior for environments whose resolved target maps to the same control plane currently used by the service. Keeping an explicit default mapping allows incremental rollout without forcing every existing environment into a new multi-target setup on day one.

**Alternatives considered**:
- Require all existing targets and environments to be backfilled with fully explicit control-plane metadata before the feature works — rejected because it increases rollout risk and breaks the incremental delivery requirement.

## Decision 5: Make target resolution failures explicit in contracts, logs, and retained records

**Decision**: Sync and status flows will emit clear target-resolution and control-plane-selection failures, and operator-visible logs plus retained audit-relevant records will identify which delivery target and control plane were attempted without exposing secrets.

**Rationale**: The updated constitution requires target-aware observability and failure reporting. The current failure mode shows low-level connection errors without clearly tying them to target resolution. Operators need actionable, target-specific visibility to diagnose missing mappings, invalid metadata, unavailable control planes, and application-not-found cases.

**Alternatives considered**:
- Return raw client errors only — rejected because they obscure which target was chosen and are not sufficient for multi-target diagnosis.
- Log full credential or connection details — rejected because secrets and raw credentials must not be exposed.

## Decision 6: Validate target-aware behavior with focused usecase tests plus multi-target integration scenarios

**Decision**: The implementation should add focused usecase tests for resolution precedence and failure paths, repository/provider tests for target-scoped client selection, and integration validation that exercises at least two different targets with distinct control-plane mappings.

**Rationale**: The risk in this feature is incorrect target selection, not only handler transport behavior. Test coverage needs to prove the usecase resolves the right target, uses the right provider, rejects invalid or cross-team mappings, and keeps one environment’s sync/status request from leaking into another target’s control plane.

**Alternatives considered**:
- Rely only on handler tests — rejected because handler tests would not meaningfully validate per-target provider selection.
- Rely only on manual environment sync checks — rejected because concurrency and failure-mode behavior need repeatable coverage.
