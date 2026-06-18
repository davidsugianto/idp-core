# Research: Environment Status Accuracy

## Decision 1: Treat cache misses as incomplete runtime state, not as valid zero summaries

**Decision**: Environment status should not silently return zero-valued pod and deployment summaries when the provisioner cache has no namespace entry. The usecase should distinguish cache-backed data from authoritative namespace reads and return an explicit unavailable error when workload state cannot be resolved.

**Rationale**: The current `GetStatus` flow reads `GetPodSummary` and `GetDeploymentSummary` independently and leaves the zero-value response fields untouched when neither cache lookup succeeds. That makes a cache miss look identical to an intentionally empty namespace, which is misleading for users.

**Alternatives considered**:
- Keep current cache-only behavior and document that zero may mean unavailable. Rejected because it preserves misleading success responses.
- Always fail on any cache miss without fallback. Rejected because valid namespaces may still be readable through existing workload queries even when summary caches are cold.

## Decision 2: Use namespace workload reads as the authoritative source for status and workload endpoints

**Decision**: `/v1/environments/{id}/status`, `/v1/environments/{id}/workloads`, and `/v1/environments/{id}/workloads/{name}` should all derive their operational state from the same target-aware provisioner repository for the resolved environment namespace.

**Rationale**: The existing environment usecase already resolves the correct delivery target through `provisionerRepositoryForEnvironment`. Using that same repository path for all workload-derived endpoints keeps multi-target behavior consistent and prevents one endpoint from reporting stale or blank data while another reports live values.

**Alternatives considered**:
- Introduce a separate environment-status persistence table. Rejected because the feature is about correctness of live operational views, not new stored counters.
- Query different repository methods per endpoint without shared semantics. Rejected because it would keep the current inconsistency between summary, list, and detail responses.

## Decision 3: Workload list responses must always carry environment and namespace context

**Decision**: `/v1/environments/{id}/workloads` should always return the requested environment ID and namespace, even when the namespace is empty, and should return an empty workload list instead of a blank object when no workloads exist.

**Rationale**: `workload.ToWorkloadStatusResponse` currently returns an empty struct when there are no workload statuses, which erases the environment context and makes an empty namespace indistinguishable from a broken resolution path.

**Alternatives considered**:
- Preserve the empty struct shape for backward compatibility. Rejected because it is the current bug.
- Omit the endpoint response entirely for empty namespaces. Rejected because users still need the namespace context and summary counts.

## Decision 4: Workload detail should be validated against the same resolved namespace contract as the workload list

**Decision**: `/v1/environments/{id}/workloads/{name}` should use the same environment lookup, target resolution, and namespace workload source as the workload list endpoint and should return a not-found failure when the named workload is absent.

**Rationale**: The current detail flow already searches the resolved namespace's deployments, but the surrounding contract should be kept aligned with the list response so a workload found in the list can be trusted in the detail view.

**Alternatives considered**:
- Add a separate direct Kubernetes lookup path for workload detail. Rejected because it would duplicate target resolution and risk diverging behavior.

## Decision 5: Adjacent environment endpoints should be reviewed for shared operational semantics, not fully redesigned

**Decision**: The planning scope includes checking other `/v1/environments` endpoints that expose workload-derived or target-resolution-dependent runtime data, but not redesigning unrelated CRUD endpoints.

**Rationale**: The user explicitly asked to "check and recheck other endpoint related to /v1/environments", and the most relevant risk is inconsistent handling across operational endpoints such as status, workloads, GitOps status, logs, and stream views.

**Alternatives considered**:
- Expand scope to every environment endpoint regardless of data path. Rejected because it would dilute the feature and exceed the spec's correctness-focused scope.
