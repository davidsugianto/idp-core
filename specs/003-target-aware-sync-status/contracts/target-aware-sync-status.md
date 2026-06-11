# Contract: Target-Aware Sync and Status Resolution

## Authentication and Authorization
- All endpoints reuse the existing authenticated `/v1/environments` API surface.
- Existing team-scoped access checks remain mandatory before any delivery-target lookup or control-plane interaction.
- The feature must not widen access to environments, targets, workloads, or logs.

## Affected Endpoints

### `POST /v1/environments/{environmentId}/sync`
Triggers a manual sync for the environment’s ArgoCD application using the control plane resolved from the environment’s assigned delivery target.

**Contract rules**:
- The system must resolve `delivery_target_id` before attempting sync.
- The request must fail if the environment is outside the caller’s allowed team scope.
- The request must fail when the delivery target is missing, unavailable for the operation, or lacks required control-plane metadata.
- The request must not fall back to an unrelated process-wide default control plane when the target cannot be resolved.
- The success path preserves the existing response shape.

**Success response**
```json
{
  "message": "sync triggered"
}
```

**Failure response shape**
```json
{
  "code": 500,
  "error": "failed to trigger sync: delivery target control plane is not configured"
}
```

**Failure semantics**:
- `404 Not Found`: environment does not exist within caller scope.
- `403 Forbidden` or existing equivalent access failure: caller cannot act on the environment.
- `500 Internal Server Error`: target resolution failed, control plane was unavailable, or the application could not be synced after selecting the target.
- `500` error bodies must sanitize secrets from upstream failures before returning the message.

### `GET /v1/environments/{environmentId}/gitops/status`
Returns GitOps sync and health information from the control plane resolved for the environment’s assigned delivery target.

**Contract rules**:
- The system must resolve `delivery_target_id` before querying GitOps status.
- The request must fail when target metadata is missing or invalid instead of querying a different target’s control plane.
- If the resolved control plane cannot find the Application, the response must describe that target-specific lookup failure.

**Success response**
```json
{
  "sync_status": "Synced",
  "health_status": "Healthy",
  "revision": "main",
  "message": ""
}
```

**Failure response shape**
```json
{
  "code": 500,
  "error": "failed to get gitops status: application not found in resolved control plane"
}
```

**Failure semantics**:
- `500` error bodies must sanitize secrets from upstream failures before returning the message.

### `GET /v1/environments/{environmentId}/status`
Returns the existing environment status response, with any ArgoCD status portion sourced from the resolved delivery target’s control plane.

**Contract rules**:
- Non-Argo status data remains unchanged.
- If the feature includes target-aware Argo status inside this endpoint, the same target-resolution guarantees and failure semantics apply.

## Delivery Target Read Contract Additions

### `GET /v1/delivery-targets`
### `GET /v1/delivery-targets/{targetId}`
Delivery target read responses must expose enough non-secret metadata for operators and admins to understand which control plane a target resolves to.

**Response additions**:
```json
{
  "id": "target-dev-a",
  "name": "dev-cluster-a",
  "cluster_name": "cluster-dev-a",
  "cluster_server": "https://10.0.0.10",
  "control_plane_name": "argocd-dev-a",
  "kubeconfig_context": "dev-a",
  "argocd_namespace": "argocd"
}
```

**Contract rules**:
- Secret references may be exposed only as safe identifiers, never raw credentials.
- Existing clients that ignore new fields continue to work.
- Read responses may also include non-secret control-plane metadata such as `control_plane_type`, `argocd_server`, or `credential_reference` when available.

## Observability and Audit Signals
- Each sync and GitOps status request must emit operator-visible logs or retained records that identify:
  - environment ID
  - delivery target ID
  - resolved control-plane name or default-selection marker
  - operation type (`sync` or `gitops_status`)
  - success or failure outcome
- Environment status reads that resolve GitOps status through the target-aware path must emit the same style of audit-safe outcome signals.
- Retained records may be stored as structured notification payloads as long as they remain audit-safe.
- Logs and records must not include kubeconfig contents, tokens, or secret material.

## Failure Modes
- Missing `delivery_target_id` on an environment that requires target-aware resolution.
- Delivery target no longer exists.
- Delivery target belongs to a different team scope or is otherwise not allowed.
- Delivery target is disabled or not usable for the requested operation.
- Delivery target lacks required control-plane metadata.
- Resolved control plane is unreachable or misconfigured.
- Resolved control plane does not contain the requested Application.

## Validation Expectations
- Validate at least two environments mapped to different delivery targets in the same running service instance.
- Validate that a request against target A never reads or mutates the Application in target B’s control plane.
- Validate that missing or invalid target metadata returns actionable target-specific failures.
- Validate that single-target/default-control-plane environments keep current behavior.
