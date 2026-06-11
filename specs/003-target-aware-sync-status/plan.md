# Implementation Plan: Target-Aware Sync and Status Resolution

**Branch**: `[master]` | **Date**: 2026-06-10 | **Spec**: [`specs/003-target-aware-sync-status/spec.md`](spec.md)

**Input**: Feature specification from `/specs/003-target-aware-sync-status/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Add target-aware sync and GitOps status behavior so each environment resolves its assigned delivery target, selects the correct Kubernetes or ArgoCD control plane for that target, and returns target-specific audit and failure signals instead of relying on a single process-wide client.

## Technical Context

**Language/Version**: Go 1.25+

**Primary Dependencies**: Gin, GORM, PostgreSQL, client-go dynamic/clientcmd, existing JWT/OIDC middleware, swaggo/OpenAPI

**Storage**: PostgreSQL for delivery-target metadata, environments, audit logs, and notifications; runtime target-scoped client selection backed by existing Kubernetes and ArgoCD client packages

**Testing**: `go test ./...`, focused usecase tests for target resolution and authorization behavior, provider/repository tests for target-scoped client selection, and integration validation across at least two targets with distinct control-plane mappings

**Target Platform**: Linux server / containerized backend API connected to PostgreSQL, Kubernetes, and one or more ArgoCD control planes

**Project Type**: Web service

**Performance Goals**:
- Target resolution adds no user-visible regression to standard sync and GitOps status flows in single-target deployments
- Sync and GitOps status requests return target-aware success or actionable failure within the same operational bounds already expected for current ArgoCD-backed calls
- Operator-visible logs and retained records identify the resolved target for 100% of validated sync and status requests

**Constraints**:
- Preserve handler → usecase → repository → model boundaries and existing dependency-injection wiring
- Reuse existing Phase 2 authn/authz baseline: JWT/OIDC/API key auth, RBAC, team scoping, and audit logging
- Do not expose kubeconfigs, tokens, or secret material in API responses, logs, notifications, or retained records
- Preserve current behavior for environments whose delivery target maps to the current default control plane
- Keep this increment backend-only; no Developer Portal UI work is required in this repository

**Scale/Scope**: One backend API instance must serve environments mapped to different delivery targets and select the correct control plane for sync and GitOps status on a per-request basis

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Self-service value is explicit: platform operators can safely manage multiple delivery targets while developers continue syncing and checking status without direct cluster access; target-selection outcomes remain reviewable through operator-visible signals.
- [x] Clean architecture is preserved: planned changes are isolated to `internal/model/environment` and `internal/model/delivery_target`, `internal/repository/gitops`, `internal/repository/provisioner`, `internal/repository/delivery_target`, `internal/usecase/environment`, `internal/handler/http/environment.go`, and wiring in `cmd/http/main.go`, with provider interfaces used for target-scoped client selection.
- [x] Security and tenant isolation are covered: existing environment/team access checks remain mandatory before target resolution; team scope and secret-handling rules apply to all target-aware sync and status flows.
- [x] Contracts and observability are covered: the plan updates sync and GitOps status behavior, delivery-target read contracts, target-resolution failures, and operator-visible audit/log signals for control-plane selection.
- [x] Delivery stays incremental: the work is sliced into target metadata and provider foundations first, then sync behavior, then status and observability validation.

## Project Structure

### Documentation (this feature)

```text
specs/003-target-aware-sync-status/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── target-aware-sync-status.md
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
│   │   └── environment.go
│   └── cron/
├── usecase/
│   └── environment/
├── repository/
│   ├── delivery_target/
│   ├── environment/
│   ├── gitops/
│   └── provisioner/
├── model/
│   ├── delivery_target/
│   └── environment/
└── pkg/
    ├── argocd/
    ├── config/
    └── kubernetes/

configs/
migrations/
deployments/
```

**Structure Decision**: Keep handlers transport-only and implement target resolution inside `internal/usecase/environment`. Extend `internal/model/delivery_target` with control-plane metadata, add provider/factory seams in `internal/repository/gitops` and `internal/repository/provisioner` so the usecase can request target-scoped repositories per environment, and wire provider implementations in `cmd/http/main.go`. Preserve existing `internal/repository/environment` and `internal/repository/delivery_target` as the sources of persisted environment and target data.

## Phase 0: Research

1. Confirm the current limitation: `cmd/http/main.go` constructs one Kubernetes client and one ArgoCD client from global config and injects singleton `provisionerRepo` and `gitopsRepo` instances into `internal/usecase/environment`.
2. Decide the clean-architecture seam for target-aware client selection; use provider interfaces returning target-scoped repositories instead of building clients in handlers or usecases.
3. Identify the missing delivery-target metadata required to distinguish workload destination cluster fields from management control-plane selection.
4. Define how default single-target behavior remains valid when a target resolves to the current shared control plane.
5. Define target-aware observability and test coverage expectations for sync, status, and failure modes.

## Phase 1: Design & Contracts

1. Extend the delivery-target data model to hold control-plane selection metadata such as operator-facing control-plane identity, Kubernetes context, ArgoCD namespace, and safe credential/config references.
2. Introduce target-scoped GitOps and provisioner provider interfaces that the environment usecase can call after resolving an environment’s delivery target.
3. Update the sync and GitOps status contracts so both endpoints explicitly resolve the target, reject invalid or missing mappings, and emit target-specific failures.
4. Define validation scenarios in `quickstart.md` that exercise two distinct targets, broken target metadata, application-not-found behavior, and unchanged single-target flows.
5. Update `CLAUDE.md` to point to this plan for future implementation context.

## Phase 2: Implementation Preview

### User Story 1 - Sync an environment through its assigned target control plane (P1)
- Extend delivery-target persistence and read models with the control-plane metadata needed for target-aware sync.
- Add provider/factory wiring so the environment usecase can resolve a target and obtain the correct GitOps repository for each sync request.
- Update `TriggerSync` to reject missing or invalid mappings, preserve existing team checks, and record target-specific outcomes.
- Keep the default control-plane path working for single-target deployments.

### User Story 2 - Read GitOps status through the correct target control plane (P2)
- Reuse the same target-resolution and provider path for `GetGitOpsStatus`.
- Update the environment status flow so any ArgoCD status returned there also comes from the resolved target control plane.
- Surface application-not-found and unreachable-control-plane failures as target-specific errors instead of opaque default-context failures.

### User Story 3 - Diagnose target resolution failures and preserve audit visibility (P3)
- Add operator-visible logs and audit-safe records that identify environment ID, delivery target ID, resolved control plane, and operation outcome.
- Ensure sync and status errors remain actionable without leaking credentials.
- Validate multi-target behavior, broken metadata paths, and cross-target isolation with focused automated coverage and quickstart scenarios.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
