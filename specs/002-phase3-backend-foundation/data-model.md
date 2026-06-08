# Data Model: Phase 3 Backend Foundation

## Overview

Phase 3 adds reusable template definitions, auditable delivery targets, movement history, and notification/live-update concepts on top of the existing Phase 1 and Phase 2 environment model.

## Entities

### Template
- **Purpose**: Reusable environment definition header owned by the platform or a team.
- **Existing code**: `internal/model/template/type.go`
- **Core fields**:
  - `id`
  - `name`
  - `slug`
  - `description`
  - `category`
  - `author`
  - `author_email`
  - `visibility` (`public`, `team`, `private`)
  - `team_id`
  - `status` (`active`, `deprecated`, `archived`)
  - `created_at`, `updated_at`
- **Validation rules**:
  - `name`, `description`, and `author` are required.
  - `slug` must remain unique.
  - `visibility` and `status` must use allowed enum values.
- **Relationships**:
  - Has many `TemplateVersion`.
  - Has many `TemplateInstance` through version usage.

### TemplateVersion
- **Purpose**: Immutable-ish release of a template definition used for validation and environment creation.
- **Existing code**: `internal/model/template_version/type.go`
- **Core fields**:
  - `id`
  - `template_id`
  - `version`
  - `description`
  - `changelog`
  - `is_latest`
  - `is_stable`
  - `status` (`draft`, `stable`, `deprecated`)
  - `created_at`, `updated_at`
- **Validation rules**:
  - `template_id` and `version` are required.
  - `(template_id, version)` must be unique.
  - Only one latest version per template.
- **Relationships**:
  - Belongs to `Template`.
  - Has many `TemplateParameter`.
  - Has many `TemplateResource`.
  - Has many `TemplateInstance`.
- **State transitions**:
  - `draft -> stable`
  - `draft -> deprecated`
  - `stable -> deprecated`
  - Deprecated versions remain readable for history.

### TemplateParameter
- **Purpose**: Structured input accepted by a template version.
- **Existing code**: `internal/model/template_parameter/type.go`
- **Core fields**:
  - `id`
  - `template_id`
  - `version_id`
  - `name`
  - `display_name`
  - `description`
  - `type`
  - `default`
  - `required`
  - `validation`
  - `order`
  - `created_at`
- **Validation rules**:
  - `name`, `display_name`, and `type` are required.
  - `name` must be unique within a version.
  - `validation` stores serialized constraints used by the template validator.
- **Relationships**:
  - Belongs to `TemplateVersion`.

### TemplateResource
- **Purpose**: Reusable resource definition emitted by a template version.
- **Existing code**: `internal/model/template_resource/type.go`
- **Core fields**:
  - `id`
  - `template_id`
  - `version_id`
  - `name`
  - `type`
  - `content`
  - `order`
  - `created_at`, `updated_at`
- **Validation rules**:
  - `name`, `type`, and `content` are required.
  - Ordering must be deterministic for rendering and validation.
- **Relationships**:
  - Belongs to `TemplateVersion`.

### TemplateInstance
- **Purpose**: Historical record linking an environment to the template version and inputs used at creation time.
- **Existing code**: `internal/model/template_instance/type.go`
- **Core fields**:
  - `id`
  - `template_id`
  - `version_id`
  - `environment_id`
  - `parameters`
  - `created_at`
- **Validation rules**:
  - `template_id`, `version_id`, and `environment_id` are required.
  - `parameters` stores the validated resolved input payload.
- **Relationships**:
  - Belongs to `Template`.
  - Belongs to `TemplateVersion`.
  - Belongs to `Environment`.

