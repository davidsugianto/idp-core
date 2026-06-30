# Contract: Centralized Image Builder

## Scope

This contract covers the first public API surface for centralized image building:
- `POST /v1/build-applications`
- `GET /v1/build-applications`
- `GET /v1/build-applications/{id}`
- `PATCH /v1/build-applications/{id}`
- `DELETE /v1/build-applications/{id}`
- `POST /v1/build-applications/{id}/builds`
- `GET /v1/build-applications/{id}/builds`
- `GET /v1/builds/{buildId}`
- `POST /v1/builds/{buildId}/retry`
- `POST /v1/builds/{buildId}/cancel`
- `GET /v1/builds/{buildId}/logs/stream`

All endpoints remain protected by the existing authenticated, team-scoped access rules.

## Shared Contract Rules

- All success responses are scoped to the authenticated caller’s team.
- All write operations emit audit-relevant records and lifecycle events.
- Trigger, retry, and cancel operations are idempotent when the client repeats the same intent.
- Long-running execution is asynchronous; request acceptance does not imply build completion.
- Failure responses must be sanitized and must not expose raw repository, registry, signing, or credential material.

## Build Application Resource

### Fields returned on detail/list responses
- `id`
- `team_id`
- `name`
- `description`
- `status`
- `repository_url`
- `application_descriptor_path`
- `runtime_family`
- `runtime_detection_mode`
- `builder_profile`
- `registry_target`
- `deployment_automation_enabled`
- `gitops_target`
- `created_at`
- `updated_at`

## `POST /v1/build-applications`

### Request expectations
- Accepts the minimum source and descriptor inputs required to register a buildable application.
- Allows optional selection of an approved builder profile, registry target, and deployment automation settings.

### Success expectations
- Returns `201` with the created build application resource.
- Records the application in an active or administratively suspended state based on policy.

### Failure expectations
- Returns `400` for invalid configuration or unsupported runtime/build options.
- Returns `401` or `403` when the caller lacks access.
- Returns `409` when the team already has a conflicting application definition.

## `GET /v1/build-applications`

### Success expectations
- Returns the caller’s team-scoped applications.
- Supports filtering by status, runtime family, registry target, and deployment automation mode.

## `GET /v1/build-applications/{id}`

### Success expectations
- Returns the buildable application detail with current configuration and latest build summary.

### Failure expectations
- Returns `404` when the application does not exist for the caller’s team.

## `PATCH /v1/build-applications/{id}`

### Success expectations
- Returns the updated application.
- Preserves immutable history of prior build attempts and emitted lifecycle events.

### Failure expectations
- Returns `400` for invalid mutable fields.
- Returns `404` when the application does not exist for the caller’s team.

## `DELETE /v1/build-applications/{id}`

### Success expectations
- Returns `200` with a deletion-complete confirmation payload (`message: "build application deleted"`).
- Applies logical deletion semantics: the application is marked deleted/suspended from future build triggers and soft-deleted from active list endpoints.
- Preserves historical build, security, and deployment records for auditability.

## Build Resource

### Fields returned on detail/list responses
- `id`
- `application_id`
- `team_id`
- `status`
- `trigger_type`
- `source_revision_requested`
- `source_revision_resolved`
- `retry_of_build_id`
- `queued_at`
- `started_at`
- `finished_at`
- `failure_reason`
- `artifact`
- `security_verification`
- `deployment_update`

## `POST /v1/build-applications/{id}/builds`

### Request expectations
- Accepts an optional source revision override and an idempotency key.
- Returns immediately after build acceptance.

### Success expectations
- Returns `201` with the accepted build resource and its initial lifecycle state.
- Emits a `build.queued` or equivalent lifecycle event.

### Failure expectations
- Returns `400` for invalid source overrides or unsupported application state.
- Returns `404` when the application does not exist for the caller’s team.
- Returns `409` when an identical accepted build intent already exists and the request is not idempotent-compatible.

## `GET /v1/build-applications/{id}/builds`

### Success expectations
- Returns reverse-chronological build history for the application.
- Includes terminal outcomes and the latest available artifact/security/deployment summaries.

## `GET /v1/builds/{buildId}`

### Success expectations
- Returns the canonical build detail, including lifecycle state, artifact info, security verification results, and deployment update status.

### Failure expectations
- Returns `404` when the build does not exist for the caller’s team.

## `POST /v1/builds/{buildId}/retry`

### Success expectations
- Returns `200` with a new build linked back to the original failed or canceled build.

### Failure expectations
- Returns `400` when retry is not allowed from the current build state.
- Returns `404` when the original build does not exist for the caller’s team.

## `POST /v1/builds/{buildId}/cancel`

### Success expectations
- Returns `200` when cancellation is accepted.
- Repeated cancellation requests for the same active build resolve idempotently.

### Failure expectations
- Returns `400` when the build is already terminal.
- Returns `404` when the build does not exist for the caller’s team.

## `GET /v1/builds/{buildId}/logs/stream`

### Success expectations
- Streams build log output while the build is active.
- Provides a terminal summary or closure signal when the build completes.

### Failure expectations
- Returns `404` when the build does not exist for the caller’s team.
- Returns a clear client-visible error when logs are unavailable.

## Lifecycle Event Contract

The system emits ordered lifecycle events for:
- `application.created`
- `application.updated`
- `application.deleted`
- `build.queued`
- `build.running`
- `build.canceled`
- `build.failed`
- `build.succeeded`
- `security.sbom_generated`
- `security.scan_completed`
- `security.signing_completed`
- `security.policy_blocked`
- `deployment.update_started`
- `deployment.update_succeeded`
- `deployment.update_failed`

Each event includes:
- `event_id`
- `event_type`
- `team_id`
- `application_id`
- `build_id` (when applicable)
- `occurred_at`
- `summary`

## External Build Resource Expectations

The implementation may rely on external build resources, but the API contract presented to users must preserve these guarantees:
- one accepted build intent maps to one canonical build record
- build progress is observable through build detail and log streaming
- security verification and deployment update results remain attached to the same build record
- team isolation remains intact regardless of how external build resources are implemented
