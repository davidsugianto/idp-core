# Implementation Plan: Environment Status Accuracy

**Branch**: `[004-fix-environment-status]` | **Date**: 2026-06-16 | **Spec**: [`specs/004-fix-environment-status/spec.md`](spec.md)

**Input**: Feature specification from `/specs/004-fix-environment-status/spec.md`

**Note**: This template is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Improve the environment operational endpoints so `/v1/environments/{id}/status`, `/v1/environments/{id}/workloads`, and `/v1/environments/{id}/workloads/{name}` return accurate namespace-scoped workload data from the environment's resolved delivery target instead of misleading zero or blank responses caused by cache misses and empty derived payloads.

## Technical Context

**Language/Version**: Go 1.25+

**Primary Dependencies**: Gin, GORM, PostgreSQL, client-go via the existing provisioner repository, existing JWT/OIDC middleware, swaggo/OpenAPI

**Storage**: PostgreSQL for persisted environment records and notifications; runtime Kubernetes reads resolved through the existing provisioner repository and delivery-target-aware provider path

**Testing**: `go test ./...`, focused `internal/usecase/environment` tests for status/workload behavior, handler-level response validation where contracts change, and end-to-end API validation against a namespace with real workloads

**Target Platform**: Linux server / containerized backend API connected to PostgreSQL, Kubernetes, and delivery-target-aware control planes

**Project Type**: Web service

**Performance Goals**:
- Environment status and workload endpoints remain within the same operational bounds as the current live GitOps status calls
- Namespace-scoped workload lookups return accurate counts for 100% of validated environments with running workloads
- Failure responses become explicit for 100% of target-resolution or workload-read failures instead of silently returning empty success payloads

**Constraints**:
- Preserve handler → usecase → repository → model boundaries and existing dependency injection patterns
- Reuse existing team-scoped access control and delivery-target resolution behavior already implemented for sync and GitOps status
- Do not leak kubeconfig paths, cluster credentials, or other secret material in API responses or logs
- Keep changes focused on correctness and consistency of existing `/v1/environments` operational endpoints rather than redesigning unrelated environment APIs
- Preserve valid zero-value responses for namespaces that are truly empty

**Scale/Scope**: Team-scoped environment operational views for namespaces on one or more delivery targets, covering environment status, workload listing, workload detail, and any related environment endpoints that depend on the same workload-derived state

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- [x] Self-service value is explicit: platform users regain a trustworthy environment health view and workload inventory without direct cluster access; operators can distinguish empty environments from unavailable workload data.
- [x] Clean architecture is preserved: planned changes are isolated to `internal/model/environment`, `internal/model/workload`, `internal/repository/provisioner`, `internal/usecase/environment`, `internal/handler/http/environment.go`, and wiring/context references in `cmd/http/main.go` / `CLAUDE.md`, with delivery-target-specific Kubernetes access continuing through repository/provider interfaces.
- [x] Security and tenant isolation are covered: existing JWT/OIDC and team-scoped environment lookup remain mandatory before any workload read, and target resolution continues to enforce environment-to-target access boundaries.
- [x] Contracts and observability are covered: the plan updates API contracts for status/workload responses, failure handling for unavailable workload data, and end-to-end validation signals for target-specific namespace reads.
- [x] Delivery stays incremental: the work is split into accurate status summaries first, then workload list/detail consistency, then validation of adjacent environment endpoints.

## Project Structure

### Documentation (this feature)

```text
specs/004-fix-environment-status/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── environment-operational-status.md
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
│   ├── environment/
│   └── provisioner/
├── model/
│   ├── environment/
│   └── workload/
└── pkg/
    ├── config/
    └── kubernetes/

configs/
deployments/
```

**Structure Decision**: Keep HTTP handlers transport-only. Implement environment status/workload correctness in `internal/usecase/environment` using the existing environment lookup plus delivery-target-aware provisioner resolution. Keep Kubernetes access inside `internal/repository/provisioner`, extend domain response models in `internal/model/environment` and `internal/model/workload` as needed for consistent status/workload contracts, and preserve existing route wiring in `cmd/http/server.go` and dependency injection in `cmd/http/main.go`.

## Phase 0: Research

1. Confirm why `/status` returns zero summaries today: `GetStatus` reads cached pod/deployment summaries and silently leaves zero-value structs when the cache has no entry.
2. Confirm why `/workloads` returns blank identifiers and null workloads today: `workload.ToWorkloadStatusResponse` returns an empty struct when there are no derived workload statuses, even if the environment and namespace are known.
3. Confirm how workload detail is currently derived and where consistency gaps exist with the workload list response.
4. Decide how to distinguish a truly empty namespace from an unavailable workload read while preserving delivery-target-aware namespace resolution.
5. Identify adjacent `/v1/environments` endpoints that reuse the same operational data path and must remain internally consistent after the fix.

## Phase 1: Design & Contracts

1. Define response semantics for accurate environment status summaries, explicit workload-read failures, and empty-but-valid namespaces.
2. Define a consistent workload list/detail contract that always carries the environment and namespace context for the requested environment.
3. Keep all workload and pod reads behind the existing provisioner repository and target-aware provider path used by the environment usecase.
4. Define validation scenarios in `quickstart.md` that cover a namespace with running workloads, a truly empty namespace, a missing workload name, and an unavailable target/workload-read path.
5. Update `CLAUDE.md` to point to this plan for future implementation context.

## Phase 2: Implementation Preview

### User Story 1 - View accurate environment status (P1)
- Update `internal/usecase/environment` so environment status derives pod and deployment summaries from authoritative namespace workload reads when cached summaries are absent or stale.
- Return explicit unavailable errors when namespace workload state cannot be resolved from the assigned delivery target.
- Preserve valid zero summaries for namespaces that truly contain no workloads.

### User Story 2 - Browse environment workloads (P2)
- Update the workload response model so `/v1/environments/{id}/workloads` always includes the requested environment ID and namespace.
- Ensure workload summaries and entries are computed from the same namespace-scoped data source used by environment status.
- Keep response semantics consistent when the namespace is empty versus when workload retrieval fails.

### User Story 3 - Inspect an individual workload and keep related endpoints consistent (P3)
- Align `/v1/environments/{id}/workloads/{name}` with the workload list contract and namespace/target resolution path.
- Return a clear not-found response for missing workload names in the resolved environment namespace.
- Review adjacent `/v1/environments` operational endpoints for consistency in namespace, target-resolution, and failure semantics.

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| None | N/A | N/A |
