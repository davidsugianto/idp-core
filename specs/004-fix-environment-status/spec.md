# Feature Specification: Environment Status Accuracy

**Feature Branch**: `004-fix-environment-status`

**Created**: 2026-06-16

**Status**: Draft

**Input**: User description: "Improve environment usecase, current situation:
1. /v1/environments/{id}/status result for pod_summary and deployment_summary value is zero.
2. /v1/environments/{id}/workloads result is zero.
3. /v1/environments/{id}/workloads/{name} related to /v1/environments/{id}/workloads result.
4. and check and recheck other endpoint related to /v1/environments"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View accurate environment status (Priority: P1)

A platform user opens an environment status view and receives pod, deployment, and Argo status data that reflects the workloads currently running in that environment namespace.

**Why this priority**: Environment status is the primary health signal for a provisioned environment. If pod and deployment summaries stay at zero while GitOps shows a successful sync, users cannot trust the platform's operational view.

**Independent Test**: Can be fully tested by requesting `/v1/environments/{id}/status` for an environment that has synced workloads and confirming the pod and deployment summaries match the live namespace state instead of returning zeros.

**Acceptance Scenarios**:

1. **Given** an environment with synced workloads in its namespace, **When** a user requests the environment status, **Then** the response includes the environment metadata, the current Argo status, and non-zero pod and deployment summary fields that match the namespace's live resources.
2. **Given** an environment whose namespace currently has no workloads, **When** a user requests the environment status, **Then** the response keeps the environment metadata and Argo status while returning zero-valued workload summaries that accurately reflect the empty namespace.
3. **Given** an environment whose workload state cannot be retrieved from the target cluster, **When** a user requests the environment status, **Then** the API returns a clear failure instead of silently reporting empty summaries as if the environment had no workloads.

---

### User Story 2 - Browse environment workloads (Priority: P2)

A platform user lists workloads for an environment and receives the namespace, workload summary, and workload entries derived from the same live environment state used by the status endpoint.

**Why this priority**: Operators need a workload inventory to understand what was actually deployed, not just the high-level environment record. Accurate workload listing also supports troubleshooting and follow-up actions from the environment detail page.

**Independent Test**: Can be fully tested by requesting `/v1/environments/{id}/workloads` for an environment with deployed workloads and confirming the environment identifier, namespace, summary counters, and workload list are populated consistently.

**Acceptance Scenarios**:

1. **Given** an environment with deployments and pods in its namespace, **When** a user requests the workload list, **Then** the response includes the correct environment identifier, namespace, summary totals, and one entry for each discovered workload.
2. **Given** an environment with no workloads in its namespace, **When** a user requests the workload list, **Then** the response returns the correct environment identifier and namespace with zero summary totals and an empty workload list.
3. **Given** the workload state cannot be retrieved from the target cluster, **When** a user requests the workload list, **Then** the API returns a clear failure instead of blank identifiers and null workload data.

---

### User Story 3 - Inspect an individual workload (Priority: P3)

A platform user opens the details for a named workload and receives information that is consistent with the workload list for the same environment.

**Why this priority**: Workload detail is a drill-down path from the workload list. It is less critical than the summary endpoints, but it must stay consistent so users can trust per-workload troubleshooting.

**Independent Test**: Can be fully tested by requesting `/v1/environments/{id}/workloads/{name}` for a workload returned by `/v1/environments/{id}/workloads` and confirming the detail response describes that same workload and remains consistent with the workload list response for name, status, replicas, and image.

**Acceptance Scenarios**:

1. **Given** a workload appears in the environment workload list, **When** a user requests that workload by name, **Then** the API returns details for the same workload in the same namespace with workload fields that align with the list response.
2. **Given** a workload name does not exist in the environment namespace, **When** a user requests workload details, **Then** the API returns a not-found style failure rather than an empty success response.
3. **Given** the environment belongs to a specific delivery target, **When** a user requests workload details, **Then** the API resolves the workload data from that environment's assigned target rather than another cluster context.

### Edge Cases

- What happens when the environment record exists but its namespace has not been created yet?
- How does the system handle a delivery target that exists in the environment record but cannot be resolved into a live cluster query context?
- What happens when GitOps reports a healthy application but the namespace contains partial or restarting workloads?
- How does the system respond when a workload was recently deleted and the detail endpoint is requested before cached views are refreshed?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST return environment status responses that preserve the stored environment metadata and include pod and deployment summaries derived from the current live state of the environment namespace.
- **FR-002**: System MUST ensure zero-valued pod or deployment summaries are returned only when the live environment namespace truly has no matching resources.
- **FR-003**: System MUST return a clear failure when environment workload summaries cannot be retrieved, instead of silently substituting empty success payloads that imply no workloads exist.
- **FR-004**: Users MUST be able to request `/v1/environments/{id}/workloads` and receive the correct environment identifier, namespace, workload summary totals, and workload entries for that environment.
- **FR-005**: System MUST ensure the workload list response and the environment status response are derived from the same environment namespace and delivery target context.
- **FR-006**: Users MUST be able to request `/v1/environments/{id}/workloads/{name}` for a workload returned by the workload list and receive details that are consistent with that list response.
- **FR-007**: System MUST return a clear not-found style failure when a requested workload name does not exist in the specified environment namespace.
- **FR-008**: System MUST resolve environment status, workload list, and workload detail queries using the environment's assigned delivery target so multi-target environments return data from the correct cluster context.
- **FR-009**: System MUST preserve existing team-scoped access controls and only expose environment operational data to callers already authorized to view that environment.
- **FR-010**: System MUST review and correct other `/v1/environments` endpoints that rely on the same environment workload or target-resolution data so related responses remain internally consistent.

### Key Entities *(include if feature involves data)*

- **Environment**: A provisioned team-scoped deployment record with stored namespace, GitOps application identity, and assigned delivery target.
- **Environment Status Summary**: The aggregate operational view returned with an environment, including pod counts, deployment counts, and GitOps state.
- **Environment Workload Summary**: The aggregate workload listing view for an environment, including workload totals and pod health counts.
- **Workload**: A deployable unit running in an environment namespace that has a name, health state, replica state, and related pod information.
- **Delivery Target Context**: The cluster-scoped resolution data associated with an environment that determines where workload and GitOps state must be queried.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: In validation environments that contain running workloads, 100% of `/v1/environments/{id}/status` responses report pod and deployment summaries that match the live environment namespace at the time of the request.
- **SC-002**: In validation environments that contain running workloads, 100% of `/v1/environments/{id}/workloads` responses return the correct environment identifier, namespace, and at least one workload entry instead of blank identifiers or null workload data.
- **SC-003**: For workload names returned by the workload list endpoint, 100% of `/v1/environments/{id}/workloads/{name}` responses describe the same workload and namespace as the list response.
- **SC-004**: When environment operational data cannot be resolved from the assigned target, users receive an explicit error response on the affected endpoint instead of a successful response with misleading zero or blank values.
- **SC-005**: Related `/v1/environments` endpoints that expose workload-derived operational data return mutually consistent environment, namespace, and target-specific values during end-to-end validation.

## Assumptions

- Existing authentication, authorization, and team-scoped access rules remain unchanged and apply to all environment endpoints in scope.
- The feature is limited to improving correctness and consistency of existing environment endpoints, not redesigning the overall environment API surface.
- Live environment workload data is expected to come from the environment's assigned delivery target and namespace rather than from static database counters.
- Existing GitOps status behavior remains part of the environment status response and should stay consistent with the environment's assigned target context.
- End-to-end validation will use environments that already have a resolvable delivery target and a synced namespace with real workloads.