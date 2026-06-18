# Contract: Environment Operational Status

## Scope

This contract covers user-visible behavior for:
- `GET /v1/environments/{id}/status`
- `GET /v1/environments/{id}/workloads`
- `GET /v1/environments/{id}/workloads/{name}`
- review of adjacent `/v1/environments` operational endpoints for consistency in target resolution and workload-derived data handling

All endpoints remain protected by the existing authenticated environment access rules.

## `GET /v1/environments/{id}/status`

### Success expectations
- Returns the requested environment's stored metadata.
- Returns pod and deployment summary values derived from the current live state of the resolved environment namespace.
- Returns the existing GitOps status block when the environment has an associated GitOps application.
- Returns zero-valued pod and deployment summaries only when the resolved namespace truly contains no matching workloads.

### Failure expectations
- Returns not found when the environment does not exist for the caller's team.
- Returns a clear client-visible error when the environment's target context cannot be resolved or when live namespace workload state cannot be retrieved.
- Must not return a successful response that implies an empty namespace when workload retrieval actually failed.

## `GET /v1/environments/{id}/workloads`

### Success expectations
- Always returns the requested `environment_id` and `namespace` for a valid environment.
- Returns summary totals and workload entries derived from the same resolved namespace and target context.
- Returns an empty workload list and zero summary totals when the namespace is valid but contains no workloads.

### Failure expectations
- Returns not found when the environment does not exist for the caller's team.
- Returns a clear error when the environment's target context cannot be resolved or workload retrieval fails.
- Must not return blank identifiers or null workload collections for a valid success response.

## `GET /v1/environments/{id}/workloads/{name}`

### Success expectations
- Returns details for the named workload from the same environment namespace represented by the workload list endpoint.
- Uses the same delivery-target-aware resolution path as the workload list and environment status endpoints.

### Failure expectations
- Returns not found when the environment does not exist for the caller's team.
- Returns not found when the named workload does not exist in the resolved environment namespace.
- Returns a clear error when target resolution or workload retrieval fails before the named workload can be evaluated.

## Consistency rules across endpoints

- Environment status, workload list, and workload detail must resolve operational data from the same environment namespace and assigned delivery target.
- Empty namespace behavior must be distinguishable from unavailable workload state.
- Adjacent environment operational endpoints must not contradict the resolved namespace or target context used by these responses.
- Error responses must remain safe for logs and API clients and must not expose secret or credential material.
