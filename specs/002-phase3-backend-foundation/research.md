# Phase 3 Research: Backend Foundation

## Decision 1: Keep template management as an extension of the existing template packages

**Decision**: Build on the current `template`, `template_version`, `template_parameter`, `template_resource`, and `template_instance` models and the existing `internal/repository/template`, `internal/usecase/template`, and `internal/handler/http/template.go` flow instead of redesigning template management from scratch.

**Rationale**: The repository already has template CRUD/version CRUD routes and usecases wired in `cmd/http/server.go`, plus dedicated models for parameters, resources, and usage history. The missing work is completion: migrations, repository reads/writes for parameters/resources/instances, validation logic, and environment creation from a published version.

**Alternatives considered**:
- Create an entirely new template subsystem in parallel — rejected because it would duplicate current code and break the incremental Phase 3 baseline.
- Fold template data into generic environment payloads only — rejected because it would lose reusable versioned definitions and historical provenance.

## Decision 2: Represent multi-cluster placement with first-class delivery targets

**Decision**: Introduce a dedicated delivery-target domain model and persist target selection on environments, while keeping `cluster_name` and `cluster_server` on the environment as resolved placement metadata.

**Rationale**: The existing environment model already stores cluster fields, but there is no entity for target availability, purpose, health, or capacity. A first-class delivery target allows admin CRUD, placement validation, movement history, and user-visible target state without overloading raw cluster fields.

**Alternatives considered**:
- Keep using only `cluster_name` as a free-form input — rejected because it cannot express approval, health, or capacity rules.
- Store target state only in config files — rejected because the feature requires auditable CRUD and API visibility.

## Decision 3: Model environment movement as a persistent workflow record

**Decision**: Add an `environment_movement` record with source target, destination target, requester, status, progress, and outcome fields, exposed through environment-scoped endpoints.

**Rationale**: The spec requires a historical record from submission through completion or failure. A persistent workflow record supports progress polling, live event delivery, audit logging, and review after completion.

**Alternatives considered**:
- Treat movement as an immediate mutation on the environment row — rejected because it loses progress and outcome history.
- Implement movement as a transient in-memory job only — rejected because the constitution requires preserved history.

## Decision 4: Use server-sent events for backend-delivered live updates

**Decision**: Expose authenticated SSE endpoints for environment status, progress, notifications, and workload logs.

**Rationale**: The service already uses Gin and HTTP handlers, and the Phase 3 scope is backend-only. SSE keeps the transport simple for UI integration later, works well for ordered one-way updates, and avoids introducing websocket state management before it is needed. Existing informer/provisioner state can feed status/progress streams, and Kubernetes log readers can feed log events.

**Alternatives considered**:
- WebSockets — rejected for this phase because bidirectional state management is extra complexity for a backend-only first increment.
- Polling-only endpoints — rejected because the spec explicitly requires live updates without repeated polling.

## Decision 5: Persist recent notifications, but keep live subscriptions ephemeral

**Decision**: Store notification records in PostgreSQL, but treat active live subscriptions as authenticated in-process sessions rather than durable rows.

**Rationale**: The spec requires notification history and reviewability, but does not require durable subscription recovery across process restarts. Keeping subscriptions ephemeral reduces schema and cleanup complexity while still allowing the API to validate and stream events securely.

**Alternatives considered**:
- Persist every subscription in PostgreSQL — rejected because it adds write churn without clear user value in the first backend release.
- Keep notifications transient only — rejected because operational events must remain reviewable.

## Decision 6: Add schema migrations for all Phase 3 persistence models

**Decision**: Create new migrations for templates, template versions, parameters, resources, instances, delivery targets, environment movements, notification history, and environment foreign-key extensions.

**Rationale**: The codebase already contains several template models, but the `migrations/` directory has no matching schema files. Planning and implementation must close that gap before the feature can be validated end to end.

**Alternatives considered**:
- Depend on GORM auto-migration implicitly — rejected because the repository already uses explicit SQL migrations.
- Defer migrations until after handler/usecase work — rejected because contract and validation planning depend on real persisted shape.