### DeliveryTarget
- **Purpose**: Approved placement destination for environments and moves.
- **Planned code**: `internal/model/delivery_target/type.go`
- **Core fields**:
  - `id`
  - `name`
  - `slug`
  - `display_name`
  - `description`
  - `purpose` (for example `dev`, `staging`, `prod`, `shared`)
  - `team_id` (optional when target is platform-wide)
  - `cluster_name`
  - `cluster_server`
  - `availability_state` (`available`, `maintenance`, `disabled`)
  - `health_state` (`healthy`, `degraded`, `unhealthy`, `unknown`)
  - `capacity_summary` (serialized summary or normalized fields)
  - `created_at`, `updated_at`
- **Validation rules**:
  - `name`, `cluster_name`, and `availability_state` are required.
  - Placement is allowed only when the target is approved and available.
- **Relationships**:
  - Referenced by `Environment` as its current target.
  - Referenced by `EnvironmentMovement` as source and destination.

### Environment
- **Purpose**: Existing environment lifecycle record extended for template and target awareness.
- **Existing code**: `internal/model/environment/type.go`
- **New planned fields**:
  - `delivery_target_id`
  - optional `template_instance_id`
- **Notes**:
  - Keep existing `cluster_name` and `cluster_server` for resolved runtime placement details.
  - Existing non-template environments remain supported.

### EnvironmentMovement
- **Purpose**: Persistent request and progress record for moving an environment between delivery targets.
- **Planned code**: `internal/model/environment_movement/type.go`
- **Core fields**:
  - `id`
  - `environment_id`
  - `source_target_id`
  - `destination_target_id`
  - `requested_by`
  - `status` (`pending`, `running`, `completed`, `failed`, `cancelled`)
  - `progress_percent`
  - `message`
  - `started_at`
  - `completed_at`
  - `created_at`, `updated_at`
- **Validation rules**:
  - Destination must differ from the current target.
  - Destination must be allowed and available.
- **Relationships**:
  - Belongs to `Environment`.
  - References two `DeliveryTarget` records.
- **State transitions**:
  - `pending -> running -> completed`
  - `pending -> running -> failed`
  - `pending -> cancelled`

### Notification
- **Purpose**: User-facing operational event retained for recent review.
- **Planned code**: `internal/model/notification/type.go`
- **Core fields**:
  - `id`
  - `user_id`
  - `team_id`
  - `environment_id` (optional)
  - `kind` (`environment`, `movement`, `target`, `template`)
  - `severity` (`info`, `warning`, `error`, `success`)
  - `title`
  - `message`
  - `payload`
  - `created_at`
  - optional `read_at`
- **Validation rules**:
  - Notification recipients must remain within current authorization scope.
- **Relationships**:
  - May reference an `Environment`.
  - Delivered via live stream and list/history APIs.

### LiveSubscription
- **Purpose**: Active authenticated stream session for status, progress, log, or notification delivery.
- **Planned code**: `internal/model/live_subscription/type.go`
- **Core fields**:
  - `id`
  - `user_id`
  - `team_id`
  - `environment_id`
  - `channel` (`status`, `progress`, `log`, `notification`)
  - `workload_name` / `container_name` for logs
  - `expires_at`
  - `last_event_id`
- **Validation rules**:
  - Subscription is allowed only for environments the caller can access.
  - Losing access closes the stream.
- **Persistence**:
  - Ephemeral transport/session model; active sessions do not need durable storage in the first release.

## Relationships Summary
- `Template 1 -> N TemplateVersion`
- `TemplateVersion 1 -> N TemplateParameter`
- `TemplateVersion 1 -> N TemplateResource`
- `TemplateVersion 1 -> N TemplateInstance`
- `Environment 1 -> 0..1 TemplateInstance`
- `DeliveryTarget 1 -> N Environment`
- `Environment 1 -> N EnvironmentMovement`
- `Environment 1 -> N Notification`

## Migration Impact
- Add missing template tables and indexes.
- Add delivery target, movement, and notification tables.
- Extend `environments` with `delivery_target_id` and optional `template_instance_id`.
- Preserve backward compatibility for existing rows by making new foreign keys nullable until backfilled by new flows.
