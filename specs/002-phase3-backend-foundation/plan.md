# Implementation Plan: Phase 3 Backend Foundation

**Branch**: `[001-phase-3-platform]` | **Date**: 2026-06-08 | **Spec**: [`specs/002-phase3-backend-foundation/spec.md`](spec.md)

**Input**: Feature specification from `/specs/002-phase3-backend-foundation/spec.md`

## Summary

Deliver the backend foundation for Phase 3 by completing template management, introducing first-class delivery targets for multi-cluster placement and movement, and adding authenticated live update channels for environment status, progress, logs, and notifications. The implementation builds on the completed Phase 1 and Phase 2 API, auth, RBAC, audit, and environment lifecycle baseline, preserves clean architecture boundaries, and keeps the release backend-only so the separate Developer Portal UI can integrate against stable contracts later.

## Technical Context

**Language/Version**: Go 1.25+

**Primary Dependencies**: Gin, GORM, PostgreSQL, client-go informers, existing JWT/OIDC middleware, swaggo/OpenAPI

**Storage**: PostgreSQL for templates, delivery targets, movement history, template usage, and notifications; in-process streaming state backed by Kubernetes/provisioner queries for live delivery

**Testing**: `go test ./...`, focused handler/usecase/repository tests for template, delivery target, and environment flows, plus validation of live-update endpoints and authorization behavior

**Target Platform**: Linux server / containerized backend API connected to PostgreSQL, Kubernetes, and ArgoCD

**Project Type**: Web service

**Performance Goals**:
- Template validation responses complete in under 3 seconds for normal requests
- Delivery target and movement status changes become stream-visible within 5 seconds of the underlying update
- Live log/status subscriptions start on the first authenticated attempt for standard validation scenarios

**Constraints**:
- Preserve handler → usecase → repository → model boundaries and existing dependency-injection wiring
- Reuse the Phase 2 authn/authz baseline: JWT/OIDC/API key auth, RBAC, team scoping, and audit logging
- Keep historical records for template usage, movement activity, and notifications instead of overwriting prior state
- Keep this increment backend-only; no Developer Portal UI work is required in this repository
- Avoid exposing secrets or cluster credentials in logs, notifications, or stream payloads

**Scale/Scope**: One backend API serving platform admins and team-scoped developers across existing environment management flows, extended with template CRUD/versioning, cluster targeting, movement tracking, and per-environment live updates

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Self-service value is explicit: platform admins can standardize templates and delivery targets, while developers can create and observe environments without direct cluster access; template usage, movement history, and notifications remain reviewable.
- [x] Clean architecture is preserved: planned changes are isolated to `internal/model`, `internal/repository`, `internal/usecase`, `internal/handler/http`, and `cmd/http` wiring.
- [x] Security and tenant isolation are covered: all new template, target, movement, and live-update flows inherit JWT/OIDC/API key auth, RBAC, team scoping, and audit/history requirements.
- [x] Contracts and observability are covered: the plan adds explicit API/stream contracts, validation responses, movement progress, notification history, and operator-visible status signals.
- [x] Delivery stays incremental: the work is sliced into template management first, then delivery target placement/movement, then live updates.

## Project Structure

### Documentation (this feature)

```text
specs/002-phase3-backend-foundation/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── phase3-backend-api.md
└── tasks.md
```

### Source Code (repository root)

```text
cmd/
├── http/
│   ├── main.go
│   └── server.go
└── cron/

internal/
├── handler/
│   ├── http/
│   │   ├── init.go
│   │   ├── environment.go
│   │   ├── template.go
│   │   ├── delivery_target.go           # new
│   │   ├── environment_movement.go      # new
│   │   └── live_update.go               # new
│   └── cron/
├── usecase/
│   ├── environment/
│   ├── template/
│   ├── delivery_target/                 # new
│   ├── environment_movement/            # new
│   ├── notification/                    # new
│   └── live_update/                     # new
├── repository/
│   ├── environment/
│   ├── template/
│   ├── provisioner/
│   ├── delivery_target/                 # new
│   ├── environment_movement/            # new
│   └── notification/                    # new
├── model/
│   ├── environment/
│   ├── template/
│   ├── template_version/
│   ├── template_parameter/
│   ├── template_resource/
│   ├── template_instance/
│   ├── delivery_target/                 # new
│   ├── environment_movement/            # new
│   ├── notification/                    # new
│   └── live_subscription/               # new transport/session model
└── pkg/

configs/
migrations/
deployments/
```

**Structure Decision**: Build on the already-started template feature in `internal/model/template*`, `internal/repository/template`, `internal/usecase/template`, and `internal/handler/http/template.go`; extend environment creation in `internal/model/environment` and `internal/handler/http/environment.go`; add first-class delivery target, movement, notification, and live-update packages rather than pushing new concerns into existing environment/template files. Wire all new usecases and handlers in `cmd/http/main.go` and `cmd/http/server.go`.

## Phase 0: Research

1. Confirm what already exists for template CRUD/versioning and what remains missing for parameters, resources, validation, and template-backed environment creation.
2. Decide how multi-cluster targeting should fit the current environment model, which already stores `cluster_name` and `cluster_server` but has no delivery-target entity or movement workflow.
3. Decide how to deliver live status, progress, log, and notification updates without frontend work in this repository; prefer a backend-native contract that works with Gin and existing informer/provisioner state.
4. Verify which historical records need persistence and which stream/session concerns can stay ephemeral.
5. Identify missing schema work, since template-related models exist in code but no matching migrations are present yet.

## Phase 1: Design & Contracts

1. Model templates as: template header, version, parameter definitions, resource definitions, and template-instance history linked to environments.
2. Model delivery targets and environment movement as first-class records so target state, placement decisions, and movement progress are auditable.
3. Extend environment creation contracts so a request can optionally reference a template version and delivery target while preserving existing non-template environment creation.
4. Define authenticated stream contracts for per-environment status/progress/log/notification delivery, including authorization failure and unsubscribe behavior.
5. Document validation scenarios in `quickstart.md` and update `CLAUDE.md` to point at this plan.

## Phase 2: Implementation Preview

### User Story 1 - Manage reusable templates (P1)
- Finish template CRUD/versioning by persisting parameters and resources per version.
- Add validation and environment-creation inputs tied to a published template version.
- Record template usage in `template_instances` so historical environment provenance survives later template changes.
- Enforce template visibility/status/team-scope rules in handlers and usecases.

### User Story 2 - Target and manage multiple clusters (P2)
- Introduce delivery target CRUD with availability/health/capacity metadata and admin-only management.
- Extend environment creation to select an approved target and persist the selected target alongside the resolved cluster fields.
- Add environment movement requests with submission, progress, completion/failure states, and history endpoints.
- Prevent placements and moves to unavailable or invalid targets.

### User Story 3 - Receive live operational updates (P3)
- Add authenticated server-sent event streams for environment status/progress/notifications and bounded workload log streaming.
- Reuse provisioner/informer state and environment ownership checks so only authorized users can subscribe.
- Persist notifications and movement progress events needed for recent operational review even after clients disconnect.
- Ensure unsubscribe/access-loss closes streams without altering the underlying environment operation.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
