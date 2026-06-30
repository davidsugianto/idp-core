# Data Model: Centralized Image Builder

## Build Application

**Purpose**: Represents a team-owned application that the platform can build from source without a Dockerfile.

**Core fields**:
- `id`
- `team_id`
- `name`
- `description`
- `status` (`active`, `suspended`, `deleting`, `deleted`)
- `repository_url`
- `repository_provider`
- `default_branch`
- `application_descriptor_path`
- `runtime_family`
- `runtime_detection_mode`
- `builder_profile_id`
- `registry_target_id`
- `deployment_automation_enabled`
- `gitops_target_id`
- `created_by`
- `updated_by`
- `created_at`
- `updated_at`
- `deleted_at`

**Relationships**:
- Belongs to one team.
- References zero or one builder profile.
- References one approved registry target.
- References zero or one GitOps deployment target.
- Has many build attempts.
- Has many lifecycle events.

**Validation rules**:
- `team_id`, `name`, `repository_url`, and `application_descriptor_path` are required.
- Application ownership and all reads/writes remain team-scoped.
- Deleting an application must preserve historical build, artifact, and audit records.
- A suspended application cannot accept new build triggers.

## Builder Profile

**Purpose**: Represents a platform-governed set of supported runtime detection behavior and approved builder options.

**Core fields**:
- `id`
- `name`
- `status`
- `supported_runtime_families`
- `runtime_detection_policy`
- `builder_reference`
- `buildpack_set_reference`
- `stack_reference`
- `registry_policy`
- `team_scope_mode`
- `created_at`
- `updated_at`

**Relationships**:
- May be assigned to many build applications.

**Validation rules**:
- Only approved builder profiles may be referenced by applications.
- A disabled profile may remain attached to historical builds but cannot be newly assigned.

## Registry Target

**Purpose**: Represents a centrally governed image publication destination.

**Core fields**:
- `id`
- `name`
- `registry_type` (`harbor`, `ghcr`, `ecr`, `gcr`)
- `host`
- `repository_namespace`
- `credential_reference`
- `status`
- `created_at`
- `updated_at`

**Relationships**:
- May be referenced by many build applications and many artifacts.

**Validation rules**:
- Only approved registry targets may be selected.
- Credentials are referenced indirectly and are never returned to API clients.

## Build Attempt

**Purpose**: Represents a single asynchronous build execution for a build application.

**Core fields**:
- `id`
- `application_id`
- `team_id`
- `sequence_number`
- `status` (`pending`, `queued`, `running`, `canceling`, `canceled`, `failed`, `succeeded`, `blocked`, `deployment_ready`)
- `trigger_type` (`manual`, `retry`, `system`)
- `triggered_by`
- `idempotency_key`
- `source_revision_requested`
- `source_revision_resolved`
- `retry_of_build_id`
- `cancel_requested_by`
- `failure_reason`
- `queued_at`
- `started_at`
- `finished_at`
- `created_at`
- `updated_at`

**Relationships**:
- Belongs to one build application.
- Has zero or one build artifact.
- Has zero or one security verification aggregate.
- Has zero or one deployment update.
- Has many lifecycle events.
- May reference another build attempt as its retry parent.

**Validation rules**:
- Every trigger request must resolve to at most one canonical build attempt for a given idempotency key.
- Terminal states are immutable except for downstream derived transitions such as `succeeded` to `deployment_ready` once deployment update completes successfully.
- Cancel is only valid from non-terminal states.

## Build Artifact

**Purpose**: Represents the published image output produced by a successful build attempt.

**Core fields**:
- `id`
- `build_id`
- `application_id`
- `registry_target_id`
- `image_repository`
- `image_tag`
- `image_digest`
- `published_image_reference`
- `published_at`
- `created_at`
- `updated_at`

**Relationships**:
- Belongs to one build attempt.
- Belongs to one registry target.
- Has one security verification aggregate.

**Validation rules**:
- Artifact records are immutable once publication succeeds.
- A deployment update can reference only an artifact with a completed security verification record.

## Security Verification

**Purpose**: Represents the compliance evidence collected for a build artifact.

**Core fields**:
- `id`
- `build_id`
- `artifact_id`
- `status` (`pending`, `passed`, `failed`, `waived`)
- `sbom_status`
- `sbom_reference`
- `scan_status`
- `scan_reference`
- `scan_summary`
- `signing_status`
- `signature_reference`
- `policy_gate_status`
- `policy_gate_reason`
- `completed_at`
- `created_at`
- `updated_at`

**Relationships**:
- Belongs to one build attempt and one artifact.

**Validation rules**:
- SBOM, scan, and signing results are required for every successful image build.
- Deployment automation must not proceed when required security verification has failed.
- Waiver support is reserved for future enhancement and must not weaken default required verification in the initial release.

## Deployment Update

**Purpose**: Represents the GitOps-first deployment source update initiated from a compliant build.

**Core fields**:
- `id`
- `build_id`
- `application_id`
- `status` (`pending`, `in_progress`, `succeeded`, `failed`)
- `gitops_target_id`
- `requested_image_reference`
- `requested_manifest_path`
- `resulting_revision`
- `failure_reason`
- `started_at`
- `finished_at`
- `created_at`
- `updated_at`

**Relationships**:
- Belongs to one build attempt.
- Belongs to one build application.

**Validation rules**:
- Deployment update creation requires successful image publication and passing security verification.
- Failed deployment updates do not invalidate the artifact but do block a deployment-ready outcome.

## Lifecycle Event

**Purpose**: Represents an ordered, auditable event emitted for application, build, security, and deployment state changes.

**Core fields**:
- `id`
- `team_id`
- `application_id`
- `build_id`
- `event_type`
- `event_source`
- `payload_summary`
- `occurred_at`
- `created_at`

**Relationships**:
- May belong to one build application.
- May belong to one build attempt.

**Validation rules**:
- Events are append-only.
- Events must be safe for operator consumption and omit secrets.
- Event ordering must support troubleshooting of asynchronous flows.

## Build Log Stream

**Purpose**: Represents the retrievable and streamable execution output for a build attempt.

**Core fields**:
- `build_id`
- `stream_state`
- `last_sequence`
- `retention_policy`
- `terminal_summary`

**Relationships**:
- Belongs to one build attempt.

**Validation rules**:
- Access remains team-scoped.
- Stream output must remain available while a build is active and terminal summaries must remain discoverable after completion.

## Relationship Summary

- Build Application `1 -> many` Build Attempt
- Build Attempt `1 -> 1` Build Artifact (optional until success)
- Build Attempt `1 -> 1` Security Verification (optional until artifact processing begins)
- Build Attempt `1 -> 1` Deployment Update (optional when deployment automation is enabled)
- Build Application `1 -> many` Lifecycle Event
- Build Attempt `1 -> many` Lifecycle Event
- Builder Profile `1 -> many` Build Application
- Registry Target `1 -> many` Build Application
- Registry Target `1 -> many` Build Artifact

## State Transitions

### Build Application
- `active -> suspended`
- `suspended -> active`
- `active|suspended -> deleting -> deleted`

### Build Attempt
- `pending -> queued -> running`
- `running -> succeeded`
- `running -> failed`
- `running -> canceling -> canceled`
- `succeeded -> blocked` when post-build security verification fails
- `succeeded -> deployment_ready` when security verification passes and deployment update succeeds

### Security Verification
- `pending -> passed`
- `pending -> failed`
- `failed -> waived` is reserved for a future policy-waiver feature and is out of scope for the initial implementation

### Deployment Update
- `pending -> in_progress -> succeeded`
- `pending|in_progress -> failed`
