# Contract: Phase 3 Backend API

## Authentication and Authorization
- All endpoints below use the existing authenticated `/v1` API surface.
- Template and delivery-target management is platform-admin scoped unless team-scoped visibility rules explicitly allow narrower access.
- Environment creation, movement, notifications, and live subscriptions reuse existing environment/team access checks.
- Stream connections must terminate with `401` or `403` when the caller is unauthenticated or loses access.

## Template Management

### `GET /v1/templates`
Lists templates visible to the caller.

**Query params**: `team_id`, `category`, `visibility`, `status`, `search`, `limit`, `offset`

### `POST /v1/templates`
Creates a template header.

**Request body**
```json
{
  "name": "golden-service",
  "description": "Reusable service environment",
  "category": "service",
  "author": "Platform Team",
  "author_email": "platform@example.com",
  "visibility": "team",
  "team_id": "team-123"
}
```

### `GET /v1/templates/{templateId}` / `PATCH /v1/templates/{templateId}`
Reads or updates template metadata and lifecycle status.

### `GET /v1/templates/{templateId}/versions`
Lists versions for a template.

### `POST /v1/templates/{templateId}/versions`
Creates a new template version.

### `GET /v1/templates/{templateId}/versions/{versionId}` / `PATCH /v1/templates/{templateId}/versions/{versionId}`
Reads or updates version metadata.

### `PUT /v1/templates/{templateId}/versions/{versionId}/parameters`
Replaces the ordered parameter definition set for a version.

**Request body**
```json
{
  "parameters": [
    {
      "name": "service_name",
      "display_name": "Service Name",
      "type": "string",
      "required": true,
      "validation": {"pattern": "^[a-z0-9-]+$"}
    }
  ]
}
```

### `PUT /v1/templates/{templateId}/versions/{versionId}/resources`
Replaces the ordered resource definition set for a version.

### `POST /v1/templates/{templateId}/versions/{versionId}/validate`
Validates prospective input before environment creation.

**Request body**
```json
{
  "inputs": {
    "service_name": "payments-api",
    "replicas": 2
  }
}
```

**Response body**
```json
{
  "valid": true,
  "errors": []
}
```

## Environment Creation and Placement

### `POST /v1/environments`
Extends the existing environment creation contract with optional template and delivery-target fields.

**Request body additions**
```json
{
  "name": "payments-dev",
  "git_repo_url": "https://github.com/example/app-config.git",
  "manifest_path": "environments/payments-dev",
  "template_version_id": "tmpl-ver-123",
  "template_inputs": {
    "service_name": "payments-api"
  },
  "delivery_target_id": "target-dev-a"
}
```

**Contract rules**
- `template_version_id` is optional so existing non-template flows continue to work.
- When `template_version_id` is present, input validation must pass before provisioning begins.
- When `delivery_target_id` is present, the target must be allowed and available.

## Delivery Targets

### `GET /v1/delivery-targets`
Lists placement targets visible to the caller.

### `POST /v1/delivery-targets`
Creates a delivery target.

### `GET /v1/delivery-targets/{targetId}` / `PATCH /v1/delivery-targets/{targetId}` / `DELETE /v1/delivery-targets/{targetId}`
Reads, updates, or removes a delivery target.

**Request body example**
```json
{
  "name": "dev-cluster-a",
  "display_name": "Development Cluster A",
  "purpose": "dev",
  "cluster_name": "cluster-dev-a",
  "cluster_server": "https://10.0.0.10",
  "availability_state": "available",
  "health_state": "healthy",
  "capacity_summary": {
    "cpu_available": "48",
    "memory_available": "128Gi"
  }
}
```

## Environment Movement

### `POST /v1/environments/{environmentId}/movements`
Creates a movement request.

**Request body**
```json
{
  "destination_target_id": "target-staging-b"
}
```

**Response body**
```json
{
  "id": "move-123",
  "status": "pending",
  "destination_target_id": "target-staging-b",
  "progress_percent": 0,
  "message": "Movement requested"
}
```

### `GET /v1/environments/{environmentId}/movements`
Lists historical movement requests for an environment.

### `GET /v1/environments/{environmentId}/movements/{movementId}`
Returns the latest movement state.

## Notifications

### `GET /v1/notifications`
Lists recent notifications visible to the caller.

**Query params**: `environment_id`, `kind`, `limit`, `offset`

## Live Update Streams

### `GET /v1/environments/{environmentId}/events/stream`
Authenticated SSE stream for status, progress, and notification events.

**Query params**: `channels=status,progress,notification`

**Headers**:
- `Accept: text/event-stream`
- Existing bearer token or equivalent authenticated context

**Event examples**
```text
event: status
data: {"environment_id":"env-123","status":"syncing","changed_at":"2026-06-08T12:00:00Z"}

event: progress
data: {"environment_id":"env-123","operation":"movement","progress_percent":45,"message":"Applying manifests"}

event: notification
data: {"notification_id":"notif-123","severity":"warning","title":"Target degraded"}
```

### `GET /v1/environments/{environmentId}/logs/stream`
Authenticated SSE stream for workload logs.

**Query params**: `workload`, `container`, `tail_lines`

**Event example**
```text
event: log
data: {"workload":"payments-api","container":"app","line":"server started","timestamp":"2026-06-08T12:00:05Z"}
```

## Failure Modes
- `400 Bad Request`: invalid template inputs, invalid movement target, malformed stream request.
- `401 Unauthorized`: missing or invalid authentication.
- `403 Forbidden`: caller lacks platform/team/environment access.
- `404 Not Found`: template, version, environment, movement, or target does not exist within scope.
- `409 Conflict`: duplicate template slug/version, invalid lifecycle transition, or movement to the same target.
- `410 Gone`: active stream expired or was invalidated after access loss.
