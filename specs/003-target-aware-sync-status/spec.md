# Feature Specification: Target-Aware Sync and Status Resolution

**Feature Branch**: `[003-target-aware-sync-status]`

**Created**: 2026-06-10

**Status**: Draft

**Input**: User description: "Modify the `idp-core` API to dynamically resolve the environment’s delivery target (e.g., cluster name, ArgoCD instance ID) and instantiate/select the appropriate ArgoCD client/context for all sync and status operations."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Sync an environment through its assigned target control plane (Priority: P1)

As a platform API consumer, I can trigger environment sync for an environment that is assigned to a delivery target, and the system executes that sync against the control plane associated with that target instead of a single global control plane.

**Why this priority**: Sync is the primary operational action that currently fails in multi-target environments, so resolving it restores the core delivery workflow.

**Independent Test**: Create or reference an environment that is assigned to a delivery target backed by a non-default control plane, trigger sync for that environment, and confirm the request succeeds or returns a target-specific failure that identifies the selected target context.

**Acceptance Scenarios**:

1. **Given** an environment is assigned to a delivery target with valid control-plane metadata, **When** a user triggers sync for that environment, **Then** the system performs the sync through that target’s resolved control plane and records the operation without using an unrelated default context.
2. **Given** two environments are assigned to different delivery targets, **When** a user triggers sync for each environment, **Then** each request uses the control plane associated with its own target and the operations do not interfere with each other.
3. **Given** an environment has no resolvable delivery target or control-plane mapping, **When** a user triggers sync, **Then** the system rejects the request with a clear error that explains the target resolution failure.

---

### User Story 2 - Read GitOps status through the correct target control plane (Priority: P2)

As a platform API consumer, I can fetch GitOps status for an environment and receive status from the control plane that manages that environment’s application.

**Why this priority**: Status visibility is the paired operational capability for sync and must be trustworthy once multi-target sync exists.

**Independent Test**: Request GitOps status for an environment whose application is managed by a non-default control plane and confirm the returned status or failure reflects that target’s control plane rather than a global fallback.

**Acceptance Scenarios**:

1. **Given** an environment is assigned to a delivery target with valid control-plane metadata, **When** a user requests GitOps status, **Then** the system queries the status from that target’s resolved control plane.
2. **Given** the resolved control plane cannot find the environment’s application, **When** a user requests GitOps status, **Then** the response reports that target-specific lookup failure without implying the application was checked in a different target.
3. **Given** target metadata is missing, invalid, or misconfigured, **When** a user requests GitOps status, **Then** the system returns a clear error describing why target resolution could not be completed.

---

### User Story 3 - Diagnose target resolution failures and preserve audit visibility (Priority: P3)

As an operator, I can identify which delivery target and control-plane selection was attempted for sync and status operations, so I can diagnose misconfiguration quickly without guessing which cluster or ArgoCD instance the system used.

**Why this priority**: Multi-target support is only safe to operate if failures are observable and auditable.

**Independent Test**: Trigger sync and status requests against environments with valid and invalid target mappings, then confirm logs, responses, and retained records make the attempted target selection visible without exposing secrets.

**Acceptance Scenarios**:

1. **Given** a sync or status request is processed, **When** the system resolves the target control plane, **Then** operator-visible logs and retained records identify the delivery target used for the operation.
2. **Given** target resolution fails because metadata is incomplete or inconsistent, **When** the system rejects the request, **Then** the response and logs explain the failure in a way operators can act on.
3. **Given** sync or status is requested by a user outside the environment’s allowed team scope, **When** the request reaches the API, **Then** existing access-control rules are enforced before any target-specific control-plane interaction occurs.

### Edge Cases

- What happens when an environment references a delivery target that has been deleted or disabled?
- How does the system handle a delivery target whose control-plane metadata points to an unavailable or unreachable ArgoCD or Kubernetes context?
- What happens when environment target metadata and delivery target metadata disagree?
- How does the system behave when a target is valid for status lookups but lacks permission to execute sync?
- What happens when multiple requests for sync and status arrive concurrently for environments mapped to different targets?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST resolve the environment’s delivery target before executing sync or GitOps status operations.
- **FR-002**: System MUST select the control plane associated with the resolved delivery target for each sync request instead of relying on a single process-wide default context.
- **FR-003**: System MUST select the control plane associated with the resolved delivery target for each GitOps status request instead of relying on a single process-wide default context.
- **FR-004**: System MUST reject sync and status requests when the environment’s delivery target cannot be resolved or lacks the metadata required to identify its control plane.
- **FR-005**: System MUST return target-specific failure messages for sync and status operations when the selected control plane is unavailable, misconfigured, or cannot find the requested application.
- **FR-006**: System MUST preserve existing authentication, authorization, and team-scoped access rules before performing any target-specific control-plane operation.
- **FR-007**: System MUST retain audit-relevant records and operator-visible logs for sync and status requests, including which delivery target or control-plane selection was attempted.
- **FR-008**: System MUST support environments assigned to different delivery targets being synced or queried for status independently within the same running service instance.
- **FR-009**: System MUST avoid exposing secrets, tokens, or raw credentials in responses, logs, or retained records when target resolution or control-plane operations fail.
- **FR-010**: System MUST preserve current behavior for environments whose delivery target resolves to the same control plane currently used in single-target deployments.

### Key Entities *(include if feature involves data)*

- **Environment Target Mapping**: The association between an environment and the delivery target whose control plane should manage sync and status actions for that environment.
- **Delivery Target Control Plane Metadata**: The stored target-identifying information needed to determine which Kubernetes or ArgoCD control plane should be used for operations on environments assigned to that target.
- **Target Resolution Attempt**: The recorded outcome of selecting a delivery target and control plane for a sync or status request, including success or failure details safe for operators to inspect.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of sync requests for environments assigned to valid non-default delivery targets are executed against the control plane mapped to those targets during validation.
- **SC-002**: 100% of GitOps status requests for environments assigned to valid non-default delivery targets return results or failures from the correct mapped control plane during validation.
- **SC-003**: 100% of sync and status requests with missing or invalid target mappings fail with actionable target-resolution errors that identify the affected delivery target or mapping issue.
- **SC-004**: Operators can distinguish which delivery target was selected for every validated sync and status request using retained records or logs without inspecting secrets.

## Assumptions

- Existing authentication, RBAC, audit logging, and team-scope enforcement remain the baseline and are reused for this feature.
- Delivery targets are the source of truth for resolving which control plane should manage an environment’s sync and status operations.
- Single-target deployments remain in scope and continue to work when all environments resolve to the same control plane.
- The backend API in this repository is the only scope for this work; Developer Portal UI changes remain out of scope and are handled in the separate `idp-ui` repository.
